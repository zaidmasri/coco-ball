package sqlite

import (
	"context"
	"fmt"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// insertInviteRow writes a PlanInvite's row using queries already bound to
// an open transaction (qtx) - callers own the transaction's begin/commit.
// Shared by PlanRepository.SaveWithInvite, the only caller: PlanInvite is an
// entity within the Plan aggregate boundary (see invite.go's doc comment),
// so its row is written in the same transaction as the Plan aggregate's own
// save and outbox event, not via a separate, non-atomic repository call.
func insertInviteRow(ctx context.Context, q *db.Queries, vi domain.ValidatedPlanInvite, now int64) error {
	invite := vi.Invite()
	if err := q.CreateInvite(ctx, db.CreateInviteParams{
		ID:          invite.ID.String(),
		PlanID:      invite.PlanID.String(),
		Email:       invite.Email,
		AccessLevel: string(invite.AccessLevel),
		Status:      string(invite.Status),
		InvitedBy:   invite.InvitedBy.String(),
		CreatedAt:   now,
	}); err != nil {
		return fmt.Errorf("failed to create invite: %w", err)
	}
	invite.CreatedAt = now
	return nil
}
