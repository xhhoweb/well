package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"well_go/internal/pkg/response"
	"well_go/internal/service"
)

// TagHandler Tag API Handler
type TagHandler struct {
	svc *service.TagService
}

// NewTagHandler 创建 TagHandler
func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

// List GET /api/v1/tags
func (h *TagHandler) List(c *gin.Context) {
	list, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, list)
}

// Hot GET /api/v1/tags/hot
func (h *TagHandler) Hot(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	list, err := h.svc.GetHot(c.Request.Context(), limit)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, list)
}

// Get GET /api/v1/tag/:tag_id
func (h *TagHandler) Get(c *gin.Context) {
	tagIDStr := c.Param("tag_id")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		response.BadRequest(c, "invalid tag_id")
		return
	}

	dto, err := h.svc.Get(c.Request.Context(), tagID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	if dto == nil {
		response.NotFound(c, "tag not found")
		return
	}

	response.Success(c, dto)
}

// GetByThread GET /api/v1/tags/thread/:tid
func (h *TagHandler) GetByThread(c *gin.Context) {
	tidStr := c.Param("tid")
	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid tid")
		return
	}

	list, err := h.svc.GetByThread(c.Request.Context(), tid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, list)
}
