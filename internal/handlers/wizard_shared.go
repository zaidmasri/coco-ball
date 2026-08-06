// Package handlers - shared helpers used across the Starting Point,
// Payroll, Sales Forecast, Operating Expenses, and Cash Flow wizard hubs.
package handlers

import (
	"log"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// parseStepPercent parses a percent-scale wizard step field (e.g. "6.5" for
// 6.5%) into its 0-1 domain scale. An empty field is valid and means 0 -
// every "growth after Year 1" step is optional, since leaving it blank
// just means no growth that year. Unlike parseStepMoney, negative values
// are also valid (a cost or headcount can't be negative, but a growth
// rate can be).
func parseStepPercent(raw string) (float64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, true
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return val / 100.0, true
}

// hubCompletion computes all five wizard hubs' completion state for
// planID at once, for pages that sit outside any single hub (Setup,
// Income Statement, Balance Sheet, Analytics) but whose sidebar still
// needs every icon's fill state to be accurate.
func (app *App) hubCompletion(planID uuid.UUID) views.HubCompletion {
	sp, err := app.Store.GetWizardSectionStatus(planID, domain.HubStartingPoint)
	if err != nil {
		log.Printf("Failed to load starting point section status for plan %s: %v", planID, err)
		sp = map[string]bool{}
	}
	pr, err := app.Store.GetWizardSectionStatus(planID, domain.HubPayroll)
	if err != nil {
		log.Printf("Failed to load payroll section status for plan %s: %v", planID, err)
		pr = map[string]bool{}
	}
	sf, err := app.Store.GetWizardSectionStatus(planID, domain.HubSalesForecast)
	if err != nil {
		log.Printf("Failed to load sales forecast section status for plan %s: %v", planID, err)
		sf = map[string]bool{}
	}
	oe, err := app.Store.GetWizardSectionStatus(planID, domain.HubOperatingExpenses)
	if err != nil {
		log.Printf("Failed to load operating expenses section status for plan %s: %v", planID, err)
		oe = map[string]bool{}
	}
	cf, err := app.Store.GetWizardSectionStatus(planID, domain.HubCashFlow)
	if err != nil {
		log.Printf("Failed to load cash flow section status for plan %s: %v", planID, err)
		cf = map[string]bool{}
	}

	return views.HubCompletion{
		StartingPoint:     views.IsStartingPointComplete(sp),
		Payroll:           views.IsPayrollComplete(pr),
		SalesForecast:     views.IsSalesForecastComplete(sf),
		OperatingExpenses: views.IsOperatingExpensesComplete(oe),
		CashFlow:          views.IsCashFlowComplete(cf),
	}
}
