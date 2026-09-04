package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// UpdateCapitalAsset is the command to change an already-complete Fixed
// Assets item's fields. StartingPointService.UpdateCapitalAsset
// re-validates the fields via domain.NewCapitalAsset and assigns the
// result ItemID — callers must not construct the entity themselves.
type UpdateCapitalAsset struct {
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

// UpdateCapitalAssetResult wraps the updated item's Result.
type UpdateCapitalAssetResult struct {
	Result *common.CapitalAssetResult
}
