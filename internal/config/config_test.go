package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any IAM_ environment variables
	clearIAMEnvVars(t)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify server defaults
	assert.Equal(t, ":8081", cfg.Server.Address)
	assert.Equal(t, 8081, cfg.Server.Port)

	// Verify database defaults
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.User)
	assert.Equal(t, "postgres", cfg.Database.Password)
	assert.Equal(t, "iam_db", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 25, cfg.Database.MaxConns)
	assert.Equal(t, 5, cfg.Database.MaxIdle)

	// Verify cache defaults
	assert.Equal(t, "none", cfg.Cache.Type)
	assert.False(t, cfg.Cache.Enabled)
	assert.Equal(t, 300, cfg.Cache.TTLSeconds)
	assert.Equal(t, 10000, cfg.Cache.MaxSize)
	assert.Equal(t, 10, cfg.Cache.CleanupMinutes)

	// Verify Redis defaults
	assert.Equal(t, "localhost:6379", cfg.Cache.Redis.Address)
	assert.Empty(t, cfg.Cache.Redis.Password)
	assert.Equal(t, 0, cfg.Cache.Redis.DB)
	assert.Equal(t, 300, cfg.Cache.Redis.TTLSeconds)
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Clear any IAM_ environment variables
	clearIAMEnvVars(t)

	// Set environment variables
	os.Setenv("IAM_SERVER_ADDRESS", ":9090")
	os.Setenv("IAM_SERVER_PORT", "9090")
	os.Setenv("IAM_DATABASE_HOST", "testdb")
	os.Setenv("IAM_DATABASE_PORT", "5433")
	os.Setenv("IAM_DATABASE_USER", "testuser")
	os.Setenv("IAM_DATABASE_PASSWORD", "testpass")
	os.Setenv("IAM_DATABASE_DBNAME", "test_iam_db")
	os.Setenv("IAM_DATABASE_SSLMODE", "require")
	os.Setenv("IAM_DATABASE_MAX_CONNS", "50")
	os.Setenv("IAM_DATABASE_MAX_IDLE", "10")
	os.Setenv("IAM_CACHE_TYPE", "redis")
	os.Setenv("IAM_CACHE_ENABLED", "true")
	os.Setenv("IAM_CACHE_TTL_SECONDS", "600")
	os.Setenv("IAM_CACHE_MAX_SIZE", "20000")
	os.Setenv("IAM_CACHE_CLEANUP_MINUTES", "15")
	os.Setenv("IAM_CACHE_REDIS_ADDRESS", "redis:6379")
	os.Setenv("IAM_CACHE_REDIS_PASSWORD", "secret")
	os.Setenv("IAM_CACHE_REDIS_DB", "1")
	os.Setenv("IAM_CACHE_REDIS_TTL_SECONDS", "600")

	defer clearIAMEnvVars(t)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify server config from env
	assert.Equal(t, ":9090", cfg.Server.Address)
	assert.Equal(t, 9090, cfg.Server.Port)

	// Verify database config from env
	assert.Equal(t, "testdb", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "test_iam_db", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, 50, cfg.Database.MaxConns)
	assert.Equal(t, 10, cfg.Database.MaxIdle)

	// Verify cache config from env
	assert.Equal(t, "redis", cfg.Cache.Type)
	assert.True(t, cfg.Cache.Enabled)
	assert.Equal(t, 600, cfg.Cache.TTLSeconds)
	assert.Equal(t, 20000, cfg.Cache.MaxSize)
	assert.Equal(t, 15, cfg.Cache.CleanupMinutes)

	// Verify Redis config from env
	assert.Equal(t, "redis:6379", cfg.Cache.Redis.Address)
	assert.Equal(t, "secret", cfg.Cache.Redis.Password)
	assert.Equal(t, 1, cfg.Cache.Redis.DB)
	assert.Equal(t, 600, cfg.Cache.Redis.TTLSeconds)
}

func TestLoad_WithPartialEnvironmentVariables(t *testing.T) {
	// Clear any IAM_ environment variables
	clearIAMEnvVars(t)

	// Set only some environment variables
	os.Setenv("IAM_DATABASE_HOST", "mydb")
	os.Setenv("IAM_CACHE_ENABLED", "true")

	defer clearIAMEnvVars(t)

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify overridden values
	assert.Equal(t, "mydb", cfg.Database.Host)
	assert.True(t, cfg.Cache.Enabled)

	// Verify defaults are still used for non-overridden values
	assert.Equal(t, ":8081", cfg.Server.Address)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "none", cfg.Cache.Type)
}

func TestLoad_CacheTypes(t *testing.T) {
	tests := []struct {
		name         string
		cacheType    string
		cacheEnabled string
		wantType     string
		wantEnabled  bool
	}{
		{
			name:         "None type",
			cacheType:    "none",
			cacheEnabled: "false",
			wantType:     "none",
			wantEnabled:  false,
		},
		{
			name:         "Memory type",
			cacheType:    "memory",
			cacheEnabled: "true",
			wantType:     "memory",
			wantEnabled:  true,
		},
		{
			name:         "Redis type",
			cacheType:    "redis",
			cacheEnabled: "true",
			wantType:     "redis",
			wantEnabled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearIAMEnvVars(t)

			os.Setenv("IAM_CACHE_TYPE", tt.cacheType)
			os.Setenv("IAM_CACHE_ENABLED", tt.cacheEnabled)

			defer clearIAMEnvVars(t)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, cfg.Cache.Type)
			assert.Equal(t, tt.wantEnabled, cfg.Cache.Enabled)
		})
	}
}

func TestLoad_DatabaseConnectionPoolSettings(t *testing.T) {
	clearIAMEnvVars(t)

	os.Setenv("IAM_DATABASE_MAX_CONNS", "100")
	os.Setenv("IAM_DATABASE_MAX_IDLE", "20")

	defer clearIAMEnvVars(t)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 100, cfg.Database.MaxConns)
	assert.Equal(t, 20, cfg.Database.MaxIdle)
}

func TestLoad_ServerAddressFormats(t *testing.T) {
	tests := []struct {
		name    string
		address string
		port    string
	}{
		{
			name:    "Default with colon",
			address: ":8081",
			port:    "8081",
		},
		{
			name:    "With IP address",
			address: "0.0.0.0:8080",
			port:    "8080",
		},
		{
			name:    "With hostname",
			address: "localhost:9000",
			port:    "9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearIAMEnvVars(t)

			os.Setenv("IAM_SERVER_ADDRESS", tt.address)
			os.Setenv("IAM_SERVER_PORT", tt.port)

			defer clearIAMEnvVars(t)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.address, cfg.Server.Address)
		})
	}
}

func TestConfig_Structs(t *testing.T) {
	// Test that config structs can be instantiated
	cfg := &Config{
		Server: ServerConfig{
			Address: ":8080",
			Port:    8080,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "user",
			Password: "pass",
			DBName:   "db",
			SSLMode:  "disable",
			MaxConns: 25,
			MaxIdle:  5,
		},
		Cache: CacheConfig{
			Type:           "memory",
			Enabled:        true,
			TTLSeconds:     300,
			MaxSize:        10000,
			CleanupMinutes: 10,
			Redis: RedisCacheConfig{
				Address:    "localhost:6379",
				Password:   "",
				DB:         0,
				TTLSeconds: 300,
			},
		},
	}

	assert.Equal(t, ":8080", cfg.Server.Address)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "memory", cfg.Cache.Type)
	assert.Equal(t, "localhost:6379", cfg.Cache.Redis.Address)
}

// Helper function to clear IAM environment variables
func clearIAMEnvVars(t *testing.T) {
	envVars := []string{
		"IAM_SERVER_ADDRESS",
		"IAM_SERVER_PORT",
		"IAM_DATABASE_HOST",
		"IAM_DATABASE_PORT",
		"IAM_DATABASE_USER",
		"IAM_DATABASE_PASSWORD",
		"IAM_DATABASE_DBNAME",
		"IAM_DATABASE_SSLMODE",
		"IAM_DATABASE_MAX_CONNS",
		"IAM_DATABASE_MAX_IDLE",
		"IAM_CACHE_TYPE",
		"IAM_CACHE_ENABLED",
		"IAM_CACHE_TTL_SECONDS",
		"IAM_CACHE_MAX_SIZE",
		"IAM_CACHE_CLEANUP_MINUTES",
		"IAM_CACHE_REDIS_ADDRESS",
		"IAM_CACHE_REDIS_PASSWORD",
		"IAM_CACHE_REDIS_DB",
		"IAM_CACHE_REDIS_TTL_SECONDS",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	t.Cleanup(func() {
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}
	})
}
