package services

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
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
		StartingPoint:     allSectionsComplete(domain.HubStartingPoint, sp),
		Payroll:           allSectionsComplete(domain.HubPayroll, pr),
		SalesForecast:     allSectionsComplete(domain.HubSalesForecast, sf),
		CashFlow:          allSectionsComplete(domain.HubCashFlow, cf),
		OperatingExpenses: oe[domain.SectionOperatingExpenses],
	}
}

// allSectionsComplete reports whether every section domain.HubSections lists
// for hub is marked complete in status - the single source of truth for hub
// membership, not a call-site-local list.
func allSectionsComplete(hub string, status map[string]bool) bool {
	for _, s := range domain.HubSections[hub] {
		if !status[s] {
			return false
		}
	}
	return true
}
