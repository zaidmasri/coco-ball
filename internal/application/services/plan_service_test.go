package services_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
)

// newTestStore creates a temp-file-backed SQLite connection (migrations
// applied) and returns a PlanRepository and UserRepository over it, its db
// path, and a cleanup func, mirroring the pattern in
// internal/interface/web/auth_test.go.
func newTestStore(t *testing.T) (plans repositories.PlanRepository, users repositories.UserRepository, sessions repositories.SessionRepository, dbPath string, cleanup func()) {
	t.Helper()

	tmpDir := filepath.Join(os.TempDir(), "northbasis_plan_service_test")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	dbPath = filepath.Join(tmpDir, uuid.New().String()+".db")

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
	return sqlite.NewPlanRepository(conn), sqlite.NewUserRepository(conn), sqlite.NewSessionRepository(conn), dbPath, cleanup
}

// newPersistedOwner creates a real User row via AuthService.CreateUser — the
// only validated entry point for constructing a User — and returns its ID,
// for tests that exercise PlanService.CreatePlan, which verifies OwnerID
// against UserRepository.GetUser, so a random unsaved uuid.UUID won't do.
func newPersistedOwner(t *testing.T, users repositories.UserRepository, sessions repositories.SessionRepository) uuid.UUID {
	t.Helper()
	authSvc := services.NewAuthService(users, sessions)
	result, err := authSvc.CreateUser(&commands.CreateUser{
		Email:     "owner@example.com",
		FirstName: "Owner",
		LastName:  "Example",
		Password:  "supersecret1",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return result.Result.ID
}

// newVerifiedOwner builds a VerifiedUser for tests that call domain.NewPlan
// directly (bypassing PlanService.CreatePlan's verification). NewPlan only
// requires proof of verification, not an actual database row.
func newVerifiedOwner(t *testing.T) domain.VerifiedUser {
	t.Helper()
	user, err := domain.NewUser("direct-owner@example.com")
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	verified, err := domain.NewVerifiedUser(user)
	if err != nil {
		t.Fatalf("NewVerifiedUser: %v", err)
	}
	return verified
}

func TestPlanService_SaveAndGet(t *testing.T) {
	plans, _, _, _, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(plans, nil)

	plan, err := domain.NewPlan("Acme Corp", 3, 2025, newVerifiedOwner(t))
	if err != nil {
		t.Fatalf("NewPlan: %v", err)
	}

	if err := svc.Save(plan); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := svc.Get(plan.ID())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name() != plan.Name() {
		t.Errorf("name mismatch: want %q got %q", plan.Name(), got.Name())
	}
}

func TestPlanService_Save_EmitsCreatedEvent(t *testing.T) {
	plans, _, _, dbPath, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(plans, nil)

	plan, _ := domain.NewPlan("Event Co", 1, 2024, newVerifiedOwner(t))
	if err := svc.Save(plan); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// SQLiteStore.Save writes domain events into the outbox_events table
	// within the same transaction (see writeOutboxEvents in sqlite.go).
	// Query it directly rather than draining an in-memory buffer.
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var eventName string
	err = db.QueryRow(`SELECT event_name FROM outbox_events ORDER BY occurred_at LIMIT 1`).Scan(&eventName)
	if err != nil {
		t.Fatalf("expected at least one outbox event after Save, got none: %v", err)
	}
	if eventName != "plan.created" {
		t.Errorf("expected plan.created event, got %q", eventName)
	}
}

func TestPlanService_Delete(t *testing.T) {
	plans, _, _, dbPath, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(plans, nil)

	plan, _ := domain.NewPlan("Delete Me", 6, 2024, newVerifiedOwner(t))
	_ = svc.Save(plan)

	if err := svc.Delete(plan.ID()); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := svc.Get(plan.ID()); err == nil {
		t.Error("expected error after delete, got nil")
	}

	// PlanRepository.Delete loads the plan, calls Plan.Delete() to record a
	// PlanDeleted event, and drains it to the outbox in the same
	// transaction as the soft-delete - assert the event actually landed.
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var eventName string
	err = db.QueryRow(`SELECT event_name FROM outbox_events WHERE event_name = 'plan.deleted' ORDER BY occurred_at DESC LIMIT 1`).Scan(&eventName)
	if err != nil {
		t.Fatalf("expected a plan.deleted outbox event, got none: %v", err)
	}
}

func TestPlanService_Get_NotFound(t *testing.T) {
	plans, _, _, _, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(plans, nil)

	missing := uuid.NewV7()
	if _, err := svc.Get(missing); err == nil {
		t.Error("expected error for missing plan, got nil")
	}
}

func TestPlanService_CreatePlan(t *testing.T) {
	plans, users, sessions, dbPath, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(plans, users)

	ownerID := newPersistedOwner(t, users, sessions)
	result, err := svc.CreatePlan(&commands.CreatePlan{
		Name:          "Command Co",
		StartingMonth: 4,
		StartingYear:  2026,
		OwnerID:       ownerID,
	})
	if err != nil {
		t.Fatalf("CreatePlan: %v", err)
	}
	if result.Result.Name != "Command Co" {
		t.Errorf("name mismatch: want %q got %q", "Command Co", result.Result.Name)
	}
	if result.Result.OwnerID != ownerID {
		t.Errorf("owner mismatch: want %v got %v", ownerID, result.Result.OwnerID)
	}

	got, err := svc.Get(result.Result.ID)
	if err != nil {
		t.Fatalf("Get after CreatePlan: %v", err)
	}
	if got.Name() != "Command Co" {
		t.Errorf("persisted name mismatch: want %q got %q", "Command Co", got.Name())
	}

	// CreatePlan (via PlanRepository.SaveNew) must grant the owner Owner
	// access atomically with the plan row itself - no separate GrantAccess
	// call, no window where the plan exists without an owner.
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open verification connection: %v", err)
	}
	defer conn.Close()
	accessRepo := sqlite.NewAccessRepository(conn)
	access, err := accessRepo.GetAccess(result.Result.ID, ownerID)
	if err != nil {
		t.Fatalf("expected plan_access row for owner, got error: %v", err)
	}
	if access.AccessLevel != domain.Owner {
		t.Errorf("expected Owner access level, got %v", access.AccessLevel)
	}
}

func TestPlanService_CreatePlan_RejectsInvalid(t *testing.T) {
	plans, users, sessions, _, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(plans, users)

	ownerID := newPersistedOwner(t, users, sessions)
	if _, err := svc.CreatePlan(&commands.CreatePlan{
		Name:          "",
		StartingMonth: 4,
		StartingYear:  2026,
		OwnerID:       ownerID,
	}); err == nil {
		t.Error("expected error for empty name, got nil")
	}
}

// TestPlanService_CreatePlan_RejectsUnknownOwner is the key coverage for the
// VerifiedUser guarantee: CreatePlan must refuse to assign an owner that
// doesn't correspond to any real User row, even though the uuid.UUID itself
// is well-formed.
func TestPlanService_CreatePlan_RejectsUnknownOwner(t *testing.T) {
	plans, users, _, _, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(plans, users)

	unknownOwnerID := uuid.NewV7()
	if _, err := svc.CreatePlan(&commands.CreatePlan{
		Name:          "Ghost Owner Co",
		StartingMonth: 4,
		StartingYear:  2026,
		OwnerID:       unknownOwnerID,
	}); err == nil {
		t.Error("expected error for unknown owner, got nil")
	}
}

func TestPlanService_UpdatePlan(t *testing.T) {
	plans, users, _, _, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(plans, users)

	plan, _ := domain.NewPlan("Original Name", 1, 2024, newVerifiedOwner(t))
	if err := svc.Save(plan); err != nil {
		t.Fatalf("Save: %v", err)
	}

	result, err := svc.UpdatePlan(&commands.UpdatePlan{
		PlanID:        plan.ID(),
		Name:          "Renamed Co",
		StartingMonth: 7,
		StartingYear:  2025,
	})
	if err != nil {
		t.Fatalf("UpdatePlan: %v", err)
	}
	if result.Result.Name != "Renamed Co" {
		t.Errorf("name mismatch: want %q got %q", "Renamed Co", result.Result.Name)
	}

	got, err := svc.Get(plan.ID())
	if err != nil {
		t.Fatalf("Get after UpdatePlan: %v", err)
	}
	if got.Name() != "Renamed Co" || got.StartingMonth() != 7 || got.StartingYear() != 2025 {
		t.Errorf("persisted update mismatch: got name=%q month=%d year=%d", got.Name(), got.StartingMonth(), got.StartingYear())
	}
}

func TestPlanService_DeletePlan(t *testing.T) {
	plans, users, _, _, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(plans, users)

	plan, _ := domain.NewPlan("Delete Via Command", 6, 2024, newVerifiedOwner(t))
	if err := svc.Save(plan); err != nil {
		t.Fatalf("Save: %v", err)
	}

	result, err := svc.DeletePlan(&commands.DeletePlan{PlanID: plan.ID()})
	if err != nil {
		t.Fatalf("DeletePlan: %v", err)
	}
	if !result.Success {
		t.Error("expected Success=true")
	}

	if _, err := svc.Get(plan.ID()); err == nil {
		t.Error("expected error after DeletePlan, got nil")
	}
}
