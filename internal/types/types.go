package types

type CreateTaskRequest struct {
	TableName string `json:"table_name" binding:"required"`
	ID        int64  `json:"id" binding:"required"`
}

type CreateTaskResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type CommonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
