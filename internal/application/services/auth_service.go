package services

import (
	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

type authService struct {
	users    repositories.UserRepository
	sessions repositories.SessionRepository
}

func NewAuthService(users repositories.UserRepository, sessions repositories.SessionRepository) interfaces.AuthService {
	return &authService{users: users, sessions: sessions}
}

func (s *authService) SaveUser(u *domain.User) error                 { return s.users.SaveUser(u) }
func (s *authService) GetUser(id uuid.UUID) (*domain.User, error)   { return s.users.GetUser(id) }
func (s *authService) GetUserByEmail(email string) (*domain.User, error) {
	return s.users.GetUserByEmail(email)
}
func (s *authService) SaveUserWithPassword(u *domain.UserWithPassword) error {
	return s.users.SaveUserWithPassword(u)
}
func (s *authService) GetUserWithPassword(email string) (*domain.UserWithPassword, error) {
	return s.users.GetUserWithPassword(email)
}
func (s *authService) SaveSession(sess *domain.Session) error { return s.sessions.SaveSession(sess) }
func (s *authService) GetSession(sessionID string) (*domain.Session, error) {
	sess, err := s.sessions.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if !sess.IsValid() {
		return nil, domain.ErrSessionExpired
	}
	return sess, nil
}
func (s *authService) DeleteSession(sessionID string) error { return s.sessions.DeleteSession(sessionID) }
