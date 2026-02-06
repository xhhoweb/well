package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var v *viper.Viper
var cfg *Config

// Config App-wide configuration
type Config struct {
	Database  DatabaseConfig  `mapstructure:"-"`
	Redis     RedisConfig     `mapstructure:"-"`
	App       AppConfig       `mapstructure:"-"`
	JWT       JWTConfig       `mapstructure:"-"`
	Cache     CacheConfig     `mapstructure:"-"`
	Snowflake SnowflakeConfig `mapstructure:"-"`
	Logging   LoggingConfig   `mapstructure:"-"`
	Security  SecurityConfig  `mapstructure:"-"`
}

// DatabaseConfig MySQL Database Configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

// RedisConfig Redis Configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// AppConfig Application Configuration
type AppConfig struct {
	Host string
	Port int
	Mode string
	BaseURL string
}

// JWTConfig JWT Configuration
type JWTConfig struct {
	Secret string
	Expiry int // Token过期时间(秒)
}

// CacheConfig Cache Configuration
type CacheConfig struct {
	L1Cap int
	L2TTL int
}

// SnowflakeConfig Snowflake Configuration
type SnowflakeConfig struct {
	WorkerID int64
}

// LoggingConfig Logging Configuration
type LoggingConfig struct {
	Level      string
	Output     string
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

// SecurityConfig Security Configuration
type SecurityConfig struct {
	AllowIPs  []string // IP白名单
	DenyIPs   []string // IP黑名单
	RateLimit int      // 频率限制
}

// Init Initialize configuration with Viper
func Init(configPath string) error {
	v = viper.New()
	cfg = &Config{}

	// 设置配置文件名（不带扩展名）
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 添加配置文件路径
	v.AddConfigPath(configPath)
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// 读取配置
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 使用默认值
			setDefaults()
		} else {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	// 环境变量覆盖
	v.SetEnvPrefix("WELL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 绑定环境变量
	bindEnvs()

	// 解析配置到结构体
	return parseConfig()
}

// setDefaults 设置默认值
func setDefaults() {
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.mode", "release")
	v.SetDefault("app.base_url", "")

	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 300)

	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.pool_size", 10)

	v.SetDefault("cache.l1_cap", 1000)
	v.SetDefault("cache.l2_ttl", 3600)

	v.SetDefault("snowflake.worker_id", 0)

	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.expiry", 86400)

	v.SetDefault("security.allow_ips", []string{"127.0.0.1", "localhost", "::1"})
	v.SetDefault("security.rate_limit", 100)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.output", "stdout")
}

// bindEnvs 绑定环境变量
func bindEnvs() {
	// Database
	v.BindEnv("database.host", "WELL_DATABASE_HOST")
	v.BindEnv("database.port", "WELL_DATABASE_PORT")
	v.BindEnv("database.username", "WELL_DATABASE_USERNAME")
	v.BindEnv("database.password", "WELL_DATABASE_PASSWORD")
	v.BindEnv("database.name", "WELL_DATABASE_NAME")

	// Redis
	v.BindEnv("redis.host", "WELL_REDIS_HOST")
	v.BindEnv("redis.port", "WELL_REDIS_PORT")
	v.BindEnv("redis.password", "WELL_REDIS_PASSWORD")

	// JWT
	v.BindEnv("jwt.secret", "WELL_JWT_SECRET")
}

// parseConfig 解析配置到结构体
func parseConfig() error {
	// Database
	cfg.Database.Host = v.GetString("database.host")
	cfg.Database.Port = v.GetInt("database.port")
	cfg.Database.Username = v.GetString("database.username")
	cfg.Database.Password = v.GetString("database.password")
	cfg.Database.Name = v.GetString("database.name")
	cfg.Database.MaxOpenConns = v.GetInt("database.max_open_conns")
	cfg.Database.MaxIdleConns = v.GetInt("database.max_idle_conns")
	cfg.Database.ConnMaxLifetime = v.GetInt("database.conn_max_lifetime")

	// Redis
	cfg.Redis.Host = v.GetString("redis.host")
	cfg.Redis.Port = v.GetInt("redis.port")
	cfg.Redis.Password = v.GetString("redis.password")
	cfg.Redis.DB = v.GetInt("redis.db")
	cfg.Redis.PoolSize = v.GetInt("redis.pool_size")

	// App
	cfg.App.Host = v.GetString("app.host")
	cfg.App.Port = v.GetInt("app.port")
	cfg.App.Mode = v.GetString("app.mode")
	cfg.App.BaseURL = strings.TrimSpace(v.GetString("app.base_url"))

	// JWT
	cfg.JWT.Secret = v.GetString("jwt.secret")
	cfg.JWT.Expiry = v.GetInt("jwt.expiry")

	// Cache
	cfg.Cache.L1Cap = v.GetInt("cache.l1_cap")
	cfg.Cache.L2TTL = v.GetInt("cache.l2_ttl")

	// Snowflake
	cfg.Snowflake.WorkerID = v.GetInt64("snowflake.worker_id")

	// Logging
	cfg.Logging.Level = v.GetString("logging.level")
	cfg.Logging.Output = v.GetString("logging.output")
	cfg.Logging.Filename = v.GetString("logging.filename")
	cfg.Logging.MaxSize = v.GetInt("logging.max_size")
	cfg.Logging.MaxAge = v.GetInt("logging.max_age")
	cfg.Logging.MaxBackups = v.GetInt("logging.max_backups")

	// Security
	cfg.Security.AllowIPs = v.GetStringSlice("security.allow_ips")
	cfg.Security.DenyIPs = v.GetStringSlice("security.deny_ips")
	cfg.Security.RateLimit = v.GetInt("security.rate_limit")

	return nil
}

// Get 获取配置实例
func Get() *Config {
	return cfg
}

// GetDSN Get MySQL DSN
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		c.Username, c.Password, c.Host, c.Port, c.Name)
}

// GetRedisAddr Get Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr Get server address
func (c *AppConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetDBName Get database name (for sqlx.Open)
func (c *DatabaseConfig) GetDBName() string {
	return c.Name
}
