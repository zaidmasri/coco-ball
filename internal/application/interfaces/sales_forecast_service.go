package interfaces

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
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
	GetProductDraft(planID uuid.UUID) (*repositories.ProductItem, error)
	GetProduct(itemID uuid.UUID) (*repositories.ProductItem, error)
	SaveProductStep(itemID uuid.UUID, product domain.Product, currentStep int, status repositories.ItemStatus) error
	ListCompleteProducts(planID uuid.UUID) ([]repositories.ProductItem, error)
	DeleteProduct(itemID uuid.UUID) error

	// Sales Growth Curve (singleton)
	GetSalesGrowthCurveRow(planID uuid.UUID) (*repositories.SalesGrowthCurveRow, error)
	SaveSalesGrowthCurveStep(planID uuid.UUID, curve domain.SalesGrowthCurve, currentStep int) error
}
