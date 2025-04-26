package main

import (
	"flag"
	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/logger"
	"github.com/igwen6w/syt-go-queue/internal/server"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
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

	// 初始化日志
	logger.Init(cfg.Logger.Level, cfg.Logger.Development)
	defer logger.Sync()

	logger.Info("API server starting",
		zap.String("app", cfg.App.Name),
		zap.String("mode", cfg.App.Mode),
		zap.String("config_file", *configFile))

	// 创建服务器
	srv := server.NewServer(&cfg)

	// 优雅关闭
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down API server...")
		srv.Stop()
	}()

	// 启动服务器
	logger.Info("Starting API server", zap.Int("port", cfg.App.Port))
	if err := srv.Run(); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
