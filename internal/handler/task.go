package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/igwen6w/syt-go-queue/internal/task"
	"github.com/igwen6w/syt-go-queue/internal/types"
	"net/http"
)

type TaskHandler struct {
	client *asynq.Client
	db     *database.Database
}

func NewTaskHandler(client *asynq.Client, db *database.Database) *TaskHandler {
	return &TaskHandler{
		client: client,
		db:     db,
	}
}

func (h *TaskHandler) CreateLLMTask(c *gin.Context) {
	var req types.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.CommonResponse{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	// 创建异步任务
	t, err := task.NewLLMTask(req.TableName, req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.CommonResponse{
			Code:    500,
			Message: "Failed to create task: " + err.Error(),
		})
		return
	}

	taskInfo, err := h.client.Enqueue(t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.CommonResponse{
			Code:    500,
			Message: "Failed to enqueue task",
		})
		return
	}

	c.JSON(http.StatusOK, types.CommonResponse{
		Code:    200,
		Message: "Success",
		Data: types.CreateTaskResponse{
			TaskID: taskInfo.ID,
			Status: "enqueued",
		},
	})
}
