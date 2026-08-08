package repositories

import (
	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type UserRepository interface {
	SaveUser(u *domain.User) error
	GetUser(id uuid.UUID) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	SaveUserWithPassword(u *domain.UserWithPassword) error
	GetUserWithPassword(email string) (*domain.UserWithPassword, error)
}
