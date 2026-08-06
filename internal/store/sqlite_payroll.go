package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

// loadPayrollInto populates a Plan's Payroll fields (Salary Roles,
// Benefits, Payroll Tax Rates) from the normalized tables, since they're no
// longer part of the plans.data JSON blob.
func (s *SQLiteStore) loadPayrollInto(plan *domain.Plan) error {
	planID := plan.ID()

	roleItems, err := s.ListCompleteSalaryRoles(planID)
	if err != nil {
		return err
	}
	roles := make([]domain.SalaryRole, len(roleItems))
	for i, item := range roleItems {
		roles[i] = item.Role
	}

	benefitItems, err := s.ListCompleteBenefits(planID)
	if err != nil {
		return err
	}
	benefits := make([]domain.Benefit, len(benefitItems))
	for i, item := range benefitItems {
		benefits[i] = item.Benefit
	}

	ratesRow, err := s.GetPayrollTaxRatesRow(planID)
	if err != nil {
		return err
	}

	plan.LoadPayrollData(roles, benefits, ratesRow.Rates)
	return nil
}

// --- Salary Roles ---

func scanSalaryRoleItem(row rowScanner) (*SalaryRoleItem, error) {
	var idStr, role, status string
	var isContractor bool
	var headcount, currentStep int
	var monthlyPay int64
	var growthYr2, growthYr3 float64

	if err := row.Scan(&idStr, &role, &isContractor, &headcount, &monthlyPay, &growthYr2, &growthYr3, &status, &currentStep); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid salary role id: %w", err)
	}

	return &SalaryRoleItem{
		ID: id,
		Role: domain.SalaryRole{
			Role:           role,
			IsContractor:   isContractor,
			Headcount:      headcount,
			MonthlyPay:     domain.Money(monthlyPay),
			GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{growthYr2, growthYr3}},
		},
		Status:      ItemStatus(status),
		CurrentStep: currentStep,
	}, nil
}

const salaryRoleColumns = "id, role, is_contractor, headcount, monthly_pay, growth_yr2, growth_yr3, status, current_step"

// CreateSalaryRoleDraft starts a new Salary Role wizard item, or resumes
// the plan's existing in-progress draft if one is already present.
func (s *SQLiteStore) CreateSalaryRoleDraft(planID uuid.UUID) (uuid.UUID, error) {
	existing, err := s.GetSalaryRoleDraft(planID)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, nil
	}

	var sortOrder int
	if err := s.db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM salary_roles WHERE plan_id = ?", planID.String()).Scan(&sortOrder); err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.New()
	now := time.Now().Unix()
	if _, err := s.db.Exec(
		`INSERT INTO salary_roles (id, plan_id, status, current_step, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id.String(), planID.String(), string(StatusDraft), sortOrder, now, now,
	); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create salary role draft: %w", err)
	}
	return id, nil
}

// GetSalaryRoleDraft returns the plan's in-progress Salary Role draft, if
// any, without creating one.
func (s *SQLiteStore) GetSalaryRoleDraft(planID uuid.UUID) (*SalaryRoleItem, error) {
	row := s.db.QueryRow(
		"SELECT "+salaryRoleColumns+" FROM salary_roles WHERE plan_id = ? AND status = ? LIMIT 1",
		planID.String(), string(StatusDraft),
	)
	item, err := scanSalaryRoleItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query salary role draft: %w", err)
	}
	return item, nil
}

// GetSalaryRole retrieves a single Salary Role wizard item by its own ID.
func (s *SQLiteStore) GetSalaryRole(itemID uuid.UUID) (*SalaryRoleItem, error) {
	row := s.db.QueryRow(
		"SELECT "+salaryRoleColumns+" FROM salary_roles WHERE id = ?",
		itemID.String(),
	)
	item, err := scanSalaryRoleItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("salary role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query salary role: %w", err)
	}
	return item, nil
}

// SaveSalaryRoleStep persists the wizard's current answer for a Salary
// Role item, advancing its step/status.
func (s *SQLiteStore) SaveSalaryRoleStep(itemID uuid.UUID, role domain.SalaryRole, currentStep int, status ItemStatus) error {
	now := time.Now().Unix()
	growthYr2 := role.GrowthAfterYr1.GrowthRatePercent(0) / 100
	growthYr3 := role.GrowthAfterYr1.GrowthRatePercent(1) / 100
	res, err := s.db.Exec(
		`UPDATE salary_roles
		 SET role=?, is_contractor=?, headcount=?, monthly_pay=?, growth_yr2=?, growth_yr3=?, status=?, current_step=?, updated_at=?
		 WHERE id = ?`,
		role.Role, role.IsContractor, role.Headcount, int64(role.MonthlyPay), growthYr2, growthYr3, string(status), currentStep, now,
		itemID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to save salary role step: %w", err)
	}
	return checkRowsAffected(res, "salary role not found")
}

// ListCompleteSalaryRoles returns every finished Salary Role for a plan, in
// the order they were added.
func (s *SQLiteStore) ListCompleteSalaryRoles(planID uuid.UUID) ([]SalaryRoleItem, error) {
	rows, err := s.db.Query(
		"SELECT "+salaryRoleColumns+" FROM salary_roles WHERE plan_id = ? AND status = ? ORDER BY sort_order ASC",
		planID.String(), string(StatusComplete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query salary roles: %w", err)
	}
	defer rows.Close()

	var items []SalaryRoleItem
	for rows.Next() {
		item, err := scanSalaryRoleItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan salary role: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate salary roles: %w", err)
	}
	return items, nil
}

// DeleteSalaryRole removes a single Salary Role item.
func (s *SQLiteStore) DeleteSalaryRole(itemID uuid.UUID) error {
	res, err := s.db.Exec("DELETE FROM salary_roles WHERE id = ?", itemID.String())
	if err != nil {
		return fmt.Errorf("failed to delete salary role: %w", err)
	}
	return checkRowsAffected(res, "salary role not found")
}

// --- Benefits ---

func scanBenefitItem(row rowScanner) (*BenefitItem, error) {
	var idStr, benefitType, status string
	var monthlyAmount int64
	var growthYr2, growthYr3 float64
	var currentStep int

	if err := row.Scan(&idStr, &benefitType, &monthlyAmount, &growthYr2, &growthYr3, &status, &currentStep); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid benefit id: %w", err)
	}

	return &BenefitItem{
		ID: id,
		Benefit: domain.Benefit{
			Type:           benefitType,
			MonthlyAmount:  domain.Money(monthlyAmount),
			GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{growthYr2, growthYr3}},
		},
		Status:      ItemStatus(status),
		CurrentStep: currentStep,
	}, nil
}

const benefitColumns = "id, type, monthly_amount, growth_yr2, growth_yr3, status, current_step"

// CreateBenefitDraft starts a new Benefit wizard item, or resumes the
// plan's existing in-progress draft if one is already present.
func (s *SQLiteStore) CreateBenefitDraft(planID uuid.UUID) (uuid.UUID, error) {
	existing, err := s.GetBenefitDraft(planID)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, nil
	}

	var sortOrder int
	if err := s.db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM benefits WHERE plan_id = ?", planID.String()).Scan(&sortOrder); err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.New()
	now := time.Now().Unix()
	if _, err := s.db.Exec(
		`INSERT INTO benefits (id, plan_id, status, current_step, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id.String(), planID.String(), string(StatusDraft), sortOrder, now, now,
	); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create benefit draft: %w", err)
	}
	return id, nil
}

// GetBenefitDraft returns the plan's in-progress Benefit draft, if any,
// without creating one.
func (s *SQLiteStore) GetBenefitDraft(planID uuid.UUID) (*BenefitItem, error) {
	row := s.db.QueryRow(
		"SELECT "+benefitColumns+" FROM benefits WHERE plan_id = ? AND status = ? LIMIT 1",
		planID.String(), string(StatusDraft),
	)
	item, err := scanBenefitItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query benefit draft: %w", err)
	}
	return item, nil
}

// GetBenefit retrieves a single Benefit wizard item by its own ID.
func (s *SQLiteStore) GetBenefit(itemID uuid.UUID) (*BenefitItem, error) {
	row := s.db.QueryRow(
		"SELECT "+benefitColumns+" FROM benefits WHERE id = ?",
		itemID.String(),
	)
	item, err := scanBenefitItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("benefit not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query benefit: %w", err)
	}
	return item, nil
}

// SaveBenefitStep persists the wizard's current answer for a Benefit item,
// advancing its step/status.
func (s *SQLiteStore) SaveBenefitStep(itemID uuid.UUID, benefit domain.Benefit, currentStep int, status ItemStatus) error {
	now := time.Now().Unix()
	growthYr2 := benefit.GrowthAfterYr1.GrowthRatePercent(0) / 100
	growthYr3 := benefit.GrowthAfterYr1.GrowthRatePercent(1) / 100
	res, err := s.db.Exec(
		`UPDATE benefits SET type=?, monthly_amount=?, growth_yr2=?, growth_yr3=?, status=?, current_step=?, updated_at=? WHERE id = ?`,
		benefit.Type, int64(benefit.MonthlyAmount), growthYr2, growthYr3, string(status), currentStep, now, itemID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to save benefit step: %w", err)
	}
	return checkRowsAffected(res, "benefit not found")
}

// ListCompleteBenefits returns every finished Benefit for a plan, in the
// order they were added.
func (s *SQLiteStore) ListCompleteBenefits(planID uuid.UUID) ([]BenefitItem, error) {
	rows, err := s.db.Query(
		"SELECT "+benefitColumns+" FROM benefits WHERE plan_id = ? AND status = ? ORDER BY sort_order ASC",
		planID.String(), string(StatusComplete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query benefits: %w", err)
	}
	defer rows.Close()

	var items []BenefitItem
	for rows.Next() {
		item, err := scanBenefitItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan benefit: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate benefits: %w", err)
	}
	return items, nil
}

// DeleteBenefit removes a single Benefit item.
func (s *SQLiteStore) DeleteBenefit(itemID uuid.UUID) error {
	res, err := s.db.Exec("DELETE FROM benefits WHERE id = ?", itemID.String())
	if err != nil {
		return fmt.Errorf("failed to delete benefit: %w", err)
	}
	return checkRowsAffected(res, "benefit not found")
}

// --- Payroll Tax Rates (singleton) ---

// GetPayrollTaxRatesRow returns a plan's Payroll Tax Rates wizard state. A
// plan that has never saved a step has no row yet; this returns a
// zero-value PayrollTaxRatesRow (CurrentStep 0) rather than creating one,
// since reads shouldn't have write side effects - the first
// SavePayrollTaxRatesStep call creates the row via upsert.
func (s *SQLiteStore) GetPayrollTaxRatesRow(planID uuid.UUID) (*PayrollTaxRatesRow, error) {
	var ss, medicare, futa, suta float64
	var currentStep int

	err := s.db.QueryRow(
		`SELECT social_security_rate, medicare_rate, futa_rate, suta_rate, current_step
		 FROM payroll_tax_rates WHERE plan_id = ?`,
		planID.String(),
	).Scan(&ss, &medicare, &futa, &suta, &currentStep)

	if errors.Is(err, sql.ErrNoRows) {
		return &PayrollTaxRatesRow{Rates: domain.PayrollTaxRates{}, CurrentStep: 0}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query payroll tax rates: %w", err)
	}

	return &PayrollTaxRatesRow{
		Rates: domain.PayrollTaxRates{
			SocialSecurityRate: ss,
			MedicareRate:       medicare,
			FUTARate:           futa,
			SUTARate:           suta,
		},
		CurrentStep: currentStep,
	}, nil
}

// SavePayrollTaxRatesStep persists the wizard's current answer for Payroll
// Tax Rates, creating the plan's singleton row on first call.
func (s *SQLiteStore) SavePayrollTaxRatesStep(planID uuid.UUID, rates domain.PayrollTaxRates, currentStep int) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO payroll_tax_rates (plan_id, social_security_rate, medicare_rate, futa_rate, suta_rate, current_step, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(plan_id) DO UPDATE SET
			social_security_rate=excluded.social_security_rate, medicare_rate=excluded.medicare_rate,
			futa_rate=excluded.futa_rate, suta_rate=excluded.suta_rate,
			current_step=excluded.current_step, updated_at=excluded.updated_at`,
		planID.String(), rates.SocialSecurityRate, rates.MedicareRate, rates.FUTARate, rates.SUTARate, currentStep, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to save payroll tax rates step: %w", err)
	}
	return nil
}
