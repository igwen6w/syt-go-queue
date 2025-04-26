package config

import (
	"testing"
	"time"
)

func TestValidateConfig(t *testing.T) {
	// 创建一个有效的配置
	validConfig := &Config{
		App: AppConfig{
			Name: "test-app",
			Mode: "development",
			Port: 8080,
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		MySQL: MySQLConfig{
			DSN:          "user:pass@tcp(localhost:3306)/db",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
		},
		Deepseek: DeepseekConfig{
			APIKey:    "test-api-key",
			BaseURL:   "https://api.example.com",
			Timeout:   30 * time.Second,
			Model:     "test-model",
			MaxTokens: 2000,
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:       true,
				MaxRequests:   10,
				Interval:      1 * time.Minute,
				Timeout:       2 * time.Minute,
				FailThreshold: 0.5,
			},
		},
		Queue: QueueConfig{
			Concurrency: 10,
			Retry:       3,
			Retention:   24 * time.Hour,
		},
		Logger: LoggerConfig{
			Level:       "info",
			Development: true,
		},
		Auth: AuthConfig{
			Enabled: true,
			Realm:   "Test Realm",
			Users: map[string]string{
				"test": "password",
			},
		},
	}

	// 测试有效配置
	if err := ValidateConfig(validConfig); err != nil {
		t.Errorf("ValidateConfig() with valid config returned error: %v", err)
	}

	// 测试各个配置部分的验证
	tests := []struct {
		name      string
		modifyFn  func(*Config)
		wantError bool
	}{
		// App 配置测试
		{
			name: "empty app name",
			modifyFn: func(c *Config) {
				c.App.Name = ""
			},
			wantError: true,
		},
		{
			name: "invalid app port (negative)",
			modifyFn: func(c *Config) {
				c.App.Port = -1
			},
			wantError: true,
		},
		{
			name: "invalid app port (too large)",
			modifyFn: func(c *Config) {
				c.App.Port = 70000
			},
			wantError: true,
		},
		{
			name: "invalid app mode",
			modifyFn: func(c *Config) {
				c.App.Mode = "invalid-mode"
			},
			wantError: true,
		},

		// Redis 配置测试
		{
			name: "empty redis addr",
			modifyFn: func(c *Config) {
				c.Redis.Addr = ""
			},
			wantError: true,
		},
		{
			name: "invalid redis addr format",
			modifyFn: func(c *Config) {
				c.Redis.Addr = "localhost"
			},
			wantError: true,
		},
		{
			name: "negative redis db",
			modifyFn: func(c *Config) {
				c.Redis.DB = -1
			},
			wantError: true,
		},

		// MySQL 配置测试
		{
			name: "empty mysql dsn",
			modifyFn: func(c *Config) {
				c.MySQL.DSN = ""
			},
			wantError: true,
		},
		{
			name: "negative mysql max idle conns",
			modifyFn: func(c *Config) {
				c.MySQL.MaxIdleConns = -1
			},
			wantError: true,
		},
		{
			name: "invalid mysql max open conns",
			modifyFn: func(c *Config) {
				c.MySQL.MaxOpenConns = 0
			},
			wantError: true,
		},

		// Deepseek 配置测试
		{
			name: "empty deepseek api key",
			modifyFn: func(c *Config) {
				c.Deepseek.APIKey = ""
			},
			wantError: true,
		},
		{
			name: "empty deepseek base url",
			modifyFn: func(c *Config) {
				c.Deepseek.BaseURL = ""
			},
			wantError: true,
		},
		{
			name: "invalid deepseek base url",
			modifyFn: func(c *Config) {
				c.Deepseek.BaseURL = "://invalid-url"
			},
			wantError: true,
		},
		{
			name: "zero deepseek timeout",
			modifyFn: func(c *Config) {
				c.Deepseek.Timeout = 0
			},
			wantError: true,
		},
		{
			name: "empty deepseek model",
			modifyFn: func(c *Config) {
				c.Deepseek.Model = ""
			},
			wantError: true,
		},
		{
			name: "invalid deepseek max tokens",
			modifyFn: func(c *Config) {
				c.Deepseek.MaxTokens = 0
			},
			wantError: true,
		},
		{
			name: "invalid circuit breaker max requests",
			modifyFn: func(c *Config) {
				c.Deepseek.CircuitBreaker.Enabled = true
				c.Deepseek.CircuitBreaker.MaxRequests = 0
			},
			wantError: true,
		},
		{
			name: "invalid circuit breaker interval",
			modifyFn: func(c *Config) {
				c.Deepseek.CircuitBreaker.Enabled = true
				c.Deepseek.CircuitBreaker.Interval = 0
			},
			wantError: true,
		},
		{
			name: "invalid circuit breaker timeout",
			modifyFn: func(c *Config) {
				c.Deepseek.CircuitBreaker.Enabled = true
				c.Deepseek.CircuitBreaker.Timeout = 0
			},
			wantError: true,
		},
		{
			name: "invalid circuit breaker fail threshold (negative)",
			modifyFn: func(c *Config) {
				c.Deepseek.CircuitBreaker.Enabled = true
				c.Deepseek.CircuitBreaker.FailThreshold = -0.1
			},
			wantError: true,
		},
		{
			name: "invalid circuit breaker fail threshold (too large)",
			modifyFn: func(c *Config) {
				c.Deepseek.CircuitBreaker.Enabled = true
				c.Deepseek.CircuitBreaker.FailThreshold = 1.1
			},
			wantError: true,
		},

		// Queue 配置测试
		{
			name: "invalid queue concurrency",
			modifyFn: func(c *Config) {
				c.Queue.Concurrency = 0
			},
			wantError: true,
		},
		{
			name: "negative queue retry",
			modifyFn: func(c *Config) {
				c.Queue.Retry = -1
			},
			wantError: true,
		},
		{
			name: "zero queue retention",
			modifyFn: func(c *Config) {
				c.Queue.Retention = 0
			},
			wantError: true,
		},

		// Logger 配置测试
		{
			name: "invalid logger level",
			modifyFn: func(c *Config) {
				c.Logger.Level = "invalid-level"
			},
			wantError: true,
		},

		// Auth 配置测试
		{
			name: "auth enabled but no users",
			modifyFn: func(c *Config) {
				c.Auth.Enabled = true
				c.Auth.Users = map[string]string{}
			},
			wantError: true,
		},
		{
			name: "auth enabled but empty realm",
			modifyFn: func(c *Config) {
				c.Auth.Enabled = true
				c.Auth.Realm = ""
			},
			wantError: true,
		},
		{
			name: "auth disabled with no users",
			modifyFn: func(c *Config) {
				c.Auth.Enabled = false
				c.Auth.Users = map[string]string{}
				c.Auth.Realm = ""
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建配置副本
			testConfig := *validConfig

			// 应用修改
			tt.modifyFn(&testConfig)

			// 验证配置
			err := ValidateConfig(&testConfig)

			// 检查结果
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateAppConfig(t *testing.T) {
	validConfig := &AppConfig{
		Name: "test-app",
		Mode: "development",
		Port: 8080,
	}

	if err := validateAppConfig(validConfig); err != nil {
		t.Errorf("validateAppConfig() with valid config returned error: %v", err)
	}

	tests := []struct {
		name      string
		config    AppConfig
		wantError bool
	}{
		{
			name: "empty name",
			config: AppConfig{
				Name: "",
				Mode: "development",
				Port: 8080,
			},
			wantError: true,
		},
		{
			name: "invalid port (zero)",
			config: AppConfig{
				Name: "test-app",
				Mode: "development",
				Port: 0,
			},
			wantError: true,
		},
		{
			name: "invalid port (negative)",
			config: AppConfig{
				Name: "test-app",
				Mode: "development",
				Port: -1,
			},
			wantError: true,
		},
		{
			name: "invalid port (too large)",
			config: AppConfig{
				Name: "test-app",
				Mode: "development",
				Port: 70000,
			},
			wantError: true,
		},
		{
			name: "invalid mode",
			config: AppConfig{
				Name: "test-app",
				Mode: "invalid-mode",
				Port: 8080,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAppConfig(&tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateAppConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateRedisConfig(t *testing.T) {
	validConfig := &RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	if err := validateRedisConfig(validConfig); err != nil {
		t.Errorf("validateRedisConfig() with valid config returned error: %v", err)
	}

	tests := []struct {
		name      string
		config    RedisConfig
		wantError bool
	}{
		{
			name: "empty addr",
			config: RedisConfig{
				Addr:     "",
				Password: "",
				DB:       0,
			},
			wantError: true,
		},
		{
			name: "invalid addr format",
			config: RedisConfig{
				Addr:     "localhost",
				Password: "",
				DB:       0,
			},
			wantError: true,
		},
		{
			name: "negative db",
			config: RedisConfig{
				Addr:     "localhost:6379",
				Password: "",
				DB:       -1,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRedisConfig(&tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateRedisConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateMySQLConfig(t *testing.T) {
	validConfig := &MySQLConfig{
		DSN:          "user:pass@tcp(localhost:3306)/db",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
	}

	if err := validateMySQLConfig(validConfig); err != nil {
		t.Errorf("validateMySQLConfig() with valid config returned error: %v", err)
	}

	tests := []struct {
		name      string
		config    MySQLConfig
		wantError bool
	}{
		{
			name: "empty dsn",
			config: MySQLConfig{
				DSN:          "",
				MaxIdleConns: 10,
				MaxOpenConns: 100,
			},
			wantError: true,
		},
		{
			name: "negative max idle conns",
			config: MySQLConfig{
				DSN:          "user:pass@tcp(localhost:3306)/db",
				MaxIdleConns: -1,
				MaxOpenConns: 100,
			},
			wantError: true,
		},
		{
			name: "invalid max open conns (zero)",
			config: MySQLConfig{
				DSN:          "user:pass@tcp(localhost:3306)/db",
				MaxIdleConns: 10,
				MaxOpenConns: 0,
			},
			wantError: true,
		},
		{
			name: "invalid max open conns (negative)",
			config: MySQLConfig{
				DSN:          "user:pass@tcp(localhost:3306)/db",
				MaxIdleConns: 10,
				MaxOpenConns: -1,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMySQLConfig(&tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateMySQLConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateDeepseekConfig(t *testing.T) {
	validConfig := &DeepseekConfig{
		APIKey:    "test-api-key",
		BaseURL:   "https://api.example.com",
		Timeout:   30 * time.Second,
		Model:     "test-model",
		MaxTokens: 2000,
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:       true,
			MaxRequests:   10,
			Interval:      1 * time.Minute,
			Timeout:       2 * time.Minute,
			FailThreshold: 0.5,
		},
	}

	if err := validateDeepseekConfig(validConfig); err != nil {
		t.Errorf("validateDeepseekConfig() with valid config returned error: %v", err)
	}

	tests := []struct {
		name      string
		config    DeepseekConfig
		wantError bool
	}{
		{
			name: "empty api key",
			config: DeepseekConfig{
				APIKey:    "",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
			},
			wantError: true,
		},
		{
			name: "empty base url",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
			},
			wantError: true,
		},
		{
			name: "invalid base url",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "://invalid-url",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
			},
			wantError: true,
		},
		{
			name: "zero timeout",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   0,
				Model:     "test-model",
				MaxTokens: 2000,
			},
			wantError: true,
		},
		{
			name: "negative timeout",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   -1 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
			},
			wantError: true,
		},
		{
			name: "empty model",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "",
				MaxTokens: 2000,
			},
			wantError: true,
		},
		{
			name: "zero max tokens",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 0,
			},
			wantError: true,
		},
		{
			name: "negative max tokens",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: -1,
			},
			wantError: true,
		},
		{
			name: "circuit breaker enabled with invalid max requests",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:       true,
					MaxRequests:   0,
					Interval:      1 * time.Minute,
					Timeout:       2 * time.Minute,
					FailThreshold: 0.5,
				},
			},
			wantError: true,
		},
		{
			name: "circuit breaker enabled with zero interval",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:       true,
					MaxRequests:   10,
					Interval:      0,
					Timeout:       2 * time.Minute,
					FailThreshold: 0.5,
				},
			},
			wantError: true,
		},
		{
			name: "circuit breaker enabled with zero timeout",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:       true,
					MaxRequests:   10,
					Interval:      1 * time.Minute,
					Timeout:       0,
					FailThreshold: 0.5,
				},
			},
			wantError: true,
		},
		{
			name: "circuit breaker enabled with negative fail threshold",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:       true,
					MaxRequests:   10,
					Interval:      1 * time.Minute,
					Timeout:       2 * time.Minute,
					FailThreshold: -0.1,
				},
			},
			wantError: true,
		},
		{
			name: "circuit breaker enabled with fail threshold > 1",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:       true,
					MaxRequests:   10,
					Interval:      1 * time.Minute,
					Timeout:       2 * time.Minute,
					FailThreshold: 1.1,
				},
			},
			wantError: true,
		},
		{
			name: "circuit breaker disabled with invalid parameters",
			config: DeepseekConfig{
				APIKey:    "test-api-key",
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				Model:     "test-model",
				MaxTokens: 2000,
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:       false,
					MaxRequests:   0,
					Interval:      0,
					Timeout:       0,
					FailThreshold: 0,
				},
			},
			wantError: false, // 断路器禁用时，其他参数不验证
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeepseekConfig(&tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateDeepseekConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateQueueConfig(t *testing.T) {
	validConfig := &QueueConfig{
		Concurrency: 10,
		Retry:       3,
		Retention:   24 * time.Hour,
	}

	if err := validateQueueConfig(validConfig); err != nil {
		t.Errorf("validateQueueConfig() with valid config returned error: %v", err)
	}

	tests := []struct {
		name      string
		config    QueueConfig
		wantError bool
	}{
		{
			name: "zero concurrency",
			config: QueueConfig{
				Concurrency: 0,
				Retry:       3,
				Retention:   24 * time.Hour,
			},
			wantError: true,
		},
		{
			name: "negative concurrency",
			config: QueueConfig{
				Concurrency: -1,
				Retry:       3,
				Retention:   24 * time.Hour,
			},
			wantError: true,
		},
		{
			name: "negative retry",
			config: QueueConfig{
				Concurrency: 10,
				Retry:       -1,
				Retention:   24 * time.Hour,
			},
			wantError: true,
		},
		{
			name: "zero retention",
			config: QueueConfig{
				Concurrency: 10,
				Retry:       3,
				Retention:   0,
			},
			wantError: true,
		},
		{
			name: "negative retention",
			config: QueueConfig{
				Concurrency: 10,
				Retry:       3,
				Retention:   -1 * time.Hour,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQueueConfig(&tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateQueueConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateLoggerConfig(t *testing.T) {
	validConfig := &LoggerConfig{
		Level:       "info",
		Development: true,
	}

	if err := validateLoggerConfig(validConfig); err != nil {
		t.Errorf("validateLoggerConfig() with valid config returned error: %v", err)
	}

	tests := []struct {
		name      string
		config    LoggerConfig
		wantError bool
	}{
		{
			name: "valid level: debug",
			config: LoggerConfig{
				Level:       "debug",
				Development: true,
			},
			wantError: false,
		},
		{
			name: "valid level: info",
			config: LoggerConfig{
				Level:       "info",
				Development: true,
			},
			wantError: false,
		},
		{
			name: "valid level: warn",
			config: LoggerConfig{
				Level:       "warn",
				Development: true,
			},
			wantError: false,
		},
		{
			name: "valid level: error",
			config: LoggerConfig{
				Level:       "error",
				Development: true,
			},
			wantError: false,
		},
		{
			name: "invalid level",
			config: LoggerConfig{
				Level:       "invalid-level",
				Development: true,
			},
			wantError: true,
		},
		{
			name: "empty level",
			config: LoggerConfig{
				Level:       "",
				Development: true,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoggerConfig(&tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateLoggerConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateAuthConfig(t *testing.T) {
	validConfig := &AuthConfig{
		Enabled: true,
		Realm:   "Test Realm",
		Users: map[string]string{
			"test": "password",
		},
	}

	if err := validateAuthConfig(validConfig); err != nil {
		t.Errorf("validateAuthConfig() with valid config returned error: %v", err)
	}

	tests := []struct {
		name      string
		config    AuthConfig
		wantError bool
	}{
		{
			name: "auth enabled with valid config",
			config: AuthConfig{
				Enabled: true,
				Realm:   "Test Realm",
				Users: map[string]string{
					"test": "password",
				},
			},
			wantError: false,
		},
		{
			name: "auth enabled but no users",
			config: AuthConfig{
				Enabled: true,
				Realm:   "Test Realm",
				Users:   map[string]string{},
			},
			wantError: true,
		},
		{
			name: "auth enabled but empty realm",
			config: AuthConfig{
				Enabled: true,
				Realm:   "",
				Users: map[string]string{
					"test": "password",
				},
			},
			wantError: true,
		},
		{
			name: "auth disabled with no users and empty realm",
			config: AuthConfig{
				Enabled: false,
				Realm:   "",
				Users:   map[string]string{},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAuthConfig(&tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateAuthConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
