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

// DistributionRepository implements repositories.DistributionRepository.
type DistributionRepository struct {
	queries *db.Queries
}

func NewDistributionRepository(conn *sql.DB) repositories.DistributionRepository {
	return &DistributionRepository{queries: db.New(conn)}
}

var _ repositories.DistributionRepository = (*DistributionRepository)(nil)

func (r *DistributionRepository) CreateDistributionDraft(planID uuid.UUID) (uuid.UUID, error) {
	ctx := context.Background()

	if existing, err := r.FindDistributionDraft(planID); err != nil {
		return uuid.UUID{}, err
	} else if existing != nil {
		return existing.ID, nil
	}

	sortOrder, err := r.queries.MaxDistributionSortOrder(ctx, planID.String())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to generate id: %w", err)
	}

	now := time.Now().Unix()
	if err := r.queries.CreateDistributionDraft(ctx, db.CreateDistributionDraftParams{
		ID:        id.String(),
		PlanID:    planID.String(),
		Status:    string(repositories.StatusDraft),
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create distribution draft: %w", err)
	}

	return id, nil
}

func (r *DistributionRepository) FindDistributionDraft(planID uuid.UUID) (*repositories.DistributionItem, error) {
	row, err := r.queries.FindDistributionDraft(context.Background(), db.FindDistributionDraftParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusDraft),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find distribution draft: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse distribution id: %w", err)
	}

	return &repositories.DistributionItem{
		ID:           id,
		Distribution: distributionFromRow(row.Name, row.MonthlyAmount, row.GrowthYr2, row.GrowthYr3),
		Status:       repositories.ItemStatus(row.Status),
		CurrentStep:  int(row.CurrentStep),
	}, nil
}

func (r *DistributionRepository) GetDistribution(itemID uuid.UUID) (*repositories.DistributionItem, error) {
	row, err := r.queries.GetDistribution(context.Background(), itemID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("distribution not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse distribution id: %w", err)
	}

	return &repositories.DistributionItem{
		ID:           id,
		Distribution: distributionFromRow(row.Name, row.MonthlyAmount, row.GrowthYr2, row.GrowthYr3),
		Status:       repositories.ItemStatus(row.Status),
		CurrentStep:  int(row.CurrentStep),
	}, nil
}

func (r *DistributionRepository) SaveDistributionStep(itemID uuid.UUID, dist domain.Distribution, currentStep int, status repositories.ItemStatus) error {
	var yr2, yr3 float64
	if len(dist.GrowthAfterYr1.RatesAfterYear1) > 0 {
		yr2 = dist.GrowthAfterYr1.RatesAfterYear1[0]
	}
	if len(dist.GrowthAfterYr1.RatesAfterYear1) > 1 {
		yr3 = dist.GrowthAfterYr1.RatesAfterYear1[1]
	}

	affected, err := r.queries.SaveDistributionStep(context.Background(), db.SaveDistributionStepParams{
		Name:          dist.Name,
		MonthlyAmount: dist.MonthlyAmount.MinorUnits(),
		GrowthYr2:     yr2,
		GrowthYr3:     yr3,
		Status:        string(status),
		CurrentStep:   int64(currentStep),
		UpdatedAt:     time.Now().Unix(),
		ID:            itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to save distribution step: %w", err)
	}
	if affected == 0 {
		return errors.New("distribution not found")
	}
	return nil
}

func (r *DistributionRepository) ListCompleteDistributions(planID uuid.UUID) ([]repositories.DistributionItem, error) {
	rows, err := r.queries.ListCompleteDistributions(context.Background(), db.ListCompleteDistributionsParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusComplete),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list distributions: %w", err)
	}

	items := make([]repositories.DistributionItem, len(rows))
	for i, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse distribution id: %w", err)
		}
		items[i] = repositories.DistributionItem{
			ID:           id,
			Distribution: distributionFromRow(row.Name, row.MonthlyAmount, row.GrowthYr2, row.GrowthYr3),
			Status:       repositories.ItemStatus(row.Status),
			CurrentStep:  int(row.CurrentStep),
		}
	}
	return items, nil
}

func (r *DistributionRepository) DeleteDistribution(itemID uuid.UUID) error {
	affected, err := r.queries.DeleteDistribution(context.Background(), db.DeleteDistributionParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete distribution: %w", err)
	}
	if affected == 0 {
		return errors.New("distribution not found")
	}
	return nil
}

func distributionFromRow(name string, monthlyAmount int64, growthYr2, growthYr3 float64) domain.Distribution {
	return domain.Distribution{
		Name:           name,
		MonthlyAmount:  mustUSD(monthlyAmount),
		GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{growthYr2, growthYr3}},
	}
}
