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

// rowScanner is satisfied by both *sql.Row and *sql.Rows, letting a single
// scan function serve both a single-row Get and a multi-row List query.
type rowScanner interface {
	Scan(dest ...any) error
}

// loadStartingPointInto populates a Plan's Starting Point fields (Fixed
// Assets, Startup Costs, Funding Sources, Cash on Hand) from the
// normalized tables, since they're no longer part of the plans.data JSON
// blob. Only status='complete' rows are loaded - an item still mid-wizard
// must not silently enter financial projection math.
func (s *SQLiteStore) loadStartingPointInto(plan *domain.Plan) error {
	planID := plan.ID()

	assetItems, err := s.ListCompleteCapitalAssets(planID)
	if err != nil {
		return err
	}
	assets := make([]domain.CapitalAsset, len(assetItems))
	for i, item := range assetItems {
		assets[i] = item.Asset
	}

	costItems, err := s.ListCompleteStartupCosts(planID)
	if err != nil {
		return err
	}
	costs := make([]domain.StartupCost, len(costItems))
	for i, item := range costItems {
		costs[i] = item.Cost
	}

	fundingItems, err := s.ListCompleteFundingSources(planID)
	if err != nil {
		return err
	}
	funding := make([]domain.FundingSource, len(fundingItems))
	for i, item := range fundingItems {
		funding[i] = item.Funding
	}

	balancesRow, err := s.GetStartingBalancesRow(planID)
	if err != nil {
		return err
	}

	plan.LoadStartingPointData(assets, costs, funding, balancesRow.Balances)
	return nil
}

// --- Fixed Assets ---

func scanCapitalAssetItem(row rowScanner) (*repositories.CapitalAssetItem, error) {
	var idStr, name, deprMethod, status string
	var purchaseCost, salvageValue int64
	var usefulLifeMonths, purchaseMonthIndex, currentStep int

	if err := row.Scan(&idStr, &name, &purchaseCost, &usefulLifeMonths, &salvageValue, &purchaseMonthIndex, &deprMethod, &status, &currentStep); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid capital asset id: %w", err)
	}

	return &repositories.CapitalAssetItem{
		ID: id,
		Asset: domain.CapitalAsset{
			Name:               name,
			PurchaseCost:       domain.Money(purchaseCost),
			UsefulLifeMonths:   usefulLifeMonths,
			SalvageValue:       domain.Money(salvageValue),
			PurchaseMonthIndex: domain.MonthIndex(purchaseMonthIndex),
			DepreciationMethod: domain.DepreciationMethod(deprMethod),
		},
		Status:      repositories.ItemStatus(status),
		CurrentStep: currentStep,
	}, nil
}

const capitalAssetColumns = "id, name, purchase_cost, useful_life_months, salvage_value, purchase_month_index, depreciation_method, status, current_step"

// CreateCapitalAssetDraft starts a new Fixed Asset wizard item, or resumes
// the plan's existing in-progress draft if one is already present (at most
// one draft per plan+section at a time).
func (s *SQLiteStore) CreateCapitalAssetDraft(planID uuid.UUID) (uuid.UUID, error) {
	existing, err := s.GetCapitalAssetDraft(planID)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, nil
	}

	var sortOrder int
	if err := s.db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM capital_assets WHERE plan_id = ?", planID.String()).Scan(&sortOrder); err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.New()
	now := time.Now().Unix()
	if _, err := s.db.Exec(
		`INSERT INTO capital_assets (id, plan_id, status, current_step, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id.String(), planID.String(), string(repositories.StatusDraft), sortOrder, now, now,
	); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create capital asset draft: %w", err)
	}
	return id, nil
}

// GetCapitalAssetDraft returns the plan's in-progress Fixed Asset draft, if
// any, without creating one. Returns (nil, nil) when there is none.
func (s *SQLiteStore) GetCapitalAssetDraft(planID uuid.UUID) (*repositories.CapitalAssetItem, error) {
	row := s.db.QueryRow(
		"SELECT "+capitalAssetColumns+" FROM capital_assets WHERE plan_id = ? AND status = ? LIMIT 1",
		planID.String(), string(repositories.StatusDraft),
	)
	item, err := scanCapitalAssetItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query capital asset draft: %w", err)
	}
	return item, nil
}

// GetCapitalAsset retrieves a single Fixed Asset wizard item by its own ID.
func (s *SQLiteStore) GetCapitalAsset(itemID uuid.UUID) (*repositories.CapitalAssetItem, error) {
	row := s.db.QueryRow(
		"SELECT "+capitalAssetColumns+" FROM capital_assets WHERE id = ?",
		itemID.String(),
	)
	item, err := scanCapitalAssetItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("capital asset not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query capital asset: %w", err)
	}
	return item, nil
}

// SaveCapitalAssetStep persists the wizard's current answer for a Fixed
// Asset item, advancing its step/status.
func (s *SQLiteStore) SaveCapitalAssetStep(itemID uuid.UUID, asset domain.CapitalAsset, currentStep int, status repositories.ItemStatus) error {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`UPDATE capital_assets
		 SET name=?, purchase_cost=?, useful_life_months=?, salvage_value=?, purchase_month_index=?, depreciation_method=?, status=?, current_step=?, updated_at=?
		 WHERE id = ?`,
		asset.Name, int64(asset.PurchaseCost), asset.UsefulLifeMonths, int64(asset.SalvageValue), int64(asset.PurchaseMonthIndex), string(asset.DepreciationMethod), string(status), currentStep, now,
		itemID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to save capital asset step: %w", err)
	}
	return checkRowsAffected(res, "capital asset not found")
}

// ListCompleteCapitalAssets returns every finished Fixed Asset for a plan,
// in the order they were added.
func (s *SQLiteStore) ListCompleteCapitalAssets(planID uuid.UUID) ([]repositories.CapitalAssetItem, error) {
	rows, err := s.db.Query(
		"SELECT "+capitalAssetColumns+" FROM capital_assets WHERE plan_id = ? AND status = ? ORDER BY sort_order ASC",
		planID.String(), string(repositories.StatusComplete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query capital assets: %w", err)
	}
	defer rows.Close()

	var items []repositories.CapitalAssetItem
	for rows.Next() {
		item, err := scanCapitalAssetItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan capital asset: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate capital assets: %w", err)
	}
	return items, nil
}

// DeleteCapitalAsset removes a single Fixed Asset item.
func (s *SQLiteStore) DeleteCapitalAsset(itemID uuid.UUID) error {
	res, err := s.db.Exec("DELETE FROM capital_assets WHERE id = ?", itemID.String())
	if err != nil {
		return fmt.Errorf("failed to delete capital asset: %w", err)
	}
	return checkRowsAffected(res, "capital asset not found")
}

// --- Startup Costs ---

func scanStartupCostItem(row rowScanner) (*repositories.StartupCostItem, error) {
	var idStr, name, status string
	var amount int64
	var currentStep int

	if err := row.Scan(&idStr, &name, &amount, &status, &currentStep); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid startup cost id: %w", err)
	}

	return &repositories.StartupCostItem{
		ID:          id,
		Cost:        domain.StartupCost{Name: name, Amount: domain.Money(amount)},
		Status:      repositories.ItemStatus(status),
		CurrentStep: currentStep,
	}, nil
}

const startupCostColumns = "id, name, amount, status, current_step"

// CreateStartupCostDraft starts a new Startup Cost wizard item, or resumes
// the plan's existing draft if one is already present.
func (s *SQLiteStore) CreateStartupCostDraft(planID uuid.UUID) (uuid.UUID, error) {
	existing, err := s.GetStartupCostDraft(planID)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, nil
	}

	var sortOrder int
	if err := s.db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM startup_costs WHERE plan_id = ?", planID.String()).Scan(&sortOrder); err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.New()
	now := time.Now().Unix()
	if _, err := s.db.Exec(
		`INSERT INTO startup_costs (id, plan_id, status, current_step, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id.String(), planID.String(), string(repositories.StatusDraft), sortOrder, now, now,
	); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create startup cost draft: %w", err)
	}
	return id, nil
}

// GetStartupCostDraft returns the plan's in-progress Startup Cost draft, if
// any, without creating one.
func (s *SQLiteStore) GetStartupCostDraft(planID uuid.UUID) (*repositories.StartupCostItem, error) {
	row := s.db.QueryRow(
		"SELECT "+startupCostColumns+" FROM startup_costs WHERE plan_id = ? AND status = ? LIMIT 1",
		planID.String(), string(repositories.StatusDraft),
	)
	item, err := scanStartupCostItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query startup cost draft: %w", err)
	}
	return item, nil
}

// GetStartupCost retrieves a single Startup Cost wizard item by its own ID.
func (s *SQLiteStore) GetStartupCost(itemID uuid.UUID) (*repositories.StartupCostItem, error) {
	row := s.db.QueryRow(
		"SELECT "+startupCostColumns+" FROM startup_costs WHERE id = ?",
		itemID.String(),
	)
	item, err := scanStartupCostItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("startup cost not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query startup cost: %w", err)
	}
	return item, nil
}

// SaveStartupCostStep persists the wizard's current answer for a Startup
// Cost item, advancing its step/status.
func (s *SQLiteStore) SaveStartupCostStep(itemID uuid.UUID, cost domain.StartupCost, currentStep int, status repositories.ItemStatus) error {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`UPDATE startup_costs SET name=?, amount=?, status=?, current_step=?, updated_at=? WHERE id = ?`,
		cost.Name, int64(cost.Amount), string(status), currentStep, now, itemID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to save startup cost step: %w", err)
	}
	return checkRowsAffected(res, "startup cost not found")
}

// ListCompleteStartupCosts returns every finished Startup Cost for a plan,
// in the order they were added.
func (s *SQLiteStore) ListCompleteStartupCosts(planID uuid.UUID) ([]repositories.StartupCostItem, error) {
	rows, err := s.db.Query(
		"SELECT "+startupCostColumns+" FROM startup_costs WHERE plan_id = ? AND status = ? ORDER BY sort_order ASC",
		planID.String(), string(repositories.StatusComplete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query startup costs: %w", err)
	}
	defer rows.Close()

	var items []repositories.StartupCostItem
	for rows.Next() {
		item, err := scanStartupCostItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan startup cost: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate startup costs: %w", err)
	}
	return items, nil
}

// DeleteStartupCost removes a single Startup Cost item.
func (s *SQLiteStore) DeleteStartupCost(itemID uuid.UUID) error {
	res, err := s.db.Exec("DELETE FROM startup_costs WHERE id = ?", itemID.String())
	if err != nil {
		return fmt.Errorf("failed to delete startup cost: %w", err)
	}
	return checkRowsAffected(res, "startup cost not found")
}

// --- Funding Sources ---

func scanFundingSourceItem(row rowScanner) (*repositories.FundingSourceItem, error) {
	var idStr, name, status string
	var amount int64
	var interestRate float64
	var termMonths, currentStep int

	if err := row.Scan(&idStr, &name, &amount, &interestRate, &termMonths, &status, &currentStep); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid funding source id: %w", err)
	}

	return &repositories.FundingSourceItem{
		ID: id,
		Funding: domain.FundingSource{
			Name:         name,
			Amount:       domain.Money(amount),
			InterestRate: interestRate,
			TermMonths:   termMonths,
		},
		Status:      repositories.ItemStatus(status),
		CurrentStep: currentStep,
	}, nil
}

const fundingSourceColumns = "id, name, amount, interest_rate, term_months, status, current_step"

// CreateFundingSourceDraft starts a new Funding Source wizard item, or
// resumes the plan's existing draft if one is already present.
func (s *SQLiteStore) CreateFundingSourceDraft(planID uuid.UUID) (uuid.UUID, error) {
	existing, err := s.GetFundingSourceDraft(planID)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return existing.ID, nil
	}

	var sortOrder int
	if err := s.db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM funding_sources WHERE plan_id = ?", planID.String()).Scan(&sortOrder); err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute sort order: %w", err)
	}

	id := uuid.New()
	now := time.Now().Unix()
	if _, err := s.db.Exec(
		`INSERT INTO funding_sources (id, plan_id, status, current_step, sort_order, created_at, updated_at)
		 VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id.String(), planID.String(), string(repositories.StatusDraft), sortOrder, now, now,
	); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create funding source draft: %w", err)
	}
	return id, nil
}

// GetFundingSourceDraft returns the plan's in-progress Funding Source
// draft, if any, without creating one.
func (s *SQLiteStore) GetFundingSourceDraft(planID uuid.UUID) (*repositories.FundingSourceItem, error) {
	row := s.db.QueryRow(
		"SELECT "+fundingSourceColumns+" FROM funding_sources WHERE plan_id = ? AND status = ? LIMIT 1",
		planID.String(), string(repositories.StatusDraft),
	)
	item, err := scanFundingSourceItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query funding source draft: %w", err)
	}
	return item, nil
}

// GetFundingSource retrieves a single Funding Source wizard item by its own
// ID.
func (s *SQLiteStore) GetFundingSource(itemID uuid.UUID) (*repositories.FundingSourceItem, error) {
	row := s.db.QueryRow(
		"SELECT "+fundingSourceColumns+" FROM funding_sources WHERE id = ?",
		itemID.String(),
	)
	item, err := scanFundingSourceItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("funding source not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query funding source: %w", err)
	}
	return item, nil
}

// SaveFundingSourceStep persists the wizard's current answer for a Funding
// Source item, advancing its step/status.
func (s *SQLiteStore) SaveFundingSourceStep(itemID uuid.UUID, funding domain.FundingSource, currentStep int, status repositories.ItemStatus) error {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`UPDATE funding_sources SET name=?, amount=?, interest_rate=?, term_months=?, status=?, current_step=?, updated_at=? WHERE id = ?`,
		funding.Name, int64(funding.Amount), funding.InterestRate, funding.TermMonths, string(status), currentStep, now, itemID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to save funding source step: %w", err)
	}
	return checkRowsAffected(res, "funding source not found")
}

// ListCompleteFundingSources returns every finished Funding Source for a
// plan, in the order they were added.
func (s *SQLiteStore) ListCompleteFundingSources(planID uuid.UUID) ([]repositories.FundingSourceItem, error) {
	rows, err := s.db.Query(
		"SELECT "+fundingSourceColumns+" FROM funding_sources WHERE plan_id = ? AND status = ? ORDER BY sort_order ASC",
		planID.String(), string(repositories.StatusComplete),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query funding sources: %w", err)
	}
	defer rows.Close()

	var items []repositories.FundingSourceItem
	for rows.Next() {
		item, err := scanFundingSourceItem(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan funding source: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate funding sources: %w", err)
	}
	return items, nil
}

// DeleteFundingSource removes a single Funding Source item.
func (s *SQLiteStore) DeleteFundingSource(itemID uuid.UUID) error {
	res, err := s.db.Exec("DELETE FROM funding_sources WHERE id = ?", itemID.String())
	if err != nil {
		return fmt.Errorf("failed to delete funding source: %w", err)
	}
	return checkRowsAffected(res, "funding source not found")
}

// --- Cash on Hand (singleton per plan) ---

// GetStartingBalancesRow returns a plan's Cash on Hand wizard state. A
// plan that has never saved any Cash on Hand step has no row yet; this
// returns a zero-value StartingBalancesRow (CurrentStep 0) rather than
// creating one, since reads shouldn't have write side effects - the first
// SaveStartingBalancesStep call creates the row via upsert.
func (s *SQLiteStore) GetStartingBalancesRow(planID uuid.UUID) (*repositories.StartingBalancesRow, error) {
	var cash, ar, pe, ap, ae int64
	var currentStep int

	err := s.db.QueryRow(
		`SELECT cash, accounts_receivable, prepaid_expenses, accounts_payable, accrued_expenses, current_step
		 FROM starting_balances WHERE plan_id = ?`,
		planID.String(),
	).Scan(&cash, &ar, &pe, &ap, &ae, &currentStep)

	if errors.Is(err, sql.ErrNoRows) {
		return &repositories.StartingBalancesRow{Balances: domain.StartingBalances{}, CurrentStep: 0}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query starting balances: %w", err)
	}

	return &repositories.StartingBalancesRow{
		Balances: domain.StartingBalances{
			Cash:               domain.Money(cash),
			AccountsReceivable: domain.Money(ar),
			PrepaidExpenses:    domain.Money(pe),
			AccountsPayable:    domain.Money(ap),
			AccruedExpenses:    domain.Money(ae),
		},
		CurrentStep: currentStep,
	}, nil
}

// SaveStartingBalancesStep persists the wizard's current answer for Cash on
// Hand, creating the plan's singleton row on first call.
func (s *SQLiteStore) SaveStartingBalancesStep(planID uuid.UUID, balances domain.StartingBalances, currentStep int) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO starting_balances (plan_id, cash, accounts_receivable, prepaid_expenses, accounts_payable, accrued_expenses, current_step, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(plan_id) DO UPDATE SET
			cash=excluded.cash, accounts_receivable=excluded.accounts_receivable, prepaid_expenses=excluded.prepaid_expenses,
			accounts_payable=excluded.accounts_payable, accrued_expenses=excluded.accrued_expenses,
			current_step=excluded.current_step, updated_at=excluded.updated_at`,
		planID.String(), int64(balances.Cash), int64(balances.AccountsReceivable), int64(balances.PrepaidExpenses), int64(balances.AccountsPayable), int64(balances.AccruedExpenses), currentStep, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to save starting balances step: %w", err)
	}
	return nil
}

// --- Section-level explicit completion ---
//
// Generalized across all four wizard hubs (Starting Point, Payroll, Sales
// Forecast, Cash Flow) via a "hub" key, rather than one hub-specific table
// and method pair each. Lives here because this file was the first (and
// for a while only) consumer; the mechanism itself is hub-agnostic.

// wizardHubSections lists the known sections for each hub, so
// GetWizardSectionStatus can report "not started" (false) for a section
// that has no wizard_sections row yet, not just omit it from the map.
var wizardHubSections = map[string][]string{
	domain.HubStartingPoint:     {domain.SectionFixedAssets, domain.SectionStartupCosts, domain.SectionFundingSources, domain.SectionCashOnHand},
	domain.HubPayroll:           {domain.SectionSalaryRoles, domain.SectionBenefits, domain.SectionPayrollTaxRates},
	domain.HubSalesForecast:     {domain.SectionProducts, domain.SectionSalesGrowthCurve},
	domain.HubOperatingExpenses: {domain.SectionOperatingExpenses},
	domain.HubCashFlow:          {domain.SectionInventoryPurchases, domain.SectionDistributions},
}

// MarkWizardSectionComplete records that the user explicitly finished a
// section of a wizard hub. This is a one-way marker: later deleting all of
// a section's items does not un-complete it.
func (s *SQLiteStore) MarkWizardSectionComplete(planID uuid.UUID, hub, section string) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO wizard_sections (plan_id, hub, section, completed_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(plan_id, hub, section) DO UPDATE SET completed_at = excluded.completed_at`,
		planID.String(), hub, section, now,
	)
	if err != nil {
		return fmt.Errorf("failed to mark section complete: %w", err)
	}
	return nil
}

// GetWizardSectionStatus returns, for every known section of hub, whether
// the user has explicitly completed it.
func (s *SQLiteStore) GetWizardSectionStatus(planID uuid.UUID, hub string) (map[string]bool, error) {
	status := make(map[string]bool, len(wizardHubSections[hub]))
	for _, section := range wizardHubSections[hub] {
		status[section] = false
	}

	rows, err := s.db.Query(
		"SELECT section, completed_at FROM wizard_sections WHERE plan_id = ? AND hub = ?",
		planID.String(), hub,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query wizard sections: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var section string
		var completedAt sql.NullInt64
		if err := rows.Scan(&section, &completedAt); err != nil {
			return nil, fmt.Errorf("failed to scan wizard section: %w", err)
		}
		status[section] = completedAt.Valid && completedAt.Int64 > 0
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate wizard sections: %w", err)
	}

	return status, nil
}

// checkRowsAffected returns notFoundErr if the exec touched zero rows.
func checkRowsAffected(res sql.Result, notFoundErr string) error {
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check result: %w", err)
	}
	if rows == 0 {
		return errors.New(notFoundErr)
	}
	return nil
}
