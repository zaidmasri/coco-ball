package commands

import "github.com/zaidmasri/business-planning-tool/internal/application/common"

// CreateUser is the command to register a new account. AuthService.
// CreateUser turns this into a validated domain.UserWithPassword via
// domain.NewUserWithPassword — callers must not construct the entity
// themselves.
type CreateUser struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
}

// CreateUserResult wraps the newly-created user's Result. The controller
// uses Result.ID to create the login session.
type CreateUserResult struct {
	Result *common.UserResult
}
