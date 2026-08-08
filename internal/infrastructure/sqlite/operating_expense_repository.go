package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// OperatingExpenseRepository implements repositories.OperatingExpenseRepository.
type OperatingExpenseRepository struct {
	queries *db.Queries
}

func NewOperatingExpenseRepository(conn *sql.DB) repositories.OperatingExpenseRepository {
	return &OperatingExpenseRepository{queries: db.New(conn)}
}

var _ repositories.OperatingExpenseRepository = (*OperatingExpenseRepository)(nil)

func (r *OperatingExpenseRepository) CreateOperatingExpenseDraft(planID uuid.UUID) (uuid.UUID, error) {
	ctx := context.Background()

	if existing, err := r.FindOperatingExpenseDraft(planID); err != nil {
		return uuid.UUID{}, err
	} else if existing != nil {
		return existing.ID, nil
	}

	sortOrder, err := r.queries.MaxOperatingExpenseSortOrder(ctx, planID.String())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to generate id: %w", err)
	}

	now := time.Now().Unix()
	if err := r.queries.CreateOperatingExpenseDraft(ctx, db.CreateOperatingExpenseDraftParams{
		ID:        id.String(),
		PlanID:    planID.String(),
		Status:    string(repositories.StatusDraft),
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create operating expense draft: %w", err)
	}

	return id, nil
}

func (r *OperatingExpenseRepository) FindOperatingExpenseDraft(planID uuid.UUID) (*repositories.OperatingExpenseItem, error) {
	row, err := r.queries.FindOperatingExpenseDraft(context.Background(), db.FindOperatingExpenseDraftParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusDraft),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find operating expense draft: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse operating expense id: %w", err)
	}

	return &repositories.OperatingExpenseItem{
		ID:          id,
		Cost:        operatingExpenseFromRow(row.Name, row.MonthlyAmount, row.GrowthType, row.AnnualRate),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *OperatingExpenseRepository) GetOperatingExpense(itemID uuid.UUID) (*repositories.OperatingExpenseItem, error) {
	row, err := r.queries.GetOperatingExpense(context.Background(), itemID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("operating expense not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get operating expense: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse operating expense id: %w", err)
	}

	return &repositories.OperatingExpenseItem{
		ID:          id,
		Cost:        operatingExpenseFromRow(row.Name, row.MonthlyAmount, row.GrowthType, row.AnnualRate),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *OperatingExpenseRepository) SaveOperatingExpenseStep(itemID uuid.UUID, cost domain.Cost, currentStep int, status repositories.ItemStatus) error {
	affected, err := r.queries.SaveOperatingExpenseStep(context.Background(), db.SaveOperatingExpenseStepParams{
		Name:          cost.Name,
		MonthlyAmount: cost.BaseAmountPerMonth.MinorUnits(),
		GrowthType:    string(cost.Growth.Type),
		AnnualRate:    cost.Growth.AnnualRate,
		Status:        string(status),
		CurrentStep:   int64(currentStep),
		UpdatedAt:     time.Now().Unix(),
		ID:            itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to save operating expense step: %w", err)
	}
	if affected == 0 {
		return errors.New("operating expense not found")
	}
	return nil
}

func (r *OperatingExpenseRepository) ListCompleteOperatingExpenses(planID uuid.UUID) ([]repositories.OperatingExpenseItem, error) {
	rows, err := r.queries.ListCompleteOperatingExpenses(context.Background(), db.ListCompleteOperatingExpensesParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusComplete),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list operating expenses: %w", err)
	}

	items := make([]repositories.OperatingExpenseItem, len(rows))
	for i, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse operating expense id: %w", err)
		}
		items[i] = repositories.OperatingExpenseItem{
			ID:          id,
			Cost:        operatingExpenseFromRow(row.Name, row.MonthlyAmount, row.GrowthType, row.AnnualRate),
			Status:      repositories.ItemStatus(row.Status),
			CurrentStep: int(row.CurrentStep),
		}
	}
	return items, nil
}

func (r *OperatingExpenseRepository) DeleteOperatingExpense(itemID uuid.UUID) error {
	affected, err := r.queries.DeleteOperatingExpense(context.Background(), db.DeleteOperatingExpenseParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete operating expense: %w", err)
	}
	if affected == 0 {
		return errors.New("operating expense not found")
	}
	return nil
}

func operatingExpenseFromRow(name string, monthlyAmount int64, growthType string, annualRate float64) domain.Cost {
	return domain.Cost{
		Name:               name,
		BaseAmountPerMonth: mustUSD(monthlyAmount),
		Growth: domain.GrowthStrategy{
			Type:       domain.GrowthType(growthType),
			AnnualRate: annualRate,
		},
	}
}
