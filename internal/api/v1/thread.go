package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"well_go/internal/pkg/response"
	"well_go/internal/service"
)

// ThreadHandler Thread API Handler
type ThreadHandler struct {
	svc *service.ThreadService
}

// NewThreadHandler 创建ThreadHandler
func NewThreadHandler(svc *service.ThreadService) *ThreadHandler {
	return &ThreadHandler{svc: svc}
}

// List GET /api/v1/threads
func (h *ThreadHandler) List(c *gin.Context) {
	fidStr := c.Query("fid")
	if fidStr == "" {
		response.BadRequest(c, "fid is required")
		return
	}

	fid, err := strconv.Atoi(fidStr)
	if err != nil {
		response.BadRequest(c, "invalid fid")
		return
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	list, err := h.svc.List(c.Request.Context(), fid, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, gin.H{
		"list":      list,
		"page":      page,
		"page_size": pageSize,
	})
}

// Get GET /api/v1/thread/:tid
func (h *ThreadHandler) Get(c *gin.Context) {
	tidStr := c.Param("tid")
	tid, err := strconv.ParseInt(tidStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid tid")
		return
	}

	dto, err := h.svc.Get(c.Request.Context(), tid)
	if err != nil {
		response.Fail(c, err)
		return
	}

	if dto == nil {
		response.NotFound(c, "thread not found")
		return
	}

	response.Success(c, dto)
}
