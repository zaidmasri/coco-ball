package interfaces

import "uuid"

// NotificationService sends transactional emails triggered by domain
// events. Its only caller is the outbox worker (internal/interface/worker),
// which resolves an event to a concrete ID before calling in — this
// interface never sees raw event payloads or JSON.
type NotificationService interface {
	// SendWelcomeEmail sends a welcome email to a newly registered user.
	SendWelcomeEmail(userID uuid.UUID) error

	// SendInviteEmail sends a collaboration-invite email for a pending
	// PlanInvite.
	SendInviteEmail(inviteID uuid.UUID) error
}
