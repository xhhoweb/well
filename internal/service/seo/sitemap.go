package seo

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"well_go/internal/repository"

	"github.com/gin-gonic/gin"
)

// SitemapConfig Sitemap配置
type SitemapConfig struct {
	BaseURL  string
	CacheTTL time.Duration // 缓存时间
	MaxURLs  int           // 单个sitemap最大URL数
}

// SitemapService Sitemap服务
type SitemapService struct {
	repo       repository.ThreadRepository
	tagRepo    repository.TagRepository
	config     *SitemapConfig
	cache      []byte
	cacheMu    sync.RWMutex
	lastModify time.Time
}

// NewSitemapService 创建Sitemap服务
func NewSitemapService(threadRepo repository.ThreadRepository, tagRepo repository.TagRepository, cfg *SitemapConfig) *SitemapService {
	return &SitemapService{
		repo:    threadRepo,
		tagRepo: tagRepo,
		config:  cfg,
	}
}

// URLEntry sitemap URL条目
type URLEntry struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

// Handler SEO处理器
func (s *SitemapService) GetIndex(ctx context.Context) ([]byte, error) {
	s.cacheMu.RLock()
	if s.cache != nil && time.Since(s.lastModify) < s.config.CacheTTL {
		defer s.cacheMu.RUnlock()
		return s.cache, nil
	}
	s.cacheMu.RUnlock()

	// 生成 sitemap 列表
	baseURL := s.config.BaseURL
	threadCount, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取主题数量失败: %w", err)
	}
	pages := (threadCount + s.config.MaxURLs - 1) / s.config.MaxURLs
	if pages < 1 {
		pages = 1
	}

	templates := make([]URLEntry, 0, pages+1)
	for i := 1; i <= pages; i++ {
		templates = append(templates, URLEntry{Loc: fmt.Sprintf("%s/sitemap-thread-%d.xml", baseURL, i), LastMod: time.Now().Format("2006-01-02")})
	}

	if s.tagRepo != nil {
		templates = append(templates, URLEntry{
			Loc:     fmt.Sprintf("%s/sitemap-tag.xml", baseURL),
			LastMod: time.Now().Format("2006-01-02"),
		})
	}

	// 直接构建 XML（避免 template 自动转义）
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`)
	for _, entry := range templates {
		buf.WriteString("  <sitemap>\n")
		buf.WriteString("    <loc>")
		buf.WriteString(entry.Loc)
		buf.WriteString("</loc>\n")
		buf.WriteString("    <lastmod>")
		buf.WriteString(entry.LastMod)
		buf.WriteString("</lastmod>\n")
		buf.WriteString("  </sitemap>\n")
	}
	buf.WriteString("</sitemapindex>")

	s.cacheMu.Lock()
	s.cache = buf.Bytes()
	s.lastModify = time.Now()
	s.cacheMu.Unlock()

	return buf.Bytes(), nil
}

// GetThreadSitemap 获取线程分片 sitemap
func (s *SitemapService) GetThreadSitemap(ctx context.Context, page int) ([]byte, error) {
	baseURL := s.config.BaseURL
	offset := (page - 1) * s.config.MaxURLs

	threads, err := s.repo.GetSitemapList(ctx, offset, s.config.MaxURLs)
	if err != nil {
		return nil, fmt.Errorf("获取线程列表失败: %w", err)
	}

	if len(threads) == 0 {
		return nil, nil
	}

	// 直接构建 XML
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`)
	for _, t := range threads {
		buf.WriteString("  <url>\n")
		buf.WriteString("    <loc>")
		buf.WriteString(fmt.Sprintf("%s/thread/%d", baseURL, t.Tid))
		buf.WriteString("</loc>\n")
		buf.WriteString("    <lastmod>")
		buf.WriteString(time.Unix(int64(t.Lastpost), 0).Format("2006-01-02"))
		buf.WriteString("</lastmod>\n")
		buf.WriteString("    <changefreq>daily</changefreq>\n")
		buf.WriteString("    <priority>0.8</priority>\n")
		buf.WriteString("  </url>\n")
	}
	buf.WriteString("</urlset>")

	return buf.Bytes(), nil
}

// GetTagSitemap 获取标签 sitemap
func (s *SitemapService) GetTagSitemap(ctx context.Context) ([]byte, error) {
	baseURL := s.config.BaseURL

	tags, err := s.tagRepo.GetSitemapList(ctx, 0, 5000)
	if err != nil {
		return nil, fmt.Errorf("获取标签列表失败: %w", err)
	}

	if len(tags) == 0 {
		return nil, nil
	}

	// 直接构建 XML
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`)
	for _, t := range tags {
		buf.WriteString("  <url>\n")
		buf.WriteString("    <loc>")
		buf.WriteString(fmt.Sprintf("%s/tag/%s", baseURL, t.Slug))
		buf.WriteString("</loc>\n")
		buf.WriteString("    <lastmod>")
		buf.WriteString(time.Now().Format("2006-01-02"))
		buf.WriteString("</lastmod>\n")
		buf.WriteString("    <changefreq>weekly</changefreq>\n")
		buf.WriteString("    <priority>0.6</priority>\n")
		buf.WriteString("  </url>\n")
	}
	buf.WriteString("</urlset>")

	return buf.Bytes(), nil
}

// GetThreadCount 获取线程总数（用于分片）
func (s *SitemapService) GetThreadCount(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// Handler SEO处理器
type Handler struct {
	svc *SitemapService
}

// NewHandler 创建SEO处理器
func NewHandler(svc *SitemapService) *Handler {
	return &Handler{svc: svc}
}

// SitemapIndex sitemap索引
func (h *Handler) SitemapIndex(c *gin.Context) {
	data, err := h.svc.GetIndex(c.Request.Context())
	if err != nil {
		c.String(500, "internal server error")
		return
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.Header("Content-Type", "application/xml")
	c.Data(200, "application/xml", data)
}

// ThreadSitemap 线程sitemap分片
func (h *Handler) ThreadSitemap(c *gin.Context) {
	page := 1
	fmt.Sscanf(c.Param("page"), "%d", &page)
	if page < 1 {
		page = 1
	}

	data, err := h.svc.GetThreadSitemap(c.Request.Context(), page)
	if err != nil {
		c.String(500, "internal server error")
		return
	}

	if data == nil {
		c.Status(404)
		return
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.Data(200, "application/xml", data)
}

// TagSitemap 标签sitemap
func (h *Handler) TagSitemap(c *gin.Context) {
	data, err := h.svc.GetTagSitemap(c.Request.Context())
	if err != nil {
		c.String(500, "internal server error")
		return
	}

	if data == nil {
		c.Status(404)
		return
	}

	c.Header("Cache-Control", "public, max-age=300")
	c.Data(200, "application/xml", data)
}
