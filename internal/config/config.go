package config

import (
	"log"
	"os"
	"strings"

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
}

type AppConfig struct {
	Name      string `mapstructure:"name"`
	Env       string `mapstructure:"env"`
	JWTSecret string `mapstructure:"jwt_secret"`
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
}

func setDefaults() {
	viper.SetDefault("app.name", "go-api-starter")
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.jwt_secret", "your-secret-key-change-in-production")
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
}

// GetConfig returns the global config
func GetConfig() *Config {
	return GlobalConfig
}
