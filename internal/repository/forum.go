package repository

import (
	"context"
	"database/sql"

	"well_go/internal/model"

	"github.com/jmoiron/sqlx"
)

// ForumRepository Forum 数据访问接口
type ForumRepository interface {
	GetByID(ctx context.Context, fid int) (*model.Forum, error)
	GetAll(ctx context.Context) ([]*model.Forum, error)
	GetTree(ctx context.Context) ([]*model.Forum, error)
	GetByParent(ctx context.Context, parent int) ([]*model.Forum, error)
	Create(ctx context.Context, forum *model.Forum) (int, error)
	Update(ctx context.Context, forum *model.Forum) error
	Delete(ctx context.Context, fid int) error
	IncThreads(ctx context.Context, fid int) error
	IncToday(ctx context.Context, fid int) error
}

// forumRepository Forum 数据访问实现
type forumRepository struct {
	db *sqlx.DB
}

// NewForumRepository 创建 ForumRepository 实例
func NewForumRepository(db *sqlx.DB) ForumRepository {
	return &forumRepository{db: db}
}

// GetByID 根据 ID 获取 Forum
func (r *forumRepository) GetByID(ctx context.Context, fid int) (*model.Forum, error) {
	var forum model.Forum
	err := r.db.GetContext(ctx, &forum, "SELECT * FROM forum WHERE fid = ?", fid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &forum, nil
}

// GetAll 获取所有 Forum
func (r *forumRepository) GetAll(ctx context.Context) ([]*model.Forum, error) {
	var forums []*model.Forum
	err := r.db.SelectContext(ctx, &forums, "SELECT * FROM forum ORDER BY `order` ASC")
	if err != nil {
		return nil, err
	}
	return forums, nil
}

// GetTree 获取论坛树（一级版块 + 子版块）
func (r *forumRepository) GetTree(ctx context.Context) ([]*model.Forum, error) {
	var forums []*model.Forum
	err := r.db.SelectContext(ctx, &forums, "SELECT * FROM forum ORDER BY `order` ASC")
	if err != nil {
		return nil, err
	}
	return forums, nil
}

// GetByParent 根据父版块获取子版块
func (r *forumRepository) GetByParent(ctx context.Context, parent int) ([]*model.Forum, error) {
	var forums []*model.Forum
	err := r.db.SelectContext(ctx, &forums, "SELECT * FROM forum WHERE parent = ? ORDER BY `order` ASC", parent)
	if err != nil {
		return nil, err
	}
	return forums, nil
}

// Create 创建 Forum
func (r *forumRepository) Create(ctx context.Context, forum *model.Forum) (int, error) {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO forum (name, parent, path, depth, `order`, threads, today, posts, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		forum.Name, forum.Parent, forum.Path, forum.Depth, forum.Order, forum.Threads, forum.Today, forum.Posts, forum.Status)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

// Update 更新 Forum
func (r *forumRepository) Update(ctx context.Context, forum *model.Forum) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE forum SET name = ?, parent = ?, path = ?, depth = ?, `order` = ?, status = ? WHERE fid = ?",
		forum.Name, forum.Parent, forum.Path, forum.Depth, forum.Order, forum.Status, forum.Fid)
	return err
}

// Delete 删除 Forum
func (r *forumRepository) Delete(ctx context.Context, fid int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM forum WHERE fid = ?", fid)
	return err
}

// IncThreads 增加主题数
func (r *forumRepository) IncThreads(ctx context.Context, fid int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE forum SET threads = threads + 1 WHERE fid = ?", fid)
	return err
}

// IncToday 增加今日主题数
func (r *forumRepository) IncToday(ctx context.Context, fid int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE forum SET today = today + 1, threads = threads + 1 WHERE fid = ?", fid)
	return err
}
