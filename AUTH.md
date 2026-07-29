# Authentication & Authorization System

## Overview

This document describes the authentication and authorization system for the Business Planning Tool. The system is built to support:

- **Plan Ownership**: Each plan has an owner (the user who created it)
- **Plan Sharing**: Owners can invite other users to view or edit their plans
- **Access Control**: Three access levels - Owner, Editor, Viewer
- **In-Memory Storage**: Currently all data is kept in memory for development

## Architecture

### Domain Models

#### User
Located in `internal/domain/user.go`, the `User` struct represents a user in the system:
- **ID**: UUID identifier
- **Email**: User's email address (unique, lowercase)

```go
user, err := domain.NewUser("alice@example.com")
```

#### AccessLevel
Represents what a user can do with a plan:
- **Owner**: Full control (create, read, update, share)
- **Editor**: Read and update (can modify plan data)
- **Viewer**: Read-only (can only view plan data)

```go
const (
    Owner    AccessLevel = "owner"
    Editor   AccessLevel = "editor"
    Viewer   AccessLevel = "viewer"
)
```

#### PlanAccess
Represents the relationship between a user and a plan:
- **PlanID**: The plan being accessed
- **UserID**: The user gaining access
- **AccessLevel**: The permission level
- **InvitedAt**: Timestamp of when access was granted

```go
access, err := domain.NewPlanAccess(planID, userID, domain.Editor)
```

### Store Interface

The `PlanStore` interface in `internal/store/memory.go` defines all persistence operations:

#### Plan Operations
- `Save(p *domain.Plan)` - Persist a plan
- `Get(id uuid.UUID)` - Retrieve a plan by ID
- `GetAll()` - Retrieve all plans (admin use only)

#### User Operations
- `SaveUser(u *domain.User)` - Persist a user
- `GetUser(id uuid.UUID)` - Retrieve user by ID
- `GetUserByEmail(email string)` - Retrieve user by email

#### Access Control Operations
- `GrantAccess(planID, userID uuid.UUID, level AccessLevel)` - Give a user access to a plan
- `GetAccess(planID, userID uuid.UUID)` - Check user's access level for a plan
- `GetPlanAccess(planID uuid.UUID)` - List all users with access to a plan
- `GetUserPlans(userID uuid.UUID)` - List all plans a user has access to

### MemoryStore Implementation

The `MemoryStore` in `internal/store/memory.go` implements the `PlanStore` interface:

- `plans map[uuid.UUID]*domain.Plan` - Stores plans by ID
- `users map[uuid.UUID]*domain.User` - Stores users by ID
- `emails map[string]uuid.UUID` - Fast lookup of users by email
- `access map[string]*domain.PlanAccess` - Stores access records using key format "planID:userID"

All operations are protected with `sync.RWMutex` for thread safety.

## Authentication Flow

### Current Implementation (Development Mode)

For development, the system automatically creates a demo user:

1. **AuthMiddleware** runs on every request
2. Checks if user is in context (normally from session/JWT)
3. If not, creates/retrieves demo user with email `demo@example.com`
4. Attaches user to request context

```go
// In handlers/auth.go
func (app *App) AuthMiddleware(next http.Handler) http.Handler {
    // Creates demo@example.com user if doesn't exist
    // Attaches user to request context
}
```

### Production Implementation (Future)

When moving to production:

1. Implement session management (cookies/JWTs)
2. Add login/signup endpoints
3. Replace demo user creation with session verification
4. Store sessions in database

## Authorization Flow

### Access Check Pattern

All handlers that access a specific plan should check authorization:

```go
user := GetUserFromContext(r)
access, err := app.Store.GetAccess(planID, user.ID())
if err != nil {
    // User doesn't have access
    return 403 Forbidden
}

if !access.CanEdit() {
    // User doesn't have edit permission
    return 403 Forbidden
}
```

### Middleware Pattern (Future)

For complex scenarios, use `RequireAccess` middleware:

```go
mux.HandleFunc("POST /plan/{id}/update", 
    app.RequireAccess(domain.Editor)(
        app.PostUpdatePlan(),
    ),
)
```

## Plan Lifecycle

### Creating a Plan

When a user creates a plan:

1. `PostSetup()` creates a new plan with the user as owner
2. Plan is saved to store
3. Owner access is automatically granted

```go
func (app *App) PostSetup() http.HandlerFunc {
    user := GetUserFromContext(r)  // Get current user
    plan, _ := domain.NewPlan(newID, name, month, year, user.ID())
    app.Store.Save(plan)
    app.Store.GrantAccess(planID, user.ID(), domain.Owner)
}
```

### Accessing a Plan

When a user views/edits a plan:

1. Check if plan exists
2. Verify user has access
3. Verify access level matches operation
4. Return data or 403 Forbidden

## Future Enhancements

### Database Migration

When moving to a database:

1. Create `users` table with email (unique)
2. Create `plans` table with owner_id foreign key
3. Create `plan_access` table for sharing
4. Implement database MemoryStore interface
5. Add database migrations

### Multi-Tenancy

Support for organizations/teams:

1. Add `Organization` model
2. Add `plan_organizations` junction table
3. Add role-based access control (RBAC)
4. Add audit logging

### Invitation Flow

Email-based plan sharing:

1. Owner invites user by email
2. System sends invitation link
3. User accepts invitation (creates account if needed)
4. Access is automatically granted

## Development Notes

### Testing Auth

In tests, manually create users and grant access:

```go
user, _ := domain.NewUser("test@example.com")
store.SaveUser(user)

plan, _ := domain.NewPlan(uuid.New(), "Test", 1, 2024, user.ID())
store.Save(plan)

store.GrantAccess(plan.ID(), user.ID(), domain.Owner)
```

### Debugging

- All access checks log denials
- User context can be verified with `GetUserFromContext(r)`
- Email lookups are case-insensitive (normalized to lowercase)

## API Reference

### Getting Current User

```go
user := handlers.GetUserFromContext(r)
if user == nil {
    // User is not authenticated
}
```

### Checking Access

```go
access, err := app.Store.GetAccess(planID, userID)
if err != nil {
    // No access
} else if access.CanEdit() {
    // Can edit
}
```

### Listing User's Plans

```go
plans, _ := app.Store.GetUserPlans(userID)
// Returns all plans user has any access to
```

### Sharing a Plan

```go
err := app.Store.GrantAccess(planID, newUserID, domain.Editor)
// Grants editor access; creates user first if needed
```
