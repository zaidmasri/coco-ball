package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type SalesGrowthCurveRepository interface {
	GetSalesGrowthCurveRow(planID uuid.UUID) (*SalesGrowthCurveRow, error)
	SaveSalesGrowthCurveStep(planID uuid.UUID, curve domain.SalesGrowthCurve, currentStep int) error
}
