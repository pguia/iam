package main

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeApp_Success(t *testing.T) {
	// Set test environment variables
	setupTestEnv(t)

	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	defer app.Close()

	// Verify all components are initialized
	assert.NotNil(t, app.Config)
	assert.NotNil(t, app.Database)
	assert.NotNil(t, app.IAMService)
	assert.NotNil(t, app.PermissionEvaluator)
	assert.NotNil(t, app.CacheService)

	// Verify configuration was loaded
	assert.Equal(t, "localhost", app.Config.Database.Host)
	assert.Equal(t, 5432, app.Config.Database.Port)

	// Verify database connection works
	err = app.Database.Ping()
	assert.NoError(t, err)
}

func TestInitializeApp_DatabaseConnection(t *testing.T) {
	setupTestEnv(t)

	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	defer app.Close()

	// Test that database is connected and migrations ran
	var tableCount int64
	err = app.Database.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name IN ('resources', 'permissions', 'roles', 'policies', 'bindings', 'conditions')").Scan(&tableCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(6), tableCount, "All 6 tables should exist after migration")
}

func TestInitializeApp_CacheInitialization(t *testing.T) {
	tests := []struct {
		name         string
		cacheType    string
		cacheEnabled string
	}{
		{
			name:         "None cache",
			cacheType:    "none",
			cacheEnabled: "false",
		},
		{
			name:         "Memory cache",
			cacheType:    "memory",
			cacheEnabled: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestEnv(t)
			os.Setenv("IAM_CACHE_TYPE", tt.cacheType)
			os.Setenv("IAM_CACHE_ENABLED", tt.cacheEnabled)

			app, err := InitializeApp()
			require.NoError(t, err)
			require.NotNil(t, app)
			defer app.Close()

			assert.NotNil(t, app.CacheService)
			assert.Equal(t, tt.cacheType, app.Config.Cache.Type)
		})
	}
}

func TestApp_Close(t *testing.T) {
	setupTestEnv(t)

	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)

	// Close the app
	err = app.Close()
	assert.NoError(t, err)

	// Verify database connection is closed
	err = app.Database.Ping()
	assert.Error(t, err, "Database should be closed")
}

func TestApp_CloseNilDatabase(t *testing.T) {
	app := &App{
		Database: nil,
	}

	err := app.Close()
	assert.NoError(t, err, "Close should handle nil database gracefully")
}

func TestInitializeApp_InvalidDatabaseConfig(t *testing.T) {
	setupTestEnv(t)

	// Set invalid database host
	os.Setenv("IAM_DATABASE_HOST", "nonexistent-host-99999")

	app, err := InitializeApp()
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "failed to initialize database")
}

func TestApp_ComponentsIntegration(t *testing.T) {
	setupTestEnv(t)

	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	defer app.Close()

	// Verify IAM service can interact with database through repositories
	// This is an integration test that verifies all components work together

	// Create a permission using the IAM service with unique name
	testID := uuid.New().String()[:8]
	permission, err := app.IAMService.CreatePermission(
		"integration.test.read."+testID,
		"Integration test permission for reading",
		"integration",
	)
	require.NoError(t, err)
	assert.NotNil(t, permission)
	assert.Contains(t, permission.Name, "integration.test.read")

	// Retrieve the permission
	retrieved, err := app.IAMService.GetPermission(permission.ID)
	require.NoError(t, err)
	assert.Equal(t, permission.ID, retrieved.ID)
	assert.Equal(t, permission.Name, retrieved.Name)
}

func TestApp_RepositoriesIntegration(t *testing.T) {
	setupTestEnv(t)

	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	defer app.Close()

	// Create a resource
	resource, err := app.IAMService.CreateResource(
		"bucket",
		"test-bucket",
		nil,
		map[string]string{"region": "us-east-1"},
	)
	require.NoError(t, err)
	assert.NotNil(t, resource)

	// Create a role with permissions with unique names
	testID := uuid.New().String()[:8]
	perm1, err := app.IAMService.CreatePermission("storage.objects.read."+testID, "Read storage objects", "storage")
	require.NoError(t, err)

	perm2, err := app.IAMService.CreatePermission("storage.objects.write."+testID, "Write storage objects", "storage")
	require.NoError(t, err)

	role, err := app.IAMService.CreateRole(
		"roles/storage.objects.admin."+testID,
		"Storage Objects Admin",
		"Full access to storage objects",
		[]uuid.UUID{perm1.ID, perm2.ID},
	)
	require.NoError(t, err)
	assert.NotNil(t, role)
	assert.Len(t, role.Permissions, 2)
}

func TestApp_PermissionEvaluator(t *testing.T) {
	setupTestEnv(t)

	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	defer app.Close()

	// Verify permission evaluator is working
	assert.NotNil(t, app.PermissionEvaluator)

	// Create test data
	resource, err := app.IAMService.CreateResource("project", "test-project", nil, nil)
	require.NoError(t, err)

	// Check permission (should be denied since no policy exists)
	allowed, _, err := app.PermissionEvaluator.CheckPermission(
		"user:test@example.com",
		resource.ID,
		"storage.buckets.read",
		nil,
	)
	require.NoError(t, err)
	assert.False(t, allowed, "Permission should be denied when no policy exists")
}

func TestApp_ConfigurationLoaded(t *testing.T) {
	setupTestEnv(t)

	// Set custom configuration
	os.Setenv("IAM_SERVER_ADDRESS", ":9090")
	os.Setenv("IAM_DATABASE_MAX_CONNS", "100")

	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	defer app.Close()

	// Verify configuration was applied
	assert.Equal(t, ":9090", app.Config.Server.Address)
	assert.Equal(t, 100, app.Config.Database.MaxConns)
}

func TestApp_DatabaseMigrations(t *testing.T) {
	setupTestEnv(t)

	app, err := InitializeApp()
	require.NoError(t, err)
	require.NotNil(t, app)
	defer app.Close()

	// Verify all expected tables exist
	expectedTables := []string{
		"resources",
		"permissions",
		"roles",
		"policies",
		"bindings",
		"conditions",
	}

	for _, tableName := range expectedTables {
		var exists bool
		err := app.Database.DB.Raw(
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = ?)",
			tableName,
		).Scan(&exists).Error
		require.NoError(t, err)
		assert.True(t, exists, "Table %s should exist", tableName)
	}
}

// Helper function to set up test environment
func setupTestEnv(t *testing.T) {
	// Clear environment variables
	clearEnvVars()

	// Set test database configuration
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	os.Setenv("IAM_DATABASE_HOST", host)
	os.Setenv("IAM_DATABASE_PORT", "5432")
	os.Setenv("IAM_DATABASE_USER", "postgres")
	os.Setenv("IAM_DATABASE_PASSWORD", "postgres")
	os.Setenv("IAM_DATABASE_DBNAME", "iam_db")
	os.Setenv("IAM_DATABASE_SSLMODE", "disable")
	os.Setenv("IAM_CACHE_TYPE", "none")
	os.Setenv("IAM_CACHE_ENABLED", "false")

	t.Cleanup(clearEnvVars)
}

func clearEnvVars() {
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
}
