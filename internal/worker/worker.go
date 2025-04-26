// Package worker 提供异步任务处理功能，包括任务处理器和工作者的实现。
// 该包使用 asynq 库来处理基于 Redis 的分布式任务队列。
package worker

import (
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/igwen6w/syt-go-queue/internal/logger"
	"github.com/igwen6w/syt-go-queue/internal/metrics"
	"github.com/igwen6w/syt-go-queue/internal/task"
	"go.uber.org/zap"
	"time"
)

// Worker 表示一个异步任务处理器，负责处理队列中的任务。
// 它封装了 asynq 服务器和路由器，用于处理不同类型的任务。
type Worker struct {
	server *asynq.Server   // asynq 服务器实例
	mux    *asynq.ServeMux // 任务路由器
}

// NewWorker 创建并返回一个新的 Worker 实例。
// 它使用提供的配置初始化 asynq 服务器，并设置任务处理路由。
//
// 参数:
//   - cfg: 应用程序配置，包含 Redis 和队列设置
//   - db: 数据库实例，用于任务处理器访问数据
//
// 返回:
//   - 配置好的 Worker 实例
func NewWorker(cfg *config.Config, db *database.Database) *Worker {
	// 记录工作者数量
	metrics.WorkerCount.Set(float64(cfg.Queue.Concurrency))

	// 创建服务器配置
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
		asynq.Config{
			Concurrency: cfg.Queue.Concurrency,
			RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
				return time.Duration(n) * time.Minute
			},
			// 添加队列大小监控
			Queues: map[string]int{
				"default": 10,
			},
			// 添加队列状态监控
			QueueStatsUpdater: asynq.QueueStatsUpdaterFunc(func(stats []asynq.QueueStats) {
				for _, stat := range stats {
					metrics.QueueSize.WithLabelValues(stat.QueueID).Set(float64(stat.Size))
					logger.Debug("Queue stats updated",
						zap.String("queue", stat.QueueID),
						zap.Int("size", stat.Size),
						zap.Int("active", stat.Active),
						zap.Int("pending", stat.Pending))
				}
			}),
		},
	)

	taskHandler := NewTaskHandler(db, cfg.Deepseek)
	mux := asynq.NewServeMux()
	mux.HandleFunc(task.TypeLLM, taskHandler.HandleLLMTask)

	return &Worker{
		server: server,
		mux:    mux,
	}
}

// Run 启动工作者并开始处理任务。
// 该方法会阻塞直到工作者被停止。
//
// 返回:
//   - 如果服务器启动失败，返回错误
func (w *Worker) Run() error {
	return w.server.Run(w.mux)
}

// Stop 优雅地停止工作者。
// 该方法会停止接受新任务，并等待正在运行的任务完成。
func (w *Worker) Stop() {
	w.server.Stop()
}

// ValidateWorkerConfig 验证工作者配置是否有效。
// 它检查并确保并发数、重试次数和保留时间等参数符合要求。
//
// 参数:
//   - cfg: 要验证的配置
//
// 返回:
//   - 如果配置无效，返回错误；否则返回 nil
func ValidateWorkerConfig(cfg *config.Config) error {
	if cfg.Queue.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive, got %d", cfg.Queue.Concurrency)
	}
	if cfg.Queue.Retry < 0 {
		return fmt.Errorf("retry must be non-negative, got %d", cfg.Queue.Retry)
	}
	if cfg.Queue.Retention <= 0 {
		return fmt.Errorf("retention must be positive, got %v", cfg.Queue.Retention)
	}
	return nil
}
