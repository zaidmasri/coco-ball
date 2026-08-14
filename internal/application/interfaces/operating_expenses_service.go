package interfaces

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
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

	// CreateOperatingExpense/UpdateOperatingExpense/DeleteOperatingExpense
	// are the only code allowed to construct or mutate the domain.Cost
	// entity for an Operating Expense item — see plan_growth.go's NewCost.
	CreateOperatingExpense(cmd *commands.CreateOperatingExpense) (*commands.CreateOperatingExpenseResult, error)
	UpdateOperatingExpense(cmd *commands.UpdateOperatingExpense) (*commands.UpdateOperatingExpenseResult, error)
	DeleteOperatingExpense(cmd *commands.DeleteOperatingExpense) (*commands.DeleteOperatingExpenseResult, error)
}
