package interfaces

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// SalesForecastSummaryResult aggregates what the hub summary page needs.
type SalesForecastSummaryResult struct {
	SectionStatus map[string]bool
	ProductCount  int
}

type SalesForecastService interface {
	// Hub summary
	GetSummary(planID uuid.UUID) (SalesForecastSummaryResult, error)
	GetHubStatus(planID uuid.UUID) (map[string]bool, error)
	MarkWizardSectionComplete(planID uuid.UUID, section string) error

	// Products
	CreateProductDraft(planID uuid.UUID) (uuid.UUID, error)
	FindProductDraft(planID uuid.UUID) (*repositories.ProductItem, error)
	GetProduct(itemID uuid.UUID) (*repositories.ProductItem, error)
	SaveProductDraftStep(itemID uuid.UUID, product domain.Product, currentStep int) error
	ListCompleteProducts(planID uuid.UUID) ([]repositories.ProductItem, error)
	DeleteProduct(itemID uuid.UUID) error

	// CreateProduct/UpdateProduct are the only code allowed to construct or
	// mutate the domain.Product entity - see plan_sales_forecast.go's
	// NewProduct.
	CreateProduct(cmd *commands.CreateProduct) (*commands.CreateProductResult, error)
	UpdateProduct(cmd *commands.UpdateProduct) (*commands.UpdateProductResult, error)

	// Sales Growth Curve (singleton)
	GetSalesGrowthCurveRow(planID uuid.UUID) (*repositories.SalesGrowthCurveRow, error)
	SaveSalesGrowthCurveStep(planID uuid.UUID, curve domain.SalesGrowthCurve, currentStep int, isComplete bool) error
}
