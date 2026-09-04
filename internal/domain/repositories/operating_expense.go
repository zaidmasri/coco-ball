package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type OperatingExpenseRepository interface {
	CreateOperatingExpenseDraft(planID uuid.UUID) (uuid.UUID, error)
	FindOperatingExpenseDraft(planID uuid.UUID) (*OperatingExpenseItem, error)
	GetOperatingExpense(itemID uuid.UUID) (*OperatingExpenseItem, error)
	// SaveOperatingExpenseDraftStep persists an in-progress, possibly
	// incomplete wizard step. Drafts are allowed to be invalid by design, so
	// this accepts a bare domain.Cost.
	SaveOperatingExpenseDraftStep(itemID uuid.UUID, cost domain.Cost, currentStep int) error
	// CompleteOperatingExpense marks an operating expense item complete. It
	// requires a domain.ValidatedCost, so it is a compile error to complete
	// an item without validating it first - see domain.NewValidatedCost.
	CompleteOperatingExpense(itemID uuid.UUID, cost domain.ValidatedCost, currentStep int) error
	ListCompleteOperatingExpenses(planID uuid.UUID) ([]OperatingExpenseItem, error)
	DeleteOperatingExpense(itemID uuid.UUID) error
}
