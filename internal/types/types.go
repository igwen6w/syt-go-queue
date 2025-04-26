package types

type CreateTaskRequest struct {
	TableName string `json:"table_name" binding:"required"`
	ID        int64  `json:"id" binding:"required"`
}

type CreateTaskResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type GetTaskStatusRequest struct {
	TaskID string `json:"task_id" binding:"required"`
}

type GetTaskStatusResponse struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	QueueName  string `json:"queue_name"`
	CreatedAt  int64  `json:"created_at"`
	RetryCount int    `json:"retry_count"`
}

type ListTasksRequest struct {
	Status    string `form:"status" json:"status"`
	QueueName string `form:"queue_name" json:"queue_name"`
	Limit     int    `form:"limit" json:"limit"`
	Offset    int    `form:"offset" json:"offset"`
}

type ListTasksResponse struct {
	Tasks      []TaskInfo `json:"tasks"`
	TotalCount int        `json:"total_count"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

type TaskInfo struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	QueueName  string `json:"queue_name"`
	CreatedAt  int64  `json:"created_at"`
	RetryCount int    `json:"retry_count"`
	Type       string `json:"type"`
}

type CommonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
