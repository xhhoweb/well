package mgt

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"well_go/internal/pkg/response"
	"well_go/internal/service"
)

// ForumHandler Forum Management API Handler
type ForumMgtHandler struct {
	svc *service.ForumService
}

// NewForumMgtHandler 创建 ForumMgtHandler
func NewForumMgtHandler(svc *service.ForumService) *ForumMgtHandler {
	return &ForumMgtHandler{svc: svc}
}

// CreateRequest 创建 Forum 请求
type CreateForumRequest struct {
	Name   string `json:"name" binding:"required"`
	Parent int    `json:"parent"`
}

// Create POST /api/mgt/forum
func (h *ForumMgtHandler) Create(c *gin.Context) {
	var req CreateForumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dto, err := h.svc.Create(c.Request.Context(), req.Name, req.Parent)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, dto)
}

// UpdateRequest 更新 Forum 请求
type UpdateForumRequest struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}

// Update PUT /api/mgt/forum/:fid
func (h *ForumMgtHandler) Update(c *gin.Context) {
	fidStr := c.Param("fid")
	fid, err := strconv.Atoi(fidStr)
	if err != nil {
		response.BadRequest(c, "invalid fid")
		return
	}

	var req UpdateForumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), fid, req.Name, req.Status); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}

// Delete DELETE /api/mgt/forum/:fid
func (h *ForumMgtHandler) Delete(c *gin.Context) {
	fidStr := c.Param("fid")
	fid, err := strconv.Atoi(fidStr)
	if err != nil {
		response.BadRequest(c, "invalid fid")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), fid); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
