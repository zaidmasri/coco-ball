package repositories

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type OperatingExpenseRepository interface {
	CreateOperatingExpenseDraft(planID uuid.UUID) (uuid.UUID, error)
	FindOperatingExpenseDraft(planID uuid.UUID) (*OperatingExpenseItem, error)
	GetOperatingExpense(itemID uuid.UUID) (*OperatingExpenseItem, error)
	SaveOperatingExpenseStep(itemID uuid.UUID, cost domain.Cost, currentStep int, status ItemStatus) error
	ListCompleteOperatingExpenses(planID uuid.UUID) ([]OperatingExpenseItem, error)
	DeleteOperatingExpense(itemID uuid.UUID) error
}
