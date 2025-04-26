package worker

import (
	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"testing"
	"time"
)

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

func TestNewWorker(t *testing.T) {
	// 初始化测试用的数据库连接
	db := sqlx.MustConnect("mysql", testConfig.MySQL.DSN)
	db.SetMaxIdleConns(testConfig.MySQL.MaxIdleConns)
	db.SetMaxOpenConns(testConfig.MySQL.MaxOpenConns)
	newDatabase := database.NewDatabase(db)
	defer db.Close()

	// 创建 worker
	worker := NewWorker(testConfig, newDatabase)
	if worker == nil {
		t.Error("NewWorker returned nil")
	}

	if worker.server == nil {
		t.Error("Worker server is nil")
	}

	if worker.mux == nil {
		t.Error("Worker mux is nil")
	}
}

func TestWorker_RunAndStop(t *testing.T) {
	// 初始化测试用的数据库连接
	db := sqlx.MustConnect("mysql", testConfig.MySQL.DSN)
	db.SetMaxIdleConns(testConfig.MySQL.MaxIdleConns)
	db.SetMaxOpenConns(testConfig.MySQL.MaxOpenConns)
	newDatabase := database.NewDatabase(db)
	defer db.Close()

	// 创建 worker
	worker := NewWorker(testConfig, newDatabase)

	// 创建一个通道来捕获错误
	errCh := make(chan error, 1)

	// 启动 worker（在 goroutine 中运行以避免阻塞）
	go func() {
		errCh <- worker.Run()
	}()

	// 给 worker 一些时间来启动
	time.Sleep(time.Second)

	// 停止 worker
	worker.Stop()

	// 等待错误或超时
	select {
	case err := <-errCh:
		// 如果是正常的关闭，错误应该为 nil
		if err != nil {
			t.Errorf("Worker.Run() failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Worker.Run() did not exit within expected time")
	}
}

func TestWorkerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Queue: config.QueueConfig{
					Concurrency: 10,
					Retry:       3,
					Retention:   24 * time.Hour,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid concurrency",
			config: &config.Config{
				Queue: config.QueueConfig{
					Concurrency: 0,
					Retry:       3,
					Retention:   24 * time.Hour,
				},
			},
			wantErr: true,
		},
		{
			name: "negative retry",
			config: &config.Config{
				Queue: config.QueueConfig{
					Concurrency: 10,
					Retry:       -1,
					Retention:   24 * time.Hour,
				},
			},
			wantErr: true,
		},
		{
			name: "zero retention",
			config: &config.Config{
				Queue: config.QueueConfig{
					Concurrency: 10,
					Retry:       3,
					Retention:   0,
				},
			},
			wantErr: true,
		},
		{
			name: "negative retention",
			config: &config.Config{
				Queue: config.QueueConfig{
					Concurrency: 10,
					Retry:       3,
					Retention:   -1 * time.Hour,
				},
			},
			wantErr: true,
		},
		{
			name: "all invalid",
			config: &config.Config{
				Queue: config.QueueConfig{
					Concurrency: 0,
					Retry:       -1,
					Retention:   0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkerConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWorkerConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
