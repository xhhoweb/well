package model

import "time"

// User 用户模型
type User struct {
	Uid       int64     `db:"uid"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	Email     string    `db:"email"`
	Avatar    string    `db:"avatar"`
	Role      int       `db:"role"`      // 0: 普通用户, 1: 管理员
	Status    int       `db:"status"`    // 0: 正常, 1: 禁用
	Dateline  int       `db:"dateline"`  // 注册时间
	Lastvisit int       `db:"lastvisit"` // 最后访问时间
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	Uid      int64  `json:"uid"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Role     int    `json:"role"`
	Status   int    `json:"status"`
	Dateline int    `json:"dateline"`
}

// UserProfile 用户详细信息（仅本人可见）
type UserProfile struct {
	UserDTO
	Lastvisit int `json:"lastvisit"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User UserDTO `json:"user"`
}
