package mgt

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"well_go/internal/core/config"
	"well_go/internal/pkg/response"
	"well_go/internal/service"
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
func Login(c *gin.Context, userSvc *service.UserService, cfg *config.JWTConfig) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 调用UserService进行真实验证
	resp, err := userSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.FailWithCode(c, 401, err.Error())
		return
	}

	response.Success(c, resp)
}

// generateJWT 生成JWT
func generateJWT(uid int64, role int, cfg *config.JWTConfig) (string, error) {
	claims := jwt.MapClaims{
		"uid":      uid,
		"username": "",
		"role":     role,
		"exp":      time.Now().Add(time.Duration(cfg.Expiry) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}
