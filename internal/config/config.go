package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Upload   UploadConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	AccessSecret          string
	RefreshSecret         string
	AccessExpiresDuration time.Duration
	RefreshExpiresDuration time.Duration
}

type UploadConfig struct {
	MaxSizeMB int64
}

// Load reads configuration from environment variables (and optionally from .env file via viper).
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read .env file if it exists (not required: e.g., production uses real env vars)
	_ = viper.ReadInConfig()

	// Set defaults
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("SERVER_MODE", "debug")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_ACCESS_EXPIRES_MINUTES", 15)
	viper.SetDefault("JWT_REFRESH_EXPIRES_DAYS", 7)
	viper.SetDefault("MAX_UPLOAD_SIZE_MB", 5)

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetInt("SERVER_PORT"),
			Mode: viper.GetString("SERVER_MODE"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetInt("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			AccessSecret:           viper.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret:          viper.GetString("JWT_REFRESH_SECRET"),
			AccessExpiresDuration:  time.Duration(viper.GetInt("JWT_ACCESS_EXPIRES_MINUTES")) * time.Minute,
			RefreshExpiresDuration: time.Duration(viper.GetInt("JWT_REFRESH_EXPIRES_DAYS")) * 24 * time.Hour,
		},
		Upload: UploadConfig{
			MaxSizeMB: viper.GetInt64("MAX_UPLOAD_SIZE_MB"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	return nil
}

// DSN returns the PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}
