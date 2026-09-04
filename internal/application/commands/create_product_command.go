package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// CreateProduct is the command that finalizes a Products wizard item into a
// valid domain.Product for the first time. SalesForecastService.
// CreateProduct turns this into a domain.Product via domain.NewProduct
// (validating the accumulated wizard field values) and assigns it ItemID —
// the wizard row's pre-existing ID from CreateProductDraft — via
// Product.SetID, so the domain entity and its wizard row share one
// identity. Callers must not construct the entity themselves.
type CreateProduct struct {
	ItemID       uuid.UUID
	Name         string
	Month1Units  int
	PricePerUnit domain.Money
	CostPerUnit  domain.Money
	CurrentStep  int
}

// CreateProductResult wraps the newly-completed item's Result.
type CreateProductResult struct {
	Result *common.ProductResult
}
