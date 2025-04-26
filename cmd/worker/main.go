package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/database"
	"github.com/igwen6w/syt-go-queue/internal/logger"
	"github.com/igwen6w/syt-go-queue/internal/worker"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var configFile = flag.String("config", "config/config.yaml", "path to config file")

func main() {
	flag.Parse()

	// 加载配置
	viper.SetConfigFile(*configFile)
	var cfg config.Config
	if err := viper.ReadInConfig(); err != nil {
		panic("Failed to read config: " + err.Error())
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		panic("Failed to unmarshal config: " + err.Error())
	}

	// 验证配置
	if err := config.ValidateConfig(&cfg); err != nil {
		panic("Invalid configuration: " + err.Error())
	}

	// 初始化日志
	logger.Init(cfg.Logger.Level, cfg.Logger.Development)
	defer logger.Sync()

	logger.Info("Worker starting",
		zap.String("app", cfg.App.Name),
		zap.String("mode", cfg.App.Mode),
		zap.String("config_file", *configFile))

	// 旧的工作者配置验证已被替换为全局配置验证

	// 初始化数据库连接
	logger.Info("Connecting to database", zap.String("dsn", maskDSN(cfg.MySQL.DSN)))
	db := sqlx.MustConnect("mysql", cfg.MySQL.DSN)
	db.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	newDatabase := database.NewDatabase(db)
	logger.Info("Database connected successfully")

	// 创建worker
	logger.Info("Creating worker",
		zap.Int("concurrency", cfg.Queue.Concurrency),
		zap.String("redis", cfg.Redis.Addr))
	w := worker.NewWorker(&cfg, newDatabase)

	// 优雅关闭
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down worker...")
		w.Stop()
	}()

	// 启动worker
	logger.Info("Starting worker...")
	if err := w.Run(); err != nil {
		logger.Fatal("Worker failed to start", zap.Error(err))
	}
}

// maskDSN 隐藏 DSN 中的敏感信息
func maskDSN(dsn string) string {
	// 简单实现，只返回一个提示信息
	// 实际应用中可以更复杂，例如保留主机名但隐藏密码
	return "[MASKED]"
}
