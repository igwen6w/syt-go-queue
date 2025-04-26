package handler

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igwen6w/syt-go-queue/internal/types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 创建 HealthHandler 测试
func TestHealthCheck(t *testing.T) {
	// 创建模拟对象
	mockDB := new(MockDatabase)
	mockClient := new(MockAsynqClient)

	// 创建健康检查处理器
	handler := &HealthHandler{
		db:     mockDB,
		client: mockClient,
	}

	// 创建 Gin 路由
	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	// 创建请求
	req, _ := http.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()

	// 发送请求
	router.ServeHTTP(resp, req)

	// 验证响应状态码
	assert.Equal(t, http.StatusOK, resp.Code)

	// 解析响应
	var response types.CommonResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证响应内容
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "Service is running", response.Message)

	// 验证响应数据
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "ok", data["status"])
	assert.Contains(t, data, "timestamp")
}

func TestLivenessCheck(t *testing.T) {
	// 创建模拟对象
	mockDB := new(MockDatabase)
	mockClient := new(MockAsynqClient)

	// 创建健康检查处理器
	handler := &HealthHandler{
		db:     mockDB,
		client: mockClient,
	}

	// 创建 Gin 路由
	router := gin.New()
	router.GET("/healthz/live", handler.LivenessCheck)

	// 创建请求
	req, _ := http.NewRequest("GET", "/healthz/live", nil)
	resp := httptest.NewRecorder()

	// 发送请求
	router.ServeHTTP(resp, req)

	// 验证响应状态码
	assert.Equal(t, http.StatusOK, resp.Code)

	// 解析响应
	var response types.CommonResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证响应内容
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "Service is alive", response.Message)

	// 验证响应数据
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "ok", data["status"])
	assert.Contains(t, data, "timestamp")
}

func TestReadinessCheck(t *testing.T) {
	// 创建模拟对象
	mockDB := new(MockDatabase)
	mockClient := new(MockAsynqClient)

	// 创建健康检查处理器
	handler := &HealthHandler{
		db:     mockDB,
		client: mockClient,
	}

	// 创建 Gin 路由
	router := gin.New()
	router.GET("/healthz/ready", handler.ReadinessCheck)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedCode   int
		expectedMsg    string
		expectedData   map[string]string
	}{
		{
			name: "service ready",
			mockSetup: func() {
				// 模拟数据库连接正常
				mockDB.On("Ping").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedCode:   200,
			expectedMsg:    "Service is ready",
			expectedData: map[string]string{
				"status":   "ok",
				"database": "ok",
			},
		},
		{
			name: "service not ready - database error",
			mockSetup: func() {
				// 模拟数据库连接错误
				mockDB.On("Ping").Return(errors.New("database connection error"))
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   503,
			expectedMsg:    "Service is not ready",
			expectedData: map[string]string{
				"status":   "error",
				"database": "error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置模拟对象
			mockDB.ExpectedCalls = nil
			mockClient.ExpectedCalls = nil

			// 设置模拟行为
			tt.mockSetup()

			// 创建请求
			req, _ := http.NewRequest("GET", "/healthz/ready", nil)
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
			assert.Equal(t, tt.expectedMsg, response.Message)

			// 验证响应数据
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, tt.expectedData["status"], data["status"])
			assert.Equal(t, tt.expectedData["database"], data["database"])
			assert.Contains(t, data, "timestamp")

			// 验证模拟对象的调用
			mockDB.AssertExpectations(t)
		})
	}
}
