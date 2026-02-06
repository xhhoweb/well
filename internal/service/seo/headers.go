package seo

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SEOHeadersConfig SEO头部配置
type SEOHeadersConfig struct {
	MaxAge     int // Cache-Control: max-age
	SMaxAge    int // Cache-Control: s-maxage
	StaleWhile int // Stale-While-Revalidate
}

// DefaultSEOHeadersConfig 默认配置
var DefaultSEOHeadersConfig = &SEOHeadersConfig{
	MaxAge:     60,  // 1分钟
	SMaxAge:    300, // 5分钟
	StaleWhile: 60,  // 1分钟
}

// SEOHeadersMiddleware SEO头部中间件
// 自动添加Cache-Control、Last-Modified、ETag等SEO相关Header
func SEOHeadersMiddleware(config *SEOHeadersConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultSEOHeadersConfig
	}

	return func(c *gin.Context) {
		// 不缓存POST、PUT、DELETE请求
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
			c.Next()
			return
		}

		// Cache-Control
		c.Header("Cache-Control", makeCacheControl(config))

		// ETag（基于路径和修改时间）
		if etag := makeETag(c.Request.URL.Path); etag != "" {
			c.Header("ETag", etag)
		}

		c.Next()
	}
}

// SEOHeadersForItem 为单个资源设置SEO头部
// 用于Thread、Forum、Tag等详情页
func SEOHeadersForItem(lastmod time.Time, config *SEOHeadersConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultSEOHeadersConfig
	}

	return func(c *gin.Context) {
		// Last-Modified
		c.Header("Last-Modified", lastmod.Format(httpTimeFormat))

		// ETag
		etag := makeETagWithTime(c.Request.URL.Path, lastmod)
		c.Header("ETag", etag)

		// Cache-Control
		c.Header("Cache-Control", makeCacheControl(config))

		c.Next()
	}
}

// SEOHeadersForList 为列表页设置SEO头部
func SEOHeadersForList(config *SEOHeadersConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultSEOHeadersConfig
	}

	return func(c *gin.Context) {
		// 列表页使用较长的缓存
		c.Header("Cache-Control", "public, max-age=120, s-maxage=600") // 2分钟本地，10分钟CDN
		c.Header("X-Content-Type-Options", "nosniff")

		c.Next()
	}
}

// DisableCache 禁用缓存（用于管理API）
func DisableCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// http时间格式
const httpTimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

func makeCacheControl(config *SEOHeadersConfig) string {
	var result string
	result += "public, max-age=" + strconv.Itoa(config.MaxAge)
	if config.SMaxAge > 0 {
		result += ", s-maxage=" + strconv.Itoa(config.SMaxAge)
	}
	if config.StaleWhile > 0 {
		result += ", stale-while-revalidate=" + strconv.Itoa(config.StaleWhile)
	}
	return result
}

func makeETag(path string) string {
	// 简单ETag生成
	hash := strconv.FormatUint(uint64(hashPath(path)), 16)
	return `"` + hash + `"`
}

func makeETagWithTime(path string, t time.Time) string {
	// 包含修改时间的ETag
	hash := strconv.FormatUint(hashPath(path+strconv.FormatInt(t.Unix(), 10)), 16)
	return `"` + hash + `"`
}

func hashPath(path string) uint64 {
	// FNV-1a hash
	var hash uint64 = 14695981039346656037
	for i := 0; i < len(path); i++ {
		hash ^= uint64(path[i])
		hash *= 1099511628211
	}
	return hash
}
