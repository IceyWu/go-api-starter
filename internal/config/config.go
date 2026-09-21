package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/confmap"
	koanfenvironment "github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
	"gopkg.in/yaml.v3"
)

// Config holds all configuration
type Config struct {
	App         AppConfig         `mapstructure:"app"`
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Log         LogConfig         `mapstructure:"log"`
	OSS         OSSConfig         `mapstructure:"oss"`
	Redis       RedisConfig       `mapstructure:"redis"`
	Transcoding TranscodingConfig `mapstructure:"transcoding"`
	CORS        CORSConfig        `mapstructure:"cors"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limit"`
	Mail        MailConfig        `mapstructure:"mail"`
	Wechat      WechatConfig      `mapstructure:"wechat"`
	WS          WebSocketConfig   `mapstructure:"ws"`
}

// MailConfig holds mail server configuration.
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

// WechatConfig holds WeChat mini-program credentials.
type WechatConfig struct {
	AppID  string `mapstructure:"appid"`  // 微信小程序 AppID
	Secret string `mapstructure:"secret"` // 微信小程序 AppSecret
}

// WebSocketConfig holds WebSocket server settings.
type WebSocketConfig struct {
	Key string `mapstructure:"key"` // 连接认证 Key
}

// CORSConfig holds CORS middleware configuration.
type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
	AllowMethods []string `mapstructure:"allow_methods"`
	AllowHeaders []string `mapstructure:"allow_headers"`
}

// RateLimitConfig holds rate limiter configuration.
type RateLimitConfig struct {
	GlobalPerMinute int `mapstructure:"global_per_minute"`
	UserPerMinute   int `mapstructure:"user_per_minute"`
	LoginPerMinute  int `mapstructure:"login_per_minute"`
	// UploadPerMinute applies only to file upload action endpoints
	// (POST /file/upload/init and /file/upload/complete).
	UploadPerMinute int `mapstructure:"upload_per_minute"`
	FallbackRPS     int `mapstructure:"fallback_rps"`
	FallbackBurst   int `mapstructure:"fallback_burst"`
}

// AppConfig holds basic application settings.
type AppConfig struct {
	Name                string `mapstructure:"name"`
	Env                 string `mapstructure:"env"`
	JWTSecret           string `mapstructure:"jwt_secret"`
	AccessTokenDays     int    `mapstructure:"access_token_days"`
	RefreshTokenDays    int    `mapstructure:"refresh_token_days"`
	UsernamePrefix      string `mapstructure:"username_prefix"`
	AdminEmail          string `mapstructure:"admin_email"`
	AdminPassword       string `mapstructure:"admin_password"`
	DocsUser            string `mapstructure:"docs_user"`
	DocsPassword        string `mapstructure:"docs_password"`
	DefaultUserPassword string `mapstructure:"default_user_password"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Mode     string `mapstructure:"mode"`
	BasePath string `mapstructure:"base_path"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Path            string `mapstructure:"path"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // seconds
}

// LogConfig holds logger settings.
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// OSSConfig holds Aliyun OSS settings.
type OSSConfig struct {
	Endpoint          string   `mapstructure:"endpoint"`
	Bucket            string   `mapstructure:"bucket"`
	BucketName        string   `mapstructure:"bucket_name"`
	Region            string   `mapstructure:"region"`
	AccessKeyID       string   `mapstructure:"access_key_id"`
	AccessKeySecret   string   `mapstructure:"access_key_secret"`
	UploadDir         string   `mapstructure:"upload_dir"`
	Domain            string   `mapstructure:"domain"`
	CallbackURL       string   `mapstructure:"callback_url"`
	MaxFileSize       int64    `mapstructure:"max_file_size"`
	AllowedExtensions []string `mapstructure:"allowed_extensions"`
	TokenExpire       int64    `mapstructure:"token_expire"`
}

// RedisConfig holds Redis connection configuration.
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

// Addr returns the Redis address in host:port format.
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type TranscodingConfig struct {
	StorageRoot         string `mapstructure:"storage_root"`
	MPSRegion           string `mapstructure:"mps_region"`
	MPSPipelineID       string `mapstructure:"mps_pipeline_id"`
	MPSTemplateOriginal string `mapstructure:"mps_template_original"`
	MPSTemplate1080p    string `mapstructure:"mps_template_1080p"`
	MPSTemplate720p     string `mapstructure:"mps_template_720p"`
	MPSTemplate480p     string `mapstructure:"mps_template_480p"`
	MPSPollIntervalSec  int    `mapstructure:"mps_poll_interval_sec"`
}

// GlobalConfig is the process-wide configuration singleton populated by Load.
var GlobalConfig *Config

const (
	defaultConfigFile = "./config/config.yaml"
	configEnvPrefix   = "GO_API_"
)

type configDocument struct {
	Common       map[string]any `yaml:"common"`
	Environments map[string]any `yaml:"-"`
}

func (d *configDocument) UnmarshalYAML(value *yaml.Node) error {
	var sections map[string]any
	if err := value.Decode(&sections); err != nil {
		return err
	}

	d.Common = map[string]any{}
	d.Environments = map[string]any{}
	for key, section := range sections {
		if key == "common" {
			if common, ok := section.(map[string]any); ok {
				d.Common = common
			}
			continue
		}
		d.Environments[key] = section
	}
	return nil
}

// Load loads common settings, the selected profile, and GO_API_* overrides.
func Load() *Config {
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = defaultConfigFile
	}

	document, err := readConfigDocument(configPath)
	if err != nil {
		log.Fatalf("failed to load configuration file %s: %v", configPath, err)
	}

	environment := normalizeEnvironment(os.Getenv("APP_ENV"))
	profile, ok := document.Environments[environment]
	if !ok {
		log.Fatalf("configuration profile %q was not found in %s", environment, configPath)
	}
	profileMap, ok := profile.(map[string]any)
	if !ok {
		log.Fatalf("configuration profile %q must be a mapping", environment)
	}

	k := koanf.New(".")
	if err := k.Load(confmap.Provider(document.Common, "."), nil); err != nil {
		log.Fatalf("failed to load common configuration: %v", err)
	}
	if err := k.Load(confmap.Provider(profileMap, "."), nil); err != nil {
		log.Fatalf("failed to load %s configuration: %v", environment, err)
	}
	if err := k.Load(koanfenvironment.Provider(".", koanfenvironment.Opt{
		Prefix: configEnvPrefix,
		TransformFunc: func(key, value string) (string, any) {
			key = strings.TrimPrefix(key, configEnvPrefix)
			key = strings.ToLower(strings.ReplaceAll(key, "__", "."))
			return key, value
		},
	}), nil); err != nil {
		log.Fatalf("failed to load environment overrides: %v", err)
	}

	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "mapstructure"}); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	log.Printf("configuration loaded from: %s", filepath.Clean(configPath))
	log.Printf("configuration profile: %s", environment)

	validationErrors := cfg.Validate()
	if validationErrors.HasErrors() {
		if cfg.App.Env == "production" || cfg.App.Env == "prod" {
			log.Fatalf("configuration validation failed: %s", validationErrors.Error())
		}
		log.Printf("WARNING: configuration issues: %s", validationErrors.Error())
	} else {
		log.Printf("configuration validation passed")
	}

	GlobalConfig = &cfg
	return &cfg
}

func readConfigDocument(configPath string) (configDocument, error) {
	contents, err := os.ReadFile(configPath)
	if err != nil {
		return configDocument{}, err
	}

	var document configDocument
	if err := yaml.Unmarshal(contents, &document); err != nil {
		return configDocument{}, err
	}
	return document, nil
}

func normalizeEnvironment(environment string) string {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "", "dev", "development":
		return "development"
	case "prod", "production":
		return "production"
	default:
		return strings.ToLower(strings.TrimSpace(environment))
	}
}

// GetConfig returns the global config.
func GetConfig() *Config {
	return GlobalConfig
}
