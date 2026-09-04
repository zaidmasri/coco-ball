package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"uuid"

	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	db "github.com/zaidmasri/business-planning-tool/internal/infrastructure/db/sqlc"
)

// softDeleteTables/hardDeleteTables mirror the table lists Delete must clear
// alongside the plan row itself, preserved from the retired hand-rolled
// store's Delete method (formerly internal/store/sqlite.go, before the
// sqlc/sql-migrate migration). Table names can't be parameterized in sqlc,
// so each gets its own named query (see sql/queries/plan.sql).

// PlanRepository implements repositories.PlanRepository.
type PlanRepository struct {
	conn    *sql.DB
	queries *db.Queries
}

// NewPlanRepository constructs a PlanRepository sharing the given
// connection with the rest of the internal/infrastructure/sqlite package.
func NewPlanRepository(conn *sql.DB) repositories.PlanRepository {
	return &PlanRepository{conn: conn, queries: db.New(conn)}
}

var _ repositories.PlanRepository = (*PlanRepository)(nil)

// Save persists a validated plan and atomically writes any accumulated
// domain events to the outbox table, in one transaction.
func (r *PlanRepository) Save(vp domain.ValidatedPlan) error {
	ctx := context.Background()

	tx, err := r.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	if err := savePlanRow(ctx, qtx, vp.Plan(), time.Now().Unix()); err != nil {
		return err
	}

	return tx.Commit()
}

// SaveNew persists a brand-new validated plan and grants its owner Owner
// access in the same transaction, so a GrantAccess failure can never leave
// an ownerless plan behind (the CreatePlan+GrantAccess split this used to
// require was two independent, non-transactional application-service
// calls). ownerID is trusted to already name a real, existing User -
// PlanService.CreatePlan's verifyOwner step covers that before this is
// called, and the plan row itself is inserted in this same transaction, so
// no existence check is needed here.
func (r *PlanRepository) SaveNew(vp domain.ValidatedPlan, ownerID uuid.UUID) error {
	ctx := context.Background()

	tx, err := r.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	p := vp.Plan()
	if err := savePlanRow(ctx, qtx, p, time.Now().Unix()); err != nil {
		return err
	}

	if err := qtx.GrantAccess(ctx, db.GrantAccessParams{
		PlanID:      p.ID().String(),
		UserID:      ownerID.String(),
		AccessLevel: string(domain.Owner),
		InvitedAt:   0,
	}); err != nil {
		return fmt.Errorf("failed to grant owner access: %w", err)
	}

	return tx.Commit()
}

// SaveWithInvite persists a PlanInvite row and the plan's data blob/outbox
// events in one transaction, so a failure between the two writes can never
// silently drop the invite's UserInvitedToPlan event (the bug this method
// replaces: InviteRepository.CreateInvite and Plan.Save used to be two
// independent, non-atomic writes).
func (r *PlanRepository) SaveWithInvite(vp domain.ValidatedPlan, invite domain.ValidatedPlanInvite) error {
	ctx := context.Background()

	tx, err := r.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	now := time.Now().Unix()

	if err := insertInviteRow(ctx, qtx, invite, now); err != nil {
		return err
	}

	if err := savePlanRow(ctx, qtx, vp.Plan(), now); err != nil {
		return err
	}

	return tx.Commit()
}

// savePlanRow writes a plan's data blob and drains/persists its accumulated
// domain events to the outbox table, using queries already bound to an
// open transaction (qtx) - callers own the transaction's begin/commit.
func savePlanRow(ctx context.Context, qtx *db.Queries, p *domain.Plan, now int64) error {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	if err := qtx.SavePlan(ctx, db.SavePlanParams{
		ID:        p.ID().String(),
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	return insertOutboxEvents(ctx, qtx, p.PullEvents(), now)
}

// Get retrieves a plan by ID, hydrating its wizard sub-entities from the
// normalized tables on top of the plans.data JSON blob.
func (r *PlanRepository) Get(id uuid.UUID) (*domain.Plan, error) {
	ctx := context.Background()

	data, err := r.queries.GetPlan(ctx, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("plan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query plan: %w", err)
	}

	var plan domain.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan: %w", err)
	}

	if err := hydratePlan(ctx, r.queries, &plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

// GetAll retrieves every non-deleted plan.
func (r *PlanRepository) GetAll() ([]*domain.Plan, error) {
	ctx := context.Background()

	rows, err := r.queries.GetAllPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %w", err)
	}

	return hydrateAllPlans(ctx, r.queries, rows)
}

// GetUserPlans retrieves every non-deleted plan the given user has access to.
func (r *PlanRepository) GetUserPlans(userID uuid.UUID) ([]*domain.Plan, error) {
	ctx := context.Background()

	rows, err := r.queries.GetUserPlans(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query user plans: %w", err)
	}

	return hydrateAllPlans(ctx, r.queries, rows)
}

// hydrateAllPlans unmarshals and hydrates a batch of plans.data blobs. Also
// used by AccessRepository.GetUserPlans, which runs the same join query
// (see sql/queries/access.sql's GetPlansForUser) but is a separate struct.
func hydrateAllPlans(ctx context.Context, queries *db.Queries, blobs [][]byte) ([]*domain.Plan, error) {
	plans := make([]*domain.Plan, 0, len(blobs))
	for _, data := range blobs {
		var plan domain.Plan
		if err := json.Unmarshal(data, &plan); err != nil {
			return nil, fmt.Errorf("failed to unmarshal plan: %w", err)
		}
		if err := hydratePlan(ctx, queries, &plan); err != nil {
			return nil, err
		}
		plans = append(plans, &plan)
	}
	return plans, nil
}

// Delete soft-deletes a plan and hard/soft-deletes every wizard sub-entity
// row belonging to it, in one transaction.
func (r *PlanRepository) Delete(id uuid.UUID) error {
	ctx := context.Background()
	planID := id.String()
	nowUnix := time.Now().Unix()
	now := sql.NullInt64{Int64: nowUnix, Valid: true}

	// Load the plan first (mirrors UserRepository.DeleteUser's precedent of
	// fetching before opening the transaction) so Delete() has an aggregate
	// to record the PlanDeleted event on - without this, the event would
	// have to be constructed directly by the repository, bypassing the
	// aggregate the same way the pre-fix code bypassed it entirely.
	plan, err := r.Get(id)
	if err != nil {
		return err
	}
	plan.Delete()

	tx, err := r.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	softDeletes := []func(context.Context, string) error{
		func(ctx context.Context, planID string) error {
			return qtx.SoftDeleteCapitalAssetsByPlan(ctx, db.SoftDeleteCapitalAssetsByPlanParams{DeletedAt: now, PlanID: planID})
		},
		func(ctx context.Context, planID string) error {
			return qtx.SoftDeleteStartupCostsByPlan(ctx, db.SoftDeleteStartupCostsByPlanParams{DeletedAt: now, PlanID: planID})
		},
		func(ctx context.Context, planID string) error {
			return qtx.SoftDeleteFundingSourcesByPlan(ctx, db.SoftDeleteFundingSourcesByPlanParams{DeletedAt: now, PlanID: planID})
		},
		func(ctx context.Context, planID string) error {
			return qtx.SoftDeleteSalaryRolesByPlan(ctx, db.SoftDeleteSalaryRolesByPlanParams{DeletedAt: now, PlanID: planID})
		},
		func(ctx context.Context, planID string) error {
			return qtx.SoftDeleteBenefitsByPlan(ctx, db.SoftDeleteBenefitsByPlanParams{DeletedAt: now, PlanID: planID})
		},
		func(ctx context.Context, planID string) error {
			return qtx.SoftDeleteProductsByPlan(ctx, db.SoftDeleteProductsByPlanParams{DeletedAt: now, PlanID: planID})
		},
		func(ctx context.Context, planID string) error {
			return qtx.SoftDeleteInventoryPurchasesByPlan(ctx, db.SoftDeleteInventoryPurchasesByPlanParams{DeletedAt: now, PlanID: planID})
		},
		func(ctx context.Context, planID string) error {
			return qtx.SoftDeleteDistributionsByPlan(ctx, db.SoftDeleteDistributionsByPlanParams{DeletedAt: now, PlanID: planID})
		},
		func(ctx context.Context, planID string) error {
			return qtx.SoftDeleteOperatingExpensesByPlan(ctx, db.SoftDeleteOperatingExpensesByPlanParams{DeletedAt: now, PlanID: planID})
		},
	}
	for _, del := range softDeletes {
		if err := del(ctx, planID); err != nil {
			return fmt.Errorf("failed to soft-delete plan children: %w", err)
		}
	}

	hardDeletes := []func(context.Context, string) error{
		qtx.DeletePlanAccessByPlan,
		qtx.DeleteStartingBalancesByPlan,
		qtx.DeletePayrollTaxRatesByPlan,
		qtx.DeleteSalesGrowthCurveByPlan,
		qtx.DeleteWizardSectionsByPlan,
	}
	for _, del := range hardDeletes {
		if err := del(ctx, planID); err != nil {
			return fmt.Errorf("failed to delete plan children: %w", err)
		}
	}

	affected, err := qtx.SoftDeletePlan(ctx, db.SoftDeletePlanParams{DeletedAt: now, ID: planID})
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}
	if affected == 0 {
		return errors.New("plan not found")
	}

	if err := insertOutboxEvents(ctx, qtx, plan.PullEvents(), nowUnix); err != nil {
		return err
	}

	return tx.Commit()
}

// hydratePlan attaches every wizard sub-entity to plan by querying the
// normalized tables directly, mirroring the retired hand-rolled store's
// loadXInto helpers (formerly internal/store/sqlite_starting_point.go,
// sqlite_payroll.go, sqlite_sales_forecast.go, sqlite_cash_flow.go,
// sqlite_operating_expenses.go).
func hydratePlan(ctx context.Context, queries *db.Queries, plan *domain.Plan) error {
	planID := plan.ID().String()
	complete := string(repositories.StatusComplete)

	assets, err := queries.ListCompleteCapitalAssets(ctx, db.ListCompleteCapitalAssetsParams{PlanID: planID, Status: complete})
	if err != nil {
		return fmt.Errorf("failed to load capital assets: %w", err)
	}
	capitalAssets := make([]domain.CapitalAsset, len(assets))
	for i, a := range assets {
		capitalAssets[i] = domain.CapitalAsset{
			Name:               a.Name,
			PurchaseCost:       mustUSD(a.PurchaseCost),
			UsefulLifeMonths:   int(a.UsefulLifeMonths),
			SalvageValue:       mustUSD(a.SalvageValue),
			PurchaseMonthIndex: domain.MonthIndex(a.PurchaseMonthIndex),
			DepreciationMethod: domain.DepreciationMethod(a.DepreciationMethod),
		}
	}

	startupCostRows, err := queries.ListCompleteStartupCosts(ctx, db.ListCompleteStartupCostsParams{PlanID: planID, Status: complete})
	if err != nil {
		return fmt.Errorf("failed to load startup costs: %w", err)
	}
	startupCosts := make([]domain.StartupCost, len(startupCostRows))
	for i, c := range startupCostRows {
		startupCosts[i] = domain.StartupCost{Name: c.Name, Amount: mustUSD(c.Amount)}
	}

	fundingRows, err := queries.ListCompleteFundingSources(ctx, db.ListCompleteFundingSourcesParams{PlanID: planID, Status: complete})
	if err != nil {
		return fmt.Errorf("failed to load funding sources: %w", err)
	}
	fundingSources := make([]domain.FundingSource, len(fundingRows))
	for i, f := range fundingRows {
		fundingSources[i] = domain.FundingSource{
			Name:         f.Name,
			Amount:       mustUSD(f.Amount),
			InterestRate: f.InterestRate,
			TermMonths:   int(f.TermMonths),
		}
	}

	var balances domain.StartingBalances
	balanceRow, err := queries.GetStartingBalancesRow(ctx, planID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to load starting balances: %w", err)
	}
	if err == nil {
		balances = domain.StartingBalances{
			Cash:               mustUSD(balanceRow.Cash),
			AccountsReceivable: mustUSD(balanceRow.AccountsReceivable),
			PrepaidExpenses:    mustUSD(balanceRow.PrepaidExpenses),
			AccountsPayable:    mustUSD(balanceRow.AccountsPayable),
			AccruedExpenses:    mustUSD(balanceRow.AccruedExpenses),
		}
	}

	plan.LoadStartingPointData(capitalAssets, startupCosts, fundingSources, balances)

	roleRows, err := queries.ListCompleteSalaryRoles(ctx, db.ListCompleteSalaryRolesParams{PlanID: planID, Status: complete})
	if err != nil {
		return fmt.Errorf("failed to load salary roles: %w", err)
	}
	salaryRoles := make([]domain.SalaryRole, len(roleRows))
	for i, s := range roleRows {
		salaryRoles[i] = domain.SalaryRole{
			Role:           s.Role,
			IsContractor:   int64ToBool(s.IsContractor),
			Headcount:      int(s.Headcount),
			MonthlyPay:     mustUSD(s.MonthlyPay),
			GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{s.GrowthYr2, s.GrowthYr3}},
		}
	}

	benefitRows, err := queries.ListCompleteBenefits(ctx, db.ListCompleteBenefitsParams{PlanID: planID, Status: complete})
	if err != nil {
		return fmt.Errorf("failed to load benefits: %w", err)
	}
	benefits := make([]domain.Benefit, len(benefitRows))
	for i, b := range benefitRows {
		benefits[i] = domain.Benefit{
			Type:           b.Type,
			MonthlyAmount:  mustUSD(b.MonthlyAmount),
			GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{b.GrowthYr2, b.GrowthYr3}},
		}
	}

	var taxRates domain.PayrollTaxRates
	taxRow, err := queries.GetPayrollTaxRatesRow(ctx, planID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to load payroll tax rates: %w", err)
	}
	if err == nil {
		taxRates = domain.PayrollTaxRates{
			SocialSecurityRate: taxRow.SocialSecurityRate,
			MedicareRate:       taxRow.MedicareRate,
			FUTARate:           taxRow.FutaRate,
			SUTARate:           taxRow.SutaRate,
		}
	}

	plan.LoadPayrollData(salaryRoles, benefits, taxRates)

	productRows, err := queries.ListCompleteProducts(ctx, db.ListCompleteProductsParams{PlanID: planID, Status: complete})
	if err != nil {
		return fmt.Errorf("failed to load products: %w", err)
	}
	products := make([]domain.Product, len(productRows))
	for i, p := range productRows {
		products[i] = domain.Product{
			Name:         p.Name,
			Month1Units:  int(p.Month1Units),
			PricePerUnit: mustUSD(p.PricePerUnit),
			CostPerUnit:  mustUSD(p.CostPerUnit),
		}
	}

	var salesGrowth domain.SalesGrowthCurve
	curveRow, err := queries.GetSalesGrowthCurveRow(ctx, planID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to load sales growth curve: %w", err)
	}
	if err == nil {
		salesGrowth = domain.SalesGrowthCurve{
			Year1QuarterlyRates: [4]float64{curveRow.Y1Q1, curveRow.Y1Q2, curveRow.Y1Q3, curveRow.Y1Q4},
			FutureYearRates:     []float64{curveRow.GrowthYr2, curveRow.GrowthYr3},
		}
	}

	plan.LoadSalesForecastData(products, salesGrowth)

	invRows, err := queries.ListCompleteInventoryPurchases(ctx, db.ListCompleteInventoryPurchasesParams{PlanID: planID, Status: complete})
	if err != nil {
		return fmt.Errorf("failed to load inventory purchases: %w", err)
	}
	inventory := make([]domain.InventoryPurchase, len(invRows))
	for i, inv := range invRows {
		inventory[i] = domain.InventoryPurchase{
			Category:       inv.Category,
			MonthlyAmount:  mustUSD(inv.MonthlyAmount),
			GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{inv.GrowthYr2, inv.GrowthYr3}},
		}
	}

	distRows, err := queries.ListCompleteDistributions(ctx, db.ListCompleteDistributionsParams{PlanID: planID, Status: complete})
	if err != nil {
		return fmt.Errorf("failed to load distributions: %w", err)
	}
	distributions := make([]domain.Distribution, len(distRows))
	for i, d := range distRows {
		distributions[i] = domain.Distribution{
			Name:           d.Name,
			MonthlyAmount:  mustUSD(d.MonthlyAmount),
			GrowthAfterYr1: domain.AnnualGrowth{RatesAfterYear1: []float64{d.GrowthYr2, d.GrowthYr3}},
		}
	}

	plan.LoadCashFlowData(inventory, distributions)

	opExRows, err := queries.ListCompleteOperatingExpenses(ctx, db.ListCompleteOperatingExpensesParams{PlanID: planID, Status: complete})
	if err != nil {
		return fmt.Errorf("failed to load operating expenses: %w", err)
	}
	opEx := make([]domain.Cost, len(opExRows))
	for i, e := range opExRows {
		id, err := uuid.Parse(e.ID)
		if err != nil {
			return fmt.Errorf("failed to parse operating expense id: %w", err)
		}
		var cost domain.Cost
		cost.SetID(id)
		cost.SetName(e.Name)
		cost.SetBaseAmountPerMonth(mustUSD(e.MonthlyAmount))
		cost.SetGrowth(domain.GrowthStrategy{
			Type:       domain.GrowthType(e.GrowthType),
			AnnualRate: e.AnnualRate,
		})
		opEx[i] = cost
	}

	plan.LoadOpExData(opEx)

	return nil
}
