package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewProductResultFromEntity builds a ProductResult from a Product value
// object.
func NewProductResultFromEntity(p domain.Product) *common.ProductResult {
	return &common.ProductResult{
		ID:           p.ID(),
		Name:         p.Name,
		Month1Units:  p.Month1Units,
		PricePerUnit: p.PricePerUnit,
		CostPerUnit:  p.CostPerUnit,
	}
}
