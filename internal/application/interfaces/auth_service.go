package interfaces

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type AuthService interface {
	SaveUser(u *domain.User) error
	GetUser(id uuid.UUID) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	SaveUserWithPassword(u *domain.UserWithPassword) error
	GetUserWithPassword(email string) (*domain.UserWithPassword, error)
	SaveSession(s *domain.Session) error
	GetSession(sessionID string) (*domain.Session, error)
	DeleteSession(sessionID string) error

	// CreateUser, UpdateUser, and DeleteUser are the three lifecycle
	// commands owned by AuthController (signup, profile name edit, account
	// deletion). They are the only entry points that may construct or
	// mutate a User — callers must not call domain.NewUserWithPassword or
	// User.ChangeName themselves.
	CreateUser(cmd *commands.CreateUser) (*commands.CreateUserResult, error)
	UpdateUser(cmd *commands.UpdateUser) (*commands.UpdateUserResult, error)
	DeleteUser(cmd *commands.DeleteUser) (*commands.DeleteUserResult, error)
}
