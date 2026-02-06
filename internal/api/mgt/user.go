package mgt

import (
	"net/http"

	"well_go/internal/model"
	"well_go/internal/service"

	"github.com/gin-gonic/gin"
)

// UserMgtHandler 用户管理API（内部使用）
// 注意：User 模块非核心内容模型，仅用于内部认证和基础用户管理
// 如无特殊需求，建议通过 thread.uid 关联用户信息
type UserMgtHandler struct {
	svc *service.UserService
}

// NewUserMgtHandler 创建用户管理处理器
func NewUserMgtHandler(svc *service.UserService) *UserMgtHandler {
	return &UserMgtHandler{svc: svc}
}

// Register 用户注册
// POST /api/mgt/user/register
func (h *UserMgtHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误",
		})
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
		"msg":  "success",
	})
}

// GetProfile 获取当前用户资料
// GET /api/mgt/user/profile
func (h *UserMgtHandler) GetProfile(c *gin.Context) {
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

// GetUIDFromContext 从上下文获取UID
func GetUIDFromContext(c *gin.Context) int64 {
	if v, exists := c.Get("uid"); exists {
		if uid, ok := v.(int64); ok {
			return uid
		}
	}
	return 0
}
