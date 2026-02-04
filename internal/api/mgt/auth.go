package mgt

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"well_go/internal/core/config"
	"well_go/internal/pkg/response"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
}

// Login POST /api/mgt/login
func Login(c *gin.Context, cfg *config.JWTConfig) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: 实现真实的用户名密码验证
	if req.Username != "admin" || req.Password != "admin123" {
		response.FailWithCode(c, 401, "invalid credentials")
		return
	}

	// 生成JWT
	claims := jwt.MapClaims{
		"uid":      1,
		"username": req.Username,
		"exp":      time.Now().Add(time.Duration(cfg.Expiry) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, LoginResponse{Token: tokenString})
}
