package mgt

import (
	"github.com/gin-gonic/gin"
	"well_go/internal/pkg/response"
	"well_go/internal/service"
)

// CacheHandler Cache Management API Handler
type CacheHandler struct {
	svc *service.ThreadService
}

// NewCacheHandler 创建CacheHandler
func NewCacheHandler(svc *service.ThreadService) *CacheHandler {
	return &CacheHandler{svc: svc}
}

// Flush POST /api/mgt/cache/flush
func (h *CacheHandler) Flush(c *gin.Context) {
	if err := h.svc.FlushCache(c.Request.Context()); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessWithMsg(c, nil, "cache flushed")
}

// Prewarm POST /api/mgt/cache/prewarm
func (h *CacheHandler) Prewarm(c *gin.Context) {
	// TODO: 实现缓存预热逻辑
	response.SuccessWithMsg(c, nil, "cache prewarm started")
}
