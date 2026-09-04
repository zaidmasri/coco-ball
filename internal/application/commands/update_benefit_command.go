package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// UpdateBenefit is the command to change an already-complete Benefits
// item's fields. PayrollService.UpdateBenefit re-validates the fields via
// domain.NewBenefit and assigns the result ItemID — callers must not
// construct the entity themselves.
type UpdateBenefit struct {
	ItemID        uuid.UUID
	Type          string
	MonthlyAmount domain.Money
	Growth        domain.AnnualGrowth
	CurrentStep   int
}

// UpdateBenefitResult wraps the updated item's Result.
type UpdateBenefitResult struct {
	Result *common.BenefitResult
}
