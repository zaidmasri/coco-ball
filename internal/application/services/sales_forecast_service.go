package services

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type salesForecastService struct {
	products         repositories.ProductRepository
	salesGrowthCurve repositories.SalesGrowthCurveRepository
	wizardProgress   repositories.WizardProgressRepository
}

func NewSalesForecastService(
	products repositories.ProductRepository,
	salesGrowthCurve repositories.SalesGrowthCurveRepository,
	wizardProgress repositories.WizardProgressRepository,
) interfaces.SalesForecastService {
	return &salesForecastService{
		products:         products,
		salesGrowthCurve: salesGrowthCurve,
		wizardProgress:   wizardProgress,
	}
}

func (s *salesForecastService) GetHubStatus(planID uuid.UUID) (map[string]bool, error) {
	return s.wizardProgress.GetWizardSectionStatus(planID, domain.HubSalesForecast)
}

func (s *salesForecastService) MarkWizardSectionComplete(planID uuid.UUID, section string) error {
	return s.wizardProgress.MarkWizardSectionComplete(planID, domain.HubSalesForecast, section)
}

func (s *salesForecastService) GetSummary(planID uuid.UUID) (interfaces.SalesForecastSummaryResult, error) {
	sectionStatus, err := s.wizardProgress.GetWizardSectionStatus(planID, domain.HubSalesForecast)
	if err != nil {
		sectionStatus = map[string]bool{}
	}
	products, err := s.products.ListCompleteProducts(planID)
	if err != nil {
		return interfaces.SalesForecastSummaryResult{}, err
	}
	return interfaces.SalesForecastSummaryResult{
		SectionStatus: sectionStatus,
		ProductCount:  len(products),
	}, nil
}

func (s *salesForecastService) CreateProductDraft(planID uuid.UUID) (uuid.UUID, error) {
	return s.products.CreateProductDraft(planID)
}
func (s *salesForecastService) FindProductDraft(planID uuid.UUID) (*repositories.ProductItem, error) {
	return s.products.FindProductDraft(planID)
}
func (s *salesForecastService) GetProduct(itemID uuid.UUID) (*repositories.ProductItem, error) {
	return s.products.GetProduct(itemID)
}
func (s *salesForecastService) SaveProductStep(itemID uuid.UUID, product domain.Product, currentStep int, status repositories.ItemStatus) error {
	return s.products.SaveProductStep(itemID, product, currentStep, status)
}
func (s *salesForecastService) ListCompleteProducts(planID uuid.UUID) ([]repositories.ProductItem, error) {
	return s.products.ListCompleteProducts(planID)
}
func (s *salesForecastService) DeleteProduct(itemID uuid.UUID) error {
	return s.products.DeleteProduct(itemID)
}

func (s *salesForecastService) GetSalesGrowthCurveRow(planID uuid.UUID) (*repositories.SalesGrowthCurveRow, error) {
	return s.salesGrowthCurve.GetSalesGrowthCurveRow(planID)
}
func (s *salesForecastService) SaveSalesGrowthCurveStep(planID uuid.UUID, curve domain.SalesGrowthCurve, currentStep int) error {
	return s.salesGrowthCurve.SaveSalesGrowthCurveStep(planID, curve, currentStep)
}
