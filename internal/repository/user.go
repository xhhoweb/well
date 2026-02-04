package repository

import (
	"context"
	"database/sql"
	"time"

	"well_go/internal/model"

	"github.com/jmoiron/sqlx"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, uid int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateLastvisit(ctx context.Context, uid int64, timestamp int) error
	Delete(ctx context.Context, uid int64) error
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

type userRepository struct {
	db *sqlx.DB
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO user (uid, username, password, email, avatar, role, status, dateline, lastvisit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		user.Uid, user.Username, user.Password, user.Email,
		user.Avatar, user.Role, user.Status, user.Dateline, user.Lastvisit)
	return err
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, uid int64) (*model.User, error) {
	query := `
		SELECT uid, username, password, email, avatar, role, status, dateline, lastvisit, created_at, updated_at
		FROM user WHERE uid = ? AND status = 0
	`
	var user model.User
	err := r.db.GetContext(ctx, &user, query, uid)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT uid, username, password, email, avatar, role, status, dateline, lastvisit, created_at, updated_at
		FROM user WHERE username = ? AND status = 0
	`
	var user model.User
	err := r.db.GetContext(ctx, &user, query, username)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT uid, username, password, email, avatar, role, status, dateline, lastvisit, created_at, updated_at
		FROM user WHERE email = ? AND status = 0
	`
	var user model.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE user SET username=?, email=?, avatar=?, role=?, status=?, updated_at=NOW()
		WHERE uid=?
	`
	_, err := r.db.ExecContext(ctx, query,
		user.Username, user.Email, user.Avatar, user.Role, user.Status, user.Uid)
	return err
}

// UpdateLastvisit 更新最后访问时间
func (r *userRepository) UpdateLastvisit(ctx context.Context, uid int64, timestamp int) error {
	query := `UPDATE user SET lastvisit=? WHERE uid=?`
	_, err := r.db.ExecContext(ctx, query, timestamp, uid)
	return err
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(ctx context.Context, uid int64) error {
	query := `UPDATE user SET status=1 WHERE uid=?`
	_, err := r.db.ExecContext(ctx, query, uid)
	return err
}

// UserCacheKey 生成用户缓存Key
func UserCacheKey(uid int64) string {
	return "user:" + time.Now().Format("20060102150405")[:8] + ":" + itoa(uid)
}

// itoa 简化版数字转字符串
func itoa(i int64) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	negative := i < 0
	if negative {
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
