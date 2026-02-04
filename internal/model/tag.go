package model

import "time"

// Tag 标签模型
type Tag struct {
	TagID    int       `db:"tag_id"`
	Name     string    `db:"name"`
	Slug     string    `db:"slug"`      // 拼音 slug
	Threads  int       `db:"threads"`    // 关联主题数
	View     int       `db:"view"`      // 浏览次数
	Status   int       `db:"status"`    // 状态
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TagDTO 标签数据传输对象
type TagDTO struct {
	TagID   int    `json:"tag_id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Threads int   `json:"threads"`
	View    int   `json:"view"`
	Status  int   `json:"status"`
}

// ThreadTag 主题标签关联
type ThreadTag struct {
	ID     int   `db:"id"`
	Tid    int64 `db:"tid"`
	TagID  int   `db:"tag_id"`
}
