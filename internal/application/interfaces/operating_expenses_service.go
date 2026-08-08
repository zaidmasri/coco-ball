package interfaces

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// OperatingExpensesSummaryResult aggregates what the hub summary page needs.
type OperatingExpensesSummaryResult struct {
	SectionStatus        map[string]bool
	OperatingExpenseCount int
}

type OperatingExpensesService interface {
	// Hub summary
	GetSummary(planID uuid.UUID) (OperatingExpensesSummaryResult, error)
	GetHubStatus(planID uuid.UUID) (map[string]bool, error)
	MarkWizardSectionComplete(planID uuid.UUID, section string) error

	// Operating Expenses
	CreateOperatingExpenseDraft(planID uuid.UUID) (uuid.UUID, error)
	FindOperatingExpenseDraft(planID uuid.UUID) (*repositories.OperatingExpenseItem, error)
	GetOperatingExpense(itemID uuid.UUID) (*repositories.OperatingExpenseItem, error)
	SaveOperatingExpenseStep(itemID uuid.UUID, cost domain.Cost, currentStep int, status repositories.ItemStatus) error
	ListCompleteOperatingExpenses(planID uuid.UUID) ([]repositories.OperatingExpenseItem, error)
	DeleteOperatingExpense(itemID uuid.UUID) error
}
