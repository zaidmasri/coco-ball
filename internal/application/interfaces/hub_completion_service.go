package interfaces

import "github.com/google/uuid"

// HubCompletion carries all five hubs' completion state for pages that
// sit outside a single hub (Setup, reports) but whose sidebar needs all
// icon fill states to be accurate.
type HubCompletion struct {
	StartingPoint     bool
	Payroll           bool
	SalesForecast     bool
	OperatingExpenses bool
	CashFlow          bool
}

type HubCompletionService interface {
	Get(planID uuid.UUID) HubCompletion
}
