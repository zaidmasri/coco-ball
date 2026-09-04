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

// CapitalAssetRepository implements repositories.CapitalAssetRepository.
type CapitalAssetRepository struct {
	queries *db.Queries
}

func NewCapitalAssetRepository(conn *sql.DB) repositories.CapitalAssetRepository {
	return &CapitalAssetRepository{queries: db.New(conn)}
}

var _ repositories.CapitalAssetRepository = (*CapitalAssetRepository)(nil)

func (r *CapitalAssetRepository) CreateCapitalAssetDraft(planID uuid.UUID) (uuid.UUID, error) {
	ctx := context.Background()

	if existing, err := r.FindCapitalAssetDraft(planID); err != nil {
		return uuid.UUID{}, err
	} else if existing != nil {
		return existing.ID, nil
	}

	sortOrder, err := r.queries.MaxCapitalAssetSortOrder(ctx, planID.String())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.NewV7()

	now := time.Now().Unix()
	if err := r.queries.CreateCapitalAssetDraft(ctx, db.CreateCapitalAssetDraftParams{
		ID:        id.String(),
		PlanID:    planID.String(),
		Status:    string(repositories.StatusDraft),
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create capital asset draft: %w", err)
	}

	return id, nil
}

func (r *CapitalAssetRepository) FindCapitalAssetDraft(planID uuid.UUID) (*repositories.CapitalAssetItem, error) {
	row, err := r.queries.FindCapitalAssetDraft(context.Background(), db.FindCapitalAssetDraftParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusDraft),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find capital asset draft: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse capital asset id: %w", err)
	}

	return &repositories.CapitalAssetItem{
		ID:          id,
		Asset:       capitalAssetFromRow(id, row.Name, row.PurchaseCost, row.UsefulLifeMonths, row.SalvageValue, row.PurchaseMonthIndex, row.DepreciationMethod),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *CapitalAssetRepository) GetCapitalAsset(itemID uuid.UUID) (*repositories.CapitalAssetItem, error) {
	row, err := r.queries.GetCapitalAsset(context.Background(), itemID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("capital asset not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get capital asset: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse capital asset id: %w", err)
	}

	return &repositories.CapitalAssetItem{
		ID:          id,
		Asset:       capitalAssetFromRow(id, row.Name, row.PurchaseCost, row.UsefulLifeMonths, row.SalvageValue, row.PurchaseMonthIndex, row.DepreciationMethod),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *CapitalAssetRepository) SaveCapitalAssetDraftStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int) error {
	return r.saveCapitalAssetStep(itemID, asset, currentStep, repositories.StatusDraft)
}

func (r *CapitalAssetRepository) CompleteCapitalAsset(itemID uuid.UUID, va domain.ValidatedCapitalAsset, currentStep int) error {
	return r.saveCapitalAssetStep(itemID, va.CapitalAsset(), currentStep, repositories.StatusComplete)
}

func (r *CapitalAssetRepository) saveCapitalAssetStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int, status repositories.ItemStatus) error {
	affected, err := r.queries.SaveCapitalAssetStep(context.Background(), db.SaveCapitalAssetStepParams{
		Name:               asset.Name,
		PurchaseCost:       asset.PurchaseCost.MinorUnits(),
		UsefulLifeMonths:   int64(asset.UsefulLifeMonths),
		SalvageValue:       asset.SalvageValue.MinorUnits(),
		PurchaseMonthIndex: int64(asset.PurchaseMonthIndex),
		DepreciationMethod: string(asset.DepreciationMethod),
		Status:             string(status),
		CurrentStep:        int64(currentStep),
		UpdatedAt:          time.Now().Unix(),
		ID:                 itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to save capital asset step: %w", err)
	}
	if affected == 0 {
		return errors.New("capital asset not found")
	}
	return nil
}

func (r *CapitalAssetRepository) ListCompleteCapitalAssets(planID uuid.UUID) ([]repositories.CapitalAssetItem, error) {
	rows, err := r.queries.ListCompleteCapitalAssets(context.Background(), db.ListCompleteCapitalAssetsParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusComplete),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list capital assets: %w", err)
	}

	items := make([]repositories.CapitalAssetItem, len(rows))
	for i, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse capital asset id: %w", err)
		}
		items[i] = repositories.CapitalAssetItem{
			ID:          id,
			Asset:       capitalAssetFromRow(id, row.Name, row.PurchaseCost, row.UsefulLifeMonths, row.SalvageValue, row.PurchaseMonthIndex, row.DepreciationMethod),
			Status:      repositories.ItemStatus(row.Status),
			CurrentStep: int(row.CurrentStep),
		}
	}
	return items, nil
}

func (r *CapitalAssetRepository) DeleteCapitalAsset(itemID uuid.UUID) error {
	affected, err := r.queries.DeleteCapitalAsset(context.Background(), db.DeleteCapitalAssetParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete capital asset: %w", err)
	}
	if affected == 0 {
		return errors.New("capital asset not found")
	}
	return nil
}

func capitalAssetFromRow(id uuid.UUID, name string, purchaseCost, usefulLifeMonths, salvageValue, purchaseMonthIndex int64, depreciationMethod string) domain.CapitalAsset {
	c := domain.CapitalAsset{
		Name:               name,
		PurchaseCost:       mustUSD(purchaseCost),
		UsefulLifeMonths:   int(usefulLifeMonths),
		SalvageValue:       mustUSD(salvageValue),
		PurchaseMonthIndex: domain.MonthIndex(purchaseMonthIndex),
		DepreciationMethod: domain.DepreciationMethod(depreciationMethod),
	}
	c.SetID(id)
	return c
}
