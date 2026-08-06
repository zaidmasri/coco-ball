package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// loadSalesForecastInto populates a Plan's Sales Forecast fields (Products,
// Sales Growth Curve) from the normalized tables, since they're no longer
// part of the plans.data JSON blob.
func (s *SQLiteStore) loadSalesForecastInto(plan *domain.Plan) error {
	planID := plan.ID()

	productItems, err := s.ListCompleteProducts(planID)
	if err != nil {
		return err
	}
	products := make([]domain.Product, len(productItems))
	for i, item := range productItems {
		products[i] = item.Product
	}

	curveRow, err := s.GetSalesGrowthCurveRow(planID)
	if err != nil {
		return err
	}

	plan.LoadSalesForecastData(products, curveRow.Curve)
	return nil
}

// --- Products ---

func scanProductItem(row rowScanner) (*ProductItem, error) {
	var idStr, name, status string
	var month1Units, currentStep int
	var pricePerUnit, costPerUnit int64

	if err := row.Scan(&idStr, &name, &month1Units, &pricePerUnit, &costPerUnit, &status, &currentStep); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}

	return &ProductItem{
		ID: id,
		Product: domain.Product{
			Name:         name,
			Month1Units:  month1Units,
			PricePerUnit: domain.Money(pricePerUnit),
			CostPerUnit:  domain.Money(costPerUnit),
		},
		Status:      ItemStatus(status),
		CurrentStep: currentStep,
	}, nil
}

const productColumns = "id, name, month1_units, price_per_unit, cost_per_unit, status, current_step"

// CreateProductDraft starts a new Product wizard item, or resumes the
// plan's existing in-progress draft if one is already present.
func (s *SQLiteStore) CreateProductDraft(planID uuid.UUID) (uuid.UUID, error) {
	existing, err := s.GetProductDraft(planID)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, nil
	}

	var sortOrder int
	if err := s.db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM products WHERE plan_id = ?", planID.String()).Scan(&sortOrder); err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.New()
	now := time.Now().Unix()
	if _, err := s.db.Exec(
		`INSERT INTO products (id, plan_id, status, current_step, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id.String(), planID.String(), string(StatusDraft), sortOrder, now, now,
	); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create product draft: %w", err)
	}
	return id, nil
}

// GetProductDraft returns the plan's in-progress Product draft, if any,
// without creating one.
func (s *SQLiteStore) GetProductDraft(planID uuid.UUID) (*ProductItem, error) {
	row := s.db.QueryRow(
		"SELECT "+productColumns+" FROM products WHERE plan_id = ? AND status = ? LIMIT 1",
		planID.String(), string(StatusDraft),
	)
	item, err := scanProductItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query product draft: %w", err)
	}
	return item, nil
}

// GetProduct retrieves a single Product wizard item by its own ID.
func (s *SQLiteStore) GetProduct(itemID uuid.UUID) (*ProductItem, error) {
	row := s.db.QueryRow(
		"SELECT "+productColumns+" FROM products WHERE id = ?",
		itemID.String(),
	)
	item, err := scanProductItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("product not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query product: %w", err)
	}
	return item, nil
}

// SaveProductStep persists the wizard's current answer for a Product item,
// advancing its step/status.
func (s *SQLiteStore) SaveProductStep(itemID uuid.UUID, product domain.Product, currentStep int, status ItemStatus) error {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`UPDATE products SET name=?, month1_units=?, price_per_unit=?, cost_per_unit=?, status=?, current_step=?, updated_at=? WHERE id = ?`,
		product.Name, product.Month1Units, int64(product.PricePerUnit), int64(product.CostPerUnit), string(status), currentStep, now, itemID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to save product step: %w", err)
	}
	return checkRowsAffected(res, "product not found")
}

// ListCompleteProducts returns every finished Product for a plan, in the
// order they were added.
func (s *SQLiteStore) ListCompleteProducts(planID uuid.UUID) ([]ProductItem, error) {
	rows, err := s.db.Query(
		"SELECT "+productColumns+" FROM products WHERE plan_id = ? AND status = ? ORDER BY sort_order ASC",
		planID.String(), string(StatusComplete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var items []ProductItem
	for rows.Next() {
		item, err := scanProductItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate products: %w", err)
	}
	return items, nil
}

// DeleteProduct removes a single Product item.
func (s *SQLiteStore) DeleteProduct(itemID uuid.UUID) error {
	res, err := s.db.Exec("DELETE FROM products WHERE id = ?", itemID.String())
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return checkRowsAffected(res, "product not found")
}

// --- Sales Growth Curve (singleton) ---

// GetSalesGrowthCurveRow returns a plan's Sales Growth Curve wizard state.
// A plan that has never saved a step has no row yet; this returns a
// zero-value SalesGrowthCurveRow (CurrentStep 0) rather than creating one,
// since reads shouldn't have write side effects - the first
// SaveSalesGrowthCurveStep call creates the row via upsert.
func (s *SQLiteStore) GetSalesGrowthCurveRow(planID uuid.UUID) (*SalesGrowthCurveRow, error) {
	var q1, q2, q3, q4, growthYr2, growthYr3 float64
	var currentStep int

	err := s.db.QueryRow(
		`SELECT y1_q1, y1_q2, y1_q3, y1_q4, growth_yr2, growth_yr3, current_step
		 FROM sales_growth_curve WHERE plan_id = ?`,
		planID.String(),
	).Scan(&q1, &q2, &q3, &q4, &growthYr2, &growthYr3, &currentStep)

	if errors.Is(err, sql.ErrNoRows) {
		return &SalesGrowthCurveRow{Curve: domain.SalesGrowthCurve{}, CurrentStep: 0}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query sales growth curve: %w", err)
	}

	return &SalesGrowthCurveRow{
		Curve: domain.SalesGrowthCurve{
			Year1QuarterlyRates: [4]float64{q1, q2, q3, q4},
			FutureYearRates:     []float64{growthYr2, growthYr3},
		},
		CurrentStep: currentStep,
	}, nil
}

// SaveSalesGrowthCurveStep persists the wizard's current answer for the
// Sales Growth Curve, creating the plan's singleton row on first call.
func (s *SQLiteStore) SaveSalesGrowthCurveStep(planID uuid.UUID, curve domain.SalesGrowthCurve, currentStep int) error {
	now := time.Now().Unix()
	growthYr2 := curve.FutureYearPercent(0) / 100
	growthYr3 := curve.FutureYearPercent(1) / 100
	_, err := s.db.Exec(
		`INSERT INTO sales_growth_curve (plan_id, y1_q1, y1_q2, y1_q3, y1_q4, growth_yr2, growth_yr3, current_step, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(plan_id) DO UPDATE SET
			y1_q1=excluded.y1_q1, y1_q2=excluded.y1_q2, y1_q3=excluded.y1_q3, y1_q4=excluded.y1_q4,
			growth_yr2=excluded.growth_yr2, growth_yr3=excluded.growth_yr3,
			current_step=excluded.current_step, updated_at=excluded.updated_at`,
		planID.String(), curve.Year1QuarterlyRates[0], curve.Year1QuarterlyRates[1], curve.Year1QuarterlyRates[2], curve.Year1QuarterlyRates[3], growthYr2, growthYr3, currentStep, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to save sales growth curve step: %w", err)
	}
	return nil
}
