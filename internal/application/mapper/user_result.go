package mapper

import (
	"github.com/zaidmasri/business-planning-tool/internal/application/common"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

// NewUserResultFromEntity builds a UserResult from a User entity.
func NewUserResultFromEntity(u *domain.User) *common.UserResult {
	return &common.UserResult{
		ID:        u.ID(),
		Email:     u.Email(),
		FirstName: u.FirstName(),
		LastName:  u.LastName(),
	}
}
