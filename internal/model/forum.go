package model

import "time"

// Forum 版块模型
type Forum struct {
	Fid      int       `db:"fid"`
	Name     string    `db:"name"`
	Parent   int       `db:"parent"`     // 父版块 ID（0 表示一级版块)
	ParentID int       `db:"-"`          // 父版块信息（关联查询）
	Path     string    `db:"path"`       // 路径链，如 "0,1,2"
	Depth    int       `db:"depth"`      // 深度
	Order    int       `db:"order"`     // 排序
	Threads  int       `db:"threads"`    // 主题数
	Today    int       `db:"today"`     // 今日主题
	Posts    int       `db:"posts"`     // 帖子数
	Status   int       `db:"status"`    // 状态（0正常1禁用）
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ForumDTO 版块数据传输对象
type ForumDTO struct {
	Fid      int      `json:"fid"`
	Name     string   `json:"name"`
	Parent   int      `json:"parent"`
	Path     string   `json:"path"`
	Depth    int      `json:"depth"`
	Order    int      `json:"order"`
	Threads  int      `json:"threads"`
	Today    int      `json:"today"`
	Posts    int      `json:"posts"`
	Status   int      `json:"status"`
}

// ForumTree 论坛树结构
type ForumTree struct {
	ForumDTO
	Children []*ForumTree `json:"children,omitempty"`
}

// ForumAccess 版块访问权限
type ForumAccess struct {
	ID       int       `db:"id"`
	Fid      int       `db:"fid"`
	GID      int       `db:"gid"`       // 用户组 ID
	Read     int       `db:"read"`      // 读权限（0禁止1允许）
	Thread   int       `db:"thread"`    // 发主题权限
	Reply    int       `db:"reply"`     // 回复权限
	Upload   int       `db:"upload"`    // 上传权限
	Download int       `db:"download"`  // 下载权限
}
