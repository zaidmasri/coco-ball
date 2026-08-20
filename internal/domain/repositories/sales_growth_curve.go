package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type SalesGrowthCurveRepository interface {
	GetSalesGrowthCurveRow(planID uuid.UUID) (*SalesGrowthCurveRow, error)
	SaveSalesGrowthCurveStep(planID uuid.UUID, curve domain.SalesGrowthCurve, currentStep int) error
}
