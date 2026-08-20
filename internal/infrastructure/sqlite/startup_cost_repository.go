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

// StartupCostRepository implements repositories.StartupCostRepository.
type StartupCostRepository struct {
	queries *db.Queries
}

func NewStartupCostRepository(conn *sql.DB) repositories.StartupCostRepository {
	return &StartupCostRepository{queries: db.New(conn)}
}

var _ repositories.StartupCostRepository = (*StartupCostRepository)(nil)

func (r *StartupCostRepository) CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error) {
	ctx := context.Background()

	if existing, err := r.FindStartupCostDraft(planID); err != nil {
		return uuid.UUID{}, err
	} else if existing != nil {
		return existing.ID, nil
	}

	sortOrder, err := r.queries.MaxStartupCostSortOrder(ctx, planID.String())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.NewV7()

	now := time.Now().Unix()
	if err := r.queries.CreateStartupCostDraft(ctx, db.CreateStartupCostDraftParams{
		ID:        id.String(),
		PlanID:    planID.String(),
		Status:    string(repositories.StatusDraft),
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create startup cost draft: %w", err)
	}

	return id, nil
}

func (r *StartupCostRepository) FindStartupCostDraft(planID uuid.UUID) (*repositories.StartupCostItem, error) {
	row, err := r.queries.FindStartupCostDraft(context.Background(), db.FindStartupCostDraftParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusDraft),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find startup cost draft: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse startup cost id: %w", err)
	}

	return &repositories.StartupCostItem{
		ID:          id,
		Cost:        domain.StartupCost{Name: row.Name, Amount: mustUSD(row.Amount)},
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *StartupCostRepository) GetStartupCost(itemID uuid.UUID) (*repositories.StartupCostItem, error) {
	row, err := r.queries.GetStartupCost(context.Background(), itemID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("startup cost not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get startup cost: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse startup cost id: %w", err)
	}

	return &repositories.StartupCostItem{
		ID:          id,
		Cost:        domain.StartupCost{Name: row.Name, Amount: mustUSD(row.Amount)},
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *StartupCostRepository) SaveStartupCostStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int, status repositories.ItemStatus) error {
	affected, err := r.queries.SaveStartupCostStep(context.Background(), db.SaveStartupCostStepParams{
		Name:        cost.Name,
		Amount:      cost.Amount.MinorUnits(),
		Status:      string(status),
		CurrentStep: int64(currentStep),
		UpdatedAt:   time.Now().Unix(),
		ID:          itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to save startup cost step: %w", err)
	}
	if affected == 0 {
		return errors.New("startup cost not found")
	}
	return nil
}

func (r *StartupCostRepository) ListCompleteStartupCosts(planID uuid.UUID) ([]repositories.StartupCostItem, error) {
	rows, err := r.queries.ListCompleteStartupCosts(context.Background(), db.ListCompleteStartupCostsParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusComplete),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list startup costs: %w", err)
	}

	items := make([]repositories.StartupCostItem, len(rows))
	for i, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse startup cost id: %w", err)
		}
		items[i] = repositories.StartupCostItem{
			ID:          id,
			Cost:        domain.StartupCost{Name: row.Name, Amount: mustUSD(row.Amount)},
			Status:      repositories.ItemStatus(row.Status),
			CurrentStep: int(row.CurrentStep),
		}
	}
	return items, nil
}

func (r *StartupCostRepository) DeleteStartupCost(itemID uuid.UUID) error {
	affected, err := r.queries.DeleteStartupCost(context.Background(), db.DeleteStartupCostParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete startup cost: %w", err)
	}
	if affected == 0 {
		return errors.New("startup cost not found")
	}
	return nil
}
