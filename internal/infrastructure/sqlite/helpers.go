package sqlite

import "github.com/zaidmasri/business-planning-tool/internal/domain/entities"

// mustUSD wraps a raw minor-units integer into a USD domain.Money. USD is
// always a supported currency, so the error from NewMoney is unreachable.
func mustUSD(minorUnits int64) entities.Money {
	m, _ := entities.NewMoney(minorUnits, entities.USD)
	return m
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func int64ToBool(n int64) bool {
	return n != 0
}
