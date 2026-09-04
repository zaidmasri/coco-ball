package common

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// ProductResult is the write-side acknowledgement shape for
// CreateProduct/UpdateProduct. It deliberately does not carry the
// repository-layer ProductItem (Status, CurrentStep) — read paths that need
// the full wizard row go through SalesForecastService.GetProduct/
// ListCompleteProducts instead. Mirrors SalaryRoleResult.
type ProductResult struct {
	ID           uuid.UUID
	Name         string
	Month1Units  int
	PricePerUnit domain.Money
	CostPerUnit  domain.Money
}
