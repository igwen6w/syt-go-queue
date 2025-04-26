package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/igwen6w/syt-go-queue/internal/handler"
	"github.com/igwen6w/syt-go-queue/internal/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
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

	// 设置认证
	if cfg.Auth.Enabled {
		logger.Info("Enabling authentication", zap.String("realm", cfg.Auth.Realm))
		// 创建认证中间件
		authorized := gin.BasicAuth(cfg.Auth.Users)
		// 全局应用认证中间件
		engine.Use(authorized)
	} else {
		logger.Warn("Authentication is disabled")
	}

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

	// 创建健康检查处理器
	healthHandler := handler.NewHealthHandler(s.db, s.client)

	// 健康检查路由 - 不需要认证
	// 如果启用了全局认证，需要先禁用这些路由的认证
	s.engine.Group("/health", func(c *gin.Context) {
		// 跳过认证中间件
		c.Next()
	}).GET("", healthHandler.HealthCheck)

	s.engine.Group("/healthz", func(c *gin.Context) {
		c.Next()
	}).GET("/live", healthHandler.LivenessCheck)

	s.engine.Group("/healthz", func(c *gin.Context) {
		c.Next()
	}).GET("/ready", healthHandler.ReadinessCheck)

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
		logger.Error("Error closing asynq client", zap.Error(err))
	}
	logger.Info("API server stopped")
}
