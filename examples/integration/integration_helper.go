package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	iamv1 "github.com/pguia/iam/api/proto/iam/v1"
)

// ChassisIntegration provides a simple way to integrate Auth + IAM services
type ChassisIntegration struct {
	authServiceURL string
	iamClient      iamv1.IAMServiceClient
	iamConn        *grpc.ClientConn
	jwtValidator   JWTValidator
}

// JWTValidator validates JWT tokens from the Auth service
type JWTValidator interface {
	ValidateToken(token string) (*UserClaims, error)
}

// UserClaims represents the user information from the JWT
type UserClaims struct {
	UserID    string
	Email     string
	ExpiresAt time.Time
}

// Config for the integration
type Config struct {
	AuthServiceURL string
	IAMServiceAddr string // e.g., "localhost:8081"
	JWTSecret      string // The access token secret from the auth service
}

// standardJWTValidator implements JWTValidator using golang-jwt
type standardJWTValidator struct {
	secret []byte
}

// CustomClaims represents the JWT claims structure from the auth service
type CustomClaims struct {
	UserID string            `json:"user_id"`
	Email  string            `json:"email"`
	Type   string            `json:"type"`
	Extra  map[string]string `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// NewJWTValidator creates a new JWT validator with the given secret
func NewJWTValidator(secret string) JWTValidator {
	return &standardJWTValidator{
		secret: []byte(secret),
	}
}

// ValidateToken validates a JWT token and extracts user claims
func (v *standardJWTValidator) ValidateToken(tokenString string) (*UserClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate token type (must be an access token)
	if claims.Type != "access" {
		return nil, fmt.Errorf("invalid token type: expected access, got %s", claims.Type)
	}

	// Convert to UserClaims
	userClaims := &UserClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	return userClaims, nil
}

// NewChassisIntegration creates a new integration helper
func NewChassisIntegration(cfg Config) (*ChassisIntegration, error) {
	// Connect to IAM service
	conn, err := grpc.NewClient(cfg.IAMServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IAM service: %w", err)
	}

	iamClient := iamv1.NewIAMServiceClient(conn)

	// Create JWT validator
	jwtValidator := NewJWTValidator(cfg.JWTSecret)

	return &ChassisIntegration{
		authServiceURL: cfg.AuthServiceURL,
		iamClient:      iamClient,
		iamConn:        conn,
		jwtValidator:   jwtValidator,
	}, nil
}

// Close closes the gRPC connection
func (ci *ChassisIntegration) Close() error {
	return ci.iamConn.Close()
}

// Middleware returns an HTTP middleware that handles both auth and authz
func (ci *ChassisIntegration) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract and validate JWT
			token := extractBearerToken(r)
			if token == "" {
				http.Error(w, "Unauthorized: no token provided", http.StatusUnauthorized)
				return
			}

			claims, err := ci.jwtValidator.ValidateToken(token)
			if err != nil {
				http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), "user_email", claims.Email)
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "chassis_integration", ci)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission returns a middleware that checks a specific permission
func (ci *ChassisIntegration) RequirePermission(resourceID, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userEmail := r.Context().Value("user_email").(string)

			allowed, reason, err := ci.CheckPermission(r.Context(), userEmail, resourceID, permission)
			if err != nil {
				http.Error(w, "Authorization check failed", http.StatusInternalServerError)
				return
			}

			if !allowed {
				http.Error(w, fmt.Sprintf("Forbidden: %s", reason), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CheckPermission checks if a user has a permission on a resource
func (ci *ChassisIntegration) CheckPermission(ctx context.Context, userEmail, resourceID, permission string) (bool, string, error) {
	principal := fmt.Sprintf("user:%s", userEmail)

	resp, err := ci.iamClient.CheckPermission(ctx, &iamv1.CheckPermissionRequest{
		Principal:  principal,
		ResourceId: resourceID,
		Permission: permission,
		Context:    nil,
	})
	if err != nil {
		return false, "", err
	}

	return resp.Allowed, resp.Reason, nil
}

// GetEffectivePermissions returns all permissions for a user on a resource
func (ci *ChassisIntegration) GetEffectivePermissions(ctx context.Context, userEmail, resourceID string) ([]string, []string, error) {
	principal := fmt.Sprintf("user:%s", userEmail)

	resp, err := ci.iamClient.GetEffectivePermissions(ctx, &iamv1.GetEffectivePermissionsRequest{
		Principal:  principal,
		ResourceId: resourceID,
	})
	if err != nil {
		return nil, nil, err
	}

	return resp.Permissions, resp.Roles, nil
}

// Helper functions

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// GetUserEmail extracts the user email from the request context
func GetUserEmail(r *http.Request) string {
	email, _ := r.Context().Value("user_email").(string)
	return email
}

// GetUserID extracts the user ID from the request context
func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value("user_id").(string)
	return id
}

// GetChassisIntegration extracts the integration from the request context
func GetChassisIntegration(r *http.Request) *ChassisIntegration {
	ci, _ := r.Context().Value("chassis_integration").(*ChassisIntegration)
	return ci
}

// RequirePermissionDynamic checks permission with dynamic resource ID
func RequirePermissionDynamic(permission string, getResourceID func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ci := GetChassisIntegration(r)
			if ci == nil {
				http.Error(w, "Internal error: integration not found", http.StatusInternalServerError)
				return
			}

			userEmail := GetUserEmail(r)
			resourceID := getResourceID(r)

			allowed, reason, err := ci.CheckPermission(r.Context(), userEmail, resourceID, permission)
			if err != nil {
				http.Error(w, "Authorization check failed", http.StatusInternalServerError)
				return
			}

			if !allowed {
				http.Error(w, fmt.Sprintf("Forbidden: %s", reason), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Example usage:
/*
func main() {
	// Initialize integration
	chassis, err := integration.NewChassisIntegration(integration.Config{
		AuthServiceURL: "http://localhost:8080",
		IAMServiceAddr: "localhost:8081",
		JWTSecret:      "your-access-token-secret", // Must match AUTH_JWT_ACCESS_TOKEN_SECRET from auth service
	})
	if err != nil {
		log.Fatal(err)
	}
	defer chassis.Close()

	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("/health", healthHandler)

	// Protected with auth only
	mux.Handle("/api/profile",
		chassis.Middleware()(
			http.HandlerFunc(profileHandler),
		),
	)

	// Protected with auth + static permission
	mux.Handle("/api/buckets/create",
		chassis.Middleware()(
			chassis.RequirePermission("project-123", "storage.buckets.create")(
				http.HandlerFunc(createBucketHandler),
			),
		),
	)

	// Protected with auth + dynamic permission
	mux.Handle("/api/buckets/{id}/delete",
		chassis.Middleware()(
			integration.RequirePermissionDynamic(
				"storage.buckets.delete",
				func(r *http.Request) string {
					return r.PathValue("id")
				},
			)(
				http.HandlerFunc(deleteBucketHandler),
			),
		),
	)

	http.ListenAndServe(":3000", mux)
}

func createBucketHandler(w http.ResponseWriter, r *http.Request) {
	userEmail := integration.GetUserEmail(r)
	// Create bucket...
	w.Write([]byte(fmt.Sprintf("Bucket created by %s", userEmail)))
}

func deleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	bucketID := r.PathValue("id")
	userEmail := integration.GetUserEmail(r)
	// Delete bucket...
	w.Write([]byte(fmt.Sprintf("Bucket %s deleted by %s", bucketID, userEmail)))
}
*/
