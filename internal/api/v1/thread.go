package v1

import (
	"context"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"well_go/internal/pkg/response"
	"well_go/internal/service"
)

// ThreadHandler Thread API Handler
type ThreadHandler struct {
	svc     *service.ThreadService
	tagSvc  *service.TagService
	userSvc *service.UserService
}

// NewThreadHandler 创建ThreadHandler
func NewThreadHandler(svc *service.ThreadService, tagSvc *service.TagService, userSvc *service.UserService) *ThreadHandler {
	return &ThreadHandler{svc: svc, tagSvc: tagSvc, userSvc: userSvc}
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

	uids := make([]int64, 0, len(list))
	for _, item := range list {
		uids = append(uids, item.Uid)
	}
	users, err := h.userSvc.GetUsersByIDs(c.Request.Context(), uids)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, gin.H{
		"list":      list,
		"users":     users,
		"page":      page,
		"page_size": pageSize,
	})
}

// Get GET /api/v1/thread/:tid
func (h *ThreadHandler) Get(c *gin.Context) {
	tidStr := c.Param("tid")
	tid, parseErr := strconv.ParseInt(tidStr, 10, 64)
	if parseErr != nil {
		response.BadRequest(c, "invalid tid")
		return
	}

	var dto *service.ThreadDTO
	var err error
	var (
		tags   []*service.TagDTO
		tagErr error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		dto, err = h.svc.Get(c.Request.Context(), tid)
	}()

	go func() {
		defer wg.Done()
		tags, tagErr = h.tagSvc.GetByThread(c.Request.Context(), tid)
	}()

	wg.Wait()

	if err != nil {
		response.Fail(c, err)
		return
	}
	if tagErr != nil {
		response.Fail(c, tagErr)
		return
	}
	if dto == nil {
		response.NotFound(c, "thread not found")
		return
	}

	// 详情页浏览量异步累计（不阻塞主流程）
	go h.svc.IncViews(context.Background(), tid)

	response.Success(c, gin.H{
		"thread": dto,
		"tags":   tags,
	})
}
