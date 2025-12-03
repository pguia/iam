package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	MaxConns int    `mapstructure:"max_conns"`
	MaxIdle  int    `mapstructure:"max_idle"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Type           string           `mapstructure:"type"` // "none", "memory", "redis"
	Enabled        bool             `mapstructure:"enabled"`
	TTLSeconds     int              `mapstructure:"ttl_seconds"`
	MaxSize        int              `mapstructure:"max_size"`
	CleanupMinutes int              `mapstructure:"cleanup_minutes"`
	Redis          RedisCacheConfig `mapstructure:"redis"`
}

// RedisCacheConfig holds Redis cache configuration
type RedisCacheConfig struct {
	Address    string `mapstructure:"address"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	TTLSeconds int    `mapstructure:"ttl_seconds"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set config file details
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/iam")

	// Set defaults
	setDefaults(v)

	// Read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; proceed with defaults and env vars
	}

	// Enable environment variable overrides
	v.SetEnvPrefix("IAM")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind all environment variables explicitly
	bindEnvVariables(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.address", ":8081")
	v.SetDefault("server.port", 8081)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.dbname", "iam_db")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_conns", 25)
	v.SetDefault("database.max_idle", 5)

	// Cache defaults (stateless by default)
	v.SetDefault("cache.type", "none")         // "none", "memory", "redis"
	v.SetDefault("cache.enabled", false)       // Disabled by default for stateless
	v.SetDefault("cache.ttl_seconds", 300)     // 5 minutes
	v.SetDefault("cache.max_size", 10000)      // 10k entries
	v.SetDefault("cache.cleanup_minutes", 10)  // cleanup every 10 minutes

	// Redis cache defaults
	v.SetDefault("cache.redis.address", "localhost:6379")
	v.SetDefault("cache.redis.password", "")
	v.SetDefault("cache.redis.db", 0)
	v.SetDefault("cache.redis.ttl_seconds", 300)
}

func bindEnvVariables(v *viper.Viper) {
	// Server
	v.BindEnv("server.address")
	v.BindEnv("server.port")

	// Database
	v.BindEnv("database.host")
	v.BindEnv("database.port")
	v.BindEnv("database.user")
	v.BindEnv("database.password")
	v.BindEnv("database.dbname")
	v.BindEnv("database.sslmode")
	v.BindEnv("database.max_conns")
	v.BindEnv("database.max_idle")

	// Cache
	v.BindEnv("cache.type")
	v.BindEnv("cache.enabled")
	v.BindEnv("cache.ttl_seconds")
	v.BindEnv("cache.max_size")
	v.BindEnv("cache.cleanup_minutes")

	// Redis Cache
	v.BindEnv("cache.redis.address")
	v.BindEnv("cache.redis.password")
	v.BindEnv("cache.redis.db")
	v.BindEnv("cache.redis.ttl_seconds")
}
