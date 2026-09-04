package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewCapitalAssetResultFromEntity builds a CapitalAssetResult from a
// CapitalAsset value object.
func NewCapitalAssetResultFromEntity(c domain.CapitalAsset) *common.CapitalAssetResult {
	return &common.CapitalAssetResult{
		ID:                 c.ID(),
		Name:               c.Name,
		PurchaseCost:       c.PurchaseCost,
		UsefulLifeMonths:   c.UsefulLifeMonths,
		SalvageValue:       c.SalvageValue,
		PurchaseMonthIndex: c.PurchaseMonthIndex,
		DepreciationMethod: c.DepreciationMethod,
		AssociatedLoan:     c.AssociatedLoan,
	}
}
