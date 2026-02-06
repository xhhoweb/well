package mgt

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"well_go/internal/pkg/apperr"
	"well_go/internal/pkg/response"
	"well_go/internal/service"
)

// ThreadNotFound 线程不存在错误
var ThreadNotFound = apperr.NewAppError(apperr.CodeThreadNotFound, "thread not found")

// ThreadHandler Thread Management API Handler
type ThreadHandler struct {
	svc    *service.ThreadService
	tagSvc *service.TagService
}

// NewThreadHandler 创建ThreadHandler
func NewThreadHandler(svc *service.ThreadService, tagSvc *service.TagService) *ThreadHandler {
	return &ThreadHandler{svc: svc, tagSvc: tagSvc}
}

// CreateRequest 创建Thread请求
type CreateRequest struct {
	Fid     int64  `json:"fid" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
	Tags    []string `json:"tags"`
}

// Create POST /api/mgt/thread
func (h *ThreadHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: 从JWT获取真实UID
	uid := int64(1)

	dto, err := h.svc.Create(c.Request.Context(), req.Fid, uid, req.Subject, req.Message)
	if err != nil {
		response.Fail(c, apperr.WrapError(err, apperr.CodeThreadCreateErr))
		return
	}

	for _, tag := range req.Tags {
		if tag == "" {
			continue
		}
		_ = h.tagSvc.AddToThread(c.Request.Context(), dto.Tid, tag)
	}

	response.Success(c, dto)
}

// UpdateRequest 更新Thread请求
type UpdateRequest struct {
	Subject string `json:"subject"`
	Status  int    `json:"status"`
	Tags    []string `json:"tags"`
}

// Update PUT /api/mgt/thread/:tid
func (h *ThreadHandler) Update(c *gin.Context) {
	tidStr := c.Param("tid")
	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid tid")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), tid, req.Subject, req.Status); err != nil {
		response.Fail(c, err)
		return
	}

	for _, tag := range req.Tags {
		if tag == "" {
			continue
		}
		_ = h.tagSvc.AddToThread(c.Request.Context(), tid, tag)
	}

	response.Success(c, nil)
}

// Delete DELETE /api/mgt/thread/:tid
func (h *ThreadHandler) Delete(c *gin.Context) {
	tidStr := c.Param("tid")
	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid tid")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), tid); err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, nil)
}
