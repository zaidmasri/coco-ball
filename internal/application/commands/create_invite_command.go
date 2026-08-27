package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateInvite is the command to invite a collaborator to a plan.
// InviteService.CreateInvite loads the Plan, turns the rest into a
// validated domain.PlanInvite via domain.NewPlanInvite, and persists the
// invite row and the Plan's UserInvitedToPlan outbox event atomically —
// callers must not construct the entity themselves.
type CreateInvite struct {
	PlanID uuid.UUID
	Email  string
	// InviterEmail is the currently-authenticated user's own email, passed
	// through from the handler's session context so CreateInvite can reject
	// a self-invite without the service needing a UserRepository dependency
	// just for this one check.
	InviterEmail string
	AccessLevel  domain.AccessLevel
	InvitedBy    uuid.UUID
}

// CreateInviteResult wraps the created invite's Result.
type CreateInviteResult struct {
	Result *common.InviteResult
}
