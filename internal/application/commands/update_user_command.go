package commands

import (
	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/common"
)

// UpdateUser is the command to change a user's first/last name.
// AuthService.UpdateUser loads the aggregate and calls User.ChangeName —
// callers must not mutate the entity themselves.
type UpdateUser struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
}

// UpdateUserResult wraps the updated user's Result.
type UpdateUserResult struct {
	Result *common.UserResult
}
