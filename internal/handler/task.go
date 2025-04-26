package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/igwen6w/syt-go-queue/internal/logger"
	"github.com/igwen6w/syt-go-queue/internal/metrics"
	"github.com/igwen6w/syt-go-queue/internal/task"
	"github.com/igwen6w/syt-go-queue/internal/types"
	"go.uber.org/zap"
	"net/http"
)

type TaskHandler struct {
	client interface {
		Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
	}
	db interface {
		Ping() error
		GetValuationRecord(ctx context.Context, tableName string, id int64) (*database.ValuationRecord, error)
		UpdateStatus(ctx context.Context, tableName string, id int64, status string) error
		UpdateFailedInfo(ctx context.Context, tableName string, id int64, failedInfo string, failedTimes int) error
		UpdateRecord(ctx context.Context, tableName string, id int64, updates map[string]interface{}) error
	}
	inspector interface {
		GetTaskInfo(queueName, taskID string) (*asynq.TaskInfo, error)
		ListPendingTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
		ListActiveTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
		ListCompletedTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
		ListRetryTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
		ListArchivedTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
	}
}

func NewTaskHandler(client *asynq.Client, db *database.Database, redisOpt asynq.RedisClientOpt) *TaskHandler {
	// 创建任务检查器，用于查询任务状态
	inspector := asynq.NewInspector(redisOpt)

	return &TaskHandler{
		client:    client,
		db:        db,
		inspector: inspector,
	}
}

func (h *TaskHandler) CreateLLMTask(c *gin.Context) {
	// 记录请求处理时间
	defer metrics.MeasureRequestDuration(c.Request.Method, c.FullPath())()

	var req types.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid create task request", zap.Error(err))
		c.JSON(http.StatusBadRequest, types.CommonResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 创建异步任务
	t, err := task.NewLLMTask(req.TableName, req.ID)
	if err != nil {
		logger.Error("Failed to create task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.CommonResponse{
			Code:    500,
			Message: "Failed to create task: " + err.Error(),
		})
		return
	}

	taskInfo, err := h.client.Enqueue(t)
	if err != nil {
		logger.Error("Failed to enqueue task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.CommonResponse{
			Code:    500,
			Message: "Failed to enqueue task",
		})
		return
	}

	logger.Info("Task created successfully",
		zap.String("task_id", taskInfo.ID),
		zap.String("table_name", req.TableName),
		zap.Int64("record_id", req.ID))

	c.JSON(http.StatusOK, types.CommonResponse{
		Code:    200,
		Message: "Success",
		Data: types.CreateTaskResponse{
			TaskID: taskInfo.ID,
			Status: "enqueued",
		},
	})
}

// GetTaskStatus 获取任务状态
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	// 记录请求处理时间
	defer metrics.MeasureRequestDuration(c.Request.Method, c.FullPath())()

	// 从路径参数获取任务ID
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, types.CommonResponse{
			Code:    400,
			Message: "Task ID is required",
		})
		return
	}

	// 使用检查器获取任务信息
	taskInfo, err := h.inspector.GetTaskInfo("default", taskID)
	if err != nil {
		logger.Error("Failed to get task info",
			zap.String("task_id", taskID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.CommonResponse{
			Code:    500,
			Message: "Failed to get task info: " + err.Error(),
		})
		return
	}

	if taskInfo == nil {
		c.JSON(http.StatusNotFound, types.CommonResponse{
			Code:    404,
			Message: "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, types.CommonResponse{
		Code:    200,
		Message: "Success",
		Data: types.GetTaskStatusResponse{
			TaskID:     taskInfo.ID,
			Status:     taskInfo.State.String(),
			QueueName:  taskInfo.Queue,
			CreatedAt:  taskInfo.NextProcessAt.Unix(),
			RetryCount: taskInfo.Retried,
		},
	})
}

// ListTasks 列出任务
func (h *TaskHandler) ListTasks(c *gin.Context) {
	// 记录请求处理时间
	defer metrics.MeasureRequestDuration(c.Request.Method, c.FullPath())()

	// 解析查询参数
	var req types.ListTasksRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.CommonResponse{
			Code:    400,
			Message: "Invalid query parameters: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 10
	} else if req.Limit > 100 {
		req.Limit = 100 // 限制最大数量
	}

	if req.Offset < 0 {
		req.Offset = 0
	}

	queueName := "default"
	if req.QueueName != "" {
		queueName = req.QueueName
	}

	// 解析状态参数
	var state asynq.TaskState
	switch req.Status {
	case "pending":
		state = asynq.TaskStatePending
	case "active":
		state = asynq.TaskStateActive
	case "completed":
		state = asynq.TaskStateCompleted
	case "failed":
		// Use retry state as there's no specific failed state
		state = asynq.TaskStateRetry
	case "retry":
		state = asynq.TaskStateRetry
	case "archived":
		state = asynq.TaskStateArchived
	default:
		state = asynq.TaskStateActive // 默认查询活跃任务
	}

	// 获取任务列表
	var tasks []*asynq.TaskInfo
	var err error

	// 使用适当的列表方法
	opts := []asynq.ListOption{asynq.PageSize(req.Limit), asynq.Page(req.Offset/req.Limit + 1)}

	switch state {
	case asynq.TaskStatePending:
		tasks, err = h.inspector.ListPendingTasks(queueName, opts...)
	case asynq.TaskStateActive:
		tasks, err = h.inspector.ListActiveTasks(queueName, opts...)
	case asynq.TaskStateCompleted:
		tasks, err = h.inspector.ListCompletedTasks(queueName, opts...)
	case asynq.TaskStateRetry:
		tasks, err = h.inspector.ListRetryTasks(queueName, opts...)
	case asynq.TaskStateArchived:
		tasks, err = h.inspector.ListArchivedTasks(queueName, opts...)
	default:
		tasks, err = h.inspector.ListActiveTasks(queueName, opts...)
	}
	if err != nil {
		logger.Error("Failed to list tasks",
			zap.String("queue", queueName),
			zap.String("state", state.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, types.CommonResponse{
			Code:    500,
			Message: "Failed to list tasks: " + err.Error(),
		})
		return
	}

	// 获取总数
	// 注意：asynq 不提供直接的计数方法，我们使用列表长度作为估计值
	totalCount := len(tasks)

	// 如果需要更准确的计数，可以使用大页面获取所有任务并计数
	// 这里简化处理，使用当前页的任务数量

	// 构建响应
	taskInfos := make([]types.TaskInfo, len(tasks))
	for i, t := range tasks {
		taskInfos[i] = types.TaskInfo{
			TaskID:     t.ID,
			Status:     t.State.String(),
			QueueName:  t.Queue,
			CreatedAt:  t.NextProcessAt.Unix(),
			RetryCount: t.Retried,
			Type:       t.Type,
		}
	}

	c.JSON(http.StatusOK, types.CommonResponse{
		Code:    200,
		Message: "Success",
		Data: types.ListTasksResponse{
			Tasks:      taskInfos,
			TotalCount: totalCount,
			Limit:      req.Limit,
			Offset:     req.Offset,
		},
	})
}
