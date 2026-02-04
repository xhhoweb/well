package service

import (
	"context"
	"fmt"
	"time"

	"well_go/internal/core/config"
	"well_go/internal/core/logger"
	"well_go/internal/core/snowflake"
	"well_go/internal/pkg/pool"
	"well_go/internal/model"
	"well_go/internal/repository"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var ErrThreadNotFound = fmt.Errorf("thread not found")

// ThreadService Thread业务服务
type ThreadService struct {
	repo     repository.ThreadRepository
	l1       *pool.SimpleCache[int64, *ThreadDTO] // L1 Cache
	l2       *redis.Client
	sf       *singleflight.Group
	l2Config *config.CacheConfig
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

// ThreadListItem 列表项
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

// NewThreadService 创建ThreadService实例
func NewThreadService(repo repository.ThreadRepository, l2 *redis.Client, l2Config *config.CacheConfig) *ThreadService {
	return &ThreadService{
		repo:     repo,
		l1:       pool.NewCache[int64, *ThreadDTO](l2Config.L1Cap),
		l2:       l2,
		sf:       &singleflight.Group{},
		l2Config: l2Config,
	}
}

// Get 获取单个Thread
func (s *ThreadService) Get(ctx context.Context, tid int64) (*ThreadDTO, error) {
	key := fmt.Sprintf("thread:%d", tid)

	// L1 Cache
	if v, ok := s.l1.Get(tid); ok {
		return v, nil
	}

	// L2 Cache
	ctxL2 := context.Background()
	if v, err := s.l2.Get(ctxL2, key).Bytes(); err == nil {
		var dto ThreadDTO
		if err := dto.UnmarshalBinary(v); err == nil {
			s.l1.Set(tid, &dto)
			return &dto, nil
		}
	}

	// SingleFlight + DB
	v, err, _ := s.sf.Do(key, func() (interface{}, error) {
		thread, err := s.repo.GetByID(ctx, tid)
		if err != nil {
			return nil, err
		}
		if thread == nil {
			return nil, nil
		}

		data, _ := s.repo.GetContentByID(ctx, tid)
		dto := &ThreadDTO{
			Tid:      thread.Tid,
			Fid:      thread.Fid,
			Uid:      thread.Uid,
			Subject:  thread.Subject,
			Views:    thread.Views,
			Replies:  thread.Replies,
			Dateline: int(thread.Dateline),
			Lastpost: int(thread.Lastpost),
			Status:   thread.Status,
		}
		if data != nil {
			dto.Message = data.Message
		}

		// Write Cache
		if bytes, err := dto.MarshalBinary(); err == nil {
			s.l2.Set(ctxL2, key, bytes, time.Duration(s.l2Config.L2TTL)*time.Second)
		}
		s.l1.Set(tid, dto)

		return dto, nil
	})

	if err != nil {
		return nil, err
	}
	return v.(*ThreadDTO), nil
}

// List 获取Thread列表
func (s *ThreadService) List(ctx context.Context, fid int, page, pageSize int) ([]*ThreadListItem, error) {
	offset := (page - 1) * pageSize
	threads, err := s.repo.GetByFid(ctx, fid, offset, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]*ThreadListItem, 0, len(threads))
	for _, t := range threads {
		list = append(list, &ThreadListItem{
			Tid:      t.Tid,
			Fid:      t.Fid,
			Uid:      t.Uid,
			Subject:  t.Subject,
			Views:    t.Views,
			Replies:  t.Replies,
			Dateline: int(t.Dateline),
			Lastpost: int(t.Lastpost),
			Status:   t.Status,
		})
	}
	return list, nil
}

// Create 创建Thread
func (s *ThreadService) Create(ctx context.Context, fid int64, uid int64, subject, message string) (*ThreadDTO, error) {
	now := time.Now().Unix()
	tid := snowflake.Generate()

	thread := &model.Thread{
		Tid:       tid,
		Fid:       int(fid),
		Uid:       uid,
		Subject:   subject,
		Views:     0,
		Replies:   0,
		Dateline:  int(now),
		Lastpost:  int(now),
		Status:    0,
	}

	content := &model.ThreadData{
		Tid:     tid,
		Message: message,
	}

	if _, err := s.repo.Create(ctx, thread, content); err != nil {
		logger.Error("create thread failed", logger.String("error", err.Error()))
		return nil, err
	}

	return &ThreadDTO{
		Tid:      tid,
		Fid:      thread.Fid,
		Uid:      thread.Uid,
		Subject:  subject,
		Views:    0,
		Replies:  0,
		Dateline: int(now),
		Lastpost: int(now),
		Status:   0,
		Message:  message,
	}, nil
}

// Update 更新Thread
func (s *ThreadService) Update(ctx context.Context, tid int64, subject string, status int) error {
	thread, err := s.repo.GetByID(ctx, tid)
	if err != nil {
		return err
	}
	if thread == nil {
		return ErrThreadNotFound
	}

	thread.Subject = subject
	thread.Status = status
	thread.Lastpost = int(time.Now().Unix())

	if err := s.repo.Update(ctx, thread); err != nil {
		return err
	}

	// Invalidate Cache
	s.l1.Remove(tid)
	key := fmt.Sprintf("thread:%d", tid)
	s.l2.Del(context.Background(), key)

	return nil
}

// Delete 删除Thread
func (s *ThreadService) Delete(ctx context.Context, tid int64) error {
	thread, err := s.repo.GetByID(ctx, tid)
	if err != nil {
		return err
	}
	if thread == nil {
		return ErrThreadNotFound
	}

	if err := s.repo.Delete(ctx, tid); err != nil {
		return err
	}

	// Invalidate Cache
	s.l1.Remove(tid)
	key := fmt.Sprintf("thread:%d", tid)
	s.l2.Del(context.Background(), key)

	return nil
}

// IncViews 增加浏览量
func (s *ThreadService) IncViews(ctx context.Context, tid int64) error {
	if err := s.repo.IncViews(ctx, tid); err != nil {
		return err
	}

	// Invalidate Cache
	s.l1.Remove(tid)
	key := fmt.Sprintf("thread:%d", tid)
	s.l2.Del(context.Background(), key)

	return nil
}

// FlushCache 刷新缓存
func (s *ThreadService) FlushCache(ctx context.Context) error {
	s.l1.Flush()
	// Redis flush 需要单独处理
	return nil
}
