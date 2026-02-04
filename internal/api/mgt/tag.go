package mgt

import (
	"github.com/gin-gonic/gin"
	"well_go/internal/pkg/response"
	"well_go/internal/service"
)

// TagHandler Tag Management API Handler
type TagMgtHandler struct {
	svc *service.TagService
}

// NewTagMgtHandler 创建 TagMgtHandler
func NewTagMgtHandler(svc *service.TagService) *TagMgtHandler {
	return &TagMgtHandler{svc: svc}
}

// CreateRequest 创建 Tag 请求
type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

// Create POST /api/mgt/tag
func (h *TagMgtHandler) Create(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dto, err := h.svc.Create(c.Request.Context(), req.Name)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, dto)
}

// Flush POST /api/mgt/cache/flush/tag
func (h *TagMgtHandler) Flush(c *gin.Context) {
	if err := h.svc.FlushCache(c.Request.Context()); err != nil {
		response.Fail(c, err)
		return
	}
	response.SuccessWithMsg(c, nil, "tag cache flushed")
}
