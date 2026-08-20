package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// AccessRepository implements repositories.AccessRepository.
type AccessRepository struct {
	queries *db.Queries
}

func NewAccessRepository(conn *sql.DB) repositories.AccessRepository {
	return &AccessRepository{queries: db.New(conn)}
}

var _ repositories.AccessRepository = (*AccessRepository)(nil)

// GrantAccess inserts/replaces a plan_access row. Existence checks for
// planID/userID are the caller's (AccessService's) responsibility - this
// method is pure translation, no business branching.
func (r *AccessRepository) GrantAccess(planID, userID uuid.UUID, level domain.AccessLevel) error {
	ctx := context.Background()

	if err := r.queries.GrantAccess(ctx, db.GrantAccessParams{
		PlanID:      planID.String(),
		UserID:      userID.String(),
		AccessLevel: string(level),
		InvitedAt:   0,
	}); err != nil {
		return fmt.Errorf("failed to grant access: %w", err)
	}
	return nil
}

func (r *AccessRepository) GetAccess(planID, userID uuid.UUID) (*domain.PlanAccess, error) {
	row, err := r.queries.GetAccess(context.Background(), db.GetAccessParams{
		PlanID: planID.String(),
		UserID: userID.String(),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAccessDenied
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get access: %w", err)
	}

	return &domain.PlanAccess{
		PlanID:      planID,
		UserID:      userID,
		AccessLevel: domain.AccessLevel(row.AccessLevel),
		InvitedAt:   row.InvitedAt,
	}, nil
}

func (r *AccessRepository) GetPlanAccess(planID uuid.UUID) ([]*domain.PlanAccess, error) {
	rows, err := r.queries.GetPlanAccess(context.Background(), planID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get plan access: %w", err)
	}

	access := make([]*domain.PlanAccess, len(rows))
	for i, row := range rows {
		userID, err := uuid.Parse(row.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user id: %w", err)
		}
		access[i] = &domain.PlanAccess{
			PlanID:      planID,
			UserID:      userID,
			AccessLevel: domain.AccessLevel(row.AccessLevel),
			InvitedAt:   row.InvitedAt,
		}
	}
	return access, nil
}
