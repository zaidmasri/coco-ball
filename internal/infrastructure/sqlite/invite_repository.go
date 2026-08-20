package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// InviteRepository implements repositories.InviteRepository.
type InviteRepository struct {
	queries *db.Queries
}

func NewInviteRepository(conn *sql.DB) repositories.InviteRepository {
	return &InviteRepository{queries: db.New(conn)}
}

var _ repositories.InviteRepository = (*InviteRepository)(nil)

func (r *InviteRepository) GetInvite(id uuid.UUID) (*domain.PlanInvite, error) {
	row, err := r.queries.GetInvite(context.Background(), id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInviteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite: %w", err)
	}

	return inviteFromRow(id, row.PlanID, row.Email, row.AccessLevel, row.Status, row.InvitedBy, row.CreatedAt, row.RespondedAt)
}

func (r *InviteRepository) GetInvitesForPlan(planID uuid.UUID) ([]*domain.PlanInvite, error) {
	rows, err := r.queries.GetInvitesForPlan(context.Background(), planID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get invites for plan: %w", err)
	}

	invites := make([]*domain.PlanInvite, 0, len(rows))
	for _, row := range rows {
		invite, err := inviteFromRow(uuid.Nil(), planID.String(), row.Email, row.AccessLevel, row.Status, row.InvitedBy, row.CreatedAt, row.RespondedAt)
		if err != nil {
			return nil, err
		}
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse invite id: %w", err)
		}
		invite.ID = id
		invites = append(invites, invite)
	}
	return invites, nil
}

func (r *InviteRepository) GetPendingInvitesForEmail(email string) ([]*domain.PlanInvite, error) {
	rows, err := r.queries.GetPendingInvitesForEmail(context.Background(), db.GetPendingInvitesForEmailParams{
		Email:  email,
		Status: string(domain.InvitePending),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pending invites: %w", err)
	}

	invites := make([]*domain.PlanInvite, 0, len(rows))
	for _, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse invite id: %w", err)
		}
		invite, err := inviteFromRow(id, row.PlanID, email, row.AccessLevel, row.Status, row.InvitedBy, row.CreatedAt, row.RespondedAt)
		if err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	return invites, nil
}

func (r *InviteRepository) UpdateInviteStatus(id uuid.UUID, status domain.InviteStatus) error {
	affected, err := r.queries.UpdateInviteStatus(context.Background(), db.UpdateInviteStatusParams{
		Status:      string(status),
		RespondedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:          id.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to update invite status: %w", err)
	}
	if affected == 0 {
		return domain.ErrInviteNotFound
	}
	return nil
}

func inviteFromRow(id uuid.UUID, planIDStr, email, accessLevel, status, invitedByStr string, createdAt int64, respondedAt sql.NullInt64) (*domain.PlanInvite, error) {
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse invite plan id: %w", err)
	}
	invitedBy, err := uuid.Parse(invitedByStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse invite invited_by: %w", err)
	}

	return &domain.PlanInvite{
		ID:          id,
		PlanID:      planID,
		Email:       email,
		AccessLevel: domain.AccessLevel(accessLevel),
		Status:      domain.InviteStatus(status),
		InvitedBy:   invitedBy,
		CreatedAt:   createdAt,
		RespondedAt: respondedAt.Int64,
	}, nil
}
