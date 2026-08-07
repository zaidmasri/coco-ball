package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// Compile-time assertions: SQLiteStore must satisfy all repository interfaces.
var (
	_ repositories.PlanRepository             = (*SQLiteStore)(nil)
	_ repositories.UserRepository             = (*SQLiteStore)(nil)
	_ repositories.SessionRepository          = (*SQLiteStore)(nil)
	_ repositories.AccessRepository           = (*SQLiteStore)(nil)
	_ repositories.InviteRepository           = (*SQLiteStore)(nil)
	_ repositories.WizardProgressRepository   = (*SQLiteStore)(nil)
	_ repositories.CapitalAssetRepository     = (*SQLiteStore)(nil)
	_ repositories.StartupCostRepository      = (*SQLiteStore)(nil)
	_ repositories.FundingSourceRepository    = (*SQLiteStore)(nil)
	_ repositories.StartingBalancesRepository = (*SQLiteStore)(nil)
	_ repositories.SalaryRoleRepository       = (*SQLiteStore)(nil)
	_ repositories.BenefitRepository          = (*SQLiteStore)(nil)
	_ repositories.PayrollTaxRatesRepository  = (*SQLiteStore)(nil)
	_ repositories.ProductRepository          = (*SQLiteStore)(nil)
	_ repositories.SalesGrowthCurveRepository = (*SQLiteStore)(nil)
	_ repositories.InventoryPurchaseRepository = (*SQLiteStore)(nil)
	_ repositories.DistributionRepository     = (*SQLiteStore)(nil)
	_ repositories.OperatingExpenseRepository = (*SQLiteStore)(nil)
)

// SQLiteStore implements all repository interfaces defined in domain/repositories.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store and runs migrations
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &SQLiteStore{db: db}

	// Run migrations
	if err := store.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// runMigrations executes all pending migrations
func (s *SQLiteStore) runMigrations() error {
	migrations := GetMigrations()

	for i, migration := range migrations {
		version := i + 1

		// Check if migration already ran
		var exists int
		err := s.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&exists)

		// First migration creates schema_migrations table, so we expect an error there
		if err != nil && version > 1 && !strings.Contains(err.Error(), "no such table") {
			return err
		}

		// Skip if already applied (only check if table exists)
		if err == nil && exists > 0 {
			continue
		}

		// Execute migration - either a raw SQL statement or a Go func
		// (used for data backfills raw SQL can't express)
		if migration.Fn != nil {
			if err := migration.Fn(s.db); err != nil {
				return fmt.Errorf("migration %d failed: %w", version, err)
			}
		} else if _, err := s.db.Exec(migration.SQL); err != nil {
			return fmt.Errorf("migration %d failed: %w", version, err)
		}

		// Mark migration as applied
		if _, err := s.db.Exec(
			"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
			version,
			time.Now().Unix(),
		); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", version, err)
		}
	}

	return nil
}

// Save stores a plan
func (s *SQLiteStore) Save(p *domain.Plan) error {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	now := time.Now().Unix()
	_, err = s.db.Exec(
		`INSERT INTO plans (id, data, created_at, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET data=excluded.data, updated_at=excluded.updated_at`,
		p.ID().String(),
		data,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	return nil
}

// Get retrieves a plan by ID
func (s *SQLiteStore) Get(id uuid.UUID) (*domain.Plan, error) {
	var data []byte

	err := s.db.QueryRow("SELECT data FROM plans WHERE id = ?", id.String()).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, errors.New("plan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query plan: %w", err)
	}

	var plan domain.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan: %w", err)
	}

	if err := s.loadStartingPointInto(&plan); err != nil {
		return nil, err
	}
	if err := s.loadPayrollInto(&plan); err != nil {
		return nil, err
	}
	if err := s.loadSalesForecastInto(&plan); err != nil {
		return nil, err
	}
	if err := s.loadCashFlowInto(&plan); err != nil {
		return nil, err
	}
	if err := s.loadOperatingExpensesInto(&plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

// GetAll retrieves all plans
func (s *SQLiteStore) GetAll() ([]*domain.Plan, error) {
	rows, err := s.db.Query("SELECT data FROM plans ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %w", err)
	}
	defer rows.Close()

	var plans []*domain.Plan
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		var plan domain.Plan
		if err := json.Unmarshal(data, &plan); err != nil {
			return nil, fmt.Errorf("failed to unmarshal plan: %w", err)
		}

		if err := s.loadStartingPointInto(&plan); err != nil {
			return nil, err
		}
		if err := s.loadPayrollInto(&plan); err != nil {
			return nil, err
		}
		if err := s.loadSalesForecastInto(&plan); err != nil {
			return nil, err
		}
		if err := s.loadCashFlowInto(&plan); err != nil {
			return nil, err
		}
		if err := s.loadOperatingExpensesInto(&plan); err != nil {
			return nil, err
		}

		plans = append(plans, &plan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate plans: %w", err)
	}

	return plans, nil
}

// Delete removes a plan and its access records
func (s *SQLiteStore) Delete(id uuid.UUID) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM plan_access WHERE plan_id = ?", id.String()); err != nil {
		return fmt.Errorf("failed to delete plan access: %w", err)
	}

	// Wizard tables aren't covered by SQLite foreign-key cascades (this
	// codebase never sets PRAGMA foreign_keys=ON), so they must be deleted
	// explicitly, matching how plan_access is handled above.
	wizardTables := []string{
		"capital_assets", "startup_costs", "funding_sources", "starting_balances",
		"salary_roles", "benefits", "payroll_tax_rates",
		"products", "sales_growth_curve",
		"operating_expenses",
		"inventory_purchases", "distributions",
		"wizard_sections",
	}
	for _, table := range wizardTables {
		if _, err := tx.Exec("DELETE FROM "+table+" WHERE plan_id = ?", id.String()); err != nil {
			return fmt.Errorf("failed to delete %s: %w", table, err)
		}
	}

	res, err := tx.Exec("DELETE FROM plans WHERE id = ?", id.String())
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}
	if rows == 0 {
		return errors.New("plan not found")
	}

	return tx.Commit()
}

// SaveUser stores a user
func (s *SQLiteStore) SaveUser(u *domain.User) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		"INSERT OR IGNORE INTO users (id, email, first_name, last_name, created_at) VALUES (?, ?, ?, ?, ?)",
		u.ID().String(),
		strings.ToLower(u.Email()),
		u.FirstName(),
		u.LastName(),
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *SQLiteStore) GetUser(id uuid.UUID) (*domain.User, error) {
	var email, firstName, lastName string

	err := s.db.QueryRow("SELECT email, first_name, last_name FROM users WHERE id = ?", id.String()).Scan(&email, &firstName, &lastName)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	user, err := domain.NewUser(email)
	if err != nil {
		return nil, err
	}

	// Manually set the ID since NewUser generates a new one
	user.SetID(id)
	user.SetFirstName(firstName)
	user.SetLastName(lastName)

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *SQLiteStore) GetUserByEmail(email string) (*domain.User, error) {
	var userID string
	lowerEmail := strings.ToLower(email)

	err := s.db.QueryRow("SELECT id FROM users WHERE LOWER(email) = ?", lowerEmail).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	return s.GetUser(id)
}

// SaveUserWithPassword stores a user with password credentials
func (s *SQLiteStore) SaveUserWithPassword(u *domain.UserWithPassword) error {
	// Save basic user info
	if err := s.SaveUser(u.User); err != nil {
		return err
	}

	// Save credentials
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO users_credentials (email, password_hash, created_at)
		 VALUES (?, ?, ?)`,
		strings.ToLower(u.Email()),
		u.PasswordHash,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to save user credentials: %w", err)
	}

	return nil
}

// GetUserWithPassword retrieves a user with password hash by email
func (s *SQLiteStore) GetUserWithPassword(email string) (*domain.UserWithPassword, error) {
	lowerEmail := strings.ToLower(email)

	var userID, passwordHash string
	err := s.db.QueryRow(
		`SELECT u.id, uc.password_hash
		 FROM users u
		 JOIN users_credentials uc ON u.email = uc.email
		 WHERE LOWER(u.email) = ?`,
		lowerEmail,
	).Scan(&userID, &passwordHash)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user credentials: %w", err)
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	user, err := s.GetUser(id)
	if err != nil {
		return nil, err
	}

	return &domain.UserWithPassword{
		User:         user,
		PasswordHash: passwordHash,
	}, nil
}

// SaveSession stores a session
func (s *SQLiteStore) SaveSession(sess *domain.Session) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO sessions (id, user_id, created_at, expires_at)
		 VALUES (?, ?, ?, ?)`,
		sess.ID,
		sess.UserID.String(),
		sess.CreatedAt.Unix(),
		sess.ExpiresAt.Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// GetSession retrieves a session by ID
func (s *SQLiteStore) GetSession(sessionID string) (*domain.Session, error) {
	var userID string
	var createdAt, expiresAt int64

	err := s.db.QueryRow(
		"SELECT user_id, created_at, expires_at FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&userID, &createdAt, &expiresAt)

	if err == sql.ErrNoRows {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	session := &domain.Session{
		ID:        sessionID,
		UserID:    id,
		CreatedAt: time.Unix(createdAt, 0),
		ExpiresAt: time.Unix(expiresAt, 0),
	}

	if !session.IsValid() {
		return nil, domain.ErrSessionExpired
	}

	return session, nil
}

// DeleteSession removes a session
func (s *SQLiteStore) DeleteSession(sessionID string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GrantAccess gives a user access to a plan
func (s *SQLiteStore) GrantAccess(planID, userID uuid.UUID, level domain.AccessLevel) error {
	if !level.IsValid() {
		return errors.New("invalid access level")
	}

	// Verify plan and user exist
	var exists int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plans WHERE id = ?", planID.String()).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("plan not found")
	}

	if err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID.String()).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return domain.ErrUserNotFound
	}

	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO plan_access (plan_id, user_id, access_level, invited_at)
		 VALUES (?, ?, ?, ?)`,
		planID.String(),
		userID.String(),
		string(level),
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to grant access: %w", err)
	}

	return nil
}

// GetAccess retrieves access for a user on a plan
func (s *SQLiteStore) GetAccess(planID, userID uuid.UUID) (*domain.PlanAccess, error) {
	var accessLevel string
	var invitedAt int64

	err := s.db.QueryRow(
		"SELECT access_level, invited_at FROM plan_access WHERE plan_id = ? AND user_id = ?",
		planID.String(),
		userID.String(),
	).Scan(&accessLevel, &invitedAt)

	if err == sql.ErrNoRows {
		return nil, domain.ErrAccessDenied
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query access: %w", err)
	}

	return &domain.PlanAccess{
		PlanID:      planID,
		UserID:      userID,
		AccessLevel: domain.AccessLevel(accessLevel),
		InvitedAt:   invitedAt,
	}, nil
}

// GetPlanAccess retrieves all access records for a plan
func (s *SQLiteStore) GetPlanAccess(planID uuid.UUID) ([]*domain.PlanAccess, error) {
	rows, err := s.db.Query(
		"SELECT user_id, access_level, invited_at FROM plan_access WHERE plan_id = ?",
		planID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query plan access: %w", err)
	}
	defer rows.Close()

	var accesses []*domain.PlanAccess
	for rows.Next() {
		var userID string
		var accessLevel string
		var invitedAt int64

		if err := rows.Scan(&userID, &accessLevel, &invitedAt); err != nil {
			return nil, fmt.Errorf("failed to scan access: %w", err)
		}

		id, err := uuid.Parse(userID)
		if err != nil {
			return nil, fmt.Errorf("invalid user id: %w", err)
		}

		accesses = append(accesses, &domain.PlanAccess{
			PlanID:      planID,
			UserID:      id,
			AccessLevel: domain.AccessLevel(accessLevel),
			InvitedAt:   invitedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate access: %w", err)
	}

	return accesses, nil
}

// GetUserPlans retrieves all plans a user has access to
func (s *SQLiteStore) GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error) {
	rows, err := s.db.Query(
		`SELECT p.data FROM plans p
		 JOIN plan_access pa ON p.id = pa.plan_id
		 WHERE pa.user_id = ?
		 ORDER BY p.created_at DESC`,
		userID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query user plans: %w", err)
	}
	defer rows.Close()

	var plans []*domain.Plan
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		var plan domain.Plan
		if err := json.Unmarshal(data, &plan); err != nil {
			return nil, fmt.Errorf("failed to unmarshal plan: %w", err)
		}

		if err := s.loadStartingPointInto(&plan); err != nil {
			return nil, err
		}
		if err := s.loadPayrollInto(&plan); err != nil {
			return nil, err
		}
		if err := s.loadSalesForecastInto(&plan); err != nil {
			return nil, err
		}
		if err := s.loadCashFlowInto(&plan); err != nil {
			return nil, err
		}
		if err := s.loadOperatingExpensesInto(&plan); err != nil {
			return nil, err
		}

		plans = append(plans, &plan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate plans: %w", err)
	}

	return plans, nil
}

// CreateInvite stores a new plan invite
func (s *SQLiteStore) CreateInvite(invite *domain.PlanInvite) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO plan_invites (id, plan_id, email, access_level, status, invited_by, created_at, responded_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, NULL)`,
		invite.ID.String(),
		invite.PlanID.String(),
		invite.Email,
		string(invite.AccessLevel),
		string(invite.Status),
		invite.InvitedBy.String(),
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create invite: %w", err)
	}

	invite.CreatedAt = now
	return nil
}

// GetInvite retrieves an invite by ID
func (s *SQLiteStore) GetInvite(id uuid.UUID) (*domain.PlanInvite, error) {
	var planIDStr, email, accessLevel, status, invitedByStr string
	var createdAt int64
	var respondedAt sql.NullInt64

	err := s.db.QueryRow(
		`SELECT plan_id, email, access_level, status, invited_by, created_at, responded_at
		 FROM plan_invites WHERE id = ?`,
		id.String(),
	).Scan(&planIDStr, &email, &accessLevel, &status, &invitedByStr, &createdAt, &respondedAt)

	if err == sql.ErrNoRows {
		return nil, domain.ErrInviteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query invite: %w", err)
	}

	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid plan id: %w", err)
	}
	invitedBy, err := uuid.Parse(invitedByStr)
	if err != nil {
		return nil, fmt.Errorf("invalid inviter id: %w", err)
	}

	return &domain.PlanInvite{
		ID:          id,
		PlanID:      planID,
		Email:       email,
		AccessLevel: domain.AccessLevel(accessLevel),
		Status:      domain.InviteStatus(status),
		InvitedBy:   invitedBy,
		CreatedAt:   createdAt,
		RespondedAt: respondedAt.Int64,
	}, nil
}

// GetInvitesForPlan retrieves every invite (any status) sent for a plan
func (s *SQLiteStore) GetInvitesForPlan(planID uuid.UUID) ([]*domain.PlanInvite, error) {
	rows, err := s.db.Query(
		`SELECT id, email, access_level, status, invited_by, created_at, responded_at
		 FROM plan_invites WHERE plan_id = ? ORDER BY created_at DESC`,
		planID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query plan invites: %w", err)
	}
	defer rows.Close()

	var invites []*domain.PlanInvite
	for rows.Next() {
		var idStr, email, accessLevel, status, invitedByStr string
		var createdAt int64
		var respondedAt sql.NullInt64

		if err := rows.Scan(&idStr, &email, &accessLevel, &status, &invitedByStr, &createdAt, &respondedAt); err != nil {
			return nil, fmt.Errorf("failed to scan invite: %w", err)
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid invite id: %w", err)
		}
		invitedBy, err := uuid.Parse(invitedByStr)
		if err != nil {
			return nil, fmt.Errorf("invalid inviter id: %w", err)
		}

		invites = append(invites, &domain.PlanInvite{
			ID:          id,
			PlanID:      planID,
			Email:       email,
			AccessLevel: domain.AccessLevel(accessLevel),
			Status:      domain.InviteStatus(status),
			InvitedBy:   invitedBy,
			CreatedAt:   createdAt,
			RespondedAt: respondedAt.Int64,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate invites: %w", err)
	}

	return invites, nil
}

// GetPendingInvitesForEmail retrieves all pending invites addressed to an email
func (s *SQLiteStore) GetPendingInvitesForEmail(email string) ([]*domain.PlanInvite, error) {
	rows, err := s.db.Query(
		`SELECT id, plan_id, access_level, status, invited_by, created_at, responded_at
		 FROM plan_invites WHERE LOWER(email) = ? AND status = ? ORDER BY created_at DESC`,
		strings.ToLower(email),
		string(domain.InvitePending),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending invites: %w", err)
	}
	defer rows.Close()

	var invites []*domain.PlanInvite
	for rows.Next() {
		var idStr, planIDStr, accessLevel, status, invitedByStr string
		var createdAt int64
		var respondedAt sql.NullInt64

		if err := rows.Scan(&idStr, &planIDStr, &accessLevel, &status, &invitedByStr, &createdAt, &respondedAt); err != nil {
			return nil, fmt.Errorf("failed to scan invite: %w", err)
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid invite id: %w", err)
		}
		planID, err := uuid.Parse(planIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid plan id: %w", err)
		}
		invitedBy, err := uuid.Parse(invitedByStr)
		if err != nil {
			return nil, fmt.Errorf("invalid inviter id: %w", err)
		}

		invites = append(invites, &domain.PlanInvite{
			ID:          id,
			PlanID:      planID,
			Email:       strings.ToLower(email),
			AccessLevel: domain.AccessLevel(accessLevel),
			Status:      domain.InviteStatus(status),
			InvitedBy:   invitedBy,
			CreatedAt:   createdAt,
			RespondedAt: respondedAt.Int64,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate invites: %w", err)
	}

	return invites, nil
}

// UpdateInviteStatus marks an invite as accepted or rejected
func (s *SQLiteStore) UpdateInviteStatus(id uuid.UUID, status domain.InviteStatus) error {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`UPDATE plan_invites SET status = ?, responded_at = ? WHERE id = ?`,
		string(status),
		now,
		id.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update invite: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}
	if rows == 0 {
		return domain.ErrInviteNotFound
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
