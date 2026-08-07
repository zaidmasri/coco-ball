package services

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type hubCompletionService struct {
	startingPoint     interfaces.StartingPointService
	payroll           interfaces.PayrollService
	salesForecast     interfaces.SalesForecastService
	cashFlow          interfaces.CashFlowService
	operatingExpenses interfaces.OperatingExpensesService
}

func NewHubCompletionService(
	startingPoint interfaces.StartingPointService,
	payroll interfaces.PayrollService,
	salesForecast interfaces.SalesForecastService,
	cashFlow interfaces.CashFlowService,
	operatingExpenses interfaces.OperatingExpensesService,
) interfaces.HubCompletionService {
	return &hubCompletionService{
		startingPoint:     startingPoint,
		payroll:           payroll,
		salesForecast:     salesForecast,
		cashFlow:          cashFlow,
		operatingExpenses: operatingExpenses,
	}
}

func (s *hubCompletionService) Get(planID uuid.UUID) interfaces.HubCompletion {
	sp, _ := s.startingPoint.GetHubStatus(planID)
	pr, _ := s.payroll.GetHubStatus(planID)
	sf, _ := s.salesForecast.GetHubStatus(planID)
	cf, _ := s.cashFlow.GetHubStatus(planID)
	oe, _ := s.operatingExpenses.GetHubStatus(planID)

	return interfaces.HubCompletion{
		StartingPoint:     allSectionsComplete(sp, domain.SectionFixedAssets, domain.SectionStartupCosts, domain.SectionFundingSources, domain.SectionCashOnHand),
		Payroll:           allSectionsComplete(pr, domain.SectionSalaryRoles, domain.SectionBenefits, domain.SectionPayrollTaxRates),
		SalesForecast:     allSectionsComplete(sf, domain.SectionProducts, domain.SectionSalesGrowthCurve),
		CashFlow:          allSectionsComplete(cf, domain.SectionInventoryPurchases, domain.SectionDistributions),
		OperatingExpenses: oe[domain.SectionOperatingExpenses],
	}
}

func allSectionsComplete(status map[string]bool, sections ...string) bool {
	for _, s := range sections {
		if !status[s] {
			return false
		}
	}
	return true
}
