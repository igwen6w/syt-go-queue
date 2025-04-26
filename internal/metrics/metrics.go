package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

var (
	// RequestCounter 记录API请求总数
	RequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "syt_go_queue_requests_total",
			Help: "The total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// RequestDuration 记录API请求处理时间
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "syt_go_queue_request_duration_seconds",
			Help:    "The request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// TaskCounter 记录任务处理总数
	TaskCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "syt_go_queue_tasks_total",
			Help: "The total number of processed tasks",
		},
		[]string{"type", "status"},
	)

	// TaskDuration 记录任务处理时间
	TaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "syt_go_queue_task_duration_seconds",
			Help:    "The task processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	// DatabaseQueryCounter 记录数据库查询总数
	DatabaseQueryCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "syt_go_queue_database_queries_total",
			Help: "The total number of database queries",
		},
		[]string{"operation", "status"},
	)

	// DatabaseQueryDuration 记录数据库查询时间
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "syt_go_queue_database_query_duration_seconds",
			Help:    "The database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// LLMAPICounter 记录LLM API调用总数
	LLMAPICounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "syt_go_queue_llm_api_calls_total",
			Help: "The total number of LLM API calls",
		},
		[]string{"status"},
	)

	// LLMAPIDuration 记录LLM API调用时间
	LLMAPIDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "syt_go_queue_llm_api_duration_seconds",
			Help:    "The LLM API call duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 从0.1秒开始，指数增长
		},
	)

	// QueueSize 记录队列大小
	QueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "syt_go_queue_size",
			Help: "The current size of the task queue",
		},
		[]string{"queue"},
	)

	// WorkerCount 记录工作者数量
	WorkerCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "syt_go_queue_workers",
			Help: "The current number of active workers",
		},
	)
)

// MeasureRequestDuration 测量请求处理时间的辅助函数
func MeasureRequestDuration(method, endpoint string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		RequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}

// MeasureTaskDuration 测量任务处理时间的辅助函数
func MeasureTaskDuration(taskType string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		TaskDuration.WithLabelValues(taskType).Observe(duration)
	}
}

// MeasureDatabaseQueryDuration 测量数据库查询时间的辅助函数
func MeasureDatabaseQueryDuration(operation string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		DatabaseQueryDuration.WithLabelValues(operation).Observe(duration)
	}
}

// MeasureLLMAPIDuration 测量LLM API调用时间的辅助函数
func MeasureLLMAPIDuration() func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		LLMAPIDuration.Observe(duration)
	}
}
