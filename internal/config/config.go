package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Redis    RedisConfig    `mapstructure:"redis"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Deepseek DeepseekConfig `mapstructure:"deepseek"`
	Queue    QueueConfig    `mapstructure:"queue"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Auth     AuthConfig     `mapstructure:"auth"`
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
	APIKey         string               `mapstructure:"api_key"`
	BaseURL        string               `mapstructure:"base_url"`
	Timeout        time.Duration        `mapstructure:"timeout"`
	Model          string               `mapstructure:"model"`
	MaxTokens      int                  `mapstructure:"max_tokens"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

type CircuitBreakerConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	MaxRequests   int           `mapstructure:"max_requests"`
	Interval      time.Duration `mapstructure:"interval"`
	Timeout       time.Duration `mapstructure:"timeout"`
	FailThreshold float64       `mapstructure:"fail_threshold"`
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

type AuthConfig struct {
	Enabled bool              `mapstructure:"enabled"`
	Users   map[string]string `mapstructure:"users"` // username -> password
	Realm   string            `mapstructure:"realm"` // Basic Auth realm
}

// ValidateConfig 验证所有配置部分
func ValidateConfig(cfg *Config) error {
	// 验证 App 配置
	if err := validateAppConfig(&cfg.App); err != nil {
		return fmt.Errorf("app config: %w", err)
	}

	// 验证 Redis 配置
	if err := validateRedisConfig(&cfg.Redis); err != nil {
		return fmt.Errorf("redis config: %w", err)
	}

	// 验证 MySQL 配置
	if err := validateMySQLConfig(&cfg.MySQL); err != nil {
		return fmt.Errorf("mysql config: %w", err)
	}

	// 验证 Deepseek 配置
	if err := validateDeepseekConfig(&cfg.Deepseek); err != nil {
		return fmt.Errorf("deepseek config: %w", err)
	}

	// 验证 Queue 配置
	if err := validateQueueConfig(&cfg.Queue); err != nil {
		return fmt.Errorf("queue config: %w", err)
	}

	// 验证 Logger 配置
	if err := validateLoggerConfig(&cfg.Logger); err != nil {
		return fmt.Errorf("logger config: %w", err)
	}

	// 验证 Auth 配置
	if err := validateAuthConfig(&cfg.Auth); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}

	return nil
}

// validateAppConfig 验证 App 配置
func validateAppConfig(cfg *AppConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("name is required")
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", cfg.Port)
	}

	validModes := map[string]bool{
		"development": true,
		"production":  true,
		"test":        true,
	}

	if !validModes[cfg.Mode] {
		return fmt.Errorf("mode must be one of [development, production, test], got %s", cfg.Mode)
	}

	return nil
}

// validateRedisConfig 验证 Redis 配置
func validateRedisConfig(cfg *RedisConfig) error {
	if cfg.Addr == "" {
		return fmt.Errorf("addr is required")
	}

	// 检查地址格式
	parts := strings.Split(cfg.Addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("addr must be in format host:port, got %s", cfg.Addr)
	}

	if cfg.DB < 0 {
		return fmt.Errorf("db must be non-negative, got %d", cfg.DB)
	}

	return nil
}

// validateMySQLConfig 验证 MySQL 配置
func validateMySQLConfig(cfg *MySQLConfig) error {
	if cfg.DSN == "" {
		return fmt.Errorf("dsn is required")
	}

	if cfg.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns must be non-negative, got %d", cfg.MaxIdleConns)
	}

	if cfg.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be positive, got %d", cfg.MaxOpenConns)
	}

	return nil
}

// validateDeepseekConfig 验证 Deepseek 配置
func validateDeepseekConfig(cfg *DeepseekConfig) error {
	if cfg.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	if cfg.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}

	// 验证 URL 格式
	_, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return fmt.Errorf("base_url is invalid: %w", err)
	}

	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", cfg.Timeout)
	}

	if cfg.Model == "" {
		return fmt.Errorf("model is required")
	}

	if cfg.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive, got %d", cfg.MaxTokens)
	}

	// 验证断路器配置
	if cfg.CircuitBreaker.Enabled {
		if cfg.CircuitBreaker.MaxRequests <= 0 {
			return fmt.Errorf("circuit_breaker.max_requests must be positive, got %d", cfg.CircuitBreaker.MaxRequests)
		}

		if cfg.CircuitBreaker.Interval <= 0 {
			return fmt.Errorf("circuit_breaker.interval must be positive, got %v", cfg.CircuitBreaker.Interval)
		}

		if cfg.CircuitBreaker.Timeout <= 0 {
			return fmt.Errorf("circuit_breaker.timeout must be positive, got %v", cfg.CircuitBreaker.Timeout)
		}

		if cfg.CircuitBreaker.FailThreshold <= 0 || cfg.CircuitBreaker.FailThreshold > 1 {
			return fmt.Errorf("circuit_breaker.fail_threshold must be between 0 and 1, got %f", cfg.CircuitBreaker.FailThreshold)
		}
	}

	return nil
}

// validateQueueConfig 验证 Queue 配置
func validateQueueConfig(cfg *QueueConfig) error {
	if cfg.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive, got %d", cfg.Concurrency)
	}

	if cfg.Retry < 0 {
		return fmt.Errorf("retry must be non-negative, got %d", cfg.Retry)
	}

	if cfg.Retention <= 0 {
		return fmt.Errorf("retention must be positive, got %v", cfg.Retention)
	}

	return nil
}

// validateLoggerConfig 验证 Logger 配置
func validateLoggerConfig(cfg *LoggerConfig) error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[cfg.Level] {
		return fmt.Errorf("level must be one of [debug, info, warn, error], got %s", cfg.Level)
	}

	return nil
}

// validateAuthConfig 验证 Auth 配置
func validateAuthConfig(cfg *AuthConfig) error {
	if cfg.Enabled {
		if len(cfg.Users) == 0 {
			return fmt.Errorf("users is required when auth is enabled")
		}

		if cfg.Realm == "" {
			return fmt.Errorf("realm is required when auth is enabled")
		}
	}

	return nil
}
