// Package mapper converts domain entities into application-layer Result
// DTOs (internal/application/common). Keeping this conversion in one place
// means a display-oriented field (e.g. a computed label) can be added to a
// Result without touching the domain entity it's built from.
package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewPlanResultFromEntity builds a PlanResult from a Plan aggregate.
func NewPlanResultFromEntity(p *domain.Plan) *common.PlanResult {
	if p == nil {
		return nil
	}

	return &common.PlanResult{
		ID:            p.ID(),
		Name:          p.Name(),
		StartingMonth: p.StartingMonth(),
		StartingYear:  p.StartingYear(),
		OwnerID:       p.OwnerID(),
	}
}
