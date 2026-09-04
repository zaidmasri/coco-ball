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

// UserRepository implements repositories.UserRepository.
type UserRepository struct {
	conn    *sql.DB
	queries *db.Queries
}

func NewUserRepository(conn *sql.DB) repositories.UserRepository {
	return &UserRepository{conn: conn, queries: db.New(conn)}
}

var _ repositories.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) GetUser(id uuid.UUID) (*domain.User, error) {
	row, err := r.queries.GetUserByID(context.Background(), id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user, err := domain.NewUser(row.Email)
	if err != nil {
		return nil, err
	}
	user.SetID(id)
	user.SetFirstName(row.FirstName)
	user.SetLastName(row.LastName)

	return user, nil
}

func (r *UserRepository) GetUserByEmail(email string) (*domain.User, error) {
	idStr, err := r.queries.GetUserIDByEmail(context.Background(), email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}

	return r.GetUser(id)
}

func (r *UserRepository) SaveUserWithPassword(vu domain.ValidatedUserWithPassword) error {
	u := vu.User()
	ctx := context.Background()

	tx, err := r.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	now := time.Now().Unix()

	if err := qtx.CreateUser(ctx, db.CreateUserParams{
		ID:        u.ID().String(),
		Email:     u.Email(),
		FirstName: u.FirstName(),
		LastName:  u.LastName(),
		CreatedAt: now,
	}); err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	if err := qtx.UpsertUserCredentials(ctx, db.UpsertUserCredentialsParams{
		Email:        u.Email(),
		PasswordHash: u.PasswordHash,
		CreatedAt:    now,
	}); err != nil {
		return fmt.Errorf("failed to save user credentials: %w", err)
	}

	if err := insertOutboxEvents(ctx, qtx, u.PullEvents(), now); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) GetUserWithPassword(email string) (*domain.UserWithPassword, error) {
	row, err := r.queries.GetUserCredentialsByEmail(context.Background(), email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user credentials: %w", err)
	}

	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user id: %w", err)
	}

	user, err := r.GetUser(id)
	if err != nil {
		return nil, err
	}

	return &domain.UserWithPassword{User: user, PasswordHash: row.PasswordHash}, nil
}

// UpdateUser persists a user's changed first/last name and atomically writes
// any accumulated domain events to the outbox table, in one transaction.
// Mirrors PlanRepository.Save's shape.
func (r *UserRepository) UpdateUser(vu domain.ValidatedUser) error {
	u := vu.User()
	ctx := context.Background()

	tx, err := r.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	now := time.Now().Unix()

	if err := qtx.UpdateUserName(ctx, db.UpdateUserNameParams{
		FirstName: u.FirstName(),
		LastName:  u.LastName(),
		ID:        u.ID().String(),
	}); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if err := insertOutboxEvents(ctx, qtx, u.PullEvents(), now); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteUser soft-deletes a user and hard-deletes every row referencing them
// that isn't cascaded by the database itself (foreign_keys is not enabled on
// the connection — see connection.go), in one transaction. Plans the user
// owns are left untouched: their plan_access row for this user is removed
// like any other, but the plan itself is not deleted or reassigned.
func (r *UserRepository) DeleteUser(id uuid.UUID) error {
	ctx := context.Background()

	user, err := r.GetUser(id)
	if err != nil {
		return err
	}

	tx, err := r.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	idStr := id.String()

	if err := qtx.DeleteSessionsByUser(ctx, idStr); err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	if err := qtx.DeletePlanAccessByUser(ctx, idStr); err != nil {
		return fmt.Errorf("failed to delete user plan access: %w", err)
	}
	if err := qtx.DeleteUserCredentialsByEmail(ctx, user.Email()); err != nil {
		return fmt.Errorf("failed to delete user credentials: %w", err)
	}

	affected, err := qtx.SoftDeleteUser(ctx, db.SoftDeleteUserParams{
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        idStr,
	})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if affected == 0 {
		return domain.ErrUserNotFound
	}

	return tx.Commit()
}
