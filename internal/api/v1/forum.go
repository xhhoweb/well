package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"well_go/internal/pkg/response"
	"well_go/internal/service"
)

// ForumHandler Forum API Handler
type ForumHandler struct {
	svc *service.ForumService
}

// NewForumHandler 创建 ForumHandler
func NewForumHandler(svc *service.ForumService) *ForumHandler {
	return &ForumHandler{svc: svc}
}

// List GET /api/v1/forums
func (h *ForumHandler) List(c *gin.Context) {
	list, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, list)
}

// Tree GET /api/v1/forums/tree
func (h *ForumHandler) Tree(c *gin.Context) {
	tree, err := h.svc.GetTree(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, tree)
}

// Get GET /api/v1/forum/:fid
func (h *ForumHandler) Get(c *gin.Context) {
	fidStr := c.Param("fid")
	fid, err := strconv.Atoi(fidStr)
	if err != nil {
		response.BadRequest(c, "invalid fid")
		return
	}

	dto, err := h.svc.Get(c.Request.Context(), fid)
	if err != nil {
		response.Fail(c, err)
		return
	}

	if dto == nil {
		response.NotFound(c, "forum not found")
		return
	}

	response.Success(c, dto)
}
