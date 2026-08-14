package interfaces

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type PlanService interface {
	// CreatePlan, UpdatePlan, and DeletePlan are the three lifecycle
	// commands owned by PlanController (creation, editing core details,
	// deletion). They are the only entry points that may construct or
	// mutate a Plan — callers must not call domain.NewPlan or
	// Plan.ChangeCoreDetails themselves. CreatePlan additionally verifies
	// cmd.OwnerID names a real, existing User (via UserRepository.GetUser)
	// before assigning them as owner — see domain.VerifiedUser.
	CreatePlan(cmd *commands.CreatePlan) (*commands.CreatePlanResult, error)
	UpdatePlan(cmd *commands.UpdatePlan) (*commands.UpdatePlanResult, error)
	DeletePlan(cmd *commands.DeletePlan) (*commands.DeletePlanResult, error)

	// Save persists an already-mutated Plan aggregate. It's the
	// general-purpose entry point every wizard controller (payroll, cash
	// flow, sales forecast, starting point, operating expenses, invites)
	// uses after fetching a Plan via Get and calling a domain mutation
	// method (AddX/ClearX/RecordUserInvited) on it directly — those
	// controllers own sub-entities within the Plan aggregate boundary and
	// are out of scope for the command wrapping above (see AGENTS.md).
	Save(p *domain.Plan) error
	Get(id uuid.UUID) (*domain.Plan, error)
	GetAll() ([]*domain.Plan, error)
	Delete(id uuid.UUID) error
	GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error)
}
