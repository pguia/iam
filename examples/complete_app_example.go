package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pguia/iam/examples/integration"
)

// This example shows a complete application using both Auth and IAM chassis

func main() {
	// Initialize the chassis integration
	chassis, err := integration.NewChassisIntegration(integration.Config{
		AuthServiceURL: "http://localhost:8080", // Auth service
		IAMServiceAddr: "localhost:8081",        // IAM service
		JWTSecret:      "your-jwt-secret",       // Shared with Auth service
	})
	if err != nil {
		log.Fatalf("Failed to initialize chassis: %v", err)
	}
	defer chassis.Close()

	// Setup HTTP server
	mux := http.NewServeMux()

	// Public endpoints (no auth required)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", homeHandler)

	// Authentication-only endpoints
	mux.Handle("/api/profile",
		chassis.Middleware()(
			http.HandlerFunc(getProfileHandler),
		),
	)

	// Endpoints with specific permissions (static resource)
	mux.Handle("/api/projects/create",
		chassis.Middleware()(
			chassis.RequirePermission("organization-123", "projects.create")(
				http.HandlerFunc(createProjectHandler),
			),
		),
	)

	// Endpoints with dynamic resource permissions
	mux.Handle("/api/projects/{id}/update",
		chassis.Middleware()(
			integration.RequirePermissionDynamic(
				"projects.update",
				func(r *http.Request) string {
					return r.PathValue("id") // Resource ID from URL
				},
			)(
				http.HandlerFunc(updateProjectHandler),
			),
		),
	)

	mux.Handle("/api/projects/{id}/delete",
		chassis.Middleware()(
			integration.RequirePermissionDynamic(
				"projects.delete",
				func(r *http.Request) string {
					return r.PathValue("id")
				},
			)(
				http.HandlerFunc(deleteProjectHandler),
			),
		),
	)

	// Endpoint that checks permissions programmatically
	mux.Handle("/api/projects/{id}",
		chassis.Middleware()(
			http.HandlerFunc(getProjectHandler),
		),
	)

	// Admin endpoint
	mux.Handle("/api/admin/users",
		chassis.Middleware()(
			chassis.RequirePermission("*", "admin.users.list")(
				http.HandlerFunc(listUsersHandler),
			),
		),
	)

	log.Println("Server starting on :3000")
	log.Println("Make sure Auth service is running on :8080")
	log.Println("Make sure IAM service is running on :8081")
	log.Fatal(http.ListenAndServe(":3000", mux))
}

// Handler implementations

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
		<h1>Chassis Integration Example</h1>
		<h2>Public Endpoints:</h2>
		<ul>
			<li>GET /health</li>
		</ul>
		<h2>Protected Endpoints (require JWT from Auth service):</h2>
		<ul>
			<li>GET /api/profile - Authentication only</li>
			<li>POST /api/projects/create - Requires 'projects.create' permission</li>
			<li>PUT /api/projects/{id}/update - Requires 'projects.update' permission</li>
			<li>DELETE /api/projects/{id}/delete - Requires 'projects.delete' permission</li>
			<li>GET /api/projects/{id} - Auth + programmatic permission check</li>
			<li>GET /api/admin/users - Requires 'admin.users.list' permission</li>
		</ul>
		<h2>How to test:</h2>
		<pre>
# 1. Login to get JWT token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "password"}'

# 2. Use token to access protected endpoints
curl http://localhost:3000/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
		</pre>
	`))
}

func getProfileHandler(w http.ResponseWriter, r *http.Request) {
	userEmail := integration.GetUserEmail(r)
	userID := integration.GetUserID(r)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"email":   userEmail,
		"message": "This endpoint only requires authentication",
	})
}

func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	userEmail := integration.GetUserEmail(r)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create project logic here...
	projectID := "project-" + generateID()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"project_id":  projectID,
		"name":        req.Name,
		"description": req.Description,
		"created_by":  userEmail,
		"message":     "Project created successfully",
	})
}

func updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	userEmail := integration.GetUserEmail(r)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update project logic here...

	json.NewEncoder(w).Encode(map[string]interface{}{
		"project_id":  projectID,
		"name":        req.Name,
		"description": req.Description,
		"updated_by":  userEmail,
		"message":     "Project updated successfully",
	})
}

func deleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	userEmail := integration.GetUserEmail(r)

	// Delete project logic here...

	json.NewEncoder(w).Encode(map[string]interface{}{
		"project_id": projectID,
		"deleted_by": userEmail,
		"message":    "Project deleted successfully",
	})
}

func getProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	userEmail := integration.GetUserEmail(r)
	chassis := integration.GetChassisIntegration(r)

	// Check if user can read this project
	allowed, reason, err := chassis.CheckPermission(r.Context(), userEmail, projectID, "projects.read")
	if err != nil {
		http.Error(w, "Permission check failed", http.StatusInternalServerError)
		return
	}

	if !allowed {
		http.Error(w, fmt.Sprintf("Forbidden: %s", reason), http.StatusForbidden)
		return
	}

	// Get effective permissions for this project
	permissions, roles, err := chassis.GetEffectivePermissions(r.Context(), userEmail, projectID)
	if err != nil {
		log.Printf("Failed to get effective permissions: %v", err)
	}

	// Get project details...
	project := map[string]interface{}{
		"id":          projectID,
		"name":        "Example Project",
		"description": "A sample project",
		"created_at":  "2024-01-01T00:00:00Z",
	}

	// Add permission info
	json.NewEncoder(w).Encode(map[string]interface{}{
		"project":     project,
		"permissions": permissions,
		"roles":       roles,
		"user_email":  userEmail,
	})
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	// This endpoint requires admin.users.list permission
	// Only admin users should be able to access this

	users := []map[string]string{
		{"id": "1", "email": "alice@example.com"},
		{"id": "2", "email": "bob@example.com"},
		{"id": "3", "email": "charlie@example.com"},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"total": len(users),
	})
}

// Helper function
func generateID() string {
	// In production, use UUID or similar
	return "123456"
}

/*
SETUP INSTRUCTIONS:

1. Start Auth Service:
   cd /Users/guilhermeguia/sandbox/chassis/auth
   docker-compose up -d

2. Start IAM Service:
   cd /Users/guilhermeguia/sandbox/chassis/iam
   docker-compose up -d

3. Register a user (Auth):
   curl -X POST http://localhost:8080/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email": "alice@example.com", "password": "password123"}'

4. Create permissions (IAM):
   # Use grpcurl or the IAM service directly to create:
   - Permission: projects.create
   - Permission: projects.read
   - Permission: projects.update
   - Permission: projects.delete
   - Permission: admin.users.list

5. Create role (IAM):
   # Create a "Project Admin" role with all project permissions

6. Grant permissions (IAM):
   # Create a policy on organization-123 that grants
   # user:alice@example.com the Project Admin role

7. Test the flow:
   # Login
   TOKEN=$(curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email": "alice@example.com", "password": "password123"}' \
     | jq -r '.access_token')

   # Create project (needs permission)
   curl -X POST http://localhost:3000/api/projects/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name": "My Project", "description": "Test project"}'

   # Get profile (auth only)
   curl http://localhost:3000/api/profile \
     -H "Authorization: Bearer $TOKEN"

EXPECTED FLOW:
1. User logs in via Auth service → gets JWT
2. User makes request to your app with JWT
3. Your app validates JWT (Auth)
4. Your app checks permission via IAM
5. If both pass → allow request
6. Otherwise → deny

PRINCIPAL FORMAT:
- Auth service stores: email = "alice@example.com"
- IAM service uses: principal = "user:alice@example.com"
- Integration helper automatically formats it correctly
*/
