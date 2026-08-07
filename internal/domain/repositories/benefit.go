package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type BenefitRepository interface {
	CreateBenefitDraft(planID uuid.UUID) (uuid.UUID, error)
	GetBenefitDraft(planID uuid.UUID) (*BenefitItem, error)
	GetBenefit(itemID uuid.UUID) (*BenefitItem, error)
	SaveBenefitStep(itemID uuid.UUID, benefit domain.Benefit, currentStep int, status ItemStatus) error
	ListCompleteBenefits(planID uuid.UUID) ([]BenefitItem, error)
	DeleteBenefit(itemID uuid.UUID) error
}
