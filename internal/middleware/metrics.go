package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/igwen6w/syt-go-queue/internal/metrics"
	"strconv"
	"time"
)

// MetricsMiddleware 为Gin请求添加Prometheus指标收集
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始计时
		start := time.Now()

		// 处理请求
		c.Next()

		// 请求结束后记录指标
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		// 记录请求计数
		metrics.RequestCounter.WithLabelValues(method, endpoint, status).Inc()

		// 记录请求处理时间
		duration := time.Since(start).Seconds()
		metrics.RequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}
