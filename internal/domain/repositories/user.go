package repositories

import (
	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
)

type UserRepository interface {
	GetUser(id uuid.UUID) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	SaveUserWithPassword(u *domain.UserWithPassword) error
	GetUserWithPassword(email string) (*domain.UserWithPassword, error)
	UpdateUser(u *domain.User) error
	DeleteUser(id uuid.UUID) error
}
