package seo

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// RobotsConfig 机器人配置
type RobotsConfig struct {
	BaseURL string
	Sitemap string
}

// RobotsService 机器人服务
type RobotsService struct {
	config *RobotsConfig
}

// NewRobotsService 创建机器人服务
func NewRobotsService(cfg *RobotsConfig) *RobotsService {
	return &RobotsService{config: cfg}
}

// GetRobots 获取robots.txt内容
func (s *RobotsService) GetRobots() []byte {
	return []byte(fmt.Sprintf("User-agent: *\nAllow: /\n\nSitemap: %s\n", s.config.Sitemap))
}

// RobotsHandler 机器人处理器
type RobotsHandler struct {
	svc *RobotsService
}

// NewRobotsHandler 创建机器人处理器
func NewRobotsHandler(svc *RobotsService) *RobotsHandler {
	return &RobotsHandler{svc: svc}
}

// Get 获取robots.txt
func (h *RobotsHandler) Get(c *gin.Context) {
	data := h.svc.GetRobots()

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=86400") // 24小时缓存
	c.Data(200, "text/plain; charset=utf-8", data)
}

// GetStaticRobots 获取静态robots.txt（零分配）
func GetStaticRobots(sitemapURL string) []byte {
	return []byte(fmt.Sprintf("User-agent: *\nAllow: /\n\nSitemap: %s\n", sitemapURL))
}

// DefaultRobots 默认robots.txt内容
var DefaultRobots = []byte("User-agent: *\nAllow: /\n\n")
