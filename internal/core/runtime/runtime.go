package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"well_go/internal/core/logger"
	"well_go/internal/repository"
	"well_go/internal/service"
)

// Runtime 运行时数据管理
type Runtime struct {
	forumList  []*service.ForumDTO
	forumTree  []*service.ForumTreeNode
	tagList    []*service.TagDTO
	config     map[string]string
	accessMaps map[int]map[int][]int // fid -> gid -> permissions
	mu         sync.RWMutex
	loadedAt   time.Time
}

// Singleton instance
var rt *Runtime
var once sync.Once

// RuntimeConfig Runtime 配置
type RuntimeConfig struct {
	ForumRepo repository.ForumRepository
	TagRepo   repository.TagRepository
	ForumSvc  *service.ForumService
	TagSvc    *service.TagService
}

// Init 初始化 Runtime
func Init(cfg *RuntimeConfig) error {
	var initErr error
	once.Do(func() {
		rt = &Runtime{
			config:     make(map[string]string),
			accessMaps: make(map[int]map[int][]int),
		}
		initErr = rt.warmup(cfg)
	})
	return initErr
}

// Get 获取 Runtime 实例
func Get() *Runtime {
	return rt
}

// warmup 预热数据
func (r *Runtime) warmup(cfg *RuntimeConfig) error {
	ctx := context.Background()
	start := time.Now()

	logger.Info("runtime warmup started")

	// 1. 预热 Forum 列表
	if cfg.ForumSvc != nil {
		list, err := cfg.ForumSvc.GetAll(ctx)
		if err != nil {
			logger.Error("warmup forum list failed", logger.String("error", err.Error()))
		} else {
			r.mu.Lock()
			r.forumList = list
			r.mu.Unlock()
			logger.Info("warmup forum list", logger.Int("count", len(list)))
		}

		// 预热 Forum 树
		tree, err := cfg.ForumSvc.GetTree(ctx)
		if err != nil {
			logger.Error("warmup forum tree failed", logger.String("error", err.Error()))
		} else {
			r.mu.Lock()
			r.forumTree = tree
			r.mu.Unlock()
			logger.Info("warmup forum tree", logger.Int("count", len(tree)))
		}
	}

	// 2. 预热 Tag 列表
	if cfg.TagSvc != nil {
		list, err := cfg.TagSvc.GetAll(ctx)
		if err != nil {
			logger.Error("warmup tag list failed", logger.String("error", err.Error()))
		} else {
			r.mu.Lock()
			r.tagList = list
			r.mu.Unlock()
			logger.Info("warmup tag list", logger.Int("count", len(list)))
		}
	}

	// 3. 预热配置（从数据库或其他来源）
	// TODO: 从数据库加载站点配置
	r.mu.Lock()
	r.config["site_name"] = "WellCMS Go"
	r.config["site_desc"] = "高性能内容管理系统"
	r.mu.Unlock()

	r.loadedAt = time.Now()

	logger.Info("runtime warmup completed", logger.Duration("duration", time.Since(start)))

	return nil
}

// Reload 重新加载所有运行时数据
func (r *Runtime) Reload(cfg *RuntimeConfig) error {
	return r.warmup(cfg)
}

// GetForumList 获取 Forum 列表
func (r *Runtime) GetForumList() []*service.ForumDTO {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.forumList
}

// GetForumTree 获取 Forum 树
func (r *Runtime) GetForumTree() []*service.ForumTreeNode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.forumTree
}

// GetTagList 获取 Tag 列表
func (r *Runtime) GetTagList() []*service.TagDTO {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tagList
}

// GetConfig 获取配置
func (r *Runtime) GetConfig(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config[key]
}

// SetConfig 设置配置
func (r *Runtime) SetConfig(key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config[key] = value
}

// GetLoadedAt 获取加载时间
func (r *Runtime) GetLoadedAt() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loadedAt
}

// FormatLoadedTime 格式化加载时间
func (r *Runtime) FormatLoadedTime() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.loadedAt.Format("2006-01-02 15:04:05")
}

// Status 返回运行时状态
func (r *Runtime) Status() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return map[string]interface{}{
		"forum_count":  len(r.forumList),
		"forum_tree":   len(r.forumTree),
		"tag_count":    len(r.tagList),
		"config_count": len(r.config),
		"loaded_at":    r.loadedAt.Format("2006-01-02 15:04:05"),
	}
}

// WarmUpLog 预热日志
func WarmUpLog() string {
	if rt == nil {
		return "runtime not initialized"
	}
	return fmt.Sprintf("Forum: %d, Tag: %d, Loaded: %s",
		len(rt.forumList), len(rt.tagList), rt.FormatLoadedTime())
}
