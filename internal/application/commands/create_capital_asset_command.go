package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateCapitalAsset is the command that finalizes a Fixed Assets wizard
// item into a valid domain.CapitalAsset for the first time.
// StartingPointService.CreateCapitalAsset turns this into a
// domain.CapitalAsset via domain.NewCapitalAsset (validating the
// accumulated wizard field values) and assigns it ItemID — the wizard
// row's pre-existing ID from CreateCapitalAssetDraft — via
// CapitalAsset.SetID, so the domain entity and its wizard row share one
// identity. Callers must not construct the entity themselves.
type CreateCapitalAsset struct {
	ItemID             uuid.UUID
	Name               string
	PurchaseCost       domain.Money
	UsefulLifeMonths   int
	SalvageValue       domain.Money
	PurchaseMonthIndex domain.MonthIndex
	DepreciationMethod domain.DepreciationMethod
	AssociatedLoan     *domain.FinancingTerm
	CurrentStep        int
}

// CreateCapitalAssetResult wraps the newly-completed item's Result.
type CreateCapitalAssetResult struct {
	Result *common.CapitalAssetResult
}
