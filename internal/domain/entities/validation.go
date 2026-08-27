package entities

import (
	"net/mail"
	"strings"
)

const (
	maxNameLength = 200

	// maxMoneyAmount caps a single line-item amount. Money's minor unit in
	// this app is a whole dollar, not a cent (see money.go's doc comment),
	// so this is a $100,000,000 ceiling — generous for any one line item in
	// a small-business plan, but tight enough to catch fat-fingered entries
	// (e.g. an extra zero or two).
	maxMoneyAmount int64 = 100_000_000

	// minGrowthRate/maxGrowthRate bound a compounding annual growth/decline
	// rate (0-1 scale, e.g. 0.05 = 5%). -100% is the floor: a rate below
	// that would flip the compounded amount's sign, which is meaningless
	// for a cost/revenue/headcount-style line item. +1000% is a generous
	// sanity cap, not a business limit — it exists to catch entries like
	// "500" typed into a field expecting "5" (i.e. 500% instead of 5%).
	minGrowthRate = -1.0
	maxGrowthRate = 10.0

	// maxPercentRate bounds a plain 0-100% rate (payroll/interest rates,
	// never negative, never over 100%).
	maxPercentRate = 1.0
)

// validateRequiredName trims name and rejects it if empty or over the
// shared max length, returning the cleaned value for callers that want it.
func validateRequiredName(name string) (string, error) {
	cleaned := strings.TrimSpace(name)
	if cleaned == "" {
		return "", ErrInvalidName
	}
	if len(cleaned) > maxNameLength {
		return "", ErrNameTooLong
	}
	return cleaned, nil
}

// validateMoneyAmount rejects a negative or implausibly large amount.
func validateMoneyAmount(m Money) error {
	if m.IsNegative() {
		return ErrNegativeAmount
	}
	if m.MinorUnits() > maxMoneyAmount {
		return ErrAmountTooLarge
	}
	return nil
}

// validateGrowthRate bounds a compounding annual growth/decline rate.
func validateGrowthRate(rate float64) error {
	if rate < minGrowthRate || rate > maxGrowthRate {
		return ErrInvalidGrowthRate
	}
	return nil
}

// validatePercentRate bounds a plain 0-100% rate (payroll tax, interest).
func validatePercentRate(rate float64) error {
	if rate < 0 || rate > maxPercentRate {
		return ErrInvalidRate
	}
	return nil
}

// validateEmailFormat checks email is a syntactically valid address, using
// the standard library's RFC 5322 parser rather than a hand-rolled regex —
// keeps the domain layer's zero-third-party-dependency rule intact.
func validateEmailFormat(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmailFormat
	}
	return nil
}
