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

// StartingBalancesRepository implements repositories.StartingBalancesRepository.
type StartingBalancesRepository struct {
	queries *db.Queries
}

func NewStartingBalancesRepository(conn *sql.DB) repositories.StartingBalancesRepository {
	return &StartingBalancesRepository{queries: db.New(conn)}
}

var _ repositories.StartingBalancesRepository = (*StartingBalancesRepository)(nil)

// GetStartingBalancesRow returns a zero-value row (not an error) when no row
// exists yet for the plan, matching the old store's read-has-no-write-side-
// effect behavior.
func (r *StartingBalancesRepository) GetStartingBalancesRow(planID uuid.UUID) (*repositories.StartingBalancesRow, error) {
	row, err := r.queries.GetStartingBalancesRow(context.Background(), planID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return &repositories.StartingBalancesRow{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get starting balances: %w", err)
	}

	return &repositories.StartingBalancesRow{
		Balances: domain.StartingBalances{
			Cash:               mustUSD(row.Cash),
			AccountsReceivable: mustUSD(row.AccountsReceivable),
			PrepaidExpenses:    mustUSD(row.PrepaidExpenses),
			AccountsPayable:    mustUSD(row.AccountsPayable),
			AccruedExpenses:    mustUSD(row.AccruedExpenses),
		},
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *StartingBalancesRepository) SaveStartingBalancesStep(planID uuid.UUID, balances domain.StartingBalances, currentStep int) error {
	now := time.Now().Unix()
	if err := r.queries.SaveStartingBalancesStep(context.Background(), db.SaveStartingBalancesStepParams{
		PlanID:             planID.String(),
		Cash:               balances.Cash.MinorUnits(),
		AccountsReceivable: balances.AccountsReceivable.MinorUnits(),
		PrepaidExpenses:    balances.PrepaidExpenses.MinorUnits(),
		AccountsPayable:    balances.AccountsPayable.MinorUnits(),
		AccruedExpenses:    balances.AccruedExpenses.MinorUnits(),
		CurrentStep:        int64(currentStep),
		CreatedAt:          now,
		UpdatedAt:          now,
	}); err != nil {
		return fmt.Errorf("failed to save starting balances: %w", err)
	}
	return nil
}
