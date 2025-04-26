package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/igwen6w/syt-go-queue/internal/logger"
	"github.com/igwen6w/syt-go-queue/internal/types"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// HealthHandler 处理健康检查相关的请求
type HealthHandler struct {
	db interface {
		Ping() error
	}
	client interface {
		Close() error
	}
}

// NewHealthHandler 创建并返回一个新的健康检查处理器
func NewHealthHandler(db interface{ Ping() error }, client interface{ Close() error }) *HealthHandler {
	return &HealthHandler{
		db:     db,
		client: client,
	}
}

// HealthCheck 处理基本的健康检查请求
// 返回服务的基本状态信息
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, types.CommonResponse{
		Code:    200,
		Message: "Service is running",
		Data: map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		},
	})
}

// LivenessCheck 处理活跃性检查请求
// 只检查服务是否在运行，不检查依赖服务
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, types.CommonResponse{
		Code:    200,
		Message: "Service is alive",
		Data: map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		},
	})
}

// ReadinessCheck 处理就绪性检查请求
// 检查服务是否准备好处理请求，包括检查依赖服务（数据库、Redis等）
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// 检查数据库连接
	dbStatus := "ok"
	if err := h.db.Ping(); err != nil {
		logger.Error("Database connection check failed", zap.Error(err))
		dbStatus = "error"
		c.JSON(http.StatusServiceUnavailable, types.CommonResponse{
			Code:    503,
			Message: "Service is not ready",
			Data: map[string]interface{}{
				"status":    "error",
				"database":  dbStatus,
				"timestamp": time.Now().Unix(),
				"error":     err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, types.CommonResponse{
		Code:    200,
		Message: "Service is ready",
		Data: map[string]interface{}{
			"status":    "ok",
			"database":  dbStatus,
			"timestamp": time.Now().Unix(),
		},
	})
}
