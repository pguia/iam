package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDatabaseConfig() *config.DatabaseConfig {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	return &config.DatabaseConfig{
		Host:     host,
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "iam_db",
		SSLMode:  "disable",
		MaxConns: 25,
		MaxIdle:  5,
	}
}

// setupTestSchema creates a unique schema for test isolation and switches to it
func setupTestSchema(t *testing.T, db *Database) string {
	schemaName := fmt.Sprintf("test_%s", uuid.New().String()[:8])

	err := db.DB.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName)).Error
	require.NoError(t, err)

	err = db.DB.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	})

	return schemaName
}

func TestNew_Success(t *testing.T) {
	cfg := getTestDatabaseConfig()

	db, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Use unique schema for isolation
	setupTestSchema(t, db)

	// Verify database is not nil
	assert.NotNil(t, db.DB)
}

func TestNew_InvalidConfig(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:     "nonexistent-host-12345",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "iam_db",
		SSLMode:  "disable",
		MaxConns: 25,
		MaxIdle:  5,
	}

	db, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestDatabase_Ping(t *testing.T) {
	cfg := getTestDatabaseConfig()

	db, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Use unique schema for isolation
	setupTestSchema(t, db)

	// Test ping
	err = db.Ping()
	assert.NoError(t, err)
}

func TestDatabase_AutoMigrate(t *testing.T) {
	cfg := getTestDatabaseConfig()

	db, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Use unique schema for isolation
	schemaName := setupTestSchema(t, db)

	// Run auto migration
	err = db.AutoMigrate()
	assert.NoError(t, err)

	// Verify tables exist by checking we can query them in the test schema
	var tableCount int64

	// Check resources table
	err = db.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'resources'", schemaName).Scan(&tableCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), tableCount)

	// Check permissions table
	err = db.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'permissions'", schemaName).Scan(&tableCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), tableCount)

	// Check roles table
	err = db.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'roles'", schemaName).Scan(&tableCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), tableCount)

	// Check policies table
	err = db.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'policies'", schemaName).Scan(&tableCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), tableCount)

	// Check bindings table
	err = db.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'bindings'", schemaName).Scan(&tableCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), tableCount)

	// Check conditions table
	err = db.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'conditions'", schemaName).Scan(&tableCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), tableCount)
}

func TestDatabase_Close(t *testing.T) {
	cfg := getTestDatabaseConfig()

	db, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Use unique schema for isolation
	setupTestSchema(t, db)

	// Close database
	err = db.Close()
	assert.NoError(t, err)

	// Verify connection is closed - ping should fail after close
	err = db.Ping()
	assert.Error(t, err)
}

func TestDatabase_ConnectionPoolSettings(t *testing.T) {
	cfg := getTestDatabaseConfig()
	cfg.MaxConns = 50
	cfg.MaxIdle = 10

	db, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Use unique schema for isolation
	setupTestSchema(t, db)

	// Get underlying SQL DB
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)

	// Verify connection pool settings
	stats := sqlDB.Stats()
	assert.Equal(t, 50, stats.MaxOpenConnections)
	// Note: MaxIdleConns might not be directly verifiable from stats,
	// but we can verify the settings were applied without error
}

func TestDatabase_ExtensionsCreated(t *testing.T) {
	cfg := getTestDatabaseConfig()

	db, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Use unique schema for isolation
	setupTestSchema(t, db)

	// Verify uuid-ossp extension exists (extensions are database-wide, not schema-specific)
	var extensionExists bool
	err = db.DB.Raw("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'uuid-ossp')").Scan(&extensionExists).Error
	assert.NoError(t, err)
	assert.True(t, extensionExists, "uuid-ossp extension should be available")

	// Verify pgcrypto extension exists
	err = db.DB.Raw("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto')").Scan(&extensionExists).Error
	assert.NoError(t, err)
	assert.True(t, extensionExists, "pgcrypto extension should be available")
}

func TestDatabase_MultipleConnections(t *testing.T) {
	cfg := getTestDatabaseConfig()

	// Create first connection
	db1, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db1)
	defer db1.Close()

	// Use unique schema for isolation
	setupTestSchema(t, db1)

	// Create second connection
	db2, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, db2)
	defer db2.Close()

	// Use unique schema for second connection
	setupTestSchema(t, db2)

	// Both should be able to ping
	assert.NoError(t, db1.Ping())
	assert.NoError(t, db2.Ping())
}

func TestDatabase_DifferentDatabases(t *testing.T) {
	tests := []struct {
		name   string
		dbname string
	}{
		{
			name:   "Default database",
			dbname: "iam_db",
		},
		{
			name:   "Postgres database",
			dbname: "postgres",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getTestDatabaseConfig()
			cfg.DBName = tt.dbname

			db, err := New(cfg)
			require.NoError(t, err)
			require.NotNil(t, db)
			defer db.Close()

			// Use unique schema for isolation
			setupTestSchema(t, db)

			// Verify connection works
			err = db.Ping()
			assert.NoError(t, err)
		})
	}
}

func TestDatabase_SSLModeOptions(t *testing.T) {
	tests := []struct {
		name    string
		sslmode string
	}{
		{
			name:    "Disable SSL",
			sslmode: "disable",
		},
		{
			name:    "Prefer SSL",
			sslmode: "prefer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getTestDatabaseConfig()
			cfg.SSLMode = tt.sslmode

			db, err := New(cfg)
			// We expect this to work with disable and might work with prefer
			// depending on server config
			if err != nil {
				// If it fails, it should be a connection error, not a panic
				assert.Contains(t, err.Error(), "failed to connect to database")
				return
			}

			require.NotNil(t, db)
			defer db.Close()

			// Use unique schema for isolation
			setupTestSchema(t, db)

			// Verify connection works
			err = db.Ping()
			assert.NoError(t, err)
		})
	}
}

func TestIsExtensionExistsError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Extension name index error",
			err:      fmt.Errorf("ERROR: duplicate key value violates unique constraint \"pg_extension_name_index\" (SQLSTATE 23505)"),
			expected: true,
		},
		{
			name:     "Extension already exists",
			err:      fmt.Errorf("ERROR: extension \"pgcrypto\" already exists"),
			expected: true,
		},
		{
			name:     "Other error",
			err:      fmt.Errorf("connection refused"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExtensionExistsError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
