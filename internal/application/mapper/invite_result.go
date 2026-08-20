package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewInviteResultFromEntity builds an InviteResult from a PlanInvite.
func NewInviteResultFromEntity(i *domain.PlanInvite) *common.InviteResult {
	if i == nil {
		return nil
	}

	return &common.InviteResult{
		ID:          i.ID,
		PlanID:      i.PlanID,
		Email:       i.Email,
		AccessLevel: i.AccessLevel,
		Status:      i.Status,
		InvitedBy:   i.InvitedBy,
	}
}
