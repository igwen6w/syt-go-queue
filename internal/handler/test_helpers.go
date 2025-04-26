package handler

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/stretchr/testify/mock"
)

// MockAsynqClient 模拟 asynq.Client
type MockAsynqClient struct {
	mock.Mock
}

func (m *MockAsynqClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	args := m.Called(task, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}

func (m *MockAsynqClient) EnqueueContext(ctx interface{}, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return nil, nil
}

func (m *MockAsynqClient) Close() error {
	return nil
}

// MockAsynqInspector 模拟 asynq.Inspector
type MockAsynqInspector struct {
	mock.Mock
}

func (m *MockAsynqInspector) GetTaskInfo(queueName, taskID string) (*asynq.TaskInfo, error) {
	args := m.Called(queueName, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}

func (m *MockAsynqInspector) ListPendingTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	args := m.Called(queueName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*asynq.TaskInfo), args.Error(1)
}

func (m *MockAsynqInspector) ListActiveTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	args := m.Called(queueName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*asynq.TaskInfo), args.Error(1)
}

func (m *MockAsynqInspector) ListCompletedTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	args := m.Called(queueName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*asynq.TaskInfo), args.Error(1)
}

func (m *MockAsynqInspector) ListRetryTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	args := m.Called(queueName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*asynq.TaskInfo), args.Error(1)
}

func (m *MockAsynqInspector) ListArchivedTasks(queueName string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	args := m.Called(queueName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*asynq.TaskInfo), args.Error(1)
}

// MockDatabase 模拟数据库
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) GetValuationRecord(ctx context.Context, tableName string, id int64) (*database.ValuationRecord, error) {
	args := m.Called(ctx, tableName, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.ValuationRecord), args.Error(1)
}

func (m *MockDatabase) UpdateStatus(ctx context.Context, tableName string, id int64, status string) error {
	args := m.Called(ctx, tableName, id, status)
	return args.Error(0)
}

func (m *MockDatabase) UpdateFailedInfo(ctx context.Context, tableName string, id int64, failedInfo string, failedTimes int) error {
	args := m.Called(ctx, tableName, id, failedInfo, failedTimes)
	return args.Error(0)
}

func (m *MockDatabase) UpdateRecord(ctx context.Context, tableName string, id int64, updates map[string]interface{}) error {
	args := m.Called(ctx, tableName, id, updates)
	return args.Error(0)
}
