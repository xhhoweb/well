package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"well_go/internal/core/config"
	"well_go/internal/core/logger"
	"well_go/internal/core/snowflake"
	"well_go/internal/model"
	"well_go/internal/pkg/pool"
	"well_go/internal/repository"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"
)

// UserService 用户服务
type UserService struct {
	repo   repository.UserRepository
	l1     *pool.BigCache // L1 Cache（零GC）
	l2     *redis.Client
	sf     *singleflight.Group
	l2Cfg  *config.CacheConfig
	jwtCfg *config.JWTConfig
}

// NewUserService 创建用户服务
func NewUserService(repo repository.UserRepository, redisClient *redis.Client, cacheCfg *config.CacheConfig, jwtCfg *config.JWTConfig) *UserService {
	l1Cache, _ := pool.NewBigCache(cacheCfg.L1Cap, time.Duration(cacheCfg.L2TTL)*time.Second)
	return &UserService{
		repo:   repo,
		l1:     l1Cache,
		l2:     redisClient,
		sf:     &singleflight.Group{},
		l2Cfg:  cacheCfg,
		jwtCfg: jwtCfg,
	}
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, username, password string) (*model.LoginResponse, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		logger.Error("login: get user error", logger.String("error", err.Error()))
		return nil, errors.New("系统错误")
	}
	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 检查状态
	if user.Status != 0 {
		return nil, errors.New("账号已被禁用")
	}

	// 更新最后访问时间
	now := int(time.Now().Unix())
	go s.repo.UpdateLastvisit(context.Background(), user.Uid, now)

	// 生成Token
	token, err := generateJWT(user.Uid, user.Role, s.jwtCfg)
	if err != nil {
		logger.Error("login: generate token error", logger.String("error", err.Error()))
		return nil, errors.New("系统错误")
	}

	dto := &model.UserDTO{
		Uid:      user.Uid,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Role:     user.Role,
		Status:   user.Status,
		Dateline: user.Dateline,
	}

	return &model.LoginResponse{
		Token: token,
		User:  *dto,
	}, nil
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *model.RegisterRequest) (*model.RegisterResponse, error) {
	// 检查用户名
	exist, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("系统错误")
	}
	if exist != nil {
		return nil, errors.New("用户名已被占用")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("register: hash password error", logger.String("error", err.Error()))
		return nil, errors.New("系统错误")
	}

	// 生成UID
	uid := snowflake.Generate()
	now := int(time.Now().Unix())

	user := &model.User{
		Uid:       uid,
		Username:  req.Username,
		Password:  string(hashedPassword),
		Email:     req.Email,
		Avatar:    fmt.Sprintf("https://api.dicebear.com/7.x/avataaars/svg?seed=%s", req.Username),
		Role:      0,
		Status:    0,
		Dateline:  now,
		Lastvisit: now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Error("register: create user error", logger.String("error", err.Error()))
		return nil, errors.New("系统错误")
	}

	return &model.RegisterResponse{
		User: model.UserDTO{
			Uid:      user.Uid,
			Username: user.Username,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Role:     user.Role,
			Status:   user.Status,
			Dateline: user.Dateline,
		},
	}, nil
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(ctx context.Context, uid int64) (*model.UserProfile, error) {
	key := fmt.Sprintf("user:profile:%d", uid)

	// L1
	if data, ok := s.l1.Get(key); ok {
		if data != nil {
			var dto model.UserDTO
			if err := json.Unmarshal(data, &dto); err == nil {
				return &model.UserProfile{
					UserDTO:   dto,
					Lastvisit: 0,
				}, nil
			}
		}
	}

	// DB
	user, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		logger.Error("getprofile: get user error", logger.String("error", err.Error()))
		return nil, errors.New("系统错误")
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	dto := &model.UserDTO{
		Uid:      user.Uid,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Role:     user.Role,
		Status:   user.Status,
		Dateline: user.Dateline,
	}

	profile := &model.UserProfile{
		UserDTO:   *dto,
		Lastvisit: user.Lastvisit,
	}

	// Write Cache
	if bytes, _ := json.Marshal(dto); bytes != nil {
		s.l1.Set(key, bytes)
	}
	ctxBg := context.Background()
	s.l2.Set(ctxBg, key, "1", time.Duration(s.l2Cfg.L2TTL)*time.Second)

	return profile, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, uid int64) (*model.UserDTO, error) {
	key := fmt.Sprintf("user:%d", uid)

	// L1
	if data, ok := s.l1.Get(key); ok {
		if data != nil {
			var dto model.UserDTO
			if err := json.Unmarshal(data, &dto); err == nil {
				return &dto, nil
			}
		}
	}

	user, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, errors.New("系统错误")
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	dto := &model.UserDTO{
		Uid:      user.Uid,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Role:     user.Role,
		Status:   user.Status,
		Dateline: user.Dateline,
	}

	// Write Cache
	if bytes, _ := json.Marshal(dto); bytes != nil {
		s.l1.Set(key, bytes)
	}

	return dto, nil
}

// generateJWT 生成JWT（简化版）
func generateJWT(uid int64, role int, cfg *config.JWTConfig) (string, error) {
	// TODO: 实现完整JWT生成
	return "", nil
}
