package services

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/application/mapper"
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
func (s *salesForecastService) SaveProductDraftStep(itemID uuid.UUID, product domain.Product, currentStep int) error {
	return s.products.SaveProductDraftStep(itemID, product, currentStep)
}
func (s *salesForecastService) ListCompleteProducts(planID uuid.UUID) ([]repositories.ProductItem, error) {
	return s.products.ListCompleteProducts(planID)
}
func (s *salesForecastService) DeleteProduct(itemID uuid.UUID) error {
	return s.products.DeleteProduct(itemID)
}

func (s *salesForecastService) CreateProduct(cmd *commands.CreateProduct) (*commands.CreateProductResult, error) {
	product, err := domain.NewProduct(cmd.Name, cmd.Month1Units, cmd.PricePerUnit, cmd.CostPerUnit)
	if err != nil {
		return nil, err
	}
	product.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedProduct(product)
	if err != nil {
		return nil, err
	}

	if err := s.products.CompleteProduct(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.CreateProductResult{Result: mapper.NewProductResultFromEntity(product)}, nil
}

func (s *salesForecastService) UpdateProduct(cmd *commands.UpdateProduct) (*commands.UpdateProductResult, error) {
	product, err := domain.NewProduct(cmd.Name, cmd.Month1Units, cmd.PricePerUnit, cmd.CostPerUnit)
	if err != nil {
		return nil, err
	}
	product.SetID(cmd.ItemID)

	validated, err := domain.NewValidatedProduct(product)
	if err != nil {
		return nil, err
	}

	if err := s.products.CompleteProduct(cmd.ItemID, validated, cmd.CurrentStep); err != nil {
		return nil, err
	}

	return &commands.UpdateProductResult{Result: mapper.NewProductResultFromEntity(product)}, nil
}

func (s *salesForecastService) GetSalesGrowthCurveRow(planID uuid.UUID) (*repositories.SalesGrowthCurveRow, error) {
	return s.salesGrowthCurve.GetSalesGrowthCurveRow(planID)
}
func (s *salesForecastService) SaveSalesGrowthCurveStep(planID uuid.UUID, curve domain.SalesGrowthCurve, currentStep int, isComplete bool) error {
	if isComplete {
		if err := domain.ValidateSalesGrowthCurve(curve); err != nil {
			return err
		}
	}
	return s.salesGrowthCurve.SaveSalesGrowthCurveStep(planID, curve, currentStep)
}
