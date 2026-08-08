// Package web - shared helpers used across every controller, and in
// particular the Starting Point, Payroll, Sales Forecast, Operating
// Expenses, and Cash Flow wizard hubs.
package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// parsePlanID extracts and parses the plan ID from the request path
func parsePlanID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue("id"))
}

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

// parseStepMoney parses a required, non-negative money field submitted by
// a single wizard step.
func parseStepMoney(raw string) (domain.Money, bool) {
	val, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || val < 0 {
		return domain.Money{}, false
	}
	m, err := domain.NewMoney(int64(val), domain.USD)
	if err != nil {
		return domain.Money{}, false
	}
	return m, true
}

func stepIndex(steps []string, current string) int {
	for i, s := range steps {
		if s == current {
			return i
		}
	}
	return -1
}

func nextStepName(steps []string, idx int) string {
	if idx+1 >= len(steps) {
		return ""
	}
	return steps[idx+1]
}

func prevStepName(steps []string, idx int) string {
	if idx <= 0 {
		return ""
	}
	return steps[idx-1]
}
