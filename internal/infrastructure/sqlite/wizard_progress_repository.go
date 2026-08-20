package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// WizardProgressRepository implements repositories.WizardProgressRepository.
type WizardProgressRepository struct {
	queries *db.Queries
}

func NewWizardProgressRepository(conn *sql.DB) repositories.WizardProgressRepository {
	return &WizardProgressRepository{queries: db.New(conn)}
}

var _ repositories.WizardProgressRepository = (*WizardProgressRepository)(nil)

func (r *WizardProgressRepository) MarkWizardSectionComplete(planID uuid.UUID, hub, section string) error {
	if err := r.queries.UpsertWizardSection(context.Background(), db.UpsertWizardSectionParams{
		PlanID:      planID.String(),
		Hub:         hub,
		Section:     section,
		CompletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to mark wizard section complete: %w", err)
	}
	return nil
}

func (r *WizardProgressRepository) GetWizardSectionStatus(planID uuid.UUID, hub string) (map[string]bool, error) {
	status := make(map[string]bool)
	for _, section := range domain.HubSections[hub] {
		status[section] = false
	}

	rows, err := r.queries.ListWizardSections(context.Background(), db.ListWizardSectionsParams{
		PlanID: planID.String(),
		Hub:    hub,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get wizard section status: %w", err)
	}

	for _, row := range rows {
		if row.CompletedAt.Valid && row.CompletedAt.Int64 > 0 {
			status[row.Section] = true
		}
	}

	return status, nil
}
