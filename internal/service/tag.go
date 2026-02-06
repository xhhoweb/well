package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"well_go/internal/core/config"
	"well_go/internal/core/logger"
	"well_go/internal/model"
	"well_go/internal/pkg/pool"
	"well_go/internal/repository"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// TagService Tag 业务服务
type TagService struct {
	repo      repository.TagRepository
	threadTag repository.ThreadTagRepository
	l1        *pool.BigCache // L1 缓存（零GC）
	l2        *redis.Client
	sf        *singleflight.Group
	config    *config.CacheConfig
}

// TagDTO 标签数据传输对象
type TagDTO struct {
	TagID   int    `json:"tag_id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Threads int    `json:"threads"`
	View    int    `json:"view"`
	Status  int    `json:"status"`
}

// NewTagService 创建 TagService 实例
func NewTagService(repo repository.TagRepository, threadTag repository.ThreadTagRepository, l2 *redis.Client, cfg *config.CacheConfig) *TagService {
	l1Cache, _ := pool.NewBigCache(cfg.L1Cap, time.Duration(cfg.L2TTL)*time.Second)
	return &TagService{
		repo:      repo,
		threadTag: threadTag,
		l1:        l1Cache,
		l2:        l2,
		sf:        &singleflight.Group{},
		config:    cfg,
	}
}

// Get 获取单个 Tag
func (s *TagService) Get(ctx context.Context, tagID int) (*TagDTO, error) {
	key := fmt.Sprintf("tag:%d", tagID)

	// L1 Cache
	if s.l1 != nil {
		if data, ok := s.l1.Get(key); ok {
			if data != nil {
				var dto TagDTO
				if err := json.Unmarshal(data, &dto); err == nil {
					return &dto, nil
				}
			}
		}
	}

	// L2 Cache
	ctxL2 := context.Background()
	if v, err := s.l2.Get(ctxL2, key).Bytes(); err == nil {
		var dto TagDTO
		if err := dto.UnmarshalBinary(v); err == nil {
			// Write L1
			if s.l1 != nil {
				if bytes, _ := json.Marshal(&dto); bytes != nil {
					s.l1.Set(key, bytes)
				}
			}
			return &dto, nil
		}
	}

	// SingleFlight + DB
	v, err, _ := s.sf.Do(key, func() (interface{}, error) {
		t, err := s.repo.GetByID(ctx, tagID)
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, nil
		}
		dto := &TagDTO{
			TagID:   t.TagID,
			Name:    t.Name,
			Slug:    t.Slug,
			Threads: t.Threads,
			View:    t.View,
			Status:  t.Status,
		}
		// Write Cache
		if bytes, err := dto.MarshalBinary(); err == nil {
			s.l2.Set(ctxL2, key, bytes, time.Duration(s.config.L2TTL)*time.Second)
		}
		// Write L1
		if s.l1 != nil {
			if bytes, _ := json.Marshal(&dto); bytes != nil {
				s.l1.Set(key, bytes)
			}
		}
		return dto, nil
	})

	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	return v.(*TagDTO), nil
}

// GetByName 根据名称获取 Tag
func (s *TagService) GetByName(ctx context.Context, name string) (*TagDTO, error) {
	t, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}
	return &TagDTO{
		TagID:   t.TagID,
		Name:    t.Name,
		Slug:    t.Slug,
		Threads: t.Threads,
		View:    t.View,
		Status:  t.Status,
	}, nil
}

// GetAll 获取所有 Tag
func (s *TagService) GetAll(ctx context.Context) ([]*TagDTO, error) {
	tags, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*TagDTO, 0, len(tags))
	for _, t := range tags {
		list = append(list, &TagDTO{
			TagID:   t.TagID,
			Name:    t.Name,
			Slug:    t.Slug,
			Threads: t.Threads,
			View:    t.View,
			Status:  t.Status,
		})
	}
	return list, nil
}

// GetHot 获取热门 Tag
func (s *TagService) GetHot(ctx context.Context, limit int) ([]*TagDTO, error) {
	tags, err := s.repo.GetHot(ctx, limit)
	if err != nil {
		return nil, err
	}

	list := make([]*TagDTO, 0, len(tags))
	for _, t := range tags {
		list = append(list, &TagDTO{
			TagID:   t.TagID,
			Name:    t.Name,
			Slug:    t.Slug,
			Threads: t.Threads,
			View:    t.View,
			Status:  t.Status,
		})
	}
	return list, nil
}

// GetByThread 获取主题关联的 Tag
func (s *TagService) GetByThread(ctx context.Context, tid int64) ([]*TagDTO, error) {
	tags, err := s.repo.GetByThread(ctx, tid)
	if err != nil {
		return nil, err
	}

	list := make([]*TagDTO, 0, len(tags))
	for _, t := range tags {
		list = append(list, &TagDTO{
			TagID:   t.TagID,
			Name:    t.Name,
			Slug:    t.Slug,
			Threads: t.Threads,
			View:    t.View,
			Status:  t.Status,
		})
	}
	return list, nil
}

// Create 创建 Tag
func (s *TagService) Create(ctx context.Context, name string) (*TagDTO, error) {
	// 检查是否已存在
	exist, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return &TagDTO{
			TagID:   exist.TagID,
			Name:    exist.Name,
			Slug:    exist.Slug,
			Threads: exist.Threads,
			View:    exist.View,
			Status:  exist.Status,
		}, nil
	}

	tag := &model.Tag{
		Name:    name,
		Slug:    "",
		Threads: 0,
		View:    0,
		Status:  0,
	}

	id, err := s.repo.Create(ctx, tag)
	if err != nil {
		logger.Error("create tag failed", logger.String("error", err.Error()))
		return nil, err
	}

	return &TagDTO{
		TagID:   id,
		Name:    name,
		Slug:    tag.Slug,
		Threads: 0,
		View:    0,
		Status:  0,
	}, nil
}

// AddToThread 将 Tag 关联到主题
func (s *TagService) AddToThread(ctx context.Context, tid int64, tagName string) error {
	// 获取或创建 Tag
	tag, err := s.repo.GetByName(ctx, tagName)
	if err != nil {
		return err
	}

	if tag == nil {
		newTag := &model.Tag{
			Name:    tagName,
			Threads: 0,
			View:    0,
			Status:  0,
		}
		id, err := s.repo.Create(ctx, newTag)
		if err != nil {
			return err
		}
		tag = &model.Tag{
			TagID:   id,
			Name:    tagName,
			Threads: 0,
			View:    0,
			Status:  0,
		}
	}

	// 检查是否已关联
	tagIDs, err := s.threadTag.GetByThread(ctx, tid)
	if err != nil {
		return err
	}
	for _, id := range tagIDs {
		if id == tag.TagID {
			return nil // 已关联
		}
	}

	// 创建关联
	tt := &model.ThreadTag{
		Tid:   tid,
		TagID: tag.TagID,
	}
	if err := s.threadTag.Create(ctx, tt); err != nil {
		return err
	}

	// 增加 Tag 关联数
	return s.repo.IncThreads(ctx, tag.TagID)
}

// RemoveFromThread 将 Tag 从主题移除
func (s *TagService) RemoveFromThread(ctx context.Context, tid int64, tagID int) error {
	// 删除关联
	tagIDs, err := s.threadTag.GetByThread(ctx, tid)
	if err != nil {
		return err
	}

	for _, id := range tagIDs {
		if id == tagID {
			if err := s.threadTag.Delete(ctx, tid, tagID); err != nil {
				return err
			}
			if err := s.repo.DecThreads(ctx, tagID); err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

// FlushCache 刷新缓存
func (s *TagService) FlushCache(ctx context.Context) error {
	if s.l1 != nil {
		s.l1.Flush()
	}
	return nil
}

// MarshalBinary 序列化
func (dto *TagDTO) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 0)

	tagID := dto.TagID
	buf = append(buf, byte(tagID))
	buf = append(buf, byte(tagID>>8))
	buf = append(buf, byte(tagID>>16))
	buf = append(buf, byte(tagID>>24))

	nameLen := len(dto.Name)
	buf = append(buf, byte(nameLen))
	buf = append(buf, []byte(dto.Name)...)

	slugLen := len(dto.Slug)
	buf = append(buf, byte(slugLen))
	buf = append(buf, []byte(dto.Slug)...)

	buf = append(buf, byte(dto.Status))

	return buf, nil
}

// UnmarshalBinary 反序列化
func (dto *TagDTO) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length")
	}

	offset := 0
	dto.TagID = int(int64(data[0]) | int64(data[1])<<8 | int64(data[2])<<16 | int64(data[3])<<24)
	offset = 4

	nameLen := int(data[offset])
	offset++
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid name length")
	}
	dto.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	slugLen := int(data[offset])
	offset++
	if len(data) < offset+slugLen {
		return fmt.Errorf("invalid slug length")
	}
	dto.Slug = string(data[offset : offset+slugLen])

	dto.Status = int(data[offset+slugLen])

	return nil
}
