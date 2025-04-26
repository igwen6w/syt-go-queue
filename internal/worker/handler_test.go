package worker

import (
	"context"
	"encoding/json"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/igwen6w/syt-go-queue/internal/task"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testConfig *config.Config

func init() {
	// 加载测试配置
	viper.SetConfigFile("../../config/config_test.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic("Failed to read config file: " + err.Error())
	}

	testConfig = &config.Config{}
	if err := viper.Unmarshal(testConfig); err != nil {
		panic("Failed to unmarshal config: " + err.Error())
	}
}

// 创建测试数据库连接
func setupTestDB(t *testing.T) (*database.Database, *sqlx.DB) {
	db := sqlx.MustConnect("mysql", testConfig.MySQL.DSN)

	// 设置连接池参数
	db.SetMaxIdleConns(testConfig.MySQL.MaxIdleConns)
	db.SetMaxOpenConns(testConfig.MySQL.MaxOpenConns)

	return database.NewDatabase(db), db
}

// 创建测试数据
func setupTestData(t *testing.T, db *sqlx.DB) {
	// 创建测试表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS test_table (
			id INT PRIMARY KEY,
			status VARCHAR(50),
			report TEXT,
			current_task_node INT,
			failed_times INT,
			user_message TEXT,
			sys_message TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// 清理旧数据
	_, err = db.Exec("DELETE FROM test_table WHERE id = 123")
	if err != nil {
		t.Fatalf("Failed to clean test data: %v", err)
	}

	// 插入测试数据
	_, err = db.Exec(`
		INSERT INTO test_table (id, status, current_task_node, failed_times)
		VALUES (123, '待处理', 1, 0)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
}

func TestTaskHandler_HandleLLMTask(t *testing.T) {
	// 创建测试数据库
	testDB, db := setupTestDB(t)
	defer db.Close()

	// 设置测试数据
	setupTestData(t, db)

	// 创建一个测试服务器模拟回调接口
	callbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer callbackServer.Close()

	// 创建测试任务处理器
	handler := NewTaskHandler(testDB, testConfig.Deepseek)

	// 创建测试任务
	payload := task.LLMPayload{
		TableName: "test_table",
		ID:        123,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	testTask := asynq.NewTask(task.TypeLLM, jsonPayload)

	// 执行任务处理
	err = handler.HandleLLMTask(context.Background(), testTask)
	if err != nil {
		t.Errorf("HandleLLMTask failed: %v", err)
	}
}

func TestTaskHandler_SendCallback(t *testing.T) {
	// 创建测试数据库
	testDB, db := setupTestDB(t)
	defer db.Close()

	// 设置测试数据
	setupTestData(t, db)

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler := NewTaskHandler(testDB, testConfig.Deepseek)

	err := handler.sendCallback(context.Background(), server.URL, "test result")
	if err != nil {
		t.Errorf("sendCallback failed: %v", err)
	}
}

func TestTaskHandler_ProcessLLM(t *testing.T) {
	// 创建测试数据库
	testDB, db := setupTestDB(t)
	defer db.Close()

	// 设置测试数据
	setupTestData(t, db)

	// 创建模拟 LLM API 的测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// 检查请求头
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("Authorization") == "" {
			t.Errorf("Missing Authorization header")
		}

		// 返回模拟的 LLM 响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// 模拟的响应数据
		response := `{
			"choices": [
				{
					"message": {
						"content": "This is a test response from the LLM API."
					}
				}
			]
		}`

		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// 创建带有测试配置的处理器
	cfg := testConfig.Deepseek
	cfg.BaseURL = server.URL // 使用测试服务器的 URL
	handler := NewTaskHandler(testDB, cfg)

	// 创建测试记录
	record := &database.ValuationRecord{
		ID:          123,
		UserMessage: "This is a test user message",
		SysMessage:  "This is a test system message",
	}

	// 测试 processLLM 方法
	result, err := handler.processLLM(context.Background(), record)
	if err != nil {
		t.Errorf("processLLM failed: %v", err)
	}

	// 验证结果
	expected := "This is a test response from the LLM API."
	if result != expected {
		t.Errorf("Expected result %q, got %q", expected, result)
	}
}
