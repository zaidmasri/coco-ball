package services

import (
	"time"

	"uuid"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	"github.com/zaidmasri/business-planning-tool/internal/application/mapper"
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

func (s *authService) GetUser(id uuid.UUID) (*domain.User, error) { return s.users.GetUser(id) }
func (s *authService) GetUserByEmail(email string) (*domain.User, error) {
	return s.users.GetUserByEmail(email)
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

func (s *authService) DeleteSession(sessionID string) error {
	return s.sessions.DeleteSession(sessionID)
}

func (s *authService) CreateSession(userID uuid.UUID, duration time.Duration) (*domain.Session, error) {
	sess, err := domain.NewSession(userID, duration)
	if err != nil {
		return nil, err
	}
	if err := s.sessions.SaveSession(sess); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *authService) CreateUser(cmd *commands.CreateUser) (*commands.CreateUserResult, error) {
	userCreds, err := domain.NewUserWithPassword(cmd.Email, cmd.FirstName, cmd.LastName, cmd.Password)
	if err != nil {
		return nil, err
	}

	validated, err := domain.NewValidatedUserWithPassword(userCreds)
	if err != nil {
		return nil, err
	}

	if err := s.users.SaveUserWithPassword(validated); err != nil {
		return nil, err
	}

	return &commands.CreateUserResult{Result: mapper.NewUserResultFromEntity(userCreds.User)}, nil
}

func (s *authService) UpdateUser(cmd *commands.UpdateUser) (*commands.UpdateUserResult, error) {
	user, err := s.users.GetUser(cmd.UserID)
	if err != nil {
		return nil, err
	}

	if err := user.ChangeName(cmd.FirstName, cmd.LastName); err != nil {
		return nil, err
	}

	validated, err := domain.NewValidatedUser(user)
	if err != nil {
		return nil, err
	}

	if err := s.users.UpdateUser(validated); err != nil {
		return nil, err
	}

	return &commands.UpdateUserResult{Result: mapper.NewUserResultFromEntity(user)}, nil
}

func (s *authService) DeleteUser(cmd *commands.DeleteUser) (*commands.DeleteUserResult, error) {
	if err := s.users.DeleteUser(cmd.UserID); err != nil {
		return nil, err
	}
	return &commands.DeleteUserResult{Success: true}, nil
}
