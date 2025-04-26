package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/circuitbreaker"
	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/igwen6w/syt-go-queue/internal/logger"
	"github.com/igwen6w/syt-go-queue/internal/metrics"
	"github.com/igwen6w/syt-go-queue/internal/task"
	"github.com/igwen6w/syt-go-queue/internal/utils"
	"github.com/pkg/errors"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

// 状态常量
const (
	StatusProcessing = "处理中" // 处理中
	StatusCompleted  = "已完成" // 已完成
	StatusFailed     = "失败"  // 失败
)

// TaskHandler 处理异步任务的组件。
// 它封装了处理不同类型任务的逻辑，如 LLM 请求处理。
type TaskHandler struct {
	db             *database.Database             // 数据库访问实例
	deepseek       config.DeepseekConfig          // Deepseek LLM API 配置
	client         *http.Client                   // HTTP 客户端，用于调用外部 API
	circuitBreaker *circuitbreaker.CircuitBreaker // 断路器，用于保护外部调用
}

// NewTaskHandler 创建并返回一个新的任务处理器实例。
//
// 参数:
//   - db: 数据库访问实例
//   - cfg: Deepseek LLM API 的配置
//
// 返回:
//   - 配置好的任务处理器实例
func NewTaskHandler(db *database.Database, cfg config.DeepseekConfig) *TaskHandler {
	// 创建 HTTP 客户端
	client := &http.Client{Timeout: cfg.Timeout}

	// 创建断路器
	var cb *circuitbreaker.CircuitBreaker
	if cfg.CircuitBreaker.Enabled {
		logger.Info("Enabling circuit breaker for LLM API",
			zap.Float64("fail_threshold", cfg.CircuitBreaker.FailThreshold),
			zap.Duration("timeout", cfg.CircuitBreaker.Timeout),
			zap.Int("max_requests", cfg.CircuitBreaker.MaxRequests))

		cb = circuitbreaker.NewCircuitBreaker(circuitbreaker.CircuitBreakerConfig{
			Name:          "llm-api",
			MaxRequests:   uint32(cfg.CircuitBreaker.MaxRequests),
			Interval:      cfg.CircuitBreaker.Interval,
			Timeout:       cfg.CircuitBreaker.Timeout,
			FailThreshold: cfg.CircuitBreaker.FailThreshold,
		})
	} else {
		logger.Warn("Circuit breaker is disabled for LLM API")
		// 使用默认断路器
		cb = circuitbreaker.DefaultLLMCircuitBreaker()
	}

	return &TaskHandler{
		db:             db,
		deepseek:       cfg,
		client:         client,
		circuitBreaker: cb,
	}
}

// sendCallback 发送回调请求到指定的 URL。
// 该方法将处理结果作为 JSON 发送到回调 URL。
//
// 参数:
//   - ctx: 上下文，用于请求的生命周期管理
//   - callbackURL: 要发送回调的 URL
//   - result: 要发送的处理结果
//
// 返回:
//   - 如果回调请求失败，返回错误
func (h *TaskHandler) sendCallback(ctx context.Context, callbackURL string, result string) error {
	// 验证回调URL是否安全
	if err := utils.ValidateCallbackURL(callbackURL); err != nil {
		return errors.Wrap(err, "callback URL validation failed")
	}

	payload := map[string]interface{}{
		"result":    result,
		"status":    "success",
		"timestamp": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal callback payload")
	}

	// 使用传入的上下文创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", callbackURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.Wrap(err, "failed to create callback request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send callback request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("callback request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// HandleLLMTask 处理 LLM 类型的异步任务。
// 该方法从数据库获取记录，调用 LLM API，并更新处理结果。
//
// 参数:
//   - ctx: 上下文，用于请求的生命周期管理
//   - t: 要处理的任务，包含记录 ID 和表名
//
// 返回:
//   - 如果任务处理失败，返回错误
func (h *TaskHandler) HandleLLMTask(ctx context.Context, t *asynq.Task) error {
	// 开始计时并记录指标
	defer metrics.MeasureTaskDuration(task.TypeLLM)()

	var p task.LLMPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		// 记录解析失败指标
		metrics.TaskCounter.WithLabelValues(task.TypeLLM, "unmarshal_error").Inc()
		return errors.Wrap(err, "failed to unmarshal payload")
	}

	// 获取任务记录
	record, err := h.db.GetValuationRecord(ctx, p.TableName, p.ID)
	if err != nil {
		// 记录获取记录失败指标
		metrics.TaskCounter.WithLabelValues(task.TypeLLM, "db_error").Inc()
		return errors.Wrap(err, "failed to get valuation record")
	}

	// 更新状态为处理中
	if err := h.db.UpdateStatus(ctx, p.TableName, p.ID, StatusProcessing); err != nil {
		// 记录更新状态失败指标
		metrics.TaskCounter.WithLabelValues(task.TypeLLM, "update_error").Inc()
		return errors.Wrap(err, "failed to update status")
	}

	// 调用 LLM API
	result, err := h.processLLM(ctx, record)
	if err != nil {
		// 记录LLM处理失败指标
		metrics.TaskCounter.WithLabelValues(task.TypeLLM, "llm_error").Inc()

		// 更新失败信息
		failedTimes := record.FailedTimes + 1

		// 更新状态和失败信息
		updates := map[string]interface{}{
			"status":       StatusFailed,
			"failed_times": failedTimes,
			"failed_info":  err.Error(),
		}
		if updateErr := h.db.UpdateRecord(ctx, p.TableName, p.ID, updates); updateErr != nil {
			return errors.Wrap(updateErr, "failed to update failure information")
		}

		return errors.Wrap(err, "failed to process LLM")
	}

	// 更新处理结果
	updates := map[string]interface{}{
		"status":            StatusCompleted,
		"report":            result,
		"current_task_node": record.CurrentTaskNode + 1,
	}
	if err := h.db.UpdateRecord(ctx, p.TableName, p.ID, updates); err != nil {
		// 记录更新结果失败指标
		metrics.TaskCounter.WithLabelValues(task.TypeLLM, "update_result_error").Inc()
		return errors.Wrap(err, "failed to update record")
	}

	// 记录任务成功指标
	metrics.TaskCounter.WithLabelValues(task.TypeLLM, "success").Inc()

	// 如果有回调URL，发送回调请求
	if record.CallbackURL != "" {
		if err := h.sendCallback(ctx, record.CallbackURL, result); err != nil {
			// 回调失败不应该影响任务完成，只记录错误
			logger.Warn("Callback failed",
				zap.String("callback_url", record.CallbackURL),
				zap.Int64("record_id", record.ID),
				zap.String("table_name", p.TableName),
				zap.Error(err))
		}
	}

	return nil
}

// processLLM 调用 LLM API 处理记录中的消息。
// 该方法使用记录中的系统消息和用户消息构建请求，
// 并调用 Deepseek API 获取响应。
//
// 参数:
//   - ctx: 上下文，用于请求的生命周期管理
//   - record: 包含要处理的消息的记录
//
// 返回:
//   - 处理结果字符串
//   - 如果处理失败，返回错误
func (h *TaskHandler) processLLM(ctx context.Context, record *database.ValuationRecord) (string, error) {
	// 记录LLM API调用指标并计时
	defer metrics.MeasureLLMAPIDuration()()

	// 创建一个带有超时的上下文，继承父上下文的取消信号
	// 如果父上下文被取消，这个上下文也会被取消
	ctx, cancel := context.WithTimeout(ctx, h.deepseek.Timeout)
	defer cancel() // 确保在函数返回前释放资源

	// 构建请求体
	payload := map[string]interface{}{
		"model": h.deepseek.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": record.SysMessage,
			},
			{
				"role":    "user",
				"content": record.UserMessage,
			},
		},
		"max_tokens": h.deepseek.MaxTokens,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		metrics.LLMAPICounter.WithLabelValues("marshal_error").Inc()
		return "", errors.Wrap(err, "failed to marshal LLM request payload")
	}

	// 使用断路器执行请求
	result, err := h.circuitBreaker.Execute(func() (interface{}, error) {
		// 创建请求
		req, err := http.NewRequestWithContext(ctx, "POST", h.deepseek.BaseURL, bytes.NewBuffer(jsonData))
		if err != nil {
			metrics.LLMAPICounter.WithLabelValues("request_error").Inc()
			return nil, errors.Wrap(err, "failed to create LLM API request")
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+h.deepseek.APIKey)

		// 发送请求
		resp, err := h.client.Do(req)
		if err != nil {
			metrics.LLMAPICounter.WithLabelValues("network_error").Inc()
			return nil, errors.Wrap(err, "failed to send LLM API request")
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			metrics.LLMAPICounter.WithLabelValues("status_error").Inc()
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return nil, errors.Wrap(readErr, "failed to read error response body")
			}
			return nil, errors.Errorf("LLM API request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
		}

		// 解析响应
		var response struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			metrics.LLMAPICounter.WithLabelValues("decode_error").Inc()
			return nil, errors.Wrap(err, "failed to decode LLM API response")
		}

		// 检查是否有响应内容
		if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
			metrics.LLMAPICounter.WithLabelValues("empty_response").Inc()
			return nil, errors.New("empty response from LLM API")
		}

		// 记录成功调用
		metrics.LLMAPICounter.WithLabelValues("success").Inc()
		return response.Choices[0].Message.Content, nil
	})

	// 处理断路器错误
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			logger.Warn("Circuit breaker is open, too many failures",
				zap.String("record_id", fmt.Sprintf("%d", record.ID)))
			return "", errors.New("service temporarily unavailable: circuit breaker is open")
		}
		return "", err
	}

	// 转换结果类型
	content, ok := result.(string)
	if !ok {
		return "", errors.New("unexpected result type from LLM API")
	}

	return content, nil
}
