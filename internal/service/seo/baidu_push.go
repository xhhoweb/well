package seo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"well_go/internal/core/logger"

	"github.com/redis/go-redis/v9"
)

// BaiduPushConfig 百度推送配置
type BaiduPushConfig struct {
	API      string        // 百度推送API
	Token    string        // 准入Token
	RedisKey string        // Redis key前缀
	RedisTTL time.Duration // 缓存时间
}

// BaiduPushResponse 百度推送响应
type BaiduPushResponse struct {
	Success     int      `json:"success"`       // 成功数量
	Remain      int      `json:"remain"`        // 剩余额度
	NotValidURL []string `json:"not_valid_url"` // 无效URL
}

// BaiduPushService 百度推送服务
// 百度不支持IndexNow，必须单独实现
type BaiduPushService struct {
	client *http.Client
	config *BaiduPushConfig
	redis  *redis.Client
}

// NewBaiduPushService 创建百度推送服务
func NewBaiduPushService(cfg *BaiduPushConfig, redisClient *redis.Client) *BaiduPushService {
	return &BaiduPushService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		config: cfg,
		redis:  redisClient,
	}
}

// ShouldPush 检查是否应该推送（防重复）
func (s *BaiduPushService) ShouldPush(ctx context.Context, url string) (bool, error) {
	key := fmt.Sprintf("%s:%s", s.config.RedisKey, url)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 0, nil
}

// MarkPushed 标记已推送
func (s *BaiduPushService) MarkPushed(ctx context.Context, url string) error {
	key := fmt.Sprintf("%s:%s", s.config.RedisKey, url)
	return s.redis.Set(ctx, key, "1", s.config.RedisTTL).Err()
}

// PushURL 推送单个URL
func (s *BaiduPushService) PushURL(ctx context.Context, url string) error {
	// 检查是否已推送
	if ok, err := s.ShouldPush(ctx, url); err != nil || !ok {
		return err
	}

	apiURL := fmt.Sprintf("%s?token=%s", s.config.API, s.config.Token)
	body := fmt.Sprintf("https://%s", strings.TrimPrefix(url, "https://"))

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 202 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("baidu push failed: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var result BaiduPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// 响应可能是纯文本，忽略解析错误
	}

	// 标记已推送
	_ = s.MarkPushed(ctx, url)
	return nil
}

// PushURLs 批量推送URL
func (s *BaiduPushService) PushURLs(ctx context.Context, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	apiURL := fmt.Sprintf("%s?token=%s", s.config.API, s.config.Token)

	var urlsStr strings.Builder
	for i, u := range urls {
		if i > 0 {
			urlsStr.WriteString("\n")
		}
		urlsStr.WriteString("https://")
		urlsStr.WriteString(strings.TrimPrefix(u, "https://"))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(urlsStr.String()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 202 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("baidu push failed: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	// 标记所有URL已推送
	for _, url := range urls {
		_ = s.MarkPushed(ctx, url)
	}

	return nil
}

// PushThread 推送帖子（供Service调用）
func (s *BaiduPushService) PushThread(ctx context.Context, baseURL string, tid int64) error {
	threadURL := fmt.Sprintf("%s/thread/%d", baseURL, tid)
	return s.PushURL(ctx, threadURL)
}

// AsyncPush 异步推送（不阻塞主流程）
func (s *BaiduPushService) AsyncPush(urls []string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.PushURLs(ctx, urls); err != nil {
			logger.Warn("baidu async push failed", logger.String("error", err.Error()))
		}
	}()
}

// AsyncPushThread 异步推送帖子
func (s *BaiduPushService) AsyncPushThread(baseURL string, tid int64) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		threadURL := fmt.Sprintf("%s/thread/%d", baseURL, tid)
		if err := s.PushURL(ctx, threadURL); err != nil {
			logger.Warn("baidu async thread push failed", logger.String("error", err.Error()))
		}
	}()
}

// GetRemain 获取剩余额度（可选）
func (s *BaiduPushService) GetRemain(ctx context.Context) (int, error) {
	apiURL := fmt.Sprintf("%s?token=%s", s.config.API, s.config.Token)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("baidu get remain failed: %d", resp.StatusCode)
	}

	var result BaiduPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Remain, nil
}

// ValidateAPI 验证API是否有效
func (s *BaiduPushService) ValidateAPI() error {
	testURL := "https://example.com"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.PushURL(ctx, testURL)
}
