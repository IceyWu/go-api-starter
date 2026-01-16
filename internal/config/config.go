package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	OSS      OSSConfig      `mapstructure:"oss"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Mail     MailConfig     `mapstructure:"mail"`
}

// MailConfig holds mail server configuration
type MailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	UseTLS   bool   `mapstructure:"use_tls"`
	Enabled  bool   `mapstructure:"enabled"`
	MockSend bool   `mapstructure:"mock_send"`
}

type AppConfig struct {
	Name      string `mapstructure:"name"`
	Env       string `mapstructure:"env"`
	JWTSecret string `mapstructure:"jwt_secret"`
	Port      int    `mapstructure:"port"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Path     string `mapstructure:"path"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Charset  string `mapstructure:"charset"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

type OSSConfig struct {
	Endpoint          string   `mapstructure:"endpoint"`
	Bucket            string   `mapstructure:"bucket"`
	BucketName        string   `mapstructure:"bucket_name"`
	Region            string   `mapstructure:"region"`
	AccessKeyID       string   `mapstructure:"access_key_id"`
	AccessKeySecret   string   `mapstructure:"access_key_secret"`
	UploadDir         string   `mapstructure:"upload_dir"`
	BasePath          string   `mapstructure:"base_path"`
	Domain            string   `mapstructure:"domain"`
	CallbackURL       string   `mapstructure:"callback_url"`
	MaxFileSize       int64    `mapstructure:"max_file_size"`
	AllowedExtensions []string `mapstructure:"allowed_extensions"`
	TokenExpire       int64    `mapstructure:"token_expire"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	Password       string        `mapstructure:"password"`
	DB             int           `mapstructure:"db"`
	PoolSize       int           `mapstructure:"pool_size"`
	MinIdleConns   int           `mapstructure:"min_idle_conns"`
	MaxRetries     int           `mapstructure:"max_retries"`
	DialTimeout    time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	ClusterMode    bool          `mapstructure:"cluster_mode"`
	ClusterAddrs   []string      `mapstructure:"cluster_addrs"`
	EnableFallback bool          `mapstructure:"enable_fallback"`
	Enabled        bool          `mapstructure:"enabled"`
}

// Addr returns the Redis address in host:port format
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

var GlobalConfig *Config

// Load loads configuration from file and environment variables
func Load() *Config {
	// Load .env file based on APP_ENV
	loadEnvFile()

	// Load base config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using defaults: %v", err)
	}

	// Bind specific environment variables BEFORE AutomaticEnv
	bindEnvVariables()

	// Environment variable support (highest priority, overrides config file)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Log configuration source
	log.Printf("Configuration loaded from: %s", viper.ConfigFileUsed())
	log.Printf("Environment: %s", cfg.App.Env)

	// Validate configuration in production
	if cfg.App.Env == "production" || cfg.App.Env == "prod" {
		errors := cfg.Validate()
		if errors.HasErrors() {
			log.Fatalf("Configuration validation failed: %s", errors.Error())
		}
		log.Printf("Configuration validation passed")
	}

	GlobalConfig = &cfg
	return &cfg
}

// loadEnvFile loads .env file based on APP_ENV
func loadEnvFile() {
	env := os.Getenv("APP_ENV")
	var envFile string

	switch env {
	case "production", "prod":
		envFile = ".env.prod"
	case "development", "dev":
		envFile = ".env.dev"
	default:
		envFile = ".env"
	}

	// Try to load the specific env file
	if err := godotenv.Load(envFile); err != nil {
		// If specific file not found, try .env
		if err := godotenv.Load(); err != nil {
			log.Printf("No .env file found, using system environment variables")
		}
	} else {
		log.Printf("Loaded environment from %s", envFile)
	}
}

// bindEnvVariables binds specific environment variables to config keys
func bindEnvVariables() {
	viper.BindEnv("app.env", "APP_ENV")
	viper.BindEnv("app.jwt_secret", "JWT_SECRET")
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.mode", "SERVER_MODE")
	viper.BindEnv("database.driver", "DB_DRIVER")
	viper.BindEnv("database.path", "DB_PATH")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.username", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("log.level", "LOG_LEVEL")
	viper.BindEnv("oss.endpoint", "ALICLOUD_OSS_ENDPOINT")
	viper.BindEnv("oss.bucket", "ALICLOUD_OSS_BUCKET")
	viper.BindEnv("oss.region", "ALICLOUD_OSS_REGION")
	viper.BindEnv("oss.access_key_id", "ALICLOUD_ACCESS_KEY_ID")
	viper.BindEnv("oss.access_key_secret", "ALICLOUD_ACCESS_KEY_SECRET")
	viper.BindEnv("oss.upload_dir", "ALICLOUD_OSS_UPLOAD_DIR")
	viper.BindEnv("oss.domain", "OSS_DOMAIN")
	viper.BindEnv("oss.callback_url", "OSS_CALLBACK_URL")
	// Redis environment variables
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("redis.enabled", "REDIS_ENABLED")
	viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")
	viper.BindEnv("redis.cluster_mode", "REDIS_CLUSTER_MODE")
	viper.BindEnv("redis.enable_fallback", "REDIS_ENABLE_FALLBACK")
	// Mail environment variables
	viper.BindEnv("mail.host", "MAIL_HOST")
	viper.BindEnv("mail.port", "MAIL_PORT")
	viper.BindEnv("mail.user", "MAIL_USER")
	viper.BindEnv("mail.password", "MAIL_PASS")
	viper.BindEnv("mail.from", "MAIL_FROM")
	viper.BindEnv("mail.use_tls", "MAIL_USE_TLS")
	viper.BindEnv("mail.enabled", "MAIL_ENABLED")
	viper.BindEnv("mail.mock_send", "MAIL_MOCK_SEND")
}

func setDefaults() {
	viper.SetDefault("app.name", "go-api-starter")
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.jwt_secret", "your-secret-key-change-in-production")
	viper.SetDefault("app.port", 9527)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "9527")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.path", "./data.db")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "123456")
	viper.SetDefault("database.dbname", "go_api_starter")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.format", "console")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("oss.upload_dir", "go_oss")
	viper.SetDefault("oss.base_path", "uploads")
	viper.SetDefault("oss.max_file_size", 10485760) // 10MB
	viper.SetDefault("oss.token_expire", 1800)      // 30 minutes
	viper.SetDefault("oss.allowed_extensions", []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".xls", ".xlsx"})
	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.dial_timeout", 5*time.Second)
	viper.SetDefault("redis.read_timeout", 3*time.Second)
	viper.SetDefault("redis.write_timeout", 3*time.Second)
	viper.SetDefault("redis.cluster_mode", false)
	viper.SetDefault("redis.enable_fallback", true)
	viper.SetDefault("redis.enabled", true)
	// Mail defaults
	viper.SetDefault("mail.host", "smtp.qq.com")
	viper.SetDefault("mail.port", 587)
	viper.SetDefault("mail.user", "")
	viper.SetDefault("mail.password", "")
	viper.SetDefault("mail.from", "")
	viper.SetDefault("mail.use_tls", true)
	viper.SetDefault("mail.enabled", false)
	viper.SetDefault("mail.mock_send", false)
}

// GetConfig returns the global config
func GetConfig() *Config {
	return GlobalConfig
}
