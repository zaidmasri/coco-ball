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

// SalesGrowthCurveRepository implements repositories.SalesGrowthCurveRepository.
type SalesGrowthCurveRepository struct {
	queries *db.Queries
}

func NewSalesGrowthCurveRepository(conn *sql.DB) repositories.SalesGrowthCurveRepository {
	return &SalesGrowthCurveRepository{queries: db.New(conn)}
}

var _ repositories.SalesGrowthCurveRepository = (*SalesGrowthCurveRepository)(nil)

func (r *SalesGrowthCurveRepository) GetSalesGrowthCurveRow(planID uuid.UUID) (*repositories.SalesGrowthCurveRow, error) {
	row, err := r.queries.GetSalesGrowthCurveRow(context.Background(), planID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return &repositories.SalesGrowthCurveRow{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sales growth curve: %w", err)
	}

	return &repositories.SalesGrowthCurveRow{
		Curve: domain.SalesGrowthCurve{
			Year1QuarterlyRates: [4]float64{row.Y1Q1, row.Y1Q2, row.Y1Q3, row.Y1Q4},
			FutureYearRates:     []float64{row.GrowthYr2, row.GrowthYr3},
		},
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *SalesGrowthCurveRepository) SaveSalesGrowthCurveStep(planID uuid.UUID, curve domain.SalesGrowthCurve, currentStep int) error {
	var yr2, yr3 float64
	if len(curve.FutureYearRates) > 0 {
		yr2 = curve.FutureYearRates[0]
	}
	if len(curve.FutureYearRates) > 1 {
		yr3 = curve.FutureYearRates[1]
	}

	now := time.Now().Unix()
	if err := r.queries.SaveSalesGrowthCurveStep(context.Background(), db.SaveSalesGrowthCurveStepParams{
		PlanID:      planID.String(),
		Y1Q1:        curve.Year1QuarterlyRates[0],
		Y1Q2:        curve.Year1QuarterlyRates[1],
		Y1Q3:        curve.Year1QuarterlyRates[2],
		Y1Q4:        curve.Year1QuarterlyRates[3],
		GrowthYr2:   yr2,
		GrowthYr3:   yr3,
		CurrentStep: int64(currentStep),
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		return fmt.Errorf("failed to save sales growth curve: %w", err)
	}
	return nil
}
