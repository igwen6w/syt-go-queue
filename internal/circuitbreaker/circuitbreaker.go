package circuitbreaker

import (
	"github.com/igwen6w/syt-go-queue/internal/logger"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"time"
)

// CircuitBreaker 封装了断路器功能
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// CircuitBreakerConfig 断路器配置
type CircuitBreakerConfig struct {
	Name          string        // 断路器名称
	MaxRequests   uint32        // 半开状态下允许的最大请求数
	Interval      time.Duration // 统计时间窗口
	Timeout       time.Duration // 断路器从开路状态转为半开状态的超时时间
	FailThreshold float64       // 触发断路器的错误率阈值 (0.0-1.0)
}

// NewCircuitBreaker 创建一个新的断路器
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= config.FailThreshold
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Info("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()))
		},
	}

	return &CircuitBreaker{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute 执行受断路器保护的函数
func (c *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return c.cb.Execute(req)
}

// State 获取断路器当前状态
func (c *CircuitBreaker) State() gobreaker.State {
	return c.cb.State()
}

// Counts 获取断路器统计信息
func (c *CircuitBreaker) Counts() gobreaker.Counts {
	return c.cb.Counts()
}

// DefaultLLMCircuitBreaker 创建默认的LLM API断路器
func DefaultLLMCircuitBreaker() *CircuitBreaker {
	return NewCircuitBreaker(CircuitBreakerConfig{
		Name:          "llm-api",
		MaxRequests:   2,
		Interval:      time.Minute,
		Timeout:       time.Minute * 2,
		FailThreshold: 0.5, // 50%错误率触发断路器
	})
}
