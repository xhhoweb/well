package seo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"well_go/internal/core/logger"

	"github.com/redis/go-redis/v9"
)

// IndexNowConfig IndexNow配置
type IndexNowConfig struct {
	Key         string        // API Key
	KeyLocation string        // Key文件路径
	Endpoint    string        // 提交端点
	RedisKey    string        //  Redis key前缀（防重复提交）
	RedisTTL    time.Duration // 缓存时间
}

// IndexNowPayload IndexNow提交内容
type IndexNowPayload struct {
	Host     string   `json:"host"`
	Key      string   `json:"key"`
	KeyOwner string   `json:"keyOwner"`
	URLList  []string `json:"urlList"`
}

// IndexNowService IndexNow服务
// 支持 Bing、Yandex、Naver 等搜索引擎
type IndexNowService struct {
	client *http.Client
	config *IndexNowConfig
	redis  *redis.Client
}

// NewIndexNowService 创建IndexNow服务
func NewIndexNowService(cfg *IndexNowConfig, redisClient *redis.Client) *IndexNowService {
	return &IndexNowService{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		config: cfg,
		redis:  redisClient,
	}
}

// GenerateKey 生成IndexNow Key
func (s *IndexNowService) GenerateKey(url string) string {
	hash := sha256.Sum256([]byte(url + s.config.Key))
	return hex.EncodeToString(hash[:])
}

// ShouldSubmit 检查是否应该提交（防重复）
func (s *IndexNowService) ShouldSubmit(ctx context.Context, url string) (bool, error) {
	key := fmt.Sprintf("%s:%s", s.config.RedisKey, url)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 0, nil
}

// MarkSubmitted 标记已提交
func (s *IndexNowService) MarkSubmitted(ctx context.Context, url string) error {
	key := fmt.Sprintf("%s:%s", s.config.RedisKey, url)
	return s.redis.Set(ctx, key, "1", s.config.RedisTTL).Err()
}

// SubmitURL 提交单个URL
func (s *IndexNowService) SubmitURL(ctx context.Context, url string) error {
	// 检查是否已提交
	if ok, err := s.ShouldSubmit(ctx, url); err != nil || !ok {
		return err
	}

	payload := IndexNowPayload{
		Host:     s.extractHost(url),
		Key:      s.GenerateKey(url),
		KeyOwner: s.config.KeyLocation,
		URLList:  []string{url},
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return fmt.Errorf("indexnow submit failed: %d", resp.StatusCode)
	}

	// 标记已提交
	_ = s.MarkSubmitted(ctx, url)
	return nil
}

// SubmitURLs 批量提交URL
func (s *IndexNowService) SubmitURLs(ctx context.Context, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	host := s.extractHost(urls[0])
	payload := IndexNowPayload{
		Host:     host,
		Key:      s.GenerateKey(urls[0]),
		KeyOwner: s.config.KeyLocation,
		URLList:  urls,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("indexnow submit failed: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	// 标记所有URL已提交
	for _, url := range urls {
		_ = s.MarkSubmitted(ctx, url)
	}

	return nil
}

func (s *IndexNowService) extractHost(url string) string {
	// 简单提取host
	for i := 0; i < len(url); i++ {
		if url[i] == '/' && i > 8 { // https:// 或 http://
			return url[:i]
		}
	}
	return url
}

// AsyncSubmit 异步提交（不阻塞主流程）
func (s *IndexNowService) AsyncSubmit(urls []string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.SubmitURLs(ctx, urls); err != nil {
			logger.Warn("indexnow async submit failed", logger.String("error", err.Error()))
		}
	}()
}

// SubmitThread 提交帖子（供Service调用）
func (s *IndexNowService) SubmitThread(ctx context.Context, tid int64) error {
	url := fmt.Sprintf("%s/thread/%d", s.config.KeyLocation, tid)
	return s.SubmitURL(ctx, url)
}
