package views

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// TemplateFuncs are helper functions available to every parsed page
// template — mainly formatting helpers for the read-only financial reports
// (Income Statement, Balance Sheet, Analytics), since Go's html/template has
// no built-in arithmetic or currency formatting.
var TemplateFuncs = template.FuncMap{
	"formatMoney":   formatMoney,
	"addMoney":      addMoney,
	"formatPercent": formatPercent,
	"formatRatio":   formatRatio,
}

// addMoney sums any number of domain.Money values, for templates that need
// to combine two report lines (e.g. Payroll + Payroll Tax).
func addMoney(values ...domain.Money) domain.Money {
	var total domain.Money
	for _, v := range values {
		total += v
	}
	return total
}

// formatMoney renders a whole-dollar Money value as US currency, e.g.
// "$1,234.00" or "-$500.00" for a negative amount.
func formatMoney(m domain.Money) string {
	neg := m < 0
	if neg {
		m = -m
	}
	s := "$" + groupThousands(int64(m)) + ".00"
	if neg {
		return "-" + s
	}
	return s
}

func groupThousands(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}

	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}

// formatPercent renders a float64 percentage (already scaled, e.g. 12.5 for
// 12.5%) to one decimal place.
func formatPercent(f float64) string {
	return fmt.Sprintf("%.1f%%", f)
}

// formatRatio renders a float64 ratio (e.g. current ratio, debt-to-equity)
// to two decimal places.
func formatRatio(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
