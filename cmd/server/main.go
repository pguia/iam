package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pguia/iam/internal/config"
	"github.com/pguia/iam/internal/database"
	"github.com/pguia/iam/internal/repository"
	"github.com/pguia/iam/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
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
		log.Fatalf("Failed to initialize cache: %v", err)
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

	log.Printf("IAM service initialized: %v", iamService)

	// TODO: Create gRPC server and register IAM service
	// This will be implemented after proto files are generated
	log.Printf("IAM service would be listening on %s", cfg.Server.Address)
	log.Println("Note: gRPC server implementation pending proto file generation")

	// For now, just keep the service running
	log.Println("IAM service is ready (core services initialized)")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
