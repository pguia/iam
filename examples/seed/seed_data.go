package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/pguia/iam/internal/config"
	"github.com/pguia/iam/internal/database"
	"github.com/pguia/iam/internal/repository"
	"github.com/pguia/iam/internal/service"
)

// This script seeds the database with common permissions and roles
// Run with: go run examples/seed_data.go

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

	// Initialize repositories
	resourceRepo := repository.NewResourceRepository(db.DB)
	permissionRepo := repository.NewPermissionRepository(db.DB)
	roleRepo := repository.NewRoleRepository(db.DB)
	policyRepo := repository.NewPolicyRepository(db.DB)
	bindingRepo := repository.NewBindingRepository(db.DB)

	// Initialize services
	cacheService := service.NewCacheService(&cfg.Cache)
	permissionEvaluator := service.NewPermissionEvaluator(
		resourceRepo,
		policyRepo,
		permissionRepo,
		cacheService,
	)
	iamService := service.NewIAMService(
		resourceRepo,
		permissionRepo,
		roleRepo,
		policyRepo,
		bindingRepo,
		permissionEvaluator,
		cacheService,
	)

	log.Println("Starting to seed IAM data...")

	// Seed permissions
	permissions := seedPermissions(iamService)
	log.Printf("Created %d permissions", len(permissions))

	// Seed roles
	roles := seedRoles(iamService, permissions)
	log.Printf("Created %d roles", len(roles))

	// Seed sample resources
	resources := seedResources(iamService)
	log.Printf("Created %d resources", len(resources))

	log.Println("Seeding completed successfully!")
}

func seedPermissions(iamService *service.IAMService) map[string]uuid.UUID {
	permissions := make(map[string]uuid.UUID)

	permissionDefs := []struct {
		name        string
		description string
		service     string
	}{
		// Storage permissions
		{"storage.buckets.list", "List storage buckets", "storage"},
		{"storage.buckets.get", "Get bucket details", "storage"},
		{"storage.buckets.create", "Create new buckets", "storage"},
		{"storage.buckets.update", "Update bucket settings", "storage"},
		{"storage.buckets.delete", "Delete buckets", "storage"},
		{"storage.objects.list", "List objects in bucket", "storage"},
		{"storage.objects.get", "Read objects from storage", "storage"},
		{"storage.objects.create", "Upload objects to storage", "storage"},
		{"storage.objects.update", "Update objects in storage", "storage"},
		{"storage.objects.delete", "Delete objects from storage", "storage"},

		// Compute permissions
		{"compute.instances.list", "List compute instances", "compute"},
		{"compute.instances.get", "Get instance details", "compute"},
		{"compute.instances.create", "Create new instances", "compute"},
		{"compute.instances.start", "Start instances", "compute"},
		{"compute.instances.stop", "Stop instances", "compute"},
		{"compute.instances.delete", "Delete instances", "compute"},

		// IAM permissions
		{"iam.roles.list", "List IAM roles", "iam"},
		{"iam.roles.get", "Get role details", "iam"},
		{"iam.roles.create", "Create custom roles", "iam"},
		{"iam.roles.update", "Update roles", "iam"},
		{"iam.roles.delete", "Delete custom roles", "iam"},
		{"iam.policies.get", "Get resource policies", "iam"},
		{"iam.policies.update", "Update resource policies", "iam"},

		// Admin permissions
		{"admin.all", "Full administrative access", "admin"},
	}

	for _, perm := range permissionDefs {
		p, err := iamService.CreatePermission(perm.name, perm.description, perm.service)
		if err != nil {
			log.Printf("Warning: Failed to create permission %s: %v", perm.name, err)
			continue
		}
		permissions[perm.name] = p.ID
		log.Printf("  ✓ Created permission: %s", perm.name)
	}

	return permissions
}

func seedRoles(iamService *service.IAMService, permissions map[string]uuid.UUID) map[string]uuid.UUID {
	roles := make(map[string]uuid.UUID)

	roleDefs := []struct {
		name        string
		title       string
		description string
		permissions []string
	}{
		{
			"roles/storage.viewer",
			"Storage Viewer",
			"Read-only access to storage resources",
			[]string{"storage.buckets.list", "storage.buckets.get", "storage.objects.list", "storage.objects.get"},
		},
		{
			"roles/storage.editor",
			"Storage Editor",
			"Read and write access to storage resources",
			[]string{
				"storage.buckets.list", "storage.buckets.get", "storage.buckets.update",
				"storage.objects.list", "storage.objects.get", "storage.objects.create",
				"storage.objects.update", "storage.objects.delete",
			},
		},
		{
			"roles/storage.admin",
			"Storage Admin",
			"Full access to storage resources",
			[]string{
				"storage.buckets.list", "storage.buckets.get", "storage.buckets.create",
				"storage.buckets.update", "storage.buckets.delete",
				"storage.objects.list", "storage.objects.get", "storage.objects.create",
				"storage.objects.update", "storage.objects.delete",
			},
		},
		{
			"roles/compute.viewer",
			"Compute Viewer",
			"Read-only access to compute resources",
			[]string{"compute.instances.list", "compute.instances.get"},
		},
		{
			"roles/compute.operator",
			"Compute Operator",
			"Can start and stop instances",
			[]string{
				"compute.instances.list", "compute.instances.get",
				"compute.instances.start", "compute.instances.stop",
			},
		},
		{
			"roles/compute.admin",
			"Compute Admin",
			"Full access to compute resources",
			[]string{
				"compute.instances.list", "compute.instances.get", "compute.instances.create",
				"compute.instances.start", "compute.instances.stop", "compute.instances.delete",
			},
		},
		{
			"roles/iam.viewer",
			"IAM Viewer",
			"Read-only access to IAM resources",
			[]string{"iam.roles.list", "iam.roles.get", "iam.policies.get"},
		},
		{
			"roles/iam.admin",
			"IAM Admin",
			"Full access to IAM resources",
			[]string{
				"iam.roles.list", "iam.roles.get", "iam.roles.create",
				"iam.roles.update", "iam.roles.delete",
				"iam.policies.get", "iam.policies.update",
			},
		},
		{
			"roles/owner",
			"Owner",
			"Full access to all resources",
			[]string{"admin.all"},
		},
	}

	for _, role := range roleDefs {
		// Get permission IDs
		var permIDs []uuid.UUID
		for _, permName := range role.permissions {
			if permID, ok := permissions[permName]; ok {
				permIDs = append(permIDs, permID)
			}
		}

		r, err := iamService.CreateRole(role.name, role.title, role.description, permIDs)
		if err != nil {
			log.Printf("Warning: Failed to create role %s: %v", role.name, err)
			continue
		}
		roles[role.name] = r.ID
		log.Printf("  ✓ Created role: %s (%d permissions)", role.name, len(permIDs))
	}

	return roles
}

func seedResources(iamService *service.IAMService) map[string]uuid.UUID {
	resources := make(map[string]uuid.UUID)

	// Create organization
	org, err := iamService.CreateResource("organization", "Example Corp", nil, map[string]string{
		"industry": "technology",
	})
	if err != nil {
		log.Printf("Warning: Failed to create organization: %v", err)
		return resources
	}
	resources["org"] = org.ID
	log.Printf("  ✓ Created organization: %s", org.Name)

	// Create projects
	project1, err := iamService.CreateResource("project", "Production", &org.ID, map[string]string{
		"environment": "production",
	})
	if err != nil {
		log.Printf("Warning: Failed to create project: %v", err)
		return resources
	}
	resources["project1"] = project1.ID
	log.Printf("  ✓ Created project: %s", project1.Name)

	project2, err := iamService.CreateResource("project", "Development", &org.ID, map[string]string{
		"environment": "development",
	})
	if err != nil {
		log.Printf("Warning: Failed to create project: %v", err)
		return resources
	}
	resources["project2"] = project2.ID
	log.Printf("  ✓ Created project: %s", project2.Name)

	// Create buckets
	bucket1, err := iamService.CreateResource("bucket", "prod-data", &project1.ID, map[string]string{
		"region": "us-east-1",
	})
	if err != nil {
		log.Printf("Warning: Failed to create bucket: %v", err)
		return resources
	}
	resources["bucket1"] = bucket1.ID
	log.Printf("  ✓ Created bucket: %s", bucket1.Name)

	bucket2, err := iamService.CreateResource("bucket", "dev-data", &project2.ID, map[string]string{
		"region": "us-west-2",
	})
	if err != nil {
		log.Printf("Warning: Failed to create bucket: %v", err)
		return resources
	}
	resources["bucket2"] = bucket2.ID
	log.Printf("  ✓ Created bucket: %s", bucket2.Name)

	return resources
}

func printSummary(iamService *service.IAMService) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Seed Data Summary")
	fmt.Println(strings.Repeat("=", 60))

	// List permissions
	permissions, _ := iamService.ListPermissions("", 100, 0)
	fmt.Printf("\nPermissions: %d\n", len(permissions))
	for _, p := range permissions {
		fmt.Printf("  - %s (%s)\n", p.Name, p.Service)
	}

	// List roles
	roles, _ := iamService.ListRoles(true, 100, 0)
	fmt.Printf("\nRoles: %d\n", len(roles))
	for _, r := range roles {
		fmt.Printf("  - %s: %s (%d permissions)\n", r.Name, r.Title, len(r.Permissions))
	}

	// List resources
	resources, _ := iamService.ListResources(nil, "", 100, 0)
	fmt.Printf("\nResources: %d\n", len(resources))
	for _, r := range resources {
		fmt.Printf("  - %s (%s)\n", r.Name, r.Type)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
}
