package services

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/application/mapper"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type planService struct {
	plans repositories.PlanRepository
	users repositories.UserRepository
}

func NewPlanService(plans repositories.PlanRepository, users repositories.UserRepository) interfaces.PlanService {
	return &planService{plans: plans, users: users}
}

// verifyOwner confirms ownerID names a real, existing User by looking it up
// via UserRepository, then wraps the result as a domain.VerifiedUser. This
// is the only place PlanService proves a claimed owner is actually a member
// of the system before letting domain.NewPlan assign them as owner.
func (s *planService) verifyOwner(ownerID uuid.UUID) (domain.VerifiedUser, error) {
	user, err := s.users.GetUser(ownerID)
	if err != nil {
		return domain.VerifiedUser{}, err
	}
	return domain.NewVerifiedUser(user)
}

func (s *planService) CreatePlan(cmd *commands.CreatePlan) (*commands.CreatePlanResult, error) {
	owner, err := s.verifyOwner(cmd.OwnerID)
	if err != nil {
		return nil, err
	}

	plan, err := domain.NewPlan(cmd.Name, cmd.StartingMonth, cmd.StartingYear, owner)
	if err != nil {
		return nil, err
	}

	validated, err := plan.Validate()
	if err != nil {
		return nil, err
	}
	if err := s.plans.SaveNew(validated, cmd.OwnerID); err != nil {
		return nil, err
	}

	return &commands.CreatePlanResult{Result: mapper.NewPlanResultFromEntity(plan)}, nil
}

func (s *planService) UpdatePlan(cmd *commands.UpdatePlan) (*commands.UpdatePlanResult, error) {
	plan, err := s.Get(cmd.PlanID)
	if err != nil {
		return nil, err
	}

	if err := plan.ChangeCoreDetails(cmd.Name, cmd.StartingMonth, cmd.StartingYear); err != nil {
		return nil, err
	}

	if err := s.Save(plan); err != nil {
		return nil, err
	}

	return &commands.UpdatePlanResult{Result: mapper.NewPlanResultFromEntity(plan)}, nil
}

func (s *planService) DeletePlan(cmd *commands.DeletePlan) (*commands.DeletePlanResult, error) {
	if err := s.Delete(cmd.PlanID); err != nil {
		return nil, err
	}
	return &commands.DeletePlanResult{Success: true}, nil
}

func (s *planService) Save(p *domain.Plan) error {
	validated, err := p.Validate()
	if err != nil {
		return err
	}
	return s.plans.Save(validated)
}
func (s *planService) Get(id uuid.UUID) (*domain.Plan, error) { return s.plans.Get(id) }
func (s *planService) GetAll() ([]*domain.Plan, error)        { return s.plans.GetAll() }
func (s *planService) Delete(id uuid.UUID) error              { return s.plans.Delete(id) }
func (s *planService) GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error) {
	return s.plans.GetUserPlans(userID)
}
