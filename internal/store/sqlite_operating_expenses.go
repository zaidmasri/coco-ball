package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// loadOperatingExpensesInto populates a Plan's Operating Expenses field
// from the normalized table, since it's no longer part of the plans.data
// JSON blob.
func (s *SQLiteStore) loadOperatingExpensesInto(plan *domain.Plan) error {
	items, err := s.ListCompleteOperatingExpenses(plan.ID())
	if err != nil {
		return err
	}
	costs := make([]domain.Cost, len(items))
	for i, item := range items {
		costs[i] = item.Cost
	}

	plan.LoadOpExData(costs)
	return nil
}

func scanOperatingExpenseItem(row rowScanner) (*repositories.OperatingExpenseItem, error) {
	var idStr, name, growthType, status string
	var monthlyAmount int64
	var annualRate float64
	var currentStep int

	if err := row.Scan(&idStr, &name, &monthlyAmount, &growthType, &annualRate, &status, &currentStep); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid operating expense id: %w", err)
	}

	return &repositories.OperatingExpenseItem{
		ID: id,
		Cost: domain.Cost{
			Name:               name,
			BaseAmountPerMonth: domain.Money(monthlyAmount),
			Growth: domain.GrowthStrategy{
				Type:       domain.GrowthType(growthType),
				AnnualRate: annualRate,
			},
		},
		Status:      repositories.ItemStatus(status),
		CurrentStep: currentStep,
	}, nil
}

const operatingExpenseColumns = "id, name, monthly_amount, growth_type, annual_rate, status, current_step"

// CreateOperatingExpenseDraft starts a new Operating Expense wizard item,
// or resumes the plan's existing in-progress draft if one is already
// present.
func (s *SQLiteStore) CreateOperatingExpenseDraft(planID uuid.UUID) (uuid.UUID, error) {
	existing, err := s.GetOperatingExpenseDraft(planID)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, nil
	}

	var sortOrder int
	if err := s.db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM operating_expenses WHERE plan_id = ?", planID.String()).Scan(&sortOrder); err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.New()
	now := time.Now().Unix()
	if _, err := s.db.Exec(
		`INSERT INTO operating_expenses (id, plan_id, status, current_step, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id.String(), planID.String(), string(repositories.StatusDraft), sortOrder, now, now,
	); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create operating expense draft: %w", err)
	}
	return id, nil
}

// GetOperatingExpenseDraft returns the plan's in-progress Operating
// Expense draft, if any, without creating one.
func (s *SQLiteStore) GetOperatingExpenseDraft(planID uuid.UUID) (*repositories.OperatingExpenseItem, error) {
	row := s.db.QueryRow(
		"SELECT "+operatingExpenseColumns+" FROM operating_expenses WHERE plan_id = ? AND status = ? LIMIT 1",
		planID.String(), string(repositories.StatusDraft),
	)
	item, err := scanOperatingExpenseItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query operating expense draft: %w", err)
	}
	return item, nil
}

// GetOperatingExpense retrieves a single Operating Expense wizard item by
// its own ID.
func (s *SQLiteStore) GetOperatingExpense(itemID uuid.UUID) (*repositories.OperatingExpenseItem, error) {
	row := s.db.QueryRow(
		"SELECT "+operatingExpenseColumns+" FROM operating_expenses WHERE id = ?",
		itemID.String(),
	)
	item, err := scanOperatingExpenseItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("operating expense not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query operating expense: %w", err)
	}
	return item, nil
}

// SaveOperatingExpenseStep persists the wizard's current answer for an
// Operating Expense item, advancing its step/status.
func (s *SQLiteStore) SaveOperatingExpenseStep(itemID uuid.UUID, cost domain.Cost, currentStep int, status repositories.ItemStatus) error {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`UPDATE operating_expenses SET name=?, monthly_amount=?, growth_type=?, annual_rate=?, status=?, current_step=?, updated_at=? WHERE id = ?`,
		cost.Name, int64(cost.BaseAmountPerMonth), string(cost.Growth.Type), cost.Growth.AnnualRate, string(status), currentStep, now, itemID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to save operating expense step: %w", err)
	}
	return checkRowsAffected(res, "operating expense not found")
}

// ListCompleteOperatingExpenses returns every finished Operating Expense
// for a plan, in the order they were added.
func (s *SQLiteStore) ListCompleteOperatingExpenses(planID uuid.UUID) ([]repositories.OperatingExpenseItem, error) {
	rows, err := s.db.Query(
		"SELECT "+operatingExpenseColumns+" FROM operating_expenses WHERE plan_id = ? AND status = ? ORDER BY sort_order ASC",
		planID.String(), string(repositories.StatusComplete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query operating expenses: %w", err)
	}
	defer rows.Close()

	var items []repositories.OperatingExpenseItem
	for rows.Next() {
		item, err := scanOperatingExpenseItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan operating expense: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate operating expenses: %w", err)
	}
	return items, nil
}

// DeleteOperatingExpense removes a single Operating Expense item.
func (s *SQLiteStore) DeleteOperatingExpense(itemID uuid.UUID) error {
	res, err := s.db.Exec("DELETE FROM operating_expenses WHERE id = ?", itemID.String())
	if err != nil {
		return fmt.Errorf("failed to delete operating expense: %w", err)
	}
	return checkRowsAffected(res, "operating expense not found")
}
