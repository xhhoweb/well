package model

import "time"

// Thread Thread主表模型
type Thread struct {
	Tid       int64     `db:"tid"`
	Fid       int       `db:"fid"`
	Uid       int64     `db:"uid"`
	Subject   string    `db:"subject"`
	Views     int       `db:"views"`
	Replies   int       `db:"replies"`
	Dateline  int       `db:"dateline"`
	Lastpost  int       `db:"lastpost"`
	Status    int       `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ThreadData Thread内容表模型
type ThreadData struct {
	Tid     int64  `db:"tid"`
	Message string `db:"message"`
}

// ThreadWithContent Thread完整信息（含内容）
type ThreadWithContent struct {
	Thread
	Message string `db:"message"`
}

// ThreadDTO Thread数据传输对象
type ThreadDTO struct {
	Tid      int64  `json:"tid"`
	Fid      int    `json:"fid"`
	Uid      int64  `json:"uid"`
	Subject  string `json:"subject"`
	Views    int    `json:"views"`
	Replies  int    `json:"replies"`
	Dateline int    `json:"dateline"`
	Lastpost int    `json:"lastpost"`
	Status   int    `json:"status"`
	Message  string `json:"message,omitempty"`
}

// ThreadListItem Thread列表项
type ThreadListItem struct {
	Tid      int64  `json:"tid"`
	Fid      int    `json:"fid"`
	Uid      int64  `json:"uid"`
	Subject  string `json:"subject"`
	Views    int    `json:"views"`
	Replies  int    `json:"replies"`
	Dateline int    `json:"dateline"`
	Lastpost int    `json:"lastpost"`
	Status   int    `json:"status"`
}
