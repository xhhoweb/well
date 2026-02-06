package repository

import (
	"context"
	"database/sql"
	"time"

	"well_go/internal/model"

	"github.com/jmoiron/sqlx"
)

// ThreadRepository Thread数据访问接口
type ThreadRepository interface {
	GetByID(ctx context.Context, tid int64) (*model.Thread, error)
	GetContentByID(ctx context.Context, tid int64) (*model.ThreadData, error)
	GetByFid(ctx context.Context, fid int, offset, limit int) ([]*model.Thread, error)
	Create(ctx context.Context, thread *model.Thread, content *model.ThreadData) (int64, error)
	Update(ctx context.Context, thread *model.Thread) error
	Delete(ctx context.Context, tid int64) error
	IncViews(ctx context.Context, tid int64) error
	IncReplies(ctx context.Context, tid int64) error
	// Sitemap 专用方法
	GetSitemapList(ctx context.Context, offset, limit int) ([]*model.Thread, error)
	Count(ctx context.Context) (int, error)
}

// threadRepository Thread数据访问实现
type threadRepository struct {
	db *sqlx.DB
}

// NewThreadRepository 创建ThreadRepository实例
func NewThreadRepository(db *sqlx.DB) ThreadRepository {
	return &threadRepository{db: db}
}

// GetByID 根据ID获取Thread
func (r *threadRepository) GetByID(ctx context.Context, tid int64) (*model.Thread, error) {
	var thread model.Thread
	err := r.db.GetContext(ctx, &thread, "SELECT tid, fid, uid, subject, views, replies, dateline, lastpost, status FROM thread WHERE tid = ?", tid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &thread, nil
}

// GetContentByID 获取Thread内容
func (r *threadRepository) GetContentByID(ctx context.Context, tid int64) (*model.ThreadData, error) {
	var data model.ThreadData
	err := r.db.GetContext(ctx, &data, "SELECT tid, message FROM thread_data WHERE tid = ?", tid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}

// GetByFid 根据Fid获取Thread列表
func (r *threadRepository) GetByFid(ctx context.Context, fid int, offset, limit int) ([]*model.Thread, error) {
	var threads []*model.Thread
	err := r.db.SelectContext(ctx, &threads,
		"SELECT tid, fid, uid, subject, views, replies, dateline, lastpost, status FROM thread WHERE fid = ? ORDER BY lastpost DESC LIMIT ?, ?",
		fid, offset, limit)
	if err != nil {
		return nil, err
	}
	return threads, nil
}

// Create 创建Thread
func (r *threadRepository) Create(ctx context.Context, thread *model.Thread, content *model.ThreadData) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 插入主表
	result, err := tx.ExecContext(ctx,
		"INSERT INTO thread (tid, fid, uid, subject, views, replies, dateline, lastpost, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		thread.Tid, thread.Fid, thread.Uid, thread.Subject, thread.Views, thread.Replies, thread.Dateline, thread.Lastpost, thread.Status)
	if err != nil {
		return 0, err
	}

	// 插入内容表
	_, err = tx.ExecContext(ctx,
		"INSERT INTO thread_data (tid, message) VALUES (?, ?)",
		content.Tid, content.Message)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	id, _ := result.LastInsertId()
	return id, nil
}

// Update 更新Thread
func (r *threadRepository) Update(ctx context.Context, thread *model.Thread) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE thread SET subject = ?, status = ?, lastpost = ? WHERE tid = ?",
		thread.Subject, thread.Status, thread.Lastpost, thread.Tid)
	return err
}

// Delete 删除Thread
func (r *threadRepository) Delete(ctx context.Context, tid int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM thread_data WHERE tid = ?", tid)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM thread WHERE tid = ?", tid)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// IncViews 增加浏览量
func (r *threadRepository) IncViews(ctx context.Context, tid int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE thread SET views = views + 1 WHERE tid = ?", tid)
	return err
}

// IncReplies 增加回复数
func (r *threadRepository) IncReplies(ctx context.Context, tid int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE thread SET replies = replies + 1, lastpost = ? WHERE tid = ?", time.Now().Unix(), tid)
	return err
}

// GetSitemapList 获取sitemap列表（只取tid和lastpost）
func (r *threadRepository) GetSitemapList(ctx context.Context, offset, limit int) ([]*model.Thread, error) {
	var threads []*model.Thread
	err := r.db.SelectContext(ctx, &threads,
		"SELECT tid, lastpost FROM thread ORDER BY tid ASC LIMIT ?, ?",
		offset, limit)
	if err != nil {
		return nil, err
	}
	return threads, nil
}

// Count 获取线程总数
func (r *threadRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM thread")
	if err != nil {
		return 0, err
	}
	return count, nil
}
