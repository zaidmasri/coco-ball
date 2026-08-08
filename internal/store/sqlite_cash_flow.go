package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
)

// loadCashFlowInto populates a Plan's discretionary Cash Flow fields
// (Inventory Purchases, Distributions) from the normalized tables, since
// they're no longer part of the plans.data JSON blob.
func (s *SQLiteStore) loadCashFlowInto(plan *domain.Plan) error {
	planID := plan.ID()

	invItems, err := s.ListCompleteInventoryPurchases(planID)
	if err != nil {
		return err
	}
	inventory := make([]domain.InventoryPurchase, len(invItems))
	for i, item := range invItems {
		inventory[i] = item.Purchase
	}

	distItems, err := s.ListCompleteDistributions(planID)
	if err != nil {
		return err
	}
	distributions := make([]domain.Distribution, len(distItems))
	for i, item := range distItems {
		distributions[i] = item.Distribution
	}

	plan.LoadCashFlowData(inventory, distributions)
	return nil
}

// --- Inventory Purchases ---

func scanInventoryPurchaseItem(row rowScanner) (*repositories.InventoryPurchaseItem, error) {
	var idStr, category, status string
	var monthlyAmount int64
	var growthYr2, growthYr3 float64
	var currentStep int

	if err := row.Scan(&idStr, &category, &monthlyAmount, &growthYr2, &growthYr3, &status, &currentStep); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid inventory purchase id: %w", err)
	}

	return &repositories.InventoryPurchaseItem{
		ID: id,
		Purchase: domain.InventoryPurchase{
			Category:       category,
			MonthlyAmount:  mustUSD(monthlyAmount),
			GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{growthYr2, growthYr3}},
		},
		Status:      repositories.ItemStatus(status),
		CurrentStep: currentStep,
	}, nil
}

const inventoryPurchaseColumns = "id, category, monthly_amount, growth_yr2, growth_yr3, status, current_step"

// CreateInventoryPurchaseDraft starts a new Inventory Purchase wizard item,
// or resumes the plan's existing in-progress draft if one is already
// present.
func (s *SQLiteStore) CreateInventoryPurchaseDraft(planID uuid.UUID) (uuid.UUID, error) {
	existing, err := s.FindInventoryPurchaseDraft(planID)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, nil
	}

	var sortOrder int
	if err := s.db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM inventory_purchases WHERE plan_id = ?", planID.String()).Scan(&sortOrder); err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to generate item id: %w", err)
	}
	now := time.Now().Unix()
	if _, err := s.db.Exec(
		`INSERT INTO inventory_purchases (id, plan_id, status, current_step, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id.String(), planID.String(), string(repositories.StatusDraft), sortOrder, now, now,
	); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create inventory purchase draft: %w", err)
	}
	return id, nil
}

// FindInventoryPurchaseDraft returns the plan's in-progress Inventory
// Purchase draft, if any, without creating one.
func (s *SQLiteStore) FindInventoryPurchaseDraft(planID uuid.UUID) (*repositories.InventoryPurchaseItem, error) {
	row := s.db.QueryRow(
		"SELECT "+inventoryPurchaseColumns+" FROM inventory_purchases WHERE plan_id = ? AND status = ? AND deleted_at IS NULL LIMIT 1",
		planID.String(), string(repositories.StatusDraft),
	)
	item, err := scanInventoryPurchaseItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory purchase draft: %w", err)
	}
	return item, nil
}

// GetInventoryPurchase retrieves a single Inventory Purchase wizard item by
// its own ID.
func (s *SQLiteStore) GetInventoryPurchase(itemID uuid.UUID) (*repositories.InventoryPurchaseItem, error) {
	row := s.db.QueryRow(
		"SELECT "+inventoryPurchaseColumns+" FROM inventory_purchases WHERE id = ? AND deleted_at IS NULL",
		itemID.String(),
	)
	item, err := scanInventoryPurchaseItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("inventory purchase not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory purchase: %w", err)
	}
	return item, nil
}

// SaveInventoryPurchaseStep persists the wizard's current answer for an
// Inventory Purchase item, advancing its step/status.
func (s *SQLiteStore) SaveInventoryPurchaseStep(itemID uuid.UUID, inv domain.InventoryPurchase, currentStep int, status repositories.ItemStatus) error {
	now := time.Now().Unix()
	growthYr2 := inv.GrowthAfterYr1.GrowthRatePercent(0) / 100
	growthYr3 := inv.GrowthAfterYr1.GrowthRatePercent(1) / 100
	res, err := s.db.Exec(
		`UPDATE inventory_purchases SET category=?, monthly_amount=?, growth_yr2=?, growth_yr3=?, status=?, current_step=?, updated_at=? WHERE id = ?`,
		inv.Category, inv.MonthlyAmount.MinorUnits(), growthYr2, growthYr3, string(status), currentStep, now, itemID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to save inventory purchase step: %w", err)
	}
	return checkRowsAffected(res, "inventory purchase not found")
}

// ListCompleteInventoryPurchases returns every finished Inventory Purchase
// for a plan, in the order they were added.
func (s *SQLiteStore) ListCompleteInventoryPurchases(planID uuid.UUID) ([]repositories.InventoryPurchaseItem, error) {
	rows, err := s.db.Query(
		"SELECT "+inventoryPurchaseColumns+" FROM inventory_purchases WHERE plan_id = ? AND status = ? AND deleted_at IS NULL ORDER BY sort_order ASC",
		planID.String(), string(repositories.StatusComplete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory purchases: %w", err)
	}
	defer rows.Close()

	var items []repositories.InventoryPurchaseItem
	for rows.Next() {
		item, err := scanInventoryPurchaseItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory purchase: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate inventory purchases: %w", err)
	}
	return items, nil
}

// DeleteInventoryPurchase soft-deletes a single Inventory Purchase item.
func (s *SQLiteStore) DeleteInventoryPurchase(itemID uuid.UUID) error {
	res, err := s.db.Exec("UPDATE inventory_purchases SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL", time.Now().Unix(), itemID.String())
	if err != nil {
		return fmt.Errorf("failed to delete inventory purchase: %w", err)
	}
	return checkRowsAffected(res, "inventory purchase not found")
}

// --- Distributions ---

func scanDistributionItem(row rowScanner) (*repositories.DistributionItem, error) {
	var idStr, name, status string
	var monthlyAmount int64
	var growthYr2, growthYr3 float64
	var currentStep int

	if err := row.Scan(&idStr, &name, &monthlyAmount, &growthYr2, &growthYr3, &status, &currentStep); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid distribution id: %w", err)
	}

	return &repositories.DistributionItem{
		ID: id,
		Distribution: domain.Distribution{
			Name:           name,
			MonthlyAmount:  mustUSD(monthlyAmount),
			GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{growthYr2, growthYr3}},
		},
		Status:      repositories.ItemStatus(status),
		CurrentStep: currentStep,
	}, nil
}

const distributionColumns = "id, name, monthly_amount, growth_yr2, growth_yr3, status, current_step"

// CreateDistributionDraft starts a new Distribution wizard item, or
// resumes the plan's existing in-progress draft if one is already present.
func (s *SQLiteStore) CreateDistributionDraft(planID uuid.UUID) (uuid.UUID, error) {
	existing, err := s.FindDistributionDraft(planID)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, nil
	}

	var sortOrder int
	if err := s.db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM distributions WHERE plan_id = ?", planID.String()).Scan(&sortOrder); err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to generate item id: %w", err)
	}
	now := time.Now().Unix()
	if _, err := s.db.Exec(
		`INSERT INTO distributions (id, plan_id, status, current_step, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id.String(), planID.String(), string(repositories.StatusDraft), sortOrder, now, now,
	); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create distribution draft: %w", err)
	}
	return id, nil
}

// FindDistributionDraft returns the plan's in-progress Distribution draft,
// if any, without creating one.
func (s *SQLiteStore) FindDistributionDraft(planID uuid.UUID) (*repositories.DistributionItem, error) {
	row := s.db.QueryRow(
		"SELECT "+distributionColumns+" FROM distributions WHERE plan_id = ? AND status = ? AND deleted_at IS NULL LIMIT 1",
		planID.String(), string(repositories.StatusDraft),
	)
	item, err := scanDistributionItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query distribution draft: %w", err)
	}
	return item, nil
}

// GetDistribution retrieves a single Distribution wizard item by its own
// ID.
func (s *SQLiteStore) GetDistribution(itemID uuid.UUID) (*repositories.DistributionItem, error) {
	row := s.db.QueryRow(
		"SELECT "+distributionColumns+" FROM distributions WHERE id = ? AND deleted_at IS NULL",
		itemID.String(),
	)
	item, err := scanDistributionItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("distribution not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query distribution: %w", err)
	}
	return item, nil
}

// SaveDistributionStep persists the wizard's current answer for a
// Distribution item, advancing its step/status.
func (s *SQLiteStore) SaveDistributionStep(itemID uuid.UUID, dist domain.Distribution, currentStep int, status repositories.ItemStatus) error {
	now := time.Now().Unix()
	growthYr2 := dist.GrowthAfterYr1.GrowthRatePercent(0) / 100
	growthYr3 := dist.GrowthAfterYr1.GrowthRatePercent(1) / 100
	res, err := s.db.Exec(
		`UPDATE distributions SET name=?, monthly_amount=?, growth_yr2=?, growth_yr3=?, status=?, current_step=?, updated_at=? WHERE id = ?`,
		dist.Name, dist.MonthlyAmount.MinorUnits(), growthYr2, growthYr3, string(status), currentStep, now, itemID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to save distribution step: %w", err)
	}
	return checkRowsAffected(res, "distribution not found")
}

// ListCompleteDistributions returns every finished Distribution for a
// plan, in the order they were added.
func (s *SQLiteStore) ListCompleteDistributions(planID uuid.UUID) ([]repositories.DistributionItem, error) {
	rows, err := s.db.Query(
		"SELECT "+distributionColumns+" FROM distributions WHERE plan_id = ? AND status = ? AND deleted_at IS NULL ORDER BY sort_order ASC",
		planID.String(), string(repositories.StatusComplete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query distributions: %w", err)
	}
	defer rows.Close()

	var items []repositories.DistributionItem
	for rows.Next() {
		item, err := scanDistributionItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan distribution: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate distributions: %w", err)
	}
	return items, nil
}

// DeleteDistribution soft-deletes a single Distribution item.
func (s *SQLiteStore) DeleteDistribution(itemID uuid.UUID) error {
	res, err := s.db.Exec("UPDATE distributions SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL", time.Now().Unix(), itemID.String())
	if err != nil {
		return fmt.Errorf("failed to delete distribution: %w", err)
	}
	return checkRowsAffected(res, "distribution not found")
}
