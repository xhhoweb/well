package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"well_go/internal/api/mgt"
	v1 "well_go/internal/api/v1"
	"well_go/internal/core/config"
	"well_go/internal/core/database"
	"well_go/internal/core/logger"
	"well_go/internal/core/runtime"
	"well_go/internal/core/snowflake"
	"well_go/internal/middleware"
	"well_go/internal/repository"
	"well_go/internal/service"

	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. 加载配置 (Viper)
	if err := config.Init("."); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Get()

	// 2. 初始化 Logger
	if err := logger.Init(&cfg.Logging); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting well_go...")

	// 3. 初始化 MySQL
	if err := database.Init(&cfg.Database); err != nil {
		logger.Error("Failed to init database", logger.String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Close()

	// 4. 初始化 Redis (L2 Cache)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	defer redisClient.Close()

	// 5. 初始化缓存配置
	cacheConfig := &config.CacheConfig{
		L1Cap: cfg.Cache.L1Cap,
		L2TTL: cfg.Cache.L2TTL,
	}

	// 6. 初始化 Snowflake
	if err := snowflake.Init(&cfg.Snowflake); err != nil {
		logger.Error("Failed to init snowflake", logger.String("error", err.Error()))
		os.Exit(1)
	}

	// 7. 初始化 Repository
	threadRepo := repository.NewThreadRepository(database.Get())
	forumRepo := repository.NewForumRepository(database.Get())
	tagRepo := repository.NewTagRepository(database.Get())
	threadTagRepo := repository.NewThreadTagRepository(database.Get())
	userRepo := repository.NewUserRepository(database.Get())

	// 8. 初始化 Service
	threadSvc := service.NewThreadService(threadRepo, redisClient, cacheConfig)
	forumSvc := service.NewForumService(forumRepo, redisClient, cacheConfig)
	tagSvc := service.NewTagService(tagRepo, threadTagRepo, redisClient, cacheConfig)
	userSvc := service.NewUserService(userRepo, redisClient, cacheConfig, &cfg.JWT)

	// 9. Runtime 预热
	rtConfig := &runtime.RuntimeConfig{
		ForumSvc: forumSvc,
		TagSvc:   tagSvc,
	}
	if err := runtime.Init(rtConfig); err != nil {
		logger.Error("Failed to init runtime", logger.String("error", err.Error()))
	}
	logger.Info("Runtime warmup: " + runtime.WarmUpLog())

	// 10. 初始化 Handler
	threadV1Handler := v1.NewThreadHandler(threadSvc)
	threadMgtHandler := mgt.NewThreadHandler(threadSvc)
	cacheMgtHandler := mgt.NewCacheHandler(threadSvc)

	forumV1Handler := v1.NewForumHandler(forumSvc)
	forumMgtHandler := mgt.NewForumMgtHandler(forumSvc)

	tagV1Handler := v1.NewTagHandler(tagSvc)
	tagMgtHandler := mgt.NewTagMgtHandler(tagSvc)

	userV1Handler := v1.NewUserHandler(userSvc)
	userMgtHandler := mgt.NewUserMgtHandler(userSvc)

	// 11. 创建 IP 限制器
	rateLimiter := middleware.NewIPLimiter(cfg.Security.RateLimit, 60)

	// 12. 注册路由
	gin.SetMode(cfg.App.Mode)
	router := gin.New()

	// Middleware
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RateLimitMW(rateLimiter))
	router.Use(middleware.CORSMiddleware())

	// Health Check (跳过 IP 检查)
	router.GET("/health", func(c *gin.Context) {
		if err := database.Ping(); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"status":    "healthy",
			"runtime":   runtime.Get().Status(),
			"timestamp": time.Now().Unix(),
		})
	})

	// Health Check (详细版 - 用于负载均衡)
	router.GET("/healthz", func(c *gin.Context) {
		status := "ok"
		checks := make(map[string]string)

		// 检查 MySQL
		if err := database.Ping(); err != nil {
			status = "error"
			checks["mysql"] = err.Error()
		} else {
			checks["mysql"] = "ok"
		}

		// 检查 Redis
		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			status = "error"
			checks["redis"] = err.Error()
		} else {
			checks["redis"] = "ok"
		}

		if status == "ok" {
			c.JSON(200, gin.H{
				"status":    "ok",
				"checks":    checks,
				"timestamp": time.Now().Unix(),
			})
		} else {
			c.JSON(503, gin.H{
				"status":    "error",
				"checks":    checks,
				"timestamp": time.Now().Unix(),
			})
		}
	})

	// Root path (跳过 IP 检查)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "well_go",
			"status":  "running",
			"version": "1.0.0",
			"runtime": runtime.WarmUpLog(),
		})
	})

	// Runtime Status
	router.GET("/runtime", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": runtime.Get().Status(),
		})
	})

	// Metrics (跳过 IP 检查)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Public API (v1) - Public 白名单（本地/内网跳过）
	v1Group := router.Group("/api/v1")
	v1Group.Use(middleware.PublicWhitelistMW())
	{
		// Thread
		v1Group.GET("/threads", threadV1Handler.List)
		v1Group.GET("/thread/:tid", threadV1Handler.Get)

		// Forum
		v1Group.GET("/forums", forumV1Handler.List)
		v1Group.GET("/forums/tree", forumV1Handler.Tree)
		v1Group.GET("/forum/:fid", forumV1Handler.Get)

		// Tag
		v1Group.GET("/tags", tagV1Handler.List)
		v1Group.GET("/tags/hot", tagV1Handler.Hot)
		v1Group.GET("/tags/thread/:tid", tagV1Handler.GetByThread)

		// User
		v1Group.GET("/user/:uid", userV1Handler.GetUser)
	}

	// Management API (mgt) - 强制 IP 白名单
	mgtGroup := router.Group("/api/mgt")
	mgtGroup.Use(middleware.AdminWhitelistMW())
	{
		mgtGroup.POST("/login", func(c *gin.Context) {
			mgt.Login(c, &cfg.JWT)
		})

		userMgt := mgtGroup.Group("/user")
		userMgt.Use(middleware.JWTMW(&cfg.JWT))
		{
			userMgt.GET("/profile", userMgtHandler.GetProfile)
		}

		// 注册不需要JWT
		mgtGroup.POST("/user/register", userMgtHandler.Register)

		threadMgt := mgtGroup.Group("/thread")
		threadMgt.Use(middleware.JWTMW(&cfg.JWT))
		{
			threadMgt.POST("", threadMgtHandler.Create)
			threadMgt.PUT("/:tid", threadMgtHandler.Update)
			threadMgt.DELETE("/:tid", threadMgtHandler.Delete)
		}

		forumMgt := mgtGroup.Group("/forum")
		forumMgt.Use(middleware.JWTMW(&cfg.JWT))
		{
			forumMgt.POST("", forumMgtHandler.Create)
			forumMgt.PUT("/:fid", forumMgtHandler.Update)
			forumMgt.DELETE("/:fid", forumMgtHandler.Delete)
		}

		tagMgt := mgtGroup.Group("/tag")
		tagMgt.Use(middleware.JWTMW(&cfg.JWT))
		{
			tagMgt.POST("", tagMgtHandler.Create)
		}

		cacheMgt := mgtGroup.Group("/cache")
		cacheMgt.Use(middleware.JWTMW(&cfg.JWT))
		{
			cacheMgt.POST("/flush", cacheMgtHandler.Flush)
			cacheMgt.POST("/prewarm", cacheMgtHandler.Prewarm)
		}
	}

	// 13. 启动 HTTP Server (带超时配置)
	srv := &http.Server{
		Addr:         cfg.App.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.App.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.App.IdleTimeout) * time.Second,
	}

	go func() {
		logger.Info("Server starting", logger.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", logger.String("error", err.Error()))
		}
	}()

	// Graceful shutdown (优雅关闭)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 1. 停止接收新请求
	// 设置一个超时，强制关闭闲置连接
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 2. 关闭所有空闲连接
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logger.String("error", err.Error()))
	}

	// 3. 关闭数据库连接
	database.Close()

	// 4. 关闭 Redis 连接
	redisClient.Close()

	// 5. 刷新日志
	logger.Sync()

	logger.Info("Server exited gracefully")
}
