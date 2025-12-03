# Basic Usage Examples

This document provides examples of common IAM operations.

## 1. Setting Up a Basic Hierarchy

```go
// Create an organization
org, _ := iamService.CreateResource("organization", "Acme Corp", nil, map[string]string{
    "industry": "technology",
})

// Create a project under the organization
project, _ := iamService.CreateResource("project", "Web Application", &org.ID, map[string]string{
    "environment": "production",
})

// Create resources under the project
bucket, _ := iamService.CreateResource("bucket", "user-uploads", &project.ID, map[string]string{
    "region": "us-east-1",
})
```

## 2. Creating Permissions

```go
// Define storage permissions
readPerm, _ := iamService.CreatePermission(
    "storage.objects.read",
    "Read objects from storage",
    "storage",
)

writePerm, _ := iamService.CreatePermission(
    "storage.objects.write",
    "Write objects to storage",
    "storage",
)

deletePerm, _ := iamService.CreatePermission(
    "storage.objects.delete",
    "Delete objects from storage",
    "storage",
)
```

## 3. Creating Roles

```go
// Create a storage viewer role
viewerRole, _ := iamService.CreateRole(
    "roles/storage.viewer",
    "Storage Viewer",
    "Can read objects from storage",
    []uuid.UUID{readPerm.ID},
)

// Create a storage admin role
adminRole, _ := iamService.CreateRole(
    "roles/storage.admin",
    "Storage Admin",
    "Full access to storage",
    []uuid.UUID{readPerm.ID, writePerm.ID, deletePerm.ID},
)
```

## 4. Creating Policies and Bindings

```go
// Create a policy on the bucket
policy, _ := iamService.CreatePolicy(bucket.ID, []domain.Binding{
    {
        RoleID: viewerRole.ID,
        Members: datatypes.JSON(`["user:alice@example.com", "group:viewers@example.com"]`),
    },
    {
        RoleID: adminRole.ID,
        Members: datatypes.JSON(`["user:bob@example.com"]`),
    },
})
```

## 5. Checking Permissions

```go
// Check if Alice can read from the bucket
allowed, reason, _ := iamService.CheckPermission(
    "user:alice@example.com",
    bucket.ID,
    "storage.objects.read",
    nil,
)

if allowed {
    fmt.Println("Access granted:", reason)
    // -> "Permission granted via role 'roles/storage.viewer' on resource 'bucket-xxx'"
} else {
    fmt.Println("Access denied:", reason)
}

// Check if Alice can delete (should fail)
allowed, reason, _ = iamService.CheckPermission(
    "user:alice@example.com",
    bucket.ID,
    "storage.objects.delete",
    nil,
)
// -> allowed: false, reason: "Permission denied: no matching policy found"
```

## 6. Hierarchical Permission Inheritance

```go
// Add a policy at the organization level
orgPolicy, _ := iamService.CreatePolicy(org.ID, []domain.Binding{
    {
        RoleID: adminRole.ID,
        Members: datatypes.JSON(`["user:charlie@example.com"]`),
    },
})

// Charlie can now access the bucket even though the binding is on the org
allowed, reason, _ := iamService.CheckPermission(
    "user:charlie@example.com",
    bucket.ID,
    "storage.objects.write",
    nil,
)
// -> allowed: true (permission inherited from organization)
```

## 7. Getting Effective Permissions

```go
// Get all effective permissions for a user on a resource
permissions, roles, _ := iamService.GetEffectivePermissions(
    "user:charlie@example.com",
    bucket.ID,
)

fmt.Println("Permissions:", permissions)
// -> ["storage.objects.read", "storage.objects.write", "storage.objects.delete"]

fmt.Println("Roles:", roles)
// -> ["roles/storage.admin"]
```

## 8. Conditional Access

```go
// Create a binding with a time-based condition
binding, _ := iamService.CreateBinding(
    bucket.ID,
    viewerRole.ID,
    []string{"user:dave@example.com"},
    &domain.Condition{
        Title:       "Business hours only",
        Description: "Access only during business hours (9 AM - 5 PM)",
        Expression:  "request.time.getHours() >= 9 && request.time.getHours() < 17",
    },
)

// During business hours
allowed, _, _ := iamService.CheckPermission(
    "user:dave@example.com",
    bucket.ID,
    "storage.objects.read",
    map[string]string{"time": "14:00"},
)
// -> allowed: true

// After hours
allowed, _, _ = iamService.CheckPermission(
    "user:dave@example.com",
    bucket.ID,
    "storage.objects.read",
    map[string]string{"time": "18:00"},
)
// -> allowed: false (condition not met)
```

## 9. Updating Policies

```go
// Get the current policy
policy, _ := iamService.GetPolicy(bucket.ID)

// Update with new bindings
updatedPolicy, _ := iamService.UpdatePolicy(
    bucket.ID,
    []domain.Binding{
        {
            RoleID: viewerRole.ID,
            Members: datatypes.JSON(`["user:alice@example.com", "user:eve@example.com"]`),
        },
    },
    policy.ETag, // For optimistic concurrency control
)
```

## 10. Listing Resources

```go
// List all projects under an organization
projects, _ := iamService.ListResources(&org.ID, "project", 10, 0)

for _, project := range projects {
    fmt.Printf("Project: %s (%s)\n", project.Name, project.ID)
}
```

## 11. Complete Example: Multi-Tenant Application

```go
package main

import (
    "fmt"
    "github.com/google/uuid"
    "github.com/pguia/iam/internal/domain"
    "gorm.io/datatypes"
)

func setupMultiTenantIAM(iamService *service.IAMService) {
    // 1. Create permissions
    readPerm, _ := iamService.CreatePermission("app.data.read", "Read data", "app")
    writePerm, _ := iamService.CreatePermission("app.data.write", "Write data", "app")
    adminPerm, _ := iamService.CreatePermission("app.admin", "Admin access", "app")

    // 2. Create roles
    readerRole, _ := iamService.CreateRole(
        "roles/app.reader",
        "Reader",
        "Can read data",
        []uuid.UUID{readPerm.ID},
    )

    writerRole, _ := iamService.CreateRole(
        "roles/app.writer",
        "Writer",
        "Can read and write data",
        []uuid.UUID{readPerm.ID, writePerm.ID},
    )

    adminRole, _ := iamService.CreateRole(
        "roles/app.admin",
        "Admin",
        "Full access",
        []uuid.UUID{readPerm.ID, writePerm.ID, adminPerm.ID},
    )

    // 3. Create tenant hierarchy
    platform, _ := iamService.CreateResource("platform", "SaaS Platform", nil, nil)

    tenant1, _ := iamService.CreateResource("tenant", "Tenant 1", &platform.ID, map[string]string{
        "plan": "enterprise",
    })

    tenant2, _ := iamService.CreateResource("tenant", "Tenant 2", &platform.ID, map[string]string{
        "plan": "basic",
    })

    // 4. Set up policies for tenant 1
    iamService.CreatePolicy(tenant1.ID, []domain.Binding{
        {
            RoleID:  adminRole.ID,
            Members: datatypes.JSON(`["user:admin@tenant1.com"]`),
        },
        {
            RoleID:  writerRole.ID,
            Members: datatypes.JSON(`["user:user1@tenant1.com", "user:user2@tenant1.com"]`),
        },
        {
            RoleID:  readerRole.ID,
            Members: datatypes.JSON(`["user:viewer@tenant1.com"]`),
        },
    })

    // 5. Set up policies for tenant 2
    iamService.CreatePolicy(tenant2.ID, []domain.Binding{
        {
            RoleID:  adminRole.ID,
            Members: datatypes.JSON(`["user:admin@tenant2.com"]`),
        },
    })

    // 6. Test permissions
    testPermissions(iamService, tenant1.ID, tenant2.ID)
}

func testPermissions(iamService *service.IAMService, tenant1ID, tenant2ID uuid.UUID) {
    tests := []struct {
        principal  string
        resourceID uuid.UUID
        permission string
        expected   bool
    }{
        {"user:admin@tenant1.com", tenant1ID, "app.admin", true},
        {"user:user1@tenant1.com", tenant1ID, "app.data.write", true},
        {"user:user1@tenant1.com", tenant1ID, "app.admin", false},
        {"user:viewer@tenant1.com", tenant1ID, "app.data.read", true},
        {"user:viewer@tenant1.com", tenant1ID, "app.data.write", false},
        // Cross-tenant access (should fail)
        {"user:admin@tenant1.com", tenant2ID, "app.admin", false},
        {"user:admin@tenant2.com", tenant1ID, "app.admin", false},
    }

    for _, test := range tests {
        allowed, reason, _ := iamService.CheckPermission(
            test.principal,
            test.resourceID,
            test.permission,
            nil,
        )

        status := "✓"
        if allowed != test.expected {
            status = "✗"
        }

        fmt.Printf("%s %s on %s: %s - %s\n",
            status, test.principal, test.permission,
            map[bool]string{true: "ALLOW", false: "DENY"}[allowed],
            reason,
        )
    }
}
```

## Best Practices

1. **Use Groups Instead of Individual Users**

   ```go
   // Good
   Members: []string{"group:engineers@example.com"}

   // Avoid
   Members: []string{"user:alice@...", "user:bob@...", "user:charlie@..."}
   ```

2. **Grant Permissions at the Highest Appropriate Level**

   ```go
   // If all projects need the same access, grant at org level
   iamService.CreatePolicy(orgID, bindings)

   // Instead of duplicating for each project
   ```

3. **Use Custom Roles for Specific Needs**

   ```go
   // Create a role that combines exactly the permissions needed
   customRole, _ := iamService.CreateRole(
       "roles/custom.deployer",
       "Deployer",
       "Can deploy but not delete",
       []uuid.UUID{readPerm.ID, writePerm.ID}, // Exclude deletePerm
   )
   ```

4. **Always Check ETag When Updating**

   ```go
   policy, _ := iamService.GetPolicy(resourceID)
   // ... make changes ...
   iamService.UpdatePolicy(resourceID, newBindings, policy.ETag)
   ```

5. **Use Conditions for Time-Based or Context-Based Access**

   ```go
   // Temporary access
   &domain.Condition{
       Expression: "request.time < timestamp('2024-12-31T23:59:59Z')",
   }

   // IP-based access
   &domain.Condition{
       Expression: "request.ip.startsWith('10.0.0.')",
   }
   ```
