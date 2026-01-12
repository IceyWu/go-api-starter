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
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
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

	// Environment variable support (highest priority, overrides config file)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables
	bindEnvVariables()

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
}

func setDefaults() {
	viper.SetDefault("app.name", "go-api-starter")
	viper.SetDefault("app.env", "development")
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
}

// GetConfig returns the global config
func GetConfig() *Config {
	return GlobalConfig
}
