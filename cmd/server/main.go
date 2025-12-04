package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pguia/iam/internal/config"
	"github.com/pguia/iam/internal/database"
	"github.com/pguia/iam/internal/repository"
	"github.com/pguia/iam/internal/service"
)

// App holds all application components
type App struct {
	Config              *config.Config
	Database            *database.Database
	IAMService          *service.IAMService
	PermissionEvaluator service.PermissionEvaluator
	CacheService        service.CacheService
}

// InitializeApp initializes all application components
func InitializeApp() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")

	// Initialize repositories
	resourceRepo := repository.NewResourceRepository(db.DB)
	permissionRepo := repository.NewPermissionRepository(db.DB)
	roleRepo := repository.NewRoleRepository(db.DB)
	policyRepo := repository.NewPolicyRepository(db.DB)
	bindingRepo := repository.NewBindingRepository(db.DB)

	// Initialize services
	cacheService, err := service.NewCache(&cfg.Cache)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}
	log.Printf("Cache initialized: type=%s, enabled=%v", cfg.Cache.Type, cfg.Cache.Enabled)

	permissionEvaluator := service.NewPermissionEvaluator(
		resourceRepo,
		policyRepo,
		permissionRepo,
		cacheService,
	)

	// Initialize IAM service
	iamService := service.NewIAMService(
		resourceRepo,
		permissionRepo,
		roleRepo,
		policyRepo,
		bindingRepo,
		permissionEvaluator,
		cacheService,
	)

	log.Printf("IAM service initialized successfully")

	return &App{
		Config:              cfg,
		Database:            db,
		IAMService:          iamService,
		PermissionEvaluator: permissionEvaluator,
		CacheService:        cacheService,
	}, nil
}

// Close cleans up application resources
func (app *App) Close() error {
	log.Println("Closing application resources...")
	if app.Database != nil {
		return app.Database.Close()
	}
	return nil
}

// Run starts the application and waits for shutdown signal
func Run(app *App) error {
	// TODO: Create gRPC server and register IAM service
	// This will be implemented after proto files are generated
	log.Printf("IAM service would be listening on %s", app.Config.Server.Address)
	log.Println("Note: gRPC server implementation pending proto file generation")

	// For now, just keep the service running
	log.Println("IAM service is ready (core services initialized)")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	return nil
}

func main() {
	app, err := InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Close()

	if err := Run(app); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
