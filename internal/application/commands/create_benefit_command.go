package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateBenefit is the command that finalizes a Benefits wizard item into a
// valid domain.Benefit for the first time. PayrollService.CreateBenefit
// turns this into a domain.Benefit via domain.NewBenefit (validating the
// accumulated wizard field values) and assigns it ItemID — the wizard row's
// pre-existing ID from CreateBenefitDraft — via Benefit.SetID, so the
// domain entity and its wizard row share one identity. Callers must not
// construct the entity themselves.
type CreateBenefit struct {
	ItemID        uuid.UUID
	Type          string
	MonthlyAmount domain.Money
	Growth        domain.AnnualGrowth
	CurrentStep   int
}

// CreateBenefitResult wraps the newly-completed item's Result.
type CreateBenefitResult struct {
	Result *common.BenefitResult
}
