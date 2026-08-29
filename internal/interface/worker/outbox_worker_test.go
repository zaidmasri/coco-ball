package worker_test

import (
	"errors"
	"testing"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/interface/worker"
)

// fakeOutboxRepository is an in-memory stand-in for
// repositories.OutboxRepository, so the worker's dispatch logic can be
// tested without a real database.
type fakeOutboxRepository struct {
	events    []repositories.OutboxEvent
	published map[string]bool
}

func newFakeOutboxRepository(events []repositories.OutboxEvent) *fakeOutboxRepository {
	return &fakeOutboxRepository{events: events, published: map[string]bool{}}
}

func (f *fakeOutboxRepository) GetUnpublished(limit int) ([]repositories.OutboxEvent, error) {
	var out []repositories.OutboxEvent
	for _, e := range f.events {
		if !f.published[e.ID] {
			out = append(out, e)
		}
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func (f *fakeOutboxRepository) MarkPublished(id string) error {
	f.published[id] = true
	return nil
}

// fakeNotificationService records every call instead of sending real email.
type fakeNotificationService struct {
	welcomeCalls []uuid.UUID
	inviteCalls  []uuid.UUID
	failUserID   uuid.UUID
}

func (f *fakeNotificationService) SendWelcomeEmail(userID uuid.UUID) error {
	if userID == f.failUserID {
		return errors.New("simulated failure")
	}
	f.welcomeCalls = append(f.welcomeCalls, userID)
	return nil
}

func (f *fakeNotificationService) SendInviteEmail(inviteID uuid.UUID) error {
	f.inviteCalls = append(f.inviteCalls, inviteID)
	return nil
}

func TestOutboxWorker_DispatchesUserRegistered(t *testing.T) {
	userID := uuid.NewV7()
	outbox := newFakeOutboxRepository([]repositories.OutboxEvent{
		{ID: "1", AggregateID: userID.String(), EventName: "user.registered", Payload: `{"Email":"a@example.com"}`},
	})
	notifications := &fakeNotificationService{}

	w := worker.NewOutboxWorker(outbox, notifications, 0)
	w.ProcessOnce()

	if len(notifications.welcomeCalls) != 1 || notifications.welcomeCalls[0] != userID {
		t.Fatalf("expected SendWelcomeEmail(%s), got %v", userID, notifications.welcomeCalls)
	}
	if !outbox.published["1"] {
		t.Error("expected event 1 to be marked published")
	}
}

func TestOutboxWorker_DispatchesUserInvitedToPlan(t *testing.T) {
	inviteID := uuid.NewV7()
	planID := uuid.NewV7()
	outbox := newFakeOutboxRepository([]repositories.OutboxEvent{
		{
			ID:          "1",
			AggregateID: planID.String(),
			EventName:   "plan.user_invited",
			Payload:     `{"InviteID":"` + inviteID.String() + `","Email":"b@example.com","AccessLevel":"editor"}`,
		},
	})
	notifications := &fakeNotificationService{}

	w := worker.NewOutboxWorker(outbox, notifications, 0)
	w.ProcessOnce()

	if len(notifications.inviteCalls) != 1 || notifications.inviteCalls[0] != inviteID {
		t.Fatalf("expected SendInviteEmail(%s), got %v", inviteID, notifications.inviteCalls)
	}
	if !outbox.published["1"] {
		t.Error("expected event 1 to be marked published")
	}
}

func TestOutboxWorker_UnknownEventIsNoOpButMarkedPublished(t *testing.T) {
	outbox := newFakeOutboxRepository([]repositories.OutboxEvent{
		{ID: "1", AggregateID: uuid.NewV7().String(), EventName: "plan.created", Payload: `{}`},
	})
	notifications := &fakeNotificationService{}

	w := worker.NewOutboxWorker(outbox, notifications, 0)
	w.ProcessOnce()

	if !outbox.published["1"] {
		t.Error("expected unknown event to be marked published (no-op, forward compatible)")
	}
}

func TestOutboxWorker_FailedEventDoesNotBlockLaterEvents(t *testing.T) {
	failingUser := uuid.NewV7()
	okUser := uuid.NewV7()
	outbox := newFakeOutboxRepository([]repositories.OutboxEvent{
		{ID: "1", AggregateID: failingUser.String(), EventName: "user.registered", Payload: `{}`},
		{ID: "2", AggregateID: okUser.String(), EventName: "user.registered", Payload: `{}`},
	})
	notifications := &fakeNotificationService{failUserID: failingUser}

	w := worker.NewOutboxWorker(outbox, notifications, 0)
	w.ProcessOnce()

	if outbox.published["1"] {
		t.Error("expected failing event to remain unpublished so it retries")
	}
	if !outbox.published["2"] {
		t.Error("expected the event after the failing one to still be processed and published")
	}
	if len(notifications.welcomeCalls) != 1 || notifications.welcomeCalls[0] != okUser {
		t.Fatalf("expected only the ok user's welcome email to be sent, got %v", notifications.welcomeCalls)
	}
}
