// Package worker holds driving adapters that trigger the application layer
// on a poll loop instead of an HTTP request — the background counterpart to
// internal/interface/web's controllers.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domevents "github.com/zaidmasri/business-planning-tool/internal/domain/events"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

const batchSize = 100

// OutboxWorker polls the transactional outbox for unpublished domain events
// and dispatches each one to the application layer. It provides
// at-least-once delivery: an event is only marked published after its
// handler succeeds, so a crash (or a failed send) between handling and
// marking leaves the row to retry on the next poll.
type OutboxWorker struct {
	outbox        repositories.OutboxRepository
	notifications interfaces.NotificationService
	interval      time.Duration
}

func NewOutboxWorker(outbox repositories.OutboxRepository, notifications interfaces.NotificationService, interval time.Duration) *OutboxWorker {
	return &OutboxWorker{outbox: outbox, notifications: notifications, interval: interval}
}

// Run blocks, polling on the configured interval until ctx is canceled.
func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.ProcessOnce()
		}
	}
}

// ProcessOnce dispatches each currently-unpublished event in turn. Exported
// so tests can drive a single batch synchronously instead of waiting on the
// ticker. A single event's failure is logged and skipped rather than
// aborting the batch, so one permanently-failing event doesn't block every
// event behind it.
func (w *OutboxWorker) ProcessOnce() {
	events, err := w.outbox.GetUnpublished(batchSize)
	if err != nil {
		log.Printf("worker: failed to fetch unpublished events: %v", err)
		return
	}

	for _, evt := range events {
		if err := w.dispatch(evt); err != nil {
			log.Printf("worker: failed to process event %s (%s): %v", evt.ID, evt.EventName, err)
			continue
		}
		if err := w.outbox.MarkPublished(evt.ID); err != nil {
			log.Printf("worker: failed to mark event %s published: %v", evt.ID, err)
		}
	}
}

// dispatch resolves one outbox row to a NotificationService call. Unknown
// event names are a no-op (and get marked published) so the outbox stays
// forward-compatible with event types this worker doesn't handle yet.
func (w *OutboxWorker) dispatch(evt repositories.OutboxEvent) error {
	switch evt.EventName {
	case domevents.UserRegisteredEventName:
		userID, err := uuid.Parse(evt.AggregateID)
		if err != nil {
			return fmt.Errorf("parse aggregate id: %w", err)
		}
		return w.notifications.SendWelcomeEmail(userID)

	case domevents.UserInvitedToPlanEventName:
		var payload struct {
			InviteID uuid.UUID
		}
		if err := json.Unmarshal([]byte(evt.Payload), &payload); err != nil {
			return fmt.Errorf("unmarshal %s payload: %w", evt.EventName, err)
		}
		return w.notifications.SendInviteEmail(payload.InviteID)

	default:
		return nil
	}
}
