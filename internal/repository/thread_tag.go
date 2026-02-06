package repository

import (
	"context"

	"well_go/internal/model"

	"github.com/jmoiron/sqlx"
)

// ThreadTagRepository ThreadTag 数据访问接口
type ThreadTagRepository interface {
	GetByThread(ctx context.Context, tid int64) ([]int, error) // 返回 tagID 列表
	GetByTag(ctx context.Context, tagID int) ([]int64, error) // 返回 tid 列表
	Create(ctx context.Context, tt *model.ThreadTag) error
	Delete(ctx context.Context, tid int64, tagID int) error
	DeleteByThread(ctx context.Context, tid int64) error
	DeleteByTag(ctx context.Context, tagID int) error
}

// threadTagRepository ThreadTag 数据访问实现
type threadTagRepository struct {
	db *sqlx.DB
}

// NewThreadTagRepository 创建 ThreadTagRepository 实例
func NewThreadTagRepository(db *sqlx.DB) ThreadTagRepository {
	return &threadTagRepository{db: db}
}

// GetByThread 获取主题关联的 tagID 列表
func (r *threadTagRepository) GetByThread(ctx context.Context, tid int64) ([]int, error) {
	var tagIDs []int
	err := r.db.SelectContext(ctx, &tagIDs, "SELECT tag_id FROM thread_tag WHERE tid = ?", tid)
	if err != nil {
		return nil, err
	}
	return tagIDs, nil
}

// GetByTag 获取标签关联的 tid 列表
func (r *threadTagRepository) GetByTag(ctx context.Context, tagID int) ([]int64, error) {
	var tids []int64
	err := r.db.SelectContext(ctx, &tids, "SELECT tid FROM thread_tag WHERE tag_id = ?", tagID)
	if err != nil {
		return nil, err
	}
	return tids, nil
}

// Create 创建关联
func (r *threadTagRepository) Create(ctx context.Context, tt *model.ThreadTag) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO thread_tag (tid, tag_id) VALUES (?, ?)",
		tt.Tid, tt.TagID)
	return err
}

// Delete 删除指定关联
func (r *threadTagRepository) Delete(ctx context.Context, tid int64, tagID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM thread_tag WHERE tid = ? AND tag_id = ?", tid, tagID)
	return err
}

// DeleteByThread 删除主题的所有关联
func (r *threadTagRepository) DeleteByThread(ctx context.Context, tid int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM thread_tag WHERE tid = ?", tid)
	return err
}

// DeleteByTag 删除标签的所有关联
func (r *threadTagRepository) DeleteByTag(ctx context.Context, tagID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM thread_tag WHERE tag_id = ?", tagID)
	return err
}
