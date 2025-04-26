// Package task 定义了异步任务的类型和载荷结构。
// 该包提供了创建不同类型任务的工具函数。
package task

import (
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
)

const TypeLLM = "llm:process"

type LLMPayload struct {
	TableName string `json:"table_name"` // 数据表名
	ID        int64  `json:"id"`         // 记录ID
}

func NewLLMTask(tableName string, id int64) (*asynq.Task, error) {
	payload, err := json.Marshal(LLMPayload{
		TableName: tableName,
		ID:        id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal LLM task payload: %w", err)
	}
	return asynq.NewTask(TypeLLM, payload), nil
}
