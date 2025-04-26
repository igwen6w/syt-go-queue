package config

import (
	"time"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Redis    RedisConfig    `mapstructure:"redis"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Deepseek DeepseekConfig `mapstructure:"deepseek"`
	Queue    QueueConfig    `mapstructure:"queue"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Mode string `mapstructure:"mode"`
	Port int    `mapstructure:"port"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MySQLConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type DeepseekConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	Timeout   time.Duration `mapstructure:"timeout"`
	Model     string        `mapstructure:"model"`
	MaxTokens int           `mapstructure:"max_tokens"`
}

type QueueConfig struct {
	Concurrency int           `mapstructure:"concurrency"`
	Retry       int           `mapstructure:"retry"`
	Retention   time.Duration `mapstructure:"retention"`
}

type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Development bool   `mapstructure:"development"`
}
