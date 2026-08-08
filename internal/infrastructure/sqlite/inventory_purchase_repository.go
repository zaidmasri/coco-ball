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

// InventoryPurchaseRepository implements repositories.InventoryPurchaseRepository.
type InventoryPurchaseRepository struct {
	queries *db.Queries
}

func NewInventoryPurchaseRepository(conn *sql.DB) repositories.InventoryPurchaseRepository {
	return &InventoryPurchaseRepository{queries: db.New(conn)}
}

var _ repositories.InventoryPurchaseRepository = (*InventoryPurchaseRepository)(nil)

func (r *InventoryPurchaseRepository) CreateInventoryPurchaseDraft(planID uuid.UUID) (uuid.UUID, error) {
	ctx := context.Background()

	if existing, err := r.FindInventoryPurchaseDraft(planID); err != nil {
		return uuid.UUID{}, err
	} else if existing != nil {
		return existing.ID, nil
	}

	sortOrder, err := r.queries.MaxInventoryPurchaseSortOrder(ctx, planID.String())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to generate id: %w", err)
	}

	now := time.Now().Unix()
	if err := r.queries.CreateInventoryPurchaseDraft(ctx, db.CreateInventoryPurchaseDraftParams{
		ID:        id.String(),
		PlanID:    planID.String(),
		Status:    string(repositories.StatusDraft),
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create inventory purchase draft: %w", err)
	}

	return id, nil
}

func (r *InventoryPurchaseRepository) FindInventoryPurchaseDraft(planID uuid.UUID) (*repositories.InventoryPurchaseItem, error) {
	row, err := r.queries.FindInventoryPurchaseDraft(context.Background(), db.FindInventoryPurchaseDraftParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusDraft),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find inventory purchase draft: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inventory purchase id: %w", err)
	}

	return &repositories.InventoryPurchaseItem{
		ID:          id,
		Purchase:    inventoryPurchaseFromRow(row.Category, row.MonthlyAmount, row.GrowthYr2, row.GrowthYr3),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *InventoryPurchaseRepository) GetInventoryPurchase(itemID uuid.UUID) (*repositories.InventoryPurchaseItem, error) {
	row, err := r.queries.GetInventoryPurchase(context.Background(), itemID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("inventory purchase not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory purchase: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inventory purchase id: %w", err)
	}

	return &repositories.InventoryPurchaseItem{
		ID:          id,
		Purchase:    inventoryPurchaseFromRow(row.Category, row.MonthlyAmount, row.GrowthYr2, row.GrowthYr3),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *InventoryPurchaseRepository) SaveInventoryPurchaseStep(itemID uuid.UUID, inv domain.InventoryPurchase, currentStep int, status repositories.ItemStatus) error {
	var yr2, yr3 float64
	if len(inv.GrowthAfterYr1.RatesAfterYear1) > 0 {
		yr2 = inv.GrowthAfterYr1.RatesAfterYear1[0]
	}
	if len(inv.GrowthAfterYr1.RatesAfterYear1) > 1 {
		yr3 = inv.GrowthAfterYr1.RatesAfterYear1[1]
	}

	affected, err := r.queries.SaveInventoryPurchaseStep(context.Background(), db.SaveInventoryPurchaseStepParams{
		Category:      inv.Category,
		MonthlyAmount: inv.MonthlyAmount.MinorUnits(),
		GrowthYr2:     yr2,
		GrowthYr3:     yr3,
		Status:        string(status),
		CurrentStep:   int64(currentStep),
		UpdatedAt:     time.Now().Unix(),
		ID:            itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to save inventory purchase step: %w", err)
	}
	if affected == 0 {
		return errors.New("inventory purchase not found")
	}
	return nil
}

func (r *InventoryPurchaseRepository) ListCompleteInventoryPurchases(planID uuid.UUID) ([]repositories.InventoryPurchaseItem, error) {
	rows, err := r.queries.ListCompleteInventoryPurchases(context.Background(), db.ListCompleteInventoryPurchasesParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusComplete),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list inventory purchases: %w", err)
	}

	items := make([]repositories.InventoryPurchaseItem, len(rows))
	for i, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse inventory purchase id: %w", err)
		}
		items[i] = repositories.InventoryPurchaseItem{
			ID:          id,
			Purchase:    inventoryPurchaseFromRow(row.Category, row.MonthlyAmount, row.GrowthYr2, row.GrowthYr3),
			Status:      repositories.ItemStatus(row.Status),
			CurrentStep: int(row.CurrentStep),
		}
	}
	return items, nil
}

func (r *InventoryPurchaseRepository) DeleteInventoryPurchase(itemID uuid.UUID) error {
	affected, err := r.queries.DeleteInventoryPurchase(context.Background(), db.DeleteInventoryPurchaseParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete inventory purchase: %w", err)
	}
	if affected == 0 {
		return errors.New("inventory purchase not found")
	}
	return nil
}

func inventoryPurchaseFromRow(category string, monthlyAmount int64, growthYr2, growthYr3 float64) domain.InventoryPurchase {
	return domain.InventoryPurchase{
		Category:       category,
		MonthlyAmount:  mustUSD(monthlyAmount),
		GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{growthYr2, growthYr3}},
	}
}
