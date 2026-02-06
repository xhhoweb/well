package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"well_go/internal/core/config"
	"well_go/internal/core/logger"
)

// IPWhitelistConfig IP 白名单配置
type IPWhitelistConfig struct {
	AllowIPs []string // 允许的 IP 列表（支持 CIDR）
	DenyIPs  []string // 拒绝的 IP 列表
}

// ipChecker IP 检查器
type ipChecker struct {
	allowNets []*net.IPNet
	denyNets  []*net.IPNet
	allowSet  map[string]bool
	denySet   map[string]bool
}

// newIPChecker 创建 IP 检查器
func newIPChecker(cfg *IPWhitelistConfig) (*ipChecker, error) {
	c := &ipChecker{
		allowSet: make(map[string]bool),
		denySet:  make(map[string]bool),
	}

	for _, ip := range cfg.AllowIPs {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		// 尝试解析为 CIDR
		if _, net, err := net.ParseCIDR(ip); err == nil {
			c.allowNets = append(c.allowNets, net)
		} else {
			// 作为普通 IP 处理
			c.allowSet[ip] = true
		}
	}

	for _, ip := range cfg.DenyIPs {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		if _, net, err := net.ParseCIDR(ip); err == nil {
			c.denyNets = append(c.denyNets, net)
		} else {
			c.denySet[ip] = true
		}
	}

	return c, nil
}

// isLocalIP 检查是否是本地 IP (支持 IPv4 和 IPv6)
func isLocalIP(ipStr string) bool {
	// 检查 localhost (IPv4 and IPv6)
	if ipStr == "localhost" || ipStr == "127.0.0.1" || ipStr == "::1" {
		return true
	}

	// 解析 IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 检查 IPv4 内网 IP
	if ipv4 := ip.To4(); ipv4 != nil {
		// 192.168.x.x
		if ipv4[0] == 192 && ipv4[1] == 168 {
			return true
		}
		// 10.x.x.x
		if ipv4[0] == 10 {
			return true
		}
		// 172.16-31.x.x
		if ipv4[0] == 172 && ipv4[1] >= 16 && ipv4[1] <= 31 {
			return true
		}
		// 127.x.x.x (loopback)
		if ipv4[0] == 127 {
			return true
		}
	}

	// 检查 IPv6 loopback
	if ip.IsLoopback() {
		return true
	}

	return false
}

// isAllowed 检查 IP 是否被允许
func (c *ipChecker) isAllowed(ipStr string) bool {
	// 首先检查是否是本地 IP（localhost/127.0.0.1/内网）
	if isLocalIP(ipStr) {
		// 如果本地 IP 在白名单中，或者白名单为空允许本地
		if c.allowSet[ipStr] {
			return true
		}
		// 如果白名单为空，默认允许本地
		if len(c.allowSet) == 0 && len(c.allowNets) == 0 {
			return true
		}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 检查黑名单
	for _, net := range c.denyNets {
		if net.Contains(ip) {
			return false
		}
	}
	if c.denySet[ipStr] {
		return false
	}

	// 检查白名单
	for _, net := range c.allowNets {
		if net.Contains(ip) {
			return true
		}
	}
	if c.allowSet[ipStr] {
		return true
	}

	return false
}

// PublicWhitelistMW Public API IP 白名单中间件
// - 本地/内网 IP 直接放行
// - 外网 IP 需要在白名单中
func PublicWhitelistMW() gin.HandlerFunc {
	cfg := config.Get()
	whitelistCfg := &IPWhitelistConfig{
		AllowIPs: cfg.Security.AllowIPs,
		DenyIPs:  cfg.Security.DenyIPs,
	}

	checker, err := newIPChecker(whitelistCfg)
	if err != nil {
		logger.Error("failed to create IP whitelist checker",
			logger.String("error", err.Error()))
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 本地/内网 IP 直接放行
		if isLocalIP(clientIP) {
			c.Next()
			return
		}

		// 如果没有任何限制，放行
		if len(whitelistCfg.AllowIPs) == 0 && len(whitelistCfg.DenyIPs) == 0 {
			c.Next()
			return
		}

		// 检查是否在白名单中
		if checker != nil && !checker.isAllowed(clientIP) {
			logger.Warn("Public IP blocked by whitelist",
				logger.String("ip", clientIP),
				logger.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "access denied: IP not in whitelist",
			})
			return
		}

		c.Next()
	}
}

// isLocalOrAllowedIP 检查是否是本地 IP 或在白名单中 (支持 IPv4 和 IPv6)
func isLocalOrAllowedIP(clientIP string, cfg *IPWhitelistConfig, checker *ipChecker) bool {
	// 1. 检查是否是 localhost/127.0.0.1/::1
	if clientIP == "localhost" || clientIP == "127.0.0.1" || clientIP == "::1" {
		return true
	}

	// 2. 解析 IP
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	// 3. 检查是否是 loopback
	if ip.IsLoopback() {
		return true
	}

	// 4. 检查 IPv4 内网 IP
	if ipv4 := ip.To4(); ipv4 != nil {
		// 192.168.x.x
		if ipv4[0] == 192 && ipv4[1] == 168 {
			return true
		}
		// 10.x.x.x
		if ipv4[0] == 10 {
			return true
		}
		// 172.16-31.x.x
		if ipv4[0] == 172 && ipv4[1] >= 16 && ipv4[1] <= 31 {
			return true
		}
		// 127.x.x.x
		if ipv4[0] == 127 {
			return true
		}
	}

	// 5. 检查是否在白名单配置中
	if checker != nil && checker.isAllowed(clientIP) {
		return true
	}

	return false
}

// AdminWhitelistMW Admin API IP 白名单中间件
// - 自动允许 localhost/127.0.0.1/内网 IP
// - 显式配置的白名单 IP 允许
// - 其他 IP 拒绝
func AdminWhitelistMW() gin.HandlerFunc {
	cfg := config.Get()
	whitelistCfg := &IPWhitelistConfig{
		AllowIPs: cfg.Security.AllowIPs,
		DenyIPs:  cfg.Security.DenyIPs,
	}

	checker, err := newIPChecker(whitelistCfg)
	if err != nil {
		logger.Error("failed to create IP whitelist checker",
			logger.String("error", err.Error()))
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		realIP := c.GetHeader("X-Real-IP")

		// 如果有 X-Real-IP 头，优先使用它（代理场景）
		if realIP != "" && isLocalOrAllowedIP(realIP, whitelistCfg, checker) {
			logger.Debug("AdminWhitelistMW: X-Real-IP local allowed",
				logger.String("ip", realIP))
			c.Next()
			return
		}

		// 检查客户端 IP
		if isLocalOrAllowedIP(clientIP, whitelistCfg, checker) {
			logger.Debug("AdminWhitelistMW: local IP allowed",
				logger.String("ip", clientIP))
			c.Next()
			return
		}

		// 拒绝外网访问
		logger.Warn("Admin access denied: IP not in whitelist",
			logger.String("ip", clientIP),
			logger.String("real_ip", realIP),
			logger.String("path", c.Request.URL.Path))
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code": 403,
			"msg":  "access denied: IP not in whitelist",
		})
	}
}

// IPLimiter IP频率限制器
type IPLimiter struct {
	mu     sync.Mutex
	visits map[string][]int64
	limit  int
	window int64
}

// NewIPLimiter 创建IP限制器
func NewIPLimiter(limit int, windowSeconds int) *IPLimiter {
	return &IPLimiter{
		visits: make(map[string][]int64),
		limit:  limit,
		window: int64(windowSeconds),
	}
}

// Allow 检查是否允许访问
func (l *IPLimiter) Allow(ip string) bool {
	now := time.Now().Unix() // 修复：使用真实时间

	l.mu.Lock()
	defer l.mu.Unlock()

	// 清理过期记录
	var valid []int64
	for _, ts := range l.visits[ip] {
		if now-ts < l.window {
			valid = append(valid, ts)
		}
	}
	l.visits[ip] = valid

	// 检查是否超限
	if len(l.visits[ip]) >= l.limit {
		return false
	}

	// 记录访问
	l.visits[ip] = append(l.visits[ip], now)
	return true
}

// RateLimitMW 频率限制中间件
func RateLimitMW(limiter *IPLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			logger.Warn("rate limit exceeded",
				logger.String("ip", ip),
				logger.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "too many requests",
			})
			return
		}

		c.Next()
	}
}
