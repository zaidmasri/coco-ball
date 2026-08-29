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

// capturingMailer implements ports.Mailer and records every send, so tests
// can assert on the recipient/subject/body without a real mail server.
type capturingMailer struct {
	sent []sentEmail
}

type sentEmail struct {
	to, subject, body string
}

func (m *capturingMailer) Send(to, subject, body string) error {
	m.sent = append(m.sent, sentEmail{to: to, subject: subject, body: body})
	return nil
}

func TestNotificationService_SendWelcomeEmail(t *testing.T) {
	_, users, sessions, dbPath, cleanup := newTestStore(t)
	defer cleanup()
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()

	planRepo := sqlite.NewPlanRepository(conn)
	inviteRepo := sqlite.NewInviteRepository(conn)

	userID := newPersistedOwner(t, users, sessions)

	mailer := &capturingMailer{}
	notifications := services.NewNotificationService(mailer, users, planRepo, inviteRepo, "https://app.example.com")

	if err := notifications.SendWelcomeEmail(userID); err != nil {
		t.Fatalf("SendWelcomeEmail: %v", err)
	}

	if len(mailer.sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(mailer.sent))
	}
	got := mailer.sent[0]
	if got.to != "owner@example.com" {
		t.Errorf("to mismatch: want %q got %q", "owner@example.com", got.to)
	}
	if got.subject != "Welcome to Northbasis" {
		t.Errorf("subject mismatch: got %q", got.subject)
	}
}

func TestNotificationService_SendWelcomeEmail_UnknownUser(t *testing.T) {
	_, users, _, dbPath, cleanup := newTestStore(t)
	defer cleanup()
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer conn.Close()

	planRepo := sqlite.NewPlanRepository(conn)
	inviteRepo := sqlite.NewInviteRepository(conn)

	mailer := &capturingMailer{}
	notifications := services.NewNotificationService(mailer, users, planRepo, inviteRepo, "https://app.example.com")

	if err := notifications.SendWelcomeEmail(uuid.NewV7()); err == nil {
		t.Error("expected error for unknown user, got nil")
	}
	if len(mailer.sent) != 0 {
		t.Errorf("expected no email sent for unknown user, got %d", len(mailer.sent))
	}
}

func TestNotificationService_SendInviteEmail(t *testing.T) {
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

	inviteResult, err := inviteSvc.CreateInvite(&commands.CreateInvite{
		PlanID:      planResult.Result.ID,
		Email:       "collaborator@example.com",
		AccessLevel: domain.Editor,
		InvitedBy:   ownerID,
	})
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}

	mailer := &capturingMailer{}
	notifications := services.NewNotificationService(mailer, users, plans, inviteRepo, "https://app.example.com")

	if err := notifications.SendInviteEmail(inviteResult.Result.ID); err != nil {
		t.Fatalf("SendInviteEmail: %v", err)
	}

	if len(mailer.sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(mailer.sent))
	}
	got := mailer.sent[0]
	if got.to != "collaborator@example.com" {
		t.Errorf("to mismatch: want %q got %q", "collaborator@example.com", got.to)
	}
	if got.subject != "Owner invited you to collaborate on Invite Co" {
		t.Errorf("subject mismatch: got %q", got.subject)
	}
}
