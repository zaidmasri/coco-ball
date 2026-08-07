package repositories

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type UserRepository interface {
	SaveUser(u *domain.User) error
	GetUser(id uuid.UUID) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	SaveUserWithPassword(u *domain.UserWithPassword) error
	GetUserWithPassword(email string) (*domain.UserWithPassword, error)
}
