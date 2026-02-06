package seo

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// CanonicalConfig Canonical配置
type CanonicalConfig struct {
	BaseURL string
}

// CanonicalService Canonical服务
type CanonicalService struct {
	baseURL string
}

// NewCanonicalService 创建Canonical服务
func NewCanonicalService(baseURL string) *CanonicalService {
	return &CanonicalService{baseURL: baseURL}
}

// GenerateURL 生成规范URL
// 规则：移除所有查询参数，只保留纯路径
func (s *CanonicalService) GenerateURL(path string) string {
	// 移除查询参数
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	return fmt.Sprintf("%s%s", s.baseURL, path)
}

// GenerateThreadURL 生成帖子规范URL
func (s *CanonicalService) GenerateThreadURL(tid int64) string {
	return fmt.Sprintf("%s/thread/%d", s.baseURL, tid)
}

// GenerateForumURL 生成版块规范URL
func (s *CanonicalService) GenerateForumURL(fid int) string {
	return fmt.Sprintf("%s/forum/%d", s.baseURL, fid)
}

// GenerateTagURL 生成标签规范URL
func (s *CanonicalService) GenerateTagURL(slug string) string {
	return fmt.Sprintf("%s/tag/%s", s.baseURL, slug)
}

// CanonicalMW 中间件：自动设置Canonical Header
func (s *CanonicalService) CanonicalMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取原始路径（不含查询参数）
		path := c.Request.URL.Path
		canonicalURL := s.GenerateURL(path)

		// 设置Canonical Header
		c.Header("Link", fmt.Sprintf("<%s>; rel=\"canonical\"", canonicalURL))

		c.Next()
	}
}

// CanonicalHandler Canonical处理器
type CanonicalHandler struct {
	svc *CanonicalService
}

// NewCanonicalHandler 创建Canonical处理器
func NewCanonicalHandler(svc *CanonicalService) *CanonicalHandler {
	return &CanonicalHandler{svc: svc}
}

// GetCanonical 获取规范URL
func (h *CanonicalHandler) GetCanonical(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = c.Request.URL.Path
	}

	canonicalURL := h.svc.GenerateURL(path)
	c.JSON(200, gin.H{
		"canonical": canonicalURL,
	})
}
