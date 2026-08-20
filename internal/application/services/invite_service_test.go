package services_test

import (
	"database/sql"
	"testing"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
)

func TestInviteService_CreateInvite(t *testing.T) {
	plans, users, sessions, dbPath, cleanup := newTestStore(t)
	defer cleanup()
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()
	inviteRepo := sqlite.NewInviteRepository(conn)

	planSvc := services.NewPlanService(plans, users)
	inviteSvc := services.NewInviteService(inviteRepo, plans)

	ownerID := newPersistedOwner(t, users, sessions)
	planResult, err := planSvc.CreatePlan(&commands.CreatePlan{
		Name:          "Invite Co",
		StartingMonth: 1,
		StartingYear:  2026,
		OwnerID:       ownerID,
	})
	if err != nil {
		t.Fatalf("CreatePlan: %v", err)
	}
	planID := planResult.Result.ID

	result, err := inviteSvc.CreateInvite(&commands.CreateInvite{
		PlanID:      planID,
		Email:       "collaborator@example.com",
		AccessLevel: domain.Editor,
		InvitedBy:   ownerID,
	})
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if result.Result.Email != "collaborator@example.com" {
		t.Errorf("email mismatch: want %q got %q", "collaborator@example.com", result.Result.Email)
	}
	if result.Result.PlanID != planID {
		t.Errorf("planID mismatch: want %v got %v", planID, result.Result.PlanID)
	}
	if result.Result.Status != domain.InvitePending {
		t.Errorf("expected pending status, got %v", result.Result.Status)
	}

	// The invite row and the Plan's UserInvitedToPlan outbox event must be
	// written atomically via PlanRepository.SaveWithInvite - assert both
	// exist, since these used to be two independent, non-atomic writes.
	stored, err := inviteRepo.GetInvite(result.Result.ID)
	if err != nil {
		t.Fatalf("expected invite row to exist, got error: %v", err)
	}
	if stored.Email != "collaborator@example.com" {
		t.Errorf("stored invite email mismatch: got %q", stored.Email)
	}

	var eventName string
	err = conn.QueryRow(`SELECT event_name FROM outbox_events WHERE event_name = 'plan.user_invited' ORDER BY occurred_at DESC LIMIT 1`).Scan(&eventName)
	if err != nil {
		t.Fatalf("expected a plan.user_invited outbox event, got none: %v", err)
	}
}

func TestInviteService_CreateInvite_RejectsInvalidEmail(t *testing.T) {
	plans, users, sessions, dbPath, cleanup := newTestStore(t)
	defer cleanup()
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()
	inviteRepo := sqlite.NewInviteRepository(conn)

	planSvc := services.NewPlanService(plans, users)
	inviteSvc := services.NewInviteService(inviteRepo, plans)

	ownerID := newPersistedOwner(t, users, sessions)
	planResult, err := planSvc.CreatePlan(&commands.CreatePlan{
		Name:          "Invite Co",
		StartingMonth: 1,
		StartingYear:  2026,
		OwnerID:       ownerID,
	})
	if err != nil {
		t.Fatalf("CreatePlan: %v", err)
	}

	if _, err := inviteSvc.CreateInvite(&commands.CreateInvite{
		PlanID:      planResult.Result.ID,
		Email:       "   ",
		AccessLevel: domain.Editor,
		InvitedBy:   ownerID,
	}); err == nil {
		t.Error("expected error for empty email, got nil")
	}
}

func TestInviteService_CreateInvite_RejectsUnknownPlan(t *testing.T) {
	plans, _, _, dbPath, cleanup := newTestStore(t)
	defer cleanup()
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()
	inviteRepo := sqlite.NewInviteRepository(conn)
	inviteSvc := services.NewInviteService(inviteRepo, plans)

	unknownPlanID := uuid.NewV7()
	invitedBy := uuid.NewV7()

	if _, err := inviteSvc.CreateInvite(&commands.CreateInvite{
		PlanID:      unknownPlanID,
		Email:       "collaborator@example.com",
		AccessLevel: domain.Editor,
		InvitedBy:   invitedBy,
	}); err == nil {
		t.Error("expected error for unknown plan, got nil")
	}
}
