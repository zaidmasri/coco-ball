package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// UpdateProduct is the command to change an already-complete Products
// item's fields. SalesForecastService.UpdateProduct re-validates the
// fields via domain.NewProduct and assigns the result ItemID — callers must
// not construct the entity themselves.
type UpdateProduct struct {
	ItemID       uuid.UUID
	Name         string
	Month1Units  int
	PricePerUnit domain.Money
	CostPerUnit  domain.Money
	CurrentStep  int
}

// UpdateProductResult wraps the updated item's Result.
type UpdateProductResult struct {
	Result *common.ProductResult
}
