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

// BenefitRepository implements repositories.BenefitRepository.
type BenefitRepository struct {
	queries *db.Queries
}

func NewBenefitRepository(conn *sql.DB) repositories.BenefitRepository {
	return &BenefitRepository{queries: db.New(conn)}
}

var _ repositories.BenefitRepository = (*BenefitRepository)(nil)

func (r *BenefitRepository) CreateBenefitDraft(planID uuid.UUID) (uuid.UUID, error) {
	ctx := context.Background()

	if existing, err := r.FindBenefitDraft(planID); err != nil {
		return uuid.UUID{}, err
	} else if existing != nil {
		return existing.ID, nil
	}

	sortOrder, err := r.queries.MaxBenefitSortOrder(ctx, planID.String())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.NewV7()

	now := time.Now().Unix()
	if err := r.queries.CreateBenefitDraft(ctx, db.CreateBenefitDraftParams{
		ID:        id.String(),
		PlanID:    planID.String(),
		Status:    string(repositories.StatusDraft),
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create benefit draft: %w", err)
	}

	return id, nil
}

func (r *BenefitRepository) FindBenefitDraft(planID uuid.UUID) (*repositories.BenefitItem, error) {
	row, err := r.queries.FindBenefitDraft(context.Background(), db.FindBenefitDraftParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusDraft),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find benefit draft: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse benefit id: %w", err)
	}

	return &repositories.BenefitItem{
		ID:          id,
		Benefit:     benefitFromRow(id, row.Type, row.MonthlyAmount, row.GrowthYr2, row.GrowthYr3),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *BenefitRepository) GetBenefit(itemID uuid.UUID) (*repositories.BenefitItem, error) {
	row, err := r.queries.GetBenefit(context.Background(), itemID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("benefit not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get benefit: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse benefit id: %w", err)
	}

	return &repositories.BenefitItem{
		ID:          id,
		Benefit:     benefitFromRow(id, row.Type, row.MonthlyAmount, row.GrowthYr2, row.GrowthYr3),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *BenefitRepository) SaveBenefitDraftStep(itemID uuid.UUID, benefit domain.Benefit, currentStep int) error {
	return r.saveBenefitStep(itemID, benefit, currentStep, repositories.StatusDraft)
}

func (r *BenefitRepository) CompleteBenefit(itemID uuid.UUID, vb domain.ValidatedBenefit, currentStep int) error {
	return r.saveBenefitStep(itemID, vb.Benefit(), currentStep, repositories.StatusComplete)
}

func (r *BenefitRepository) saveBenefitStep(itemID uuid.UUID, benefit domain.Benefit, currentStep int, status repositories.ItemStatus) error {
	var yr2, yr3 float64
	if len(benefit.GrowthAfterYr1.RatesAfterYear1) > 0 {
		yr2 = benefit.GrowthAfterYr1.RatesAfterYear1[0]
	}
	if len(benefit.GrowthAfterYr1.RatesAfterYear1) > 1 {
		yr3 = benefit.GrowthAfterYr1.RatesAfterYear1[1]
	}

	affected, err := r.queries.SaveBenefitStep(context.Background(), db.SaveBenefitStepParams{
		Type:          benefit.Type,
		MonthlyAmount: benefit.MonthlyAmount.MinorUnits(),
		GrowthYr2:     yr2,
		GrowthYr3:     yr3,
		Status:        string(status),
		CurrentStep:   int64(currentStep),
		UpdatedAt:     time.Now().Unix(),
		ID:            itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to save benefit step: %w", err)
	}
	if affected == 0 {
		return errors.New("benefit not found")
	}
	return nil
}

func (r *BenefitRepository) ListCompleteBenefits(planID uuid.UUID) ([]repositories.BenefitItem, error) {
	rows, err := r.queries.ListCompleteBenefits(context.Background(), db.ListCompleteBenefitsParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusComplete),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list benefits: %w", err)
	}

	items := make([]repositories.BenefitItem, len(rows))
	for i, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse benefit id: %w", err)
		}
		items[i] = repositories.BenefitItem{
			ID:          id,
			Benefit:     benefitFromRow(id, row.Type, row.MonthlyAmount, row.GrowthYr2, row.GrowthYr3),
			Status:      repositories.ItemStatus(row.Status),
			CurrentStep: int(row.CurrentStep),
		}
	}
	return items, nil
}

func (r *BenefitRepository) DeleteBenefit(itemID uuid.UUID) error {
	affected, err := r.queries.DeleteBenefit(context.Background(), db.DeleteBenefitParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete benefit: %w", err)
	}
	if affected == 0 {
		return errors.New("benefit not found")
	}
	return nil
}

func benefitFromRow(id uuid.UUID, benefitType string, monthlyAmount int64, growthYr2, growthYr3 float64) domain.Benefit {
	b := domain.Benefit{
		Type:           benefitType,
		MonthlyAmount:  mustUSD(monthlyAmount),
		GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{growthYr2, growthYr3}},
	}
	b.SetID(id)
	return b
}
