package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type PayrollTaxRatesRepository interface {
	GetPayrollTaxRatesRow(planID uuid.UUID) (*PayrollTaxRatesRow, error)
	SavePayrollTaxRatesStep(planID uuid.UUID, rates domain.PayrollTaxRates, currentStep int) error
}
