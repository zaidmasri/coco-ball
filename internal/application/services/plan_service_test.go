package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/fakes"
	"github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

func TestPlanService_SaveAndGet(t *testing.T) {
	repo := fakes.NewPlanRepository()
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
	repo := fakes.NewPlanRepository()
	svc := services.NewPlanService(repo)

	ownerID, _ := uuid.NewV7()
	plan, _ := domain.NewPlan("Event Co", 1, 2024, ownerID)
	if err := svc.Save(plan); err != nil {
		t.Fatalf("Save: %v", err)
	}

	evts := repo.DrainEvents()
	if len(evts) == 0 {
		t.Fatal("expected at least one domain event after Save, got none")
	}
	if evts[0].EventName() != "plan.created" {
		t.Errorf("expected plan.created event, got %q", evts[0].EventName())
	}
}

func TestPlanService_Delete(t *testing.T) {
	repo := fakes.NewPlanRepository()
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
	repo := fakes.NewPlanRepository()
	svc := services.NewPlanService(repo)

	missing, _ := uuid.NewV7()
	if _, err := svc.Get(missing); err == nil {
		t.Error("expected error for missing plan, got nil")
	}
}
