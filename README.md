# IAM Chassis

[![Tests](https://github.com/guipguia/iam/actions/workflows/test.yml/badge.svg)](https://github.com/guipguia/iam/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/guipguia/iam/branch/main/graph/badge.svg)](https://codecov.io/gh/guipguia/iam)
[![Go Report Card](https://goreportcard.com/badge/github.com/guipguia/iam)](https://goreportcard.com/report/github.com/guipguia/iam)
[![Go Version](https://img.shields.io/github/go-mod/go-version/guipguia/iam)](https://github.com/guipguia/iam)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A generic, reusable Identity and Access Management (IAM) service inspired by Google Cloud IAM. This chassis provides fine-grained access control with hierarchical resource management, role-based permissions, and conditional access.

## Features

- **✅ Stateless Architecture**: Fully stateless by default - horizontally scalable with no local state
- **Hierarchical Resource Model**: Resources can be organized in a tree structure with permission inheritance
- **Role-Based Access Control (RBAC)**: Define roles with specific permissions
- **Policy-Based Authorization**: Attach policies to resources with multiple bindings
- **Conditional Access**: Support for attribute-based access control (ABAC) with CEL expressions
- **Flexible Caching**: Choose none (stateless), Valkey (distributed), or memory (dev only)
- **gRPC API**: High-performance gRPC service with 22 methods
- **PostgreSQL Backend**: Robust database with GORM ORM
- **Docker & Kubernetes Ready**: Easy deployment with Docker Compose or K8s
- **Horizontal Scaling**: Run multiple replicas behind a load balancer
- **Auth Integration**: Ready-to-use integration helper for combining with Auth service

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     gRPC API Layer                       │
│  (CheckPermission, Policy/Role/Resource Management)      │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                   Service Layer                          │
│  ┌──────────────────┐  ┌────────────────────────────┐  │
│  │  IAM Service     │  │  Permission Evaluator      │  │
│  │  - CRUD ops      │  │  - Check permissions       │  │
│  │  - Policy mgmt   │  │  - Hierarchical eval       │  │
│  └──────────────────┘  └────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────┐  │
│  │            Cache Service                         │  │
│  │  - In-memory cache for permission checks        │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                Repository Layer                          │
│  - ResourceRepository   - PermissionRepository          │
│  - RoleRepository       - PolicyRepository              │
│  - BindingRepository                                    │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                 Domain Models                            │
│  - Resource     - Permission     - Role                 │
│  - Policy       - Binding        - Condition            │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              PostgreSQL Database                         │
└─────────────────────────────────────────────────────────┘
```

## Core Concepts

### Resource

A hierarchical entity that can have policies attached. Resources can have parent-child relationships, enabling permission inheritance.

**Example hierarchy:**

```
Organization (org-123)
  ├── Project (project-456)
  │     ├── Bucket (bucket-789)
  │     └── Instance (instance-abc)
  └── Project (project-def)
```

### Permission

A specific action that can be performed on a resource.

**Format:** `service.resource.action`

**Examples:**

- `storage.buckets.create`
- `storage.objects.read`
- `compute.instances.start`

### Role

A collection of permissions that can be assigned to principals.

**Types:**

- **Predefined Roles**: Built-in roles (e.g., `roles/storage.admin`)
- **Custom Roles**: User-defined roles

**Example:**

```
Role: roles/storage.admin
Permissions:
  - storage.buckets.create
  - storage.buckets.delete
  - storage.objects.read
  - storage.objects.write
```

### Policy

A set of bindings attached to a resource that defines who has what access.

**Components:**

- Resource ID
- Bindings (array of role assignments)
- Version & ETag (for concurrency control)

### Binding

Associates a role with a list of members (principals).

**Components:**

- Role
- Members (e.g., `user:alice@example.com`, `group:admins`)
- Optional condition (CEL expression)

**Example:**

```json
{
  "role": "roles/storage.admin",
  "members": ["user:alice@example.com", "group:storage-admins@example.com"],
  "condition": {
    "title": "Restrict to business hours",
    "expression": "request.time.getHours() >= 9 && request.time.getHours() < 17"
  }
}
```

### Principal

An identity that can be granted access. Format: `type:identifier`

**Types:**

- `user:email@example.com`
- `group:group-name@example.com`
- `serviceAccount:sa@project.iam.gserviceaccount.com`
- `domain:example.com`

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 15+ (PostgreSQL 18 in Docker)
- Protocol Buffers compiler (`protoc`)
- Docker & Docker Compose (optional)
- Valkey 7+ (optional, for distributed caching - open source Redis alternative)

### Installation

1. Clone the repository:

```bash
cd /path/to/chassis/iam
```

2. Install dependencies:

```bash
go mod download
```

3. Install protoc plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

4. Generate proto files:

```bash
make proto
```

### Configuration

Copy the example configuration:

```bash
cp config.yaml.example config.yaml
# OR
cp .env.example .env
```

Edit the configuration to match your environment:

```yaml
server:
  address: ":8081"
  port: 8081

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: iam_db
  sslmode: disable
  max_conns: 25
  max_idle: 5

cache:
  # Cache type: "none" (stateless), "memory" (single instance only), "redis" (distributed, Valkey-compatible)
  type: none
  enabled: false # For stateless deployments, keep disabled
  ttl_seconds: 300 # 5 minutes
  max_size: 10000 # Maximum number of cache entries (memory only)
  cleanup_minutes: 10 # Run cleanup every 10 minutes (memory only)

  # Valkey/Redis configuration (for distributed caching)
  # Note: Using "redis" type for Valkey (protocol-compatible)
  redis:
    address: localhost:6379
    password: ""
    db: 0
    ttl_seconds: 300
```

### Running with Docker Compose

The easiest way to get started:

```bash
docker-compose up -d
```

This will start:

- PostgreSQL database
- Valkey (open source Redis alternative for distributed caching)
- IAM service

By default, the IAM service is configured to use Valkey caching. For a stateless deployment without caching, edit `docker-compose.yml` and set `IAM_CACHE_TYPE: none` and `IAM_CACHE_ENABLED: "false"`.

### Running Locally

1. Start PostgreSQL:

```bash
docker run -d \
  --name iam-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=iam_db \
  -p 5432:5432 \
  postgres:15-alpine
```

2. Build and run the service:

```bash
make build
./iam-server
```

### Seeding Sample Data

To populate the database with common permissions and roles:

```bash
go run examples/seed/seed_data.go
```

This creates:

- Common permissions (storage._, compute._, iam.\*)
- Predefined roles (viewer, editor, admin)
- Sample resources and policies

## Quick Start

For a complete working example:

```bash
# 1. Start services with Docker Compose
docker-compose up -d

# 2. Wait for services to be healthy
docker-compose ps

# 3. Run the seed script
go run examples/seed/seed_data.go

# 4. Run the complete application example
go run examples/complete_app_example.go

# 5. Test the endpoints
curl http://localhost:3000/health
curl -H "Authorization: Bearer <your-jwt-token>" http://localhost:3000/api/profile
```

## API Usage

### Permission Checking

Check if a principal has permission on a resource:

```protobuf
CheckPermission {
  principal: "user:alice@example.com"
  resource_id: "project-123"
  permission: "storage.buckets.create"
  context: {}
}
-> { allowed: true, reason: "Permission granted via role 'roles/storage.admin'" }
```

### Creating a Resource

```protobuf
CreateResource {
  type: "project"
  name: "My Project"
  parent_id: "org-123"
  attributes: {
    "environment": "production"
  }
}
-> Resource { id: "project-456", ... }
```

### Creating a Role

```protobuf
CreateRole {
  name: "roles/custom.viewer"
  title: "Custom Viewer"
  description: "Can view resources"
  permission_ids: ["perm-1", "perm-2"]
}
-> Role { id: "role-789", ... }
```

### Creating a Policy

```protobuf
CreatePolicy {
  resource_id: "project-456"
  bindings: [
    {
      role_id: "role-789"
      members: ["user:alice@example.com", "group:viewers"]
    }
  ]
}
-> Policy { id: "policy-abc", etag: "xyz", ... }
```

## Permission Evaluation

The IAM service evaluates permissions hierarchically:

1. Check if the principal has the permission on the requested resource
2. If not found, check the parent resource
3. Continue up the hierarchy until permission is found or root is reached
4. Cache positive results for performance

**Example:**

```
Organization (org-123)
  └── Project (project-456)
        └── Bucket (bucket-789)

Policy on org-123:
  - user:alice@example.com has role:admin

Permission Check:
  principal: user:alice@example.com
  resource: bucket-789
  permission: storage.buckets.read

Evaluation:
  1. Check bucket-789 policy -> No matching binding
  2. Check project-456 policy -> No matching binding
  3. Check org-123 policy -> Found! user:alice has role:admin
  4. role:admin includes storage.buckets.read -> ALLOWED
```

## Integration Examples

### Using the Integration Helper (Recommended)

The easiest way to integrate IAM with the Auth service is using the provided integration helper:

```go
import (
    "github.com/guipguia/iam/examples/integration"
)

// Initialize the chassis integration
chassis, err := integration.NewChassisIntegration(integration.Config{
    AuthServiceURL: "http://localhost:8080",        // Auth service URL
    IAMServiceAddr: "localhost:8081",                // IAM service address
    JWTSecret:      "your-access-token-secret",      // Must match AUTH_JWT_ACCESS_TOKEN_SECRET
})
if err != nil {
    log.Fatal(err)
}
defer chassis.Close()

// Use as HTTP middleware for auth + authz
mux := http.NewServeMux()

// Protected endpoint with permission check
mux.Handle("/api/buckets/create",
    chassis.Middleware()(
        chassis.RequirePermission("project-123", "storage.buckets.create")(
            http.HandlerFunc(createBucketHandler),
        ),
    ),
)

// Dynamic permission check based on URL parameters
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
```

See [`examples/complete_app_example.go`](examples/complete_app_example.go) for a full working example.

### Direct gRPC Client Usage

For direct gRPC access without the integration helper:

```go
import (
    iamv1 "github.com/guipguia/iam/api/proto/iam/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// Connect to IAM service
conn, _ := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
defer conn.Close()

client := iamv1.NewIAMServiceClient(conn)

// Check permission
resp, _ := client.CheckPermission(ctx, &iamv1.CheckPermissionRequest{
    Principal:  "user:alice@example.com",
    ResourceId: "project-123",
    Permission: "storage.buckets.create",
})

if resp.Allowed {
    // Allow action
} else {
    // Deny action
}
```

## Database Schema

Key tables:

- `resources`: Hierarchical resources with parent_id
- `permissions`: Available permissions
- `roles`: Role definitions
- `role_permissions`: Many-to-many relationship
- `policies`: Policies attached to resources
- `bindings`: Role assignments to principals
- `conditions`: Conditional access expressions

## Development

### Running Tests

```bash
make test
```

### Building

```bash
make build
```

### Cleaning

```bash
make clean
```

### Code Formatting

```bash
make fmt
```

## Performance Considerations

- **Flexible Caching**:
  - **Stateless mode** (default): No caching, fully stateless and horizontally scalable
  - **Memory cache**: Fast in-memory caching for single-instance deployments (not horizontally scalable)
  - **Valkey cache**: Distributed caching for multi-replica deployments (open source, BSD-3 licensed)
  - Default TTL: 5 minutes for permission checks
- **Connection Pooling**: Database connections are pooled (25 max, 5 idle by default)
- **Hierarchical Queries**: Uses PostgreSQL recursive CTEs for efficient hierarchy traversal
- **Batch Operations**: Support for batch permission checks via `BatchCheckPermissions`
- **Horizontal Scaling**: Run multiple replicas behind a load balancer (use Valkey cache or no cache)

## Security Best Practices

1. **Least Privilege**: Grant only the permissions necessary
2. **Use Groups**: Assign roles to groups, not individual users
3. **Regular Audits**: Review policies and bindings regularly
4. **Conditional Access**: Use conditions for time-based or context-based restrictions
5. **Versioning**: Use etag for optimistic concurrency control

## Additional Documentation

- **[STATELESS_DEPLOYMENT.md](STATELESS_DEPLOYMENT.md)**: Comprehensive guide for stateless, horizontally scalable deployments
- **[INTEGRATION_WITH_AUTH.md](INTEGRATION_WITH_AUTH.md)**: Detailed integration guide with the Auth chassis
- **[examples/](examples/)**: Working code examples
  - `complete_app_example.go`: Full application example with Auth + IAM integration
  - `integration/`: Reusable integration helper package
  - `seed/`: Database seeding utilities

## Deployment Options

### Stateless (Recommended for Production)

Run without caching for full horizontal scalability:

```yaml
cache:
  type: none
  enabled: false
```

Deploy multiple replicas behind a load balancer. Each replica is completely stateless.

### Single Instance with Memory Cache

Fast caching for development or single-instance production:

```yaml
cache:
  type: memory
  enabled: true
  ttl_seconds: 300
  max_size: 10000
```

⚠️ **Not suitable for multiple replicas** - cache is not shared between instances.

### Multi-Replica with Valkey Cache

Distributed caching for high-traffic production deployments:

```yaml
cache:
  type: redis # Use "redis" type for Valkey (protocol-compatible)
  enabled: true
  redis:
    address: valkey:6379
    ttl_seconds: 300
```

Deploy multiple IAM replicas sharing a single Valkey instance for cache coherence. Valkey is fully compatible with the Redis protocol and is 100% open source (BSD-3 license).

## Roadmap

- [x] Complete gRPC server implementation with 22 methods
- [x] Stateless architecture with flexible caching
- [x] Auth service integration helper
- [x] Docker and Kubernetes deployment configs
- [ ] CEL expression evaluation for conditions
- [ ] Audit logging
- [ ] Policy simulation/dry-run
- [ ] Terraform provider
- [ ] REST API gateway
- [ ] Performance metrics and monitoring
- [ ] Policy recommendations
- [ ] Bulk import/export

## License

MIT License

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
