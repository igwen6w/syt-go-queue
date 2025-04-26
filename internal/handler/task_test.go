package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 设置 Gin 测试模式
func init() {
	gin.SetMode(gin.TestMode)
}

func TestCreateLLMTask(t *testing.T) {
	// 创建模拟对象
	mockClient := new(MockAsynqClient)
	mockDB := new(MockDatabase)
	mockInspector := new(MockAsynqInspector)

	// 创建任务处理器
	handler := &TaskHandler{
		client:    mockClient,
		db:        mockDB,
		inspector: mockInspector,
	}

	// 创建 Gin 路由
	router := gin.New()
	router.POST("/api/tasks/llm", handler.CreateLLMTask)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedCode   int
		expectedMsg    string
	}{
		{
			name: "valid request",
			requestBody: types.CreateTaskRequest{
				TableName: "test_table",
				ID:        123,
			},
			mockSetup: func() {
				// 模拟成功的任务创建
				mockClient.On("Enqueue", mock.Anything, mock.Anything).Return(&asynq.TaskInfo{
					ID:    "task123",
					Queue: "default",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCode:   200,
			expectedMsg:    "Success",
		},
		{
			name: "invalid request - missing table name",
			requestBody: map[string]interface{}{
				"id": 123,
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   400,
			expectedMsg:    "Key: 'CreateTaskRequest.TableName' Error:Field validation for 'TableName' failed on the 'required' tag",
		},
		{
			name: "invalid request - missing id",
			requestBody: map[string]interface{}{
				"table_name": "test_table",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   400,
			expectedMsg:    "Key: 'CreateTaskRequest.ID' Error:Field validation for 'ID' failed on the 'required' tag",
		},
		{
			name: "enqueue error",
			requestBody: types.CreateTaskRequest{
				TableName: "test_table",
				ID:        123,
			},
			mockSetup: func() {
				// 模拟任务入队失败
				mockClient.On("Enqueue", mock.Anything, mock.Anything).Return(nil, errors.New("enqueue error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   500,
			expectedMsg:    "Failed to enqueue task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置模拟对象
			mockClient.ExpectedCalls = nil
			mockDB.ExpectedCalls = nil
			mockInspector.ExpectedCalls = nil

			// 设置模拟行为
			tt.mockSetup()

			// 创建请求
			jsonData, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/tasks/llm", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			// 发送请求
			router.ServeHTTP(resp, req)

			// 验证响应状态码
			assert.Equal(t, tt.expectedStatus, resp.Code)

			// 解析响应
			var response types.CommonResponse
			err := json.Unmarshal(resp.Body.Bytes(), &response)
			assert.NoError(t, err)

			// 验证响应内容
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Contains(t, response.Message, tt.expectedMsg)

			// 验证模拟对象的调用
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetTaskStatus(t *testing.T) {
	// 创建模拟对象
	mockClient := new(MockAsynqClient)
	mockDB := new(MockDatabase)
	mockInspector := new(MockAsynqInspector)

	// 创建任务处理器
	handler := &TaskHandler{
		client:    mockClient,
		db:        mockDB,
		inspector: mockInspector,
	}

	// 创建 Gin 路由
	router := gin.New()
	router.GET("/api/tasks/:id", handler.GetTaskStatus)

	tests := []struct {
		name           string
		taskID         string
		mockSetup      func()
		expectedStatus int
		expectedCode   int
		expectedMsg    string
	}{
		{
			name:   "valid task id",
			taskID: "task123",
			mockSetup: func() {
				// 模拟成功获取任务信息
				mockInspector.On("GetTaskInfo", "default", "task123").Return(&asynq.TaskInfo{
					ID:      "task123",
					Queue:   "default",
					State:   asynq.TaskStateActive,
					Retried: 0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCode:   200,
			expectedMsg:    "Success",
		},
		{
			name:   "task not found",
			taskID: "nonexistent",
			mockSetup: func() {
				// 模拟任务不存在
				mockInspector.On("GetTaskInfo", "default", "nonexistent").Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedCode:   404,
			expectedMsg:    "Task not found",
		},
		{
			name:   "inspector error",
			taskID: "task123",
			mockSetup: func() {
				// 模拟检查器错误
				mockInspector.On("GetTaskInfo", "default", "task123").Return(nil, errors.New("inspector error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   500,
			expectedMsg:    "Failed to get task info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置模拟对象
			mockClient.ExpectedCalls = nil
			mockDB.ExpectedCalls = nil
			mockInspector.ExpectedCalls = nil

			// 设置模拟行为
			tt.mockSetup()

			// 创建请求
			req, _ := http.NewRequest("GET", "/api/tasks/"+tt.taskID, nil)
			resp := httptest.NewRecorder()

			// 发送请求
			router.ServeHTTP(resp, req)

			// 验证响应状态码
			assert.Equal(t, tt.expectedStatus, resp.Code)

			// 解析响应
			var response types.CommonResponse
			err := json.Unmarshal(resp.Body.Bytes(), &response)
			assert.NoError(t, err)

			// 验证响应内容
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Contains(t, response.Message, tt.expectedMsg)

			// 验证模拟对象的调用
			mockInspector.AssertExpectations(t)
		})
	}
}

func TestListTasks(t *testing.T) {
	// 创建模拟对象
	mockClient := new(MockAsynqClient)
	mockDB := new(MockDatabase)
	mockInspector := new(MockAsynqInspector)

	// 创建任务处理器
	handler := &TaskHandler{
		client:    mockClient,
		db:        mockDB,
		inspector: mockInspector,
	}

	// 创建 Gin 路由
	router := gin.New()
	router.GET("/api/tasks", handler.ListTasks)

	// 创建测试任务列表
	testTasks := []*asynq.TaskInfo{
		{
			ID:      "task1",
			Queue:   "default",
			State:   asynq.TaskStateActive,
			Type:    "llm:process",
			Retried: 0,
		},
		{
			ID:      "task2",
			Queue:   "default",
			State:   asynq.TaskStateActive,
			Type:    "llm:process",
			Retried: 1,
		},
	}

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func()
		expectedStatus int
		expectedCode   int
		expectedMsg    string
		expectedCount  int
	}{
		{
			name:        "default parameters",
			queryParams: "",
			mockSetup: func() {
				// 模拟成功获取任务列表
				mockInspector.On("ListActiveTasks", "default", mock.Anything).Return(testTasks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCode:   200,
			expectedMsg:    "Success",
			expectedCount:  2,
		},
		{
			name:        "custom parameters",
			queryParams: "?status=pending&queue_name=custom&limit=5&offset=10",
			mockSetup: func() {
				// 模拟成功获取任务列表
				mockInspector.On("ListPendingTasks", "custom", mock.Anything).Return(testTasks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCode:   200,
			expectedMsg:    "Success",
			expectedCount:  2,
		},
		{
			name:        "list error",
			queryParams: "",
			mockSetup: func() {
				// 模拟列表获取失败
				mockInspector.On("ListActiveTasks", "default", mock.Anything).Return(nil, errors.New("list error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   500,
			expectedMsg:    "Failed to list tasks",
			expectedCount:  0,
		},

		{
			name:        "invalid limit",
			queryParams: "?limit=-1",
			mockSetup: func() {
				// 应该使用默认限制 10
				mockInspector.On("ListActiveTasks", "default", mock.Anything).Return(testTasks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCode:   200,
			expectedMsg:    "Success",
			expectedCount:  2,
		},
		{
			name:        "limit too large",
			queryParams: "?limit=200",
			mockSetup: func() {
				// 应该限制为最大 100
				mockInspector.On("ListActiveTasks", "default", mock.Anything).Return(testTasks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCode:   200,
			expectedMsg:    "Success",
			expectedCount:  2,
		},
		{
			name:        "negative offset",
			queryParams: "?offset=-10",
			mockSetup: func() {
				// 应该使用默认偏移 0
				mockInspector.On("ListActiveTasks", "default", mock.Anything).Return(testTasks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCode:   200,
			expectedMsg:    "Success",
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置模拟对象
			mockClient.ExpectedCalls = nil
			mockDB.ExpectedCalls = nil
			mockInspector.ExpectedCalls = nil

			// 设置模拟行为
			tt.mockSetup()

			// 创建请求
			req, _ := http.NewRequest("GET", "/api/tasks"+tt.queryParams, nil)
			resp := httptest.NewRecorder()

			// 发送请求
			router.ServeHTTP(resp, req)

			// 验证响应状态码
			assert.Equal(t, tt.expectedStatus, resp.Code)

			// 解析响应
			var response types.CommonResponse
			err := json.Unmarshal(resp.Body.Bytes(), &response)
			assert.NoError(t, err)

			// 验证响应内容
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.Contains(t, response.Message, tt.expectedMsg)

			// 如果是成功响应，验证任务数量
			if tt.expectedStatus == http.StatusOK {
				listResponse, ok := response.Data.(map[string]interface{})
				assert.True(t, ok)
				tasks, ok := listResponse["tasks"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, tasks, tt.expectedCount)
			}

			// 验证模拟对象的调用
			mockInspector.AssertExpectations(t)
		})
	}
}
