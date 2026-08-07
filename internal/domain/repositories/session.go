package repositories

import "github.com/zaidmasri/business-planning-tool/internal/domain"

type SessionRepository interface {
	SaveSession(s *domain.Session) error
	GetSession(sessionID string) (*domain.Session, error)
	DeleteSession(sessionID string) error
}
