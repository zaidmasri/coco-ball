package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// SessionRepository implements repositories.SessionRepository.
type SessionRepository struct {
	queries *db.Queries
}

func NewSessionRepository(conn *sql.DB) repositories.SessionRepository {
	return &SessionRepository{queries: db.New(conn)}
}

var _ repositories.SessionRepository = (*SessionRepository)(nil)

func (r *SessionRepository) SaveSession(s *domain.Session) error {
	if err := r.queries.SaveSession(context.Background(), db.SaveSessionParams{
		ID:        s.ID,
		UserID:    s.UserID.String(),
		CreatedAt: s.CreatedAt.Unix(),
		ExpiresAt: s.ExpiresAt.Unix(),
	}); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetSession(sessionID string) (*domain.Session, error) {
	row, err := r.queries.GetSession(context.Background(), sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	userID, err := uuid.Parse(row.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse session user id: %w", err)
	}

	return &domain.Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Unix(row.CreatedAt, 0),
		ExpiresAt: time.Unix(row.ExpiresAt, 0),
	}, nil
}

func (r *SessionRepository) DeleteSession(sessionID string) error {
	if err := r.queries.DeleteSession(context.Background(), sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
