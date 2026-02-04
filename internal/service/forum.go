package service

import (
	"context"
	"fmt"
	"time"

	"well_go/internal/core/config"
	"well_go/internal/core/logger"
	"well_go/internal/pkg/pool"
	"well_go/internal/model"
	"well_go/internal/repository"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// ForumService Forum 业务服务
type ForumService struct {
	repo   repository.ForumRepository
	l1     *pool.SimpleCache[int, *ForumDTO] // L1 缓存
	l2     *redis.Client
	sf     *singleflight.Group
	config *config.CacheConfig
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

// ForumTreeNode 论坛树节点
type ForumTreeNode struct {
	ForumDTO
	Children []*ForumTreeNode `json:"children,omitempty"`
}

// NewForumService 创建 ForumService 实例
func NewForumService(repo repository.ForumRepository, l2 *redis.Client, cfg *config.CacheConfig) *ForumService {
	return &ForumService{
		repo:   repo,
		l1:     pool.NewCache[int, *ForumDTO](cfg.L1Cap),
		l2:     l2,
		sf:     &singleflight.Group{},
		config: cfg,
	}
}

// Get 获取单个 Forum
func (s *ForumService) Get(ctx context.Context, fid int) (*ForumDTO, error) {
	key := fmt.Sprintf("forum:%d", fid)

	// L1 Cache
	if v, ok := s.l1.Get(fid); ok {
		return v, nil
	}

	// L2 Cache
	ctxL2 := context.Background()
	if v, err := s.l2.Get(ctxL2, key).Bytes(); err == nil {
		var dto ForumDTO
		if err := dto.UnmarshalBinary(v); err == nil {
			s.l1.Set(fid, &dto)
			return &dto, nil
		}
	}

	// SingleFlight + DB
	v, err, _ := s.sf.Do(key, func() (interface{}, error) {
		f, err := s.repo.GetByID(ctx, fid)
		if err != nil {
			return nil, err
		}
		if f == nil {
			return nil, nil
		}
		dto := &ForumDTO{
			Fid:     f.Fid,
			Name:    f.Name,
			Parent:  f.Parent,
			Path:    f.Path,
			Depth:   f.Depth,
			Order:   f.Order,
			Threads: f.Threads,
			Today:   f.Today,
			Posts:   f.Posts,
			Status:  f.Status,
		}
		// Write Cache
		if bytes, err := dto.MarshalBinary(); err == nil {
			s.l2.Set(ctxL2, key, bytes, time.Duration(s.config.L2TTL)*time.Second)
		}
		s.l1.Set(fid, dto)
		return dto, nil
	})

	if err != nil {
		return nil, err
	}
	return v.(*ForumDTO), nil
}

// GetAll 获取所有 Forum
func (s *ForumService) GetAll(ctx context.Context) ([]*ForumDTO, error) {
	forums, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*ForumDTO, 0, len(forums))
	for _, f := range forums {
		list = append(list, &ForumDTO{
			Fid:     f.Fid,
			Name:    f.Name,
			Parent:  f.Parent,
			Path:    f.Path,
			Depth:   f.Depth,
			Order:   f.Order,
			Threads: f.Threads,
			Today:   f.Today,
			Posts:   f.Posts,
			Status:  f.Status,
		})
	}
	return list, nil
}

// GetTree 获取论坛树
func (s *ForumService) GetTree(ctx context.Context) ([]*ForumTreeNode, error) {
	forums, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// 构建树
	nodeMap := make(map[int]*ForumTreeNode)
	nodes := make([]*ForumTreeNode, 0, len(forums))

	for _, f := range forums {
		node := &ForumTreeNode{
			ForumDTO: ForumDTO{
				Fid:     f.Fid,
				Name:    f.Name,
				Parent:  f.Parent,
				Path:    f.Path,
				Depth:   f.Depth,
				Order:   f.Order,
				Threads: f.Threads,
				Today:   f.Today,
				Posts:   f.Posts,
				Status:  f.Status,
			},
			Children: nil,
		}
		nodeMap[f.Fid] = node
		nodes = append(nodes, node)
	}

	for _, node := range nodes {
		if node.Parent > 0 {
			if parent, ok := nodeMap[node.Parent]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	// 返回一级版块
	var roots []*ForumTreeNode
	for _, node := range nodes {
		if node.Parent == 0 {
			roots = append(roots, node)
		}
	}

	return roots, nil
}

// Create 创建 Forum
func (s *ForumService) Create(ctx context.Context, name string, parent int) (*ForumDTO, error) {
	// 计算 path 和 depth
	path := "0"
	depth := 0
	if parent > 0 {
		p, err := s.repo.GetByID(ctx, parent)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, fmt.Errorf("parent forum not found")
		}
		path = p.Path + "," + fmt.Sprintf("%d", parent)
		depth = p.Depth + 1
	}

	forum := &model.Forum{
		Name:    name,
		Parent:  parent,
		Path:    path,
		Depth:   depth,
		Order:   0,
		Threads: 0,
		Today:   0,
		Posts:   0,
		Status:  0,
	}

	id, err := s.repo.Create(ctx, forum)
	if err != nil {
		logger.Error("create forum failed", logger.String("error", err.Error()))
		return nil, err
	}

	return &ForumDTO{
		Fid:     id,
		Name:    name,
		Parent:  parent,
		Path:    path,
		Depth:   depth,
		Status:  0,
	}, nil
}

// Update 更新 Forum
func (s *ForumService) Update(ctx context.Context, fid int, name string, status int) error {
	forum, err := s.repo.GetByID(ctx, fid)
	if err != nil {
		return err
	}
	if forum == nil {
		return fmt.Errorf("forum not found")
	}

	forum.Name = name
	forum.Status = status

	if err := s.repo.Update(ctx, forum); err != nil {
		return err
	}

	// Invalidate Cache
	s.l1.Remove(fid)
	key := fmt.Sprintf("forum:%d", fid)
	s.l2.Del(context.Background(), key)

	return nil
}

// Delete 删除 Forum
func (s *ForumService) Delete(ctx context.Context, fid int) error {
	forum, err := s.repo.GetByID(ctx, fid)
	if err != nil {
		return err
	}
	if forum == nil {
		return fmt.Errorf("forum not found")
	}

	if err := s.repo.Delete(ctx, fid); err != nil {
		return err
	}

	// Invalidate Cache
	s.l1.Remove(fid)
	s.l1.Flush() // 简单起见，删除时刷新整个缓存
	key := fmt.Sprintf("forum:%d", fid)
	s.l2.Del(context.Background(), key)

	return nil
}

// FlushCache 刷新缓存
func (s *ForumService) FlushCache(ctx context.Context) error {
	s.l1.Flush()
	return nil
}

// MarshalBinary 序列化
func (dto *ForumDTO) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 0)
	buf = append(buf, byte(dto.Fid))
	buf = append(buf, byte(dto.Fid>>8))
	buf = append(buf, byte(dto.Fid>>16))
	buf = append(buf, byte(dto.Fid>>24))
	buf = append(buf, byte(dto.Fid>>32))
	buf = append(buf, byte(dto.Fid>>40))
	buf = append(buf, byte(dto.Fid>>48))
	buf = append(buf, byte(dto.Fid>>56))
	
	nameLen := len(dto.Name)
	buf = append(buf, byte(nameLen))
	buf = append(buf, []byte(dto.Name)...)
	
	buf = append(buf, byte(dto.Parent))
	buf = append(buf, byte(dto.Parent>>8))
	buf = append(buf, byte(dto.Parent>>16))
	buf = append(buf, byte(dto.Parent>>24))
	
	buf = append(buf, byte(dto.Status))
	
	return buf, nil
}

// UnmarshalBinary 反序列化
func (dto *ForumDTO) UnmarshalBinary(data []byte) error {
	if len(data) < 10 {
		return fmt.Errorf("invalid data length")
	}
	
	offset := 0
	dto.Fid = int(int64(data[0]) | int64(data[1])<<8 | int64(data[2])<<16 | int64(data[3])<<24 | 
		int64(data[4])<<32 | int64(data[5])<<40 | int64(data[6])<<48 | int64(data[7])<<56)
	offset = 8
	
	nameLen := int(data[offset])
	offset++
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid name length")
	}
	dto.Name = string(data[offset : offset+nameLen])
	offset += nameLen
	
	dto.Parent = int(int64(data[offset]) | int64(data[offset+1])<<8 | int64(data[offset+2])<<16 | int64(data[offset+3])<<24)
	offset += 4
	
	dto.Status = int(data[offset])
	
	return nil
}
