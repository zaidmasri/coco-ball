// Package domain: financial projection engine.
//
// This file turns the raw plan inputs (products + sales growth curve,
// payroll, operating expenses, capital assets, funding sources, and
// discretionary cash flow items) into month-by-month and year-by-year
// financial statements.
//
// Simplifying assumptions (documented here since they affect every number
// this file produces):
//
//   - Income tax is not modeled. NetIncome is pre-tax. There is no tax-rate
//     input anywhere in the app, so showing a fabricated rate would be worse
//     than clearly showing $0 and leaving tax planning to the user/accountant.
//   - A FundingSource is treated as an amortizing loan only if it has both a
//     positive InterestRate and a positive TermMonths; otherwise its full
//     Amount is treated as an equity contribution (received once, in cash,
//     before month 0, with no debt service).
//   - Accounts Receivable, Prepaid Expenses, and Accounts Payable/Accrued
//     Expenses are held static at their Starting Point balances for the
//     entire projection (no revenue-driven AR/AP growth is modeled).
//   - "Additional Inventory" cash-flow purchases are treated as an
//     asset-for-asset swap (cash -> inventory) with no depletion modeled, so
//     they never touch the Income Statement, only the Balance Sheet.
//   - Startup Costs are treated as a pre-operations use of funds: they
//     reduce the plan's opening cash balance but never appear as an
//     Income Statement expense.
//   - Starting Point's Cash/AR/Prepaid/AP/AccruedExpenses balances describe
//     an already-operating business, not new money raised for this plan.
//     Their net value (assets minus liabilities) is folded into the opening
//     Retained Earnings baseline as implied pre-existing equity, since
//     nothing else in the domain model accounts for it.
//
// These choices keep the accounting identity Assets = Liabilities + Equity
// true by construction (up to whole-dollar rounding) for every month of the
// projection — see TestBalanceSheetSnapshots_Balance in projection_test.go.
package entities

import "math"

// ProjectionYears is how many years of financial statements the app
// generates (matches the 3-year horizon used throughout the SCORE template).
const ProjectionYears = 3

// ProjectionMonths is ProjectionYears expressed in months.
const ProjectionMonths = ProjectionYears * 12

// MonthlyFinancials is the fully computed financial position for a single
// month of the projection, including running/cumulative balances needed to
// build balance sheet snapshots without re-walking the whole history.
type MonthlyFinancials struct {
	Month MonthIndex

	Revenue     Money
	COGS        Money
	GrossProfit Money

	PayrollCost  Money // salaries + benefits, compounded by their growth schedules
	PayrollTax   Money // employer-side payroll taxes on W-2 (non-contractor) pay
	OpEx         Money
	Depreciation Money

	InterestExpense  Money
	PrincipalPayment Money

	OperatingIncome    Money // GrossProfit - OpEx - PayrollCost - PayrollTax - Depreciation
	NetIncomeBeforeTax Money // OperatingIncome - InterestExpense
	NetIncome          Money // == NetIncomeBeforeTax; no tax is modeled

	CapitalExpenditure Money
	InventoryPurchase  Money
	Distribution       Money
	NetCashFlow        Money

	// Running/cumulative balances as of the end of this month.
	CashBalance             Money
	AccumulatedDepreciation Money
	FixedAssetsGross        Money
	InventoryBalance        Money
	LoanBalance             Money
	RetainedEarnings        Money
}

// AnnualFinancials aggregates MonthlyFinancials income-statement lines over
// a single year.
type AnnualFinancials struct {
	Year int

	Revenue         Money
	COGS            Money
	GrossProfit     Money
	PayrollCost     Money
	PayrollTax      Money
	OpEx            Money
	Depreciation    Money
	InterestExpense Money
	OperatingIncome Money
	NetIncome       Money
}

// BalanceSheetSnapshot is the plan's financial position as of the end of a
// given year.
type BalanceSheetSnapshot struct {
	Year int

	Cash               Money
	AccountsReceivable Money
	Inventory          Money
	PrepaidExpenses    Money
	CurrentAssets      Money

	FixedAssetsGross        Money
	AccumulatedDepreciation Money
	FixedAssetsNet          Money

	TotalAssets Money

	AccountsPayable    Money
	AccruedExpenses    Money
	CurrentLiabilities Money
	LoansPayable       Money
	TotalLiabilities   Money

	OwnersEquity     Money
	RetainedEarnings Money
	TotalEquity      Money

	TotalLiabilitiesAndEquity Money
}

// FinancialRatios holds the standard liquidity/safety/profitability ratios
// for a single year. Ratios that are undefined for the year (e.g. Sales
// Growth in Year 1, or any ratio with a zero denominator) are left at 0 —
// templates should treat 0 for a growth/ratio value that can't be computed
// as "not applicable" rather than a literal measured zero.
type FinancialRatios struct {
	Year int

	CurrentRatio float64
	QuickRatio   float64

	DebtToEquity float64
	DSCR         float64

	SalesGrowthPercent       float64
	GrossProfitMarginPercent float64
	NetProfitMarginPercent   float64
	ReturnOnEquityPercent    float64
}

// BreakevenAnalysis is the Year 1 breakeven calculation: the annual sales
// level at which Gross Margin exactly covers fixed (payroll + operating)
// expenses.
type BreakevenAnalysis struct {
	GrossMargin        Money
	TotalSales         Money
	GrossMarginPercent float64

	PayrollFixedCost          Money
	OperatingExpenseFixedCost Money
	TotalFixedExpenses        Money

	BreakevenAnnual  Money
	BreakevenMonthly Money
}

// isAmortizingLoan reports whether a funding source should be treated as a
// debt instrument with a monthly payment schedule, as opposed to an equity
// contribution.
func isAmortizingLoan(f FundingSource) bool {
	return f.InterestRate > 0 && f.TermMonths > 0
}

// loanScheduleEntry is one month of a standard fixed-payment amortization
// schedule.
type loanScheduleEntry struct {
	Interest      Money
	Principal     Money
	EndingBalance Money
}

// amortizationSchedule computes a standard fixed-payment loan amortization
// schedule for `horizon` months (zero-padded after payoff or past term).
func amortizationSchedule(principal Money, annualRate float64, termMonths, horizon int) []loanScheduleEntry {
	schedule := make([]loanScheduleEntry, horizon)
	if termMonths <= 0 || !principal.IsPositive() {
		return schedule
	}

	monthlyRate := annualRate / 12.0
	p := float64(principal.MinorUnits())

	var payment float64
	if monthlyRate == 0 {
		payment = p / float64(termMonths)
	} else {
		payment = p * monthlyRate / (1 - math.Pow(1+monthlyRate, -float64(termMonths)))
	}

	balance := p
	for m := 0; m < horizon; m++ {
		if m >= termMonths || balance <= 0 {
			continue
		}

		interest := balance * monthlyRate
		principalPortion := payment - interest
		if principalPortion > balance {
			principalPortion = balance
		}
		balance -= principalPortion

		schedule[m] = loanScheduleEntry{
			Interest:      fromFloatUSD(interest),
			Principal:     fromFloatUSD(principalPortion),
			EndingBalance: fromFloatUSD(balance),
		}
	}

	return schedule
}

// annualGrowthMultiplier compounds an AnnualGrowth schedule (Year 2 / Year 3
// rates) up through the given month. Years beyond the schedule's rates hold
// flat (no further compounding) rather than erroring.
func annualGrowthMultiplier(g AnnualGrowth, month int) float64 {
	yearsPassed := month / 12
	multiplier := 1.0
	for y := 1; y <= yearsPassed; y++ {
		idx := y - 1
		if idx >= len(g.RatesAfterYear1) {
			break
		}
		multiplier *= 1 + g.RatesAfterYear1[idx]
	}
	return multiplier
}

// unitGrowthMultipliers computes the global unit-sales growth multiplier for
// each month of the horizon: Year 1 compounds monthly using the quarterly
// rates, subsequent years compound monthly using their single annual-labeled
// (but literally monthly, per the form's own copy) rate.
func (p *Plan) unitGrowthMultipliers(horizon int) []float64 {
	multipliers := make([]float64, horizon)
	if horizon == 0 {
		return multipliers
	}

	multipliers[0] = 1.0
	curve := p.salesGrowth

	for m := 1; m < horizon; m++ {
		var rate float64
		if m < 12 {
			quarter := m / 3
			if quarter > 3 {
				quarter = 3
			}
			rate = curve.Year1QuarterlyRates[quarter]
		} else {
			yearIdx := m/12 - 1
			if yearIdx >= 0 && yearIdx < len(curve.FutureYearRates) {
				rate = curve.FutureYearRates[yearIdx]
			}
		}
		multipliers[m] = multipliers[m-1] * (1 + rate)
	}

	return multipliers
}

// monthlyProductRevenueAndCOGS sums revenue and COGS across every product
// line for a single month, given that month's global unit-growth multiplier.
func (p *Plan) monthlyProductRevenueAndCOGS(multiplier float64) (revenue, cogs Money) {
	for _, prod := range p.products {
		units := float64(prod.Month1Units) * multiplier
		revenue = revenue.Add(fromFloatUSD(units * float64(prod.PricePerUnit.MinorUnits())))
		cogs = cogs.Add(fromFloatUSD(units * float64(prod.CostPerUnit.MinorUnits())))
	}
	return revenue, cogs
}

// monthlyPayrollCost sums salary + benefit costs (each compounded by its own
// Year 2/3 growth schedule) and the employer payroll tax owed on the W-2
// (non-contractor) portion of salaries for a single month.
func (p *Plan) monthlyPayrollCost(month int) (payrollCost, payrollTax Money) {
	var w2Base Money

	for _, role := range p.salaryRoles {
		amt := fromFloatUSD(float64(role.MonthlyPay.MinorUnits()) * annualGrowthMultiplier(role.GrowthAfterYr1, month))
		payrollCost = payrollCost.Add(amt)
		if !role.IsContractor {
			w2Base = w2Base.Add(amt)
		}
	}

	for _, b := range p.benefits {
		amt := fromFloatUSD(float64(b.MonthlyAmount.MinorUnits()) * annualGrowthMultiplier(b.GrowthAfterYr1, month))
		payrollCost = payrollCost.Add(amt)
	}

	rates := p.payrollTaxRates
	taxRate := rates.SocialSecurityRate + rates.MedicareRate + rates.FUTARate + rates.SUTARate
	payrollTax = fromFloatUSD(float64(w2Base.MinorUnits()) * taxRate)

	return payrollCost, payrollTax
}

// ProjectMonths computes the full month-by-month projection for the given
// horizon (in months), including running cash, depreciation, inventory,
// loan-balance, and retained-earnings balances.
func (p *Plan) ProjectMonths(horizon int) []MonthlyFinancials {
	if horizon <= 0 {
		return nil
	}

	multipliers := p.unitGrowthMultipliers(horizon)

	var loanSchedules [][]loanScheduleEntry
	var openingCash Money
	for _, f := range p.fundingSources {
		openingCash = openingCash.Add(f.Amount)
		if isAmortizingLoan(f) {
			loanSchedules = append(loanSchedules, amortizationSchedule(f.Amount, f.InterestRate, f.TermMonths, horizon))
		}
	}
	openingCash = openingCash.Add(p.startingBalances.Cash)

	var openingRetainedEarnings Money

	// Startup costs are a pre-operations use of cash with no offsetting asset
	// (unlike a capital purchase). To keep Assets = Liabilities + Equity,
	// they must reduce Equity by the same amount they reduce Cash — modeled
	// here as an opening deficit against Retained Earnings.
	for _, sc := range p.startupCosts {
		openingCash = openingCash.Sub(sc.Amount)
		openingRetainedEarnings = openingRetainedEarnings.Sub(sc.Amount)
	}

	// StartingBalances (Cash/AR/Prepaid/AP/AccruedExpenses) describe an
	// already-operating business's balance sheet snapshot, not new money
	// raised for this plan. Their net value is pre-existing equity that
	// predates FundingSources and this projection's Retained Earnings, so it
	// must be folded into the opening equity baseline for the books to
	// balance (Cash/AR/Prepaid are assets with no other offsetting entry;
	// AP/AccruedExpenses are liabilities already counted separately).
	sb := p.startingBalances
	openingRetainedEarnings = openingRetainedEarnings.
		Add(sb.Cash).
		Add(sb.AccountsReceivable).
		Add(sb.PrepaidExpenses).
		Sub(sb.AccountsPayable).
		Sub(sb.AccruedExpenses)

	capexByMonth := make([]Money, horizon)
	for _, asset := range p.futurePurchases {
		m := int(asset.PurchaseMonthIndex)
		if m >= 0 && m < horizon {
			capexByMonth[m] = capexByMonth[m].Add(asset.PurchaseCost)
		}
	}

	results := make([]MonthlyFinancials, horizon)

	runningCash := openingCash
	runningRetainedEarnings := openingRetainedEarnings
	var runningAccumDepr, runningFixedAssetsGross, runningInventory Money

	for m := 0; m < horizon; m++ {
		revenue, cogs := p.monthlyProductRevenueAndCOGS(multipliers[m])
		// Fold in the legacy flat/annual-step Revenue & COGS API (kept for
		// backward compatibility with AddRevenue/AddCOGS); no current
		// handler populates those slices, so in practice this adds 0.
		revenue = revenue.Add(p.MonthlyRevenue(MonthIndex(m)))
		cogs = cogs.Add(p.MonthlyCOGS(MonthIndex(m)))

		payrollCost, payrollTax := p.monthlyPayrollCost(m)
		opex := p.MonthlyOpEx(MonthIndex(m))
		depr := p.MonthlyDepreciation(MonthIndex(m))

		var interest, principal, loanBalance Money
		for _, schedule := range loanSchedules {
			interest = interest.Add(schedule[m].Interest)
			principal = principal.Add(schedule[m].Principal)
			loanBalance = loanBalance.Add(schedule[m].EndingBalance)
		}

		var inventoryPurchase, distribution Money
		for _, inv := range p.inventoryPlan {
			inventoryPurchase = inventoryPurchase.Add(fromFloatUSD(float64(inv.MonthlyAmount.MinorUnits()) * annualGrowthMultiplier(inv.GrowthAfterYr1, m)))
		}
		for _, d := range p.distributions {
			distribution = distribution.Add(fromFloatUSD(float64(d.MonthlyAmount.MinorUnits()) * annualGrowthMultiplier(d.GrowthAfterYr1, m)))
		}

		grossProfit := revenue.Sub(cogs)
		operatingIncome := grossProfit.Sub(opex).Sub(payrollCost).Sub(payrollTax).Sub(depr)
		netIncome := operatingIncome.Sub(interest)

		capex := capexByMonth[m]
		netCashFlow := netIncome.Add(depr).Sub(capex).Sub(inventoryPurchase).Sub(principal).Sub(distribution)
		runningCash = runningCash.Add(netCashFlow)

		runningAccumDepr = runningAccumDepr.Add(depr)
		runningFixedAssetsGross = runningFixedAssetsGross.Add(capex)
		runningInventory = runningInventory.Add(inventoryPurchase)
		runningRetainedEarnings = runningRetainedEarnings.Add(netIncome).Sub(distribution)

		results[m] = MonthlyFinancials{
			Month:              MonthIndex(m),
			Revenue:            revenue,
			COGS:               cogs,
			GrossProfit:        grossProfit,
			PayrollCost:        payrollCost,
			PayrollTax:         payrollTax,
			OpEx:               opex,
			Depreciation:       depr,
			InterestExpense:    interest,
			PrincipalPayment:   principal,
			OperatingIncome:    operatingIncome,
			NetIncomeBeforeTax: netIncome,
			NetIncome:          netIncome,
			CapitalExpenditure: capex,
			InventoryPurchase:  inventoryPurchase,
			Distribution:       distribution,
			NetCashFlow:        netCashFlow,

			CashBalance:             runningCash,
			AccumulatedDepreciation: runningAccumDepr,
			FixedAssetsGross:        runningFixedAssetsGross,
			InventoryBalance:        runningInventory,
			LoanBalance:             loanBalance,
			RetainedEarnings:        runningRetainedEarnings,
		}
	}

	return results
}

// AnnualSummaries aggregates the income-statement lines of ProjectMonths
// into `years` yearly totals.
func (p *Plan) AnnualSummaries(years int) []AnnualFinancials {
	months := p.ProjectMonths(years * 12)
	summaries := make([]AnnualFinancials, years)

	for y := 0; y < years; y++ {
		s := AnnualFinancials{Year: y + 1}
		for m := y * 12; m < (y+1)*12 && m < len(months); m++ {
			mm := months[m]
			s.Revenue = s.Revenue.Add(mm.Revenue)
			s.COGS = s.COGS.Add(mm.COGS)
			s.GrossProfit = s.GrossProfit.Add(mm.GrossProfit)
			s.PayrollCost = s.PayrollCost.Add(mm.PayrollCost)
			s.PayrollTax = s.PayrollTax.Add(mm.PayrollTax)
			s.OpEx = s.OpEx.Add(mm.OpEx)
			s.Depreciation = s.Depreciation.Add(mm.Depreciation)
			s.InterestExpense = s.InterestExpense.Add(mm.InterestExpense)
			s.OperatingIncome = s.OperatingIncome.Add(mm.OperatingIncome)
			s.NetIncome = s.NetIncome.Add(mm.NetIncome)
		}
		summaries[y] = s
	}

	return summaries
}

// BalanceSheetSnapshots computes the plan's financial position as of the
// end of each of the given number of years.
func (p *Plan) BalanceSheetSnapshots(years int) []BalanceSheetSnapshot {
	months := p.ProjectMonths(years * 12)

	var openingEquity Money
	for _, f := range p.fundingSources {
		if !isAmortizingLoan(f) {
			openingEquity = openingEquity.Add(f.Amount)
		}
	}

	balances := p.startingBalances
	snapshots := make([]BalanceSheetSnapshot, years)

	for y := 0; y < years; y++ {
		idx := (y+1)*12 - 1
		if idx >= len(months) {
			idx = len(months) - 1
		}
		mm := months[idx]

		currentAssets := mm.CashBalance.Add(balances.AccountsReceivable).Add(mm.InventoryBalance).Add(balances.PrepaidExpenses)
		fixedAssetsNet := mm.FixedAssetsGross.Sub(mm.AccumulatedDepreciation)
		totalAssets := currentAssets.Add(fixedAssetsNet)

		currentLiabilities := balances.AccountsPayable.Add(balances.AccruedExpenses)
		totalLiabilities := currentLiabilities.Add(mm.LoanBalance)

		totalEquity := openingEquity.Add(mm.RetainedEarnings)

		snapshots[y] = BalanceSheetSnapshot{
			Year: y + 1,

			Cash:               mm.CashBalance,
			AccountsReceivable: balances.AccountsReceivable,
			Inventory:          mm.InventoryBalance,
			PrepaidExpenses:    balances.PrepaidExpenses,
			CurrentAssets:      currentAssets,

			FixedAssetsGross:        mm.FixedAssetsGross,
			AccumulatedDepreciation: mm.AccumulatedDepreciation,
			FixedAssetsNet:          fixedAssetsNet,

			TotalAssets: totalAssets,

			AccountsPayable:    balances.AccountsPayable,
			AccruedExpenses:    balances.AccruedExpenses,
			CurrentLiabilities: currentLiabilities,
			LoansPayable:       mm.LoanBalance,
			TotalLiabilities:   totalLiabilities,

			OwnersEquity:     openingEquity,
			RetainedEarnings: mm.RetainedEarnings,
			TotalEquity:      totalEquity,

			TotalLiabilitiesAndEquity: totalLiabilities.Add(totalEquity),
		}
	}

	return snapshots
}

// FinancialRatiosSeries computes the standard liquidity/safety/profitability
// ratios for each of the given number of years.
func (p *Plan) FinancialRatiosSeries(years int) []FinancialRatios {
	annuals := p.AnnualSummaries(years)
	balances := p.BalanceSheetSnapshots(years)
	months := p.ProjectMonths(years * 12)
	ratios := make([]FinancialRatios, years)

	for y := 0; y < years; y++ {
		a := annuals[y]
		b := balances[y]
		r := FinancialRatios{Year: y + 1}

		if !b.CurrentLiabilities.IsZero() {
			r.CurrentRatio = float64(b.CurrentAssets.MinorUnits()) / float64(b.CurrentLiabilities.MinorUnits())
			r.QuickRatio = float64(b.Cash.Add(b.AccountsReceivable).MinorUnits()) / float64(b.CurrentLiabilities.MinorUnits())
		}
		if !b.TotalEquity.IsZero() {
			r.DebtToEquity = float64(b.TotalLiabilities.MinorUnits()) / float64(b.TotalEquity.MinorUnits())
			r.ReturnOnEquityPercent = float64(a.NetIncome.MinorUnits()) / float64(b.TotalEquity.MinorUnits()) * 100
		}

		var annualPrincipal Money
		for m := y * 12; m < (y+1)*12 && m < len(months); m++ {
			annualPrincipal = annualPrincipal.Add(months[m].PrincipalPayment)
		}
		debtService := a.InterestExpense.Add(annualPrincipal)
		if !debtService.IsZero() {
			r.DSCR = float64(a.OperatingIncome.Add(a.Depreciation).MinorUnits()) / float64(debtService.MinorUnits())
		}

		if y > 0 && !annuals[y-1].Revenue.IsZero() {
			r.SalesGrowthPercent = float64(a.Revenue.Sub(annuals[y-1].Revenue).MinorUnits()) / float64(annuals[y-1].Revenue.MinorUnits()) * 100
		}
		if !a.Revenue.IsZero() {
			r.GrossProfitMarginPercent = float64(a.GrossProfit.MinorUnits()) / float64(a.Revenue.MinorUnits()) * 100
			r.NetProfitMarginPercent = float64(a.NetIncome.MinorUnits()) / float64(a.Revenue.MinorUnits()) * 100
		}

		ratios[y] = r
	}

	return ratios
}

// ProductFinancials is one product line's revenue and COGS for each year of
// the projection.
type ProductFinancials struct {
	Name          string
	RevenueByYear []Money
	COGSByYear    []Money
}

// ProductFinancialsSeries breaks the aggregated product revenue/COGS back
// out per product, for `years` years, so reports can show a line per
// product instead of just the combined total.
func (p *Plan) ProductFinancialsSeries(years int) []ProductFinancials {
	horizon := years * 12
	multipliers := p.unitGrowthMultipliers(horizon)

	result := make([]ProductFinancials, len(p.products))
	for i, prod := range p.products {
		result[i] = ProductFinancials{
			Name:          prod.Name,
			RevenueByYear: make([]Money, years),
			COGSByYear:    make([]Money, years),
		}
		for y := 0; y < years; y++ {
			for m := y * 12; m < (y+1)*12 && m < horizon; m++ {
				units := float64(prod.Month1Units) * multipliers[m]
				result[i].RevenueByYear[y] = result[i].RevenueByYear[y].Add(fromFloatUSD(units * float64(prod.PricePerUnit.MinorUnits())))
				result[i].COGSByYear[y] = result[i].COGSByYear[y].Add(fromFloatUSD(units * float64(prod.CostPerUnit.MinorUnits())))
			}
		}
	}

	return result
}

// OpExLine is one operating-expense line item's cost for each year of the
// projection.
type OpExLine struct {
	Name         string
	AmountByYear []Money
}

// OpExAnnualBreakdown breaks the aggregated OpEx total back out per
// category, for `years` years.
func (p *Plan) OpExAnnualBreakdown(years int) []OpExLine {
	horizon := years * 12

	lines := make([]OpExLine, len(p.opEx))
	for i, cost := range p.opEx {
		lines[i] = OpExLine{Name: cost.Name, AmountByYear: make([]Money, years)}
		for y := 0; y < years; y++ {
			for m := y * 12; m < (y+1)*12 && m < horizon; m++ {
				lines[i].AmountByYear[y] = lines[i].AmountByYear[y].Add(cost.ProjectedAmount(MonthIndex(m)))
			}
		}
	}

	return lines
}

// AssetDepreciation is a single capital asset's depreciation schedule
// summary, used for the Amortization & Depreciation report.
type AssetDepreciation struct {
	Name              string
	UsefulLifeYears   int
	Year1Depreciation Money
}

// AssetDepreciationBreakdown returns Year 1 depreciation for each capital
// asset in the plan.
func (p *Plan) AssetDepreciationBreakdown() []AssetDepreciation {
	result := make([]AssetDepreciation, len(p.futurePurchases))
	for i, asset := range p.futurePurchases {
		var year1 Money
		for m := 0; m < 12; m++ {
			year1 = year1.Add(asset.DepreciationForMonth(MonthIndex(m)))
		}
		result[i] = AssetDepreciation{
			Name:              asset.Name,
			UsefulLifeYears:   asset.UsefulLifeYears(),
			Year1Depreciation: year1,
		}
	}
	return result
}

// LoanYearSummary is the aggregated (across all amortizing loans)
// interest/principal paid and ending balance for a single year, used for
// the Amortization report.
type LoanYearSummary struct {
	Year          int
	Interest      Money
	Principal     Money
	EndingBalance Money
}

// LoanAmortizationSummary aggregates interest, principal, and ending
// balance across every amortizing loan, for each of `years` years.
func (p *Plan) LoanAmortizationSummary(years int) []LoanYearSummary {
	horizon := years * 12
	months := p.ProjectMonths(horizon)

	result := make([]LoanYearSummary, years)
	for y := 0; y < years; y++ {
		var interest, principal Money
		for m := y * 12; m < (y+1)*12 && m < len(months); m++ {
			interest = interest.Add(months[m].InterestExpense)
			principal = principal.Add(months[m].PrincipalPayment)
		}

		idx := (y+1)*12 - 1
		if idx >= len(months) {
			idx = len(months) - 1
		}

		result[y] = LoanYearSummary{
			Year:          y + 1,
			Interest:      interest,
			Principal:     principal,
			EndingBalance: months[idx].LoanBalance,
		}
	}

	return result
}

// FundingBreakdown splits total funding raised into its equity and
// amortizing-loan (debt) portions, per the same isAmortizingLoan rule used
// throughout the projection engine.
func (p *Plan) FundingBreakdown() (equity, debt Money) {
	for _, f := range p.fundingSources {
		if isAmortizingLoan(f) {
			debt = debt.Add(f.Amount)
		} else {
			equity = equity.Add(f.Amount)
		}
	}
	return equity, debt
}

// AverageLoanInterestRate returns the unweighted average annual interest
// rate across every amortizing loan (0 if there are none).
func (p *Plan) AverageLoanInterestRate() float64 {
	var sum float64
	var count int
	for _, f := range p.fundingSources {
		if isAmortizingLoan(f) {
			sum += f.InterestRate
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// Breakeven computes the Year 1 breakeven sales analysis: the annual (and
// monthly) revenue level at which Gross Margin exactly covers fixed
// (payroll + operating) expenses.
func (p *Plan) Breakeven() BreakevenAnalysis {
	annuals := p.AnnualSummaries(1)
	y1 := annuals[0]

	b := BreakevenAnalysis{
		GrossMargin:               y1.GrossProfit,
		TotalSales:                y1.Revenue,
		PayrollFixedCost:          y1.PayrollCost.Add(y1.PayrollTax),
		OperatingExpenseFixedCost: y1.OpEx,
	}
	b.TotalFixedExpenses = b.PayrollFixedCost.Add(b.OperatingExpenseFixedCost)

	if !y1.Revenue.IsZero() {
		b.GrossMarginPercent = float64(y1.GrossProfit.MinorUnits()) / float64(y1.Revenue.MinorUnits()) * 100
	}
	if b.GrossMarginPercent > 0 {
		b.BreakevenAnnual = fromFloatUSD(float64(b.TotalFixedExpenses.MinorUnits()) / (b.GrossMarginPercent / 100))
		b.BreakevenMonthly = b.BreakevenAnnual.Div(12)
	}

	return b
}
