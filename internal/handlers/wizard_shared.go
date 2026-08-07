// Package handlers - shared helpers used across the Starting Point,
// Payroll, Sales Forecast, Operating Expenses, and Cash Flow wizard hubs.
package handlers

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
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

// hubCompletion returns the completion state of all five wizard hubs for
// pages that sit outside any single hub (Setup, Income Statement, Balance
// Sheet, Analytics) but whose sidebar still needs every icon's fill state.
func (app *App) hubCompletion(planID uuid.UUID) views.HubCompletion {
	return toViewsHubCompletion(app.HubSvc.Get(planID))
}
