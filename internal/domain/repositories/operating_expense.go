package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type OperatingExpenseRepository interface {
	CreateOperatingExpenseDraft(planID uuid.UUID) (uuid.UUID, error)
	GetOperatingExpenseDraft(planID uuid.UUID) (*OperatingExpenseItem, error)
	GetOperatingExpense(itemID uuid.UUID) (*OperatingExpenseItem, error)
	SaveOperatingExpenseStep(itemID uuid.UUID, cost domain.Cost, currentStep int, status ItemStatus) error
	ListCompleteOperatingExpenses(planID uuid.UUID) ([]OperatingExpenseItem, error)
	DeleteOperatingExpense(itemID uuid.UUID) error
}
