package repository

import (
	"context"
	"database/sql"
	"strings"

	"well_go/internal/model"

	"github.com/jmoiron/sqlx"
)

// TagRepository Tag 数据访问接口
type TagRepository interface {
	GetByID(ctx context.Context, tagID int) (*model.Tag, error)
	GetByName(ctx context.Context, name string) (*model.Tag, error)
	GetBySlug(ctx context.Context, slug string) (*model.Tag, error)
	GetAll(ctx context.Context) ([]*model.Tag, error)
	GetHot(ctx context.Context, limit int) ([]*model.Tag, error)
	GetByThread(ctx context.Context, tid int64) ([]*model.Tag, error)
	Create(ctx context.Context, tag *model.Tag) (int, error)
	Update(ctx context.Context, tag *model.Tag) error
	Delete(ctx context.Context, tagID int) error
	IncThreads(ctx context.Context, tagID int) error
	IncView(ctx context.Context, tagID int) error
	// Sitemap 专用方法
	GetSitemapList(ctx context.Context, offset, limit int) ([]*model.Tag, error)
}

// tagRepository Tag 数据访问实现
type tagRepository struct {
	db *sqlx.DB
}

// NewTagRepository 创建 TagRepository 实例
func NewTagRepository(db *sqlx.DB) TagRepository {
	return &tagRepository{db: db}
}

// GetByID 根据 ID 获取 Tag
func (r *tagRepository) GetByID(ctx context.Context, tagID int) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.GetContext(ctx, &tag, "SELECT * FROM tag WHERE tag_id = ?", tagID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// GetByName 根据名称获取 Tag
func (r *tagRepository) GetByName(ctx context.Context, name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.GetContext(ctx, &tag, "SELECT * FROM tag WHERE name = ?", name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// GetBySlug 根据 slug 获取 Tag
func (r *tagRepository) GetBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.GetContext(ctx, &tag, "SELECT * FROM tag WHERE slug = ?", slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// GetAll 获取所有 Tag
func (r *tagRepository) GetAll(ctx context.Context) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.SelectContext(ctx, &tags, "SELECT * FROM tag ORDER BY threads DESC")
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// GetHot 获取热门 Tag
func (r *tagRepository) GetHot(ctx context.Context, limit int) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.SelectContext(ctx, &tags, "SELECT * FROM tag ORDER BY threads DESC, view DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// GetByThread 获取主题关联的 Tag
func (r *tagRepository) GetByThread(ctx context.Context, tid int64) ([]*model.Tag, error) {
	var tags []*model.Tag
	query := `
		SELECT t.* FROM tag t
		INNER JOIN thread_tag tt ON t.tag_id = tt.tag_id
		WHERE tt.tid = ?
		ORDER BY t.threads DESC
	`
	err := r.db.SelectContext(ctx, &tags, query, tid)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// Create 创建 Tag
func (r *tagRepository) Create(ctx context.Context, tag *model.Tag) (int, error) {
	// 生成 slug
	if tag.Slug == "" {
		tag.Slug = generateSlug(tag.Name)
	}

	result, err := r.db.ExecContext(ctx,
		"INSERT INTO tag (name, slug, threads, view, status) VALUES (?, ?, ?, ?, ?)",
		tag.Name, tag.Slug, tag.Threads, tag.View, tag.Status)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

// Update 更新 Tag
func (r *tagRepository) Update(ctx context.Context, tag *model.Tag) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE tag SET name = ?, slug = ?, status = ? WHERE tag_id = ?",
		tag.Name, tag.Slug, tag.Status, tag.TagID)
	return err
}

// Delete 删除 Tag
func (r *tagRepository) Delete(ctx context.Context, tagID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tag WHERE tag_id = ?", tagID)
	return err
}

// IncThreads 增加关联主题数
func (r *tagRepository) IncThreads(ctx context.Context, tagID int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE tag SET threads = threads + 1 WHERE tag_id = ?", tagID)
	return err
}

// IncView 增加浏览数
func (r *tagRepository) IncView(ctx context.Context, tagID int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE tag SET view = view + 1 WHERE tag_id = ?", tagID)
	return err
}

// GetSitemapList 获取sitemap列表
func (r *tagRepository) GetSitemapList(ctx context.Context, offset, limit int) ([]*model.Tag, error) {
	var tags []*model.Tag
	err := r.db.SelectContext(ctx, &tags,
		"SELECT tag_id, name, slug FROM tag WHERE status = 0 ORDER BY threads DESC LIMIT ?, ?",
		offset, limit)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// generateSlug 生成拼音 slug
func generateSlug(name string) string {
	// 简单实现：转小写，空格变横线
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// 移除特殊字符
	var result []byte
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		}
	}
	return string(result)
}
