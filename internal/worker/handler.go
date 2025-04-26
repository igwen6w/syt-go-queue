package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/igwen6w/syt-go-queue/internal/task"
	"github.com/pkg/errors"
	"io"
	"net
	"net/http"
	"time"
)

// TaskHandler 处理异步任务的组件。
// 它封装了处理不同类型任务的逻辑，如 LLM 请求处理。
type TaskHandler struct {
	db       *database.Database    // 数据库访问实例
	deepseek config.DeepseekConfig // Deepseek LLM API 配置
	client   *http.Client          // HTTP 客户端，用于调用外部 API
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
	return &TaskHandler{
		db:       db,
		deepseek: cfg,
		client:   &http.Client{Timeout: cfg.Timeout},
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
	var p task.LLMPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return errors.Wrap(err, "failed to unmarshal payload")
	}

	// 获取任务记录
	record, err := h.db.GetValuationRecord(ctx, p.TableName, p.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get valuation record")
	}

	// 更新状态为处理中
	if err := h.db.UpdateStatus(ctx, p.TableName, p.ID, "处理中"); err != nil {
		return errors.Wrap(err, "failed to update status")
	}

	// 调用 LLM API
	result, err := h.processLLM(ctx, record)
	if err != nil {
		// 更新失败信息
		failedTimes := record.FailedTimes + 1
		if err := h.db.UpdateFailedInfo(ctx, p.TableName, p.ID, err.Error(), failedTimes); err != nil {
			return errors.Wrap(err, "failed to update failed info")
		}
		return errors.Wrap(err, "failed to process LLM")
	}

	// 更新处理结果
	updates := map[string]interface{}{
		"status":            "已完成",
		"report":            result,
		"current_task_node": record.CurrentTaskNode + 1,
	}
	if err := h.db.UpdateRecord(ctx, p.TableName, p.ID, updates); err != nil {
		return errors.Wrap(err, "failed to update record")
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
	// 创建一个带有超时的上下文
	ctx, cancel := context.WithTimeout(ctx, h.deepseek.Timeout)
	defer cancel()

	// 构建请求体
	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
				"role":    "system",
				"role": "system",
				"content": record.SysMessage,
			},
				"role":    "user",
				"role": "user",
				"content": record.UserMessage,
			},
		},
		"max_tokens":  2000,
		"max_tokens": 2000,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal LLM request payload")
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", h.deepseek.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", errors.Wrap(err, "failed to create LLM API request")
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.deepseek.APIKey)

	// 发送请求
	resp, err := h.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send LLM API request")
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", errors.Errorf("LLM API request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
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
		return "", errors.Wrap(err, "failed to decode LLM API response")
	}

	// 检查是否有响应内容
	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		return "", errors.New("empty response from LLM API")
	}

	return response.Choices[0].Message.Content, nil
}
