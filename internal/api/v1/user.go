package v1

import (
	"net/http"

	"well_go/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户公开API
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetUser 获取用户信息（公开）
// GET /api/v1/user/:uid
func (h *UserHandler) GetUser(c *gin.Context) {
	uid := ParseID(c.Param("uid"))
	if uid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的用户ID",
		})
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": user,
		"msg":  "success",
	})
}

// GetUserProfile 获取用户详细资料（需要登录）
// GET /api/v1/user/profile
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	// 从JWT中获取用户ID
	uid := GetUIDFromContext(c)
	if uid <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "未登录",
		})
		return
	}

	profile, err := h.svc.GetProfile(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": profile,
		"msg":  "success",
	})
}

// ParseID 解析ID
func ParseID(s string) int64 {
	var id int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		id = id*10 + int64(c-'0')
	}
	return id
}

// GetUIDFromContext 从上下文获取UID
func GetUIDFromContext(c *gin.Context) int64 {
	if v, exists := c.Get("uid"); exists {
		if uid, ok := v.(int64); ok {
			return uid
		}
	}
	return 0
}
