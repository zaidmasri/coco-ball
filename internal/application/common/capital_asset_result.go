package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CapitalAssetResult is the write-side acknowledgement shape for
// CreateCapitalAsset/UpdateCapitalAsset. It deliberately does not carry the
// repository-layer CapitalAssetItem (Status, CurrentStep) — read paths
// that need the full wizard row go through StartingPointService.
// GetCapitalAsset/ListCompleteCapitalAssets instead. Mirrors
// SalaryRoleResult.
type CapitalAssetResult struct {
	ID                 uuid.UUID
	Name               string
	PurchaseCost       domain.Money
	UsefulLifeMonths   int
	SalvageValue       domain.Money
	PurchaseMonthIndex domain.MonthIndex
	DepreciationMethod domain.DepreciationMethod
	AssociatedLoan     *domain.FinancingTerm
}
