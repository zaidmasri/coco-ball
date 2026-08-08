package services_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
)

// newTestStore creates a temp-file-backed SQLite connection (migrations
// applied) and returns a PlanRepository over it, its db path, and a cleanup
// func, mirroring the pattern in internal/handlers/auth_test.go.
func newTestStore(t *testing.T) (repo repositories.PlanRepository, dbPath string, cleanup func()) {
	t.Helper()

	tmpDir := filepath.Join(os.TempDir(), "northbasis_plan_service_test")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	dbPath = filepath.Join(tmpDir, uuid.NewString()+".db")

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
	return sqlite.NewPlanRepository(conn), dbPath, cleanup
}

func TestPlanService_SaveAndGet(t *testing.T) {
	repo, _, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(repo)

	ownerID, _ := uuid.NewV7()
	plan, err := domain.NewPlan("Acme Corp", 3, 2025, ownerID)
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
	repo, dbPath, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(repo)

	ownerID, _ := uuid.NewV7()
	plan, _ := domain.NewPlan("Event Co", 1, 2024, ownerID)
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
	err = db.QueryRow(`SELECT event_name FROM outbox_events ORDER BY created_at LIMIT 1`).Scan(&eventName)
	if err != nil {
		t.Fatalf("expected at least one outbox event after Save, got none: %v", err)
	}
	if eventName != "plan.created" {
		t.Errorf("expected plan.created event, got %q", eventName)
	}
}

func TestPlanService_Delete(t *testing.T) {
	repo, _, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(repo)

	ownerID, _ := uuid.NewV7()
	plan, _ := domain.NewPlan("Delete Me", 6, 2024, ownerID)
	_ = svc.Save(plan)

	if err := svc.Delete(plan.ID()); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := svc.Get(plan.ID()); err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestPlanService_Get_NotFound(t *testing.T) {
	repo, _, cleanup := newTestStore(t)
	defer cleanup()
	svc := services.NewPlanService(repo)

	missing, _ := uuid.NewV7()
	if _, err := svc.Get(missing); err == nil {
		t.Error("expected error for missing plan, got nil")
	}
}
