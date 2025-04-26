package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

var (
	// Log 全局日志实例
	Log  *zap.Logger
	once sync.Once
)

// Init 初始化日志
func Init(level string, development bool) {
	once.Do(func() {
		// 解析日志级别
		var logLevel zapcore.Level
		switch level {
		case "debug":
			logLevel = zapcore.DebugLevel
		case "info":
			logLevel = zapcore.InfoLevel
		case "warn":
			logLevel = zapcore.WarnLevel
		case "error":
			logLevel = zapcore.ErrorLevel
		default:
			logLevel = zapcore.InfoLevel
		}

		// 创建日志配置
		config := zap.Config{
			Level:       zap.NewAtomicLevelAt(logLevel),
			Development: development,
			Encoding:    "json",
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "time",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				FunctionKey:    zapcore.OmitKey,
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		// 如果是开发环境，使用更友好的控制台输出
		if development {
			config.Encoding = "console"
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		// 创建日志实例
		var err error
		Log, err = config.Build(zap.AddCallerSkip(1))
		if err != nil {
			// 如果初始化失败，使用一个基本的日志配置
			Log = zap.NewExample()
			Log.Error("Failed to initialize logger", zap.Error(err))
		}
	})
}

// Debug 输出调试级别日志
func Debug(msg string, fields ...zap.Field) {
	ensureLogger()
	Log.Debug(msg, fields...)
}

// Info 输出信息级别日志
func Info(msg string, fields ...zap.Field) {
	ensureLogger()
	Log.Info(msg, fields...)
}

// Warn 输出警告级别日志
func Warn(msg string, fields ...zap.Field) {
	ensureLogger()
	Log.Warn(msg, fields...)
}

// Error 输出错误级别日志
func Error(msg string, fields ...zap.Field) {
	ensureLogger()
	Log.Error(msg, fields...)
}

// Fatal 输出致命错误日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	ensureLogger()
	Log.Fatal(msg, fields...)
}

// ensureLogger 确保日志实例已初始化
func ensureLogger() {
	if Log == nil {
		// 如果日志未初始化，使用默认配置
		Init("info", false)
	}
}

// Sync 同步日志缓冲区
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
