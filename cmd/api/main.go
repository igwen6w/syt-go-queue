package main

import (
	"flag"
	"github.com/igwen6w/syt-go-queue/internal/config"
	"github.com/igwen6w/syt-go-queue/internal/server"
	"github.com/spf13/viper"
	"log"
)

var configFile = flag.String("config", "config/config.yaml", "path to config file")

func main() {
	flag.Parse()

	// 加载配置
	viper.SetConfigFile(*configFile)
	var cfg config.Config
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	// 创建并启动服务器
	srv := server.NewServer(&cfg)
	defer srv.Stop()

	log.Printf("Starting API server on port %d...", cfg.App.Port)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
