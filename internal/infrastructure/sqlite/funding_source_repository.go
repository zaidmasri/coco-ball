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

// FundingSourceRepository implements repositories.FundingSourceRepository.
type FundingSourceRepository struct {
	queries *db.Queries
}

func NewFundingSourceRepository(conn *sql.DB) repositories.FundingSourceRepository {
	return &FundingSourceRepository{queries: db.New(conn)}
}

var _ repositories.FundingSourceRepository = (*FundingSourceRepository)(nil)

func (r *FundingSourceRepository) CreateFundingSourceDraft(planID uuid.UUID) (uuid.UUID, error) {
	ctx := context.Background()

	if existing, err := r.FindFundingSourceDraft(planID); err != nil {
		return uuid.UUID{}, err
	} else if existing != nil {
		return existing.ID, nil
	}

	sortOrder, err := r.queries.MaxFundingSourceSortOrder(ctx, planID.String())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.NewV7()

	now := time.Now().Unix()
	if err := r.queries.CreateFundingSourceDraft(ctx, db.CreateFundingSourceDraftParams{
		ID:        id.String(),
		PlanID:    planID.String(),
		Status:    string(repositories.StatusDraft),
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create funding source draft: %w", err)
	}

	return id, nil
}

func (r *FundingSourceRepository) FindFundingSourceDraft(planID uuid.UUID) (*repositories.FundingSourceItem, error) {
	row, err := r.queries.FindFundingSourceDraft(context.Background(), db.FindFundingSourceDraftParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusDraft),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find funding source draft: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse funding source id: %w", err)
	}

	return &repositories.FundingSourceItem{
		ID:          id,
		Funding:     fundingSourceFromRow(id, row.Name, row.Amount, row.InterestRate, row.TermMonths),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *FundingSourceRepository) GetFundingSource(itemID uuid.UUID) (*repositories.FundingSourceItem, error) {
	row, err := r.queries.GetFundingSource(context.Background(), itemID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("funding source not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get funding source: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse funding source id: %w", err)
	}

	return &repositories.FundingSourceItem{
		ID:          id,
		Funding:     fundingSourceFromRow(id, row.Name, row.Amount, row.InterestRate, row.TermMonths),
		Status:      repositories.ItemStatus(row.Status),
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *FundingSourceRepository) SaveFundingSourceDraftStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int) error {
	return r.saveFundingSourceStep(itemID, funding, currentStep, repositories.StatusDraft)
}

func (r *FundingSourceRepository) CompleteFundingSource(itemID uuid.UUID, vf domain.ValidatedFundingSource, currentStep int) error {
	return r.saveFundingSourceStep(itemID, vf.FundingSource(), currentStep, repositories.StatusComplete)
}

func (r *FundingSourceRepository) saveFundingSourceStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int, status repositories.ItemStatus) error {
	affected, err := r.queries.SaveFundingSourceStep(context.Background(), db.SaveFundingSourceStepParams{
		Name:         funding.Name,
		Amount:       funding.Amount.MinorUnits(),
		InterestRate: funding.InterestRate,
		TermMonths:   int64(funding.TermMonths),
		Status:       string(status),
		CurrentStep:  int64(currentStep),
		UpdatedAt:    time.Now().Unix(),
		ID:           itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to save funding source step: %w", err)
	}
	if affected == 0 {
		return errors.New("funding source not found")
	}
	return nil
}

func (r *FundingSourceRepository) ListCompleteFundingSources(planID uuid.UUID) ([]repositories.FundingSourceItem, error) {
	rows, err := r.queries.ListCompleteFundingSources(context.Background(), db.ListCompleteFundingSourcesParams{
		PlanID: planID.String(),
		Status: string(repositories.StatusComplete),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list funding sources: %w", err)
	}

	items := make([]repositories.FundingSourceItem, len(rows))
	for i, row := range rows {
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse funding source id: %w", err)
		}
		items[i] = repositories.FundingSourceItem{
			ID:          id,
			Funding:     fundingSourceFromRow(id, row.Name, row.Amount, row.InterestRate, row.TermMonths),
			Status:      repositories.ItemStatus(row.Status),
			CurrentStep: int(row.CurrentStep),
		}
	}
	return items, nil
}

func (r *FundingSourceRepository) DeleteFundingSource(itemID uuid.UUID) error {
	affected, err := r.queries.DeleteFundingSource(context.Background(), db.DeleteFundingSourceParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        itemID.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete funding source: %w", err)
	}
	if affected == 0 {
		return errors.New("funding source not found")
	}
	return nil
}

func fundingSourceFromRow(id uuid.UUID, name string, amount int64, interestRate float64, termMonths int64) domain.FundingSource {
	f := domain.FundingSource{
		Name:         name,
		Amount:       mustUSD(amount),
		InterestRate: interestRate,
		TermMonths:   int(termMonths),
	}
	f.SetID(id)
	return f
}
