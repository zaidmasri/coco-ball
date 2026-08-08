package repositories

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type PayrollTaxRatesRepository interface {
	GetPayrollTaxRatesRow(planID uuid.UUID) (*PayrollTaxRatesRow, error)
	SavePayrollTaxRatesStep(planID uuid.UUID, rates domain.PayrollTaxRates, currentStep int) error
}
