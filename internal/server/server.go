package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/igwen6w/syt-go-queue/internal/handler"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	engine *gin.Engine
	cfg    *config.Config
	client *asynq.Client
	db     *database.Database
}

func NewServer(cfg *config.Config) *Server {
	// 初始化 Redis 客户端
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 初始化 MySQL 连接
	db := sqlx.MustConnect("mysql", cfg.MySQL.DSN)
	db.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)

	// 初始化数据库实例
	newDatabase := database.NewDatabase(db)

	// 初始化 Gin 引擎
	engine := gin.Default()

	server := &Server{
		engine: engine,
		cfg:    cfg,
		client: client,
		db:     newDatabase,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// 创建任务处理器
	taskHandler := handler.NewTaskHandler(s.client, s.db)

	// API 路由
	api := s.engine.Group("/api")
	{
		api.POST("/tasks/llm", taskHandler.CreateLLMTask)
		// 可以添加更多路由...
	}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.cfg.App.Port)
	return s.engine.Run(addr)
}

func (s *Server) Stop() {
	if err := s.client.Close(); err != nil {
		// 使用标准日志记录错误，后续会改进为结构化日志
		fmt.Printf("Error closing asynq client: %v\n", err)
	}
}
