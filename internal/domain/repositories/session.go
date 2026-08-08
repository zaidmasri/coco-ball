package repositories

import domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"

type SessionRepository interface {
	SaveSession(s *domain.Session) error
	GetSession(sessionID string) (*domain.Session, error)
	DeleteSession(sessionID string) error
}
