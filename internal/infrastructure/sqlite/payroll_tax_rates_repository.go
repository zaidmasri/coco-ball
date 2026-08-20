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

// PayrollTaxRatesRepository implements repositories.PayrollTaxRatesRepository.
type PayrollTaxRatesRepository struct {
	queries *db.Queries
}

func NewPayrollTaxRatesRepository(conn *sql.DB) repositories.PayrollTaxRatesRepository {
	return &PayrollTaxRatesRepository{queries: db.New(conn)}
}

var _ repositories.PayrollTaxRatesRepository = (*PayrollTaxRatesRepository)(nil)

func (r *PayrollTaxRatesRepository) GetPayrollTaxRatesRow(planID uuid.UUID) (*repositories.PayrollTaxRatesRow, error) {
	row, err := r.queries.GetPayrollTaxRatesRow(context.Background(), planID.String())
	if errors.Is(err, sql.ErrNoRows) {
		return &repositories.PayrollTaxRatesRow{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payroll tax rates: %w", err)
	}

	return &repositories.PayrollTaxRatesRow{
		Rates: domain.PayrollTaxRates{
			SocialSecurityRate: row.SocialSecurityRate,
			MedicareRate:       row.MedicareRate,
			FUTARate:           row.FutaRate,
			SUTARate:           row.SutaRate,
		},
		CurrentStep: int(row.CurrentStep),
	}, nil
}

func (r *PayrollTaxRatesRepository) SavePayrollTaxRatesStep(planID uuid.UUID, rates domain.PayrollTaxRates, currentStep int) error {
	now := time.Now().Unix()
	if err := r.queries.SavePayrollTaxRatesStep(context.Background(), db.SavePayrollTaxRatesStepParams{
		PlanID:             planID.String(),
		SocialSecurityRate: rates.SocialSecurityRate,
		MedicareRate:       rates.MedicareRate,
		FutaRate:           rates.FUTARate,
		SutaRate:           rates.SUTARate,
		CurrentStep:        int64(currentStep),
		CreatedAt:          now,
		UpdatedAt:          now,
	}); err != nil {
		return fmt.Errorf("failed to save payroll tax rates: %w", err)
	}
	return nil
}
