package services_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
)

// newAuthTestStore creates a temp-file-backed SQLite connection (migrations
// applied) and returns the repositories AuthService/PlanService tests need
// over it, and a cleanup func, mirroring newTestStore in plan_service_test.go.
func newAuthTestStore(t *testing.T) (users repositories.UserRepository, sessions repositories.SessionRepository, plans repositories.PlanRepository, access repositories.AccessRepository, cleanup func()) {
	t.Helper()

	tmpDir := filepath.Join(os.TempDir(), "northbasis_auth_service_test")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	dbPath := filepath.Join(tmpDir, uuid.NewString()+".db")

	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		t.Fatalf("NewConnection: %v", err)
	}
	if err := sqlite.RunMigrations(conn); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	cleanup = func() {
		conn.Close()
		os.Remove(dbPath)
	}
	return sqlite.NewUserRepository(conn), sqlite.NewSessionRepository(conn), sqlite.NewPlanRepository(conn), sqlite.NewAccessRepository(conn), cleanup
}

func TestAuthService_CreateUser(t *testing.T) {
	users, sessions, _, _, cleanup := newAuthTestStore(t)
	defer cleanup()
	svc := services.NewAuthService(users, sessions)

	result, err := svc.CreateUser(&commands.CreateUser{
		Email:     "jane@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
		Password:  "supersecret",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if result.Result.Email != "jane@example.com" || result.Result.FirstName != "Jane" || result.Result.LastName != "Doe" {
		t.Fatalf("unexpected result: %+v", result.Result)
	}

	if _, err := svc.GetUserWithPassword("jane@example.com"); err != nil {
		t.Fatalf("expected user to be persisted, GetUserWithPassword: %v", err)
	}
}

func TestAuthService_CreateUser_RejectsWeakPassword(t *testing.T) {
	users, sessions, _, _, cleanup := newAuthTestStore(t)
	defer cleanup()
	svc := services.NewAuthService(users, sessions)

	_, err := svc.CreateUser(&commands.CreateUser{
		Email:     "jane@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
		Password:  "short",
	})
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestAuthService_UpdateUser(t *testing.T) {
	users, sessions, _, _, cleanup := newAuthTestStore(t)
	defer cleanup()
	svc := services.NewAuthService(users, sessions)

	created, err := svc.CreateUser(&commands.CreateUser{
		Email: "jane@example.com", FirstName: "Jane", LastName: "Doe", Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	result, err := svc.UpdateUser(&commands.UpdateUser{
		UserID: created.Result.ID, FirstName: "Janet", LastName: "Smith",
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if result.Result.FirstName != "Janet" || result.Result.LastName != "Smith" {
		t.Fatalf("unexpected result: %+v", result.Result)
	}

	fetched, err := svc.GetUser(created.Result.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if fetched.FirstName() != "Janet" || fetched.LastName() != "Smith" {
		t.Fatalf("expected persisted name Janet Smith, got %s %s", fetched.FirstName(), fetched.LastName())
	}
}

func TestAuthService_UpdateUser_RejectsEmptyName(t *testing.T) {
	users, sessions, _, _, cleanup := newAuthTestStore(t)
	defer cleanup()
	svc := services.NewAuthService(users, sessions)

	created, err := svc.CreateUser(&commands.CreateUser{
		Email: "jane@example.com", FirstName: "Jane", LastName: "Doe", Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if _, err := svc.UpdateUser(&commands.UpdateUser{UserID: created.Result.ID, FirstName: "", LastName: "Smith"}); !errors.Is(err, domain.ErrInvalidUserName) {
		t.Fatalf("expected ErrInvalidUserName, got %v", err)
	}
}

func TestAuthService_DeleteUser(t *testing.T) {
	users, sessions, _, _, cleanup := newAuthTestStore(t)
	defer cleanup()
	svc := services.NewAuthService(users, sessions)

	created, err := svc.CreateUser(&commands.CreateUser{
		Email: "jane@example.com", FirstName: "Jane", LastName: "Doe", Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	result, err := svc.DeleteUser(&commands.DeleteUser{UserID: created.Result.ID})
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected Success=true")
	}

	if _, err := svc.GetUser(created.Result.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound after delete, got %v", err)
	}
	if _, err := svc.GetUserWithPassword("jane@example.com"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected credentials gone after delete, got %v", err)
	}
}

// TestAuthService_CreateUser_ReusesEmailAfterDelete confirms that once an
// account is deleted, a new account can be created with the same email -
// the users table's uniqueness is scoped to active (non-deleted) rows via a
// partial index, not a blanket UNIQUE column constraint, precisely so a
// soft-deleted row doesn't permanently squat on its email.
func TestAuthService_CreateUser_ReusesEmailAfterDelete(t *testing.T) {
	users, sessions, _, _, cleanup := newAuthTestStore(t)
	defer cleanup()
	svc := services.NewAuthService(users, sessions)

	first, err := svc.CreateUser(&commands.CreateUser{
		Email: "jane@example.com", FirstName: "Jane", LastName: "Doe", Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("CreateUser (first): %v", err)
	}
	if _, err := svc.DeleteUser(&commands.DeleteUser{UserID: first.Result.ID}); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	second, err := svc.CreateUser(&commands.CreateUser{
		Email: "jane@example.com", FirstName: "Jane", LastName: "Reused", Password: "othersecret",
	})
	if err != nil {
		t.Fatalf("CreateUser (second): %v", err)
	}
	if second.Result.ID == first.Result.ID {
		t.Fatalf("expected a new user ID, got the deleted one back")
	}

	fetched, err := svc.GetUser(second.Result.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if fetched.LastName() != "Reused" {
		t.Fatalf("expected the second account's data, got %+v", fetched)
	}

	if _, err := svc.GetUserWithPassword("jane@example.com"); err != nil {
		t.Fatalf("expected GetUserWithPassword to find the new account: %v", err)
	}
}

// TestAuthService_DeleteUser_OwningPlanIsAllowed confirms the deliberate
// product decision (no ownership guard): deleting a user who owns a plan
// succeeds, removing their plan_access row but leaving the plan itself
// intact - see AGENTS.md's "Application Layer: go-ddd Comparison" section.
func TestAuthService_DeleteUser_OwningPlanIsAllowed(t *testing.T) {
	users, sessions, plans, access, cleanup := newAuthTestStore(t)
	defer cleanup()
	svc := services.NewAuthService(users, sessions)

	created, err := svc.CreateUser(&commands.CreateUser{
		Email: "owner@example.com", FirstName: "Jane", LastName: "Doe", Password: "supersecret",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	user, err := users.GetUser(created.Result.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	verified, err := domain.NewVerifiedUser(user)
	if err != nil {
		t.Fatalf("NewVerifiedUser: %v", err)
	}
	plan, err := domain.NewPlan("Acme Corp", 1, 2026, verified)
	if err != nil {
		t.Fatalf("NewPlan: %v", err)
	}
	validated, err := plan.Validate()
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if err := plans.Save(validated); err != nil {
		t.Fatalf("Save plan: %v", err)
	}
	if err := access.GrantAccess(plan.ID(), user.ID(), domain.Owner); err != nil {
		t.Fatalf("GrantAccess: %v", err)
	}

	if _, err := svc.DeleteUser(&commands.DeleteUser{UserID: created.Result.ID}); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	if _, err := plans.Get(plan.ID()); err != nil {
		t.Fatalf("expected plan to survive user deletion, Get: %v", err)
	}
	if _, err := access.GetAccess(plan.ID(), user.ID()); !errors.Is(err, domain.ErrAccessDenied) {
		t.Fatalf("expected plan_access row to be gone, got %v", err)
	}
}
