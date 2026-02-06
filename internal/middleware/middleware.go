package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"well_go/internal/core/config"
	"well_go/internal/core/logger"
)

// LoggerMiddleware 请求日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("request",
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.String("query", query),
			logger.Int("status", status),
			logger.Duration("latency", latency),
			logger.String("client_ip", c.ClientIP()),
		)
	}
}

// RecoveryMiddleware 异常恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					logger.String("error", fmt.Sprintf("%v", err)))
				c.AbortWithStatusJSON(500, gin.H{
					"code": 500,
					"msg":  "internal server error",
				})
			}
		}()
		c.Next()
	}
}

// TimeoutMiddleware 请求超时中间件
// 默认超时时间 30 秒
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// 创建一个通道来接收处理结果
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		// 等待完成或超时
		select {
		case <-done:
			// 正常完成
		case <-ctx.Done():
			// 超时
			c.AbortWithStatusJSON(504, gin.H{
				"code": 504,
				"msg":  "request timeout",
			})
		}
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// UserClaims 自定义 JWT Claims
type UserClaims struct {
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

// JWTMW JWT中间件
func JWTMW(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"code": 401,
				"msg":  "unauthorized",
			})
			return
		}

		// 验证 Bearer 前缀
		if !strings.HasPrefix(token, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{
				"code": 401,
				"msg":  "invalid token format: missing 'Bearer ' prefix",
			})
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")

		// 解析JWT
		claims, err := ParseJWT(token, cfg.Secret)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"code": 401,
				"msg":  "invalid token",
			})
			return
		}

		// 提取用户信息
		if uid, ok := claims["uid"].(float64); ok {
			c.Set("uid", int64(uid))
		}
		if role, ok := claims["role"].(float64); ok {
			c.Set("role", int(role))
		}
		if username, ok := claims["username"].(string); ok {
			c.Set("username", username)
		}

		c.Next()
	}
}

// ParseJWT 解析JWT
func ParseJWT(tokenString, secret string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return map[string]interface{}(claims), nil
	}
	return nil, fmt.Errorf("invalid token")
}

// GenerateToken 生成 JWT Token
func GenerateToken(uid int64, username string, role int, cfg *config.JWTConfig) (string, error) {
	claims := UserClaims{
		UID:      uid,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.Expiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "well_go",
			Subject:   fmt.Sprintf("%d", uid),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}
