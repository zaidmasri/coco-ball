# Business Planning Tool (NorthBasis)

## Project Overview

This project is converting the **SCORE Financial Projections Template** (an Excel-based business planning tool) into a modern web application. The tool helps entrepreneurs and small business owners create detailed financial projections, including:

- **Starting Point** - Initial balance sheet setup with fixed assets, startup costs, funding sources
- **Payroll Planning** - Employee/contractor salary projections with payroll taxes and benefits
- **Sales Forecasting** - Revenue stream projections with growth scenarios
- **Operating Expenses** - OpEx tracking with growth strategies
- **Cost of Goods Sold (COGS)** - Product cost projections
- **Cash Flow Analysis** - Monthly cash flow projections
- **Income Statement** - P&L projections
- **Balance Sheet** - Asset, liability, and equity projections
- **Financial Ratios & Breakeven Analysis** - Key financial metrics and analysis tools

## Tech Stack

### Backend
- **Language**: Go 1.26.1
- **Server**: Standard library `net/http` with custom router (Go 1.22+ features)
- **Data Storage**: In-memory store (development) - needs database implementation for production
- **Domain Model**: Value-driven domain model with proper validation and error handling

### Frontend
- **Templating**: Go `html/template` (server-side rendering)
- **Static Assets**: Embedded filesystem (Go 1.16+)
- **CSS/JS**: Static files served via embedded FS

### Core Dependencies
- `github.com/google/uuid` - For plan ID generation

## Completed Features ✅

### Core Infrastructure
- [x] Project structure (cmd/cli, internal/{domain,handlers,middleware,store,templates,static})
- [x] HTTP server with proper timeouts and middleware
- [x] Static file serving with embedded filesystem
- [x] Error page handling with 404 catch-all
- [x] Template caching and rendering pipeline
- [x] Logger middleware (moved to internal/middleware)
- [x] Route registration encapsulation (internal/handlers/routes.go)

### Domain Model
- [x] Plan aggregate root with immutable ID
- [x] Revenue streams with growth strategies (Flat, AnnualStepPercent)
- [x] Operating expenses (OpEx) with growth tracking
- [x] Cost of Goods Sold (COGS) with growth tracking
- [x] Capital asset depreciation (Straight-Line, Double-Declining)
- [x] Financing terms and funding sources
- [x] Startup costs tracking
- [x] Starting balance sheet state
- [x] Comprehensive validation and error types

### Data Storage
- [x] In-memory plan store with GetAll(), GetByID(), Save() methods
- [x] Plan store interface for future database implementation
- [x] SQLite database implementation with persistent storage

### Authentication & Authorization
- [x] User authentication system (login/logout/signup)
- [x] Password hashing and verification
- [x] Session management and middleware
- [x] User registration/signup flow
- [x] Session-based auth with cookies
- [x] User profile page (GET `/profile`)
- [x] Logout as POST (proper state-modifying HTTP semantics)
- [x] User Authorization - Ensure users can only access their own plans (access control middleware enforces Owner/Editor/Viewer roles)

### Routes & Pages
- [x] Home page (root `/`) - List of existing plans
- [x] Login page (GET/POST `/login`) - User authentication
- [x] Signup page (GET/POST `/signup`) - User registration
- [x] Profile page (GET `/profile`) - User profile display
- [x] Logout endpoint (POST `/logout`) - Session termination
- [x] Plan setup form (POST `/plan/setup`) - Create new plan
- [x] Setup page (GET/POST `/plan/{id}/setup`) - Edit plan core details (name, starting month/year)
- [x] Starting Point page (GET/POST `/plan/{id}/starting-point`) - Startup costs, funding sources, initial balances, capital assets
- [x] Payroll page (GET `/plan/{id}/payroll`) - Employee/contractor salary planning
- [x] Sales Forecast page (GET `/plan/{id}/sales-forecast`) - Revenue projections
- [x] Operating Expenses page (GET `/plan/{id}/operating-expenses`) - OpEx tracking
- [x] Cash Flow page (GET `/plan/{id}/cash-flow`) - Cash flow projections
- [x] Income Statement page (GET `/plan/{id}/income-statement`) - P&L projections
- [x] Balance Sheet page (GET `/plan/{id}/balance-sheet`) - Balance sheet projections
- [x] Analytics/Home page (GET `/plan/{id}/analytics`) - Key metrics and dashboards

### Form Handlers (POST Endpoints)
- [x] POST `/plan/{id}/payroll` - Save payroll data (salary roles, benefits, employer tax rates)
- [x] POST `/plan/{id}/sales-forecast` - Save sales data (products + global unit-growth curve)
- [x] POST `/plan/{id}/operating-expenses` - Save OpEx data
- [x] POST `/plan/{id}/cash-flow` - Save discretionary cash flow (additional inventory, distributions)
- [x] POST `/plan/{id}/delete` - Delete a plan (requires Owner access)

### Financial Calculations & Reporting
- [x] Payroll tax calculations (employer SS/Medicare/FUTA/SUTA on W-2 wages, contractors exempt)
- [x] Product-based COGS calculations (per-unit cost × projected units sold)
- [x] Depreciation calculations (pre-existing: Straight-Line, Double-Declining)
- [x] Cash flow projections (monthly, with running cash balance)
- [x] Income statement generation (Income Statement page now shows real per-product/per-category numbers for 3 years)
- [x] Balance sheet generation (Balance Sheet page now shows real numbers; balances within rounding — see Tech Decisions Log)
- [x] Financial ratios calculations (current, quick, debt-to-equity, DSCR, sales growth, gross/net margin, ROE)
- [x] Breakeven analysis (Year 1 gross-margin-based breakeven, annual + monthly)

### UI Components
- [x] Base layout template with navigation sidebar
- [x] Form components and styling
- [x] Plan listing view
- [x] Multi-step form flows for data entry
- [x] Error pages

## Remaining Work 🚧

### Authentication & Authorization ⭐ CRITICAL
- [x] User authentication system (login/logout)
- [x] Password hashing and verification
- [x] Session management and middleware
- [x] User registration/signup flow
- [x] Session-based auth tokens
- [ ] Password reset functionality
- [ ] Email verification
- [ ] Two-factor authentication (optional)

### User Management
- [ ] User model and database table
- [ ] User profile management
- [ ] User preferences/settings
- [ ] Workspace/organization support (optional multi-user teams)
- [ ] User role-based access control (Admin, Editor, Viewer)
- [ ] Audit logging for plan changes

### Database Implementation
- [ ] PostgreSQL/MySQL schema design
- [ ] Database driver integration (github.com/lib/pq or similar)
- [ ] Migration system
- [ ] Connection pooling and health checks
- [ ] Persistent plan storage (replace in-memory store)
- [ ] User data persistence
- [ ] Plan versioning/history (optional)

### Data Persistence
- [ ] Implement database store interface
- [ ] Save revenue streams to database
- [ ] Save operating expenses to database
- [ ] Save COGS to database
- [ ] Save capital assets/depreciation to database
- [ ] Save funding sources and startup costs
- [ ] Plan update/edit persistence

### Financial Calculations & Reporting (remaining)
- [ ] Income tax modeling (currently deliberately unmodeled — see Tech Decisions Log)
- [ ] Export functionality (PDF, Excel) — the report pages have a browser "Print / Save PDF" button but no server-side export
- [ ] "Diagnostic Tool" tab's editable Industry Norms column is still client-side-only (not saved to the plan)
- [ ] COGS Calculator tab on Analytics remains an intentional client-side scratchpad (doesn't save to plan — by design)

### Remaining Generated-Report Endpoints
- [ ] POST `/plan/{id}/income-statement` - Generate/save income statement (blocked on Financial Calculations below)
- [ ] POST `/plan/{id}/balance-sheet` - Generate/save balance sheet (blocked on Financial Calculations below)

Note: there is no dedicated `/cogs` route — per-product COGS is captured per-unit
directly on the Sales Forecast form (`prod_cogs[]`) and saved via
`POST /plan/{id}/sales-forecast`.

### Input Validation & Error Handling
- [ ] Form validation on server-side (currently minimal)
- [ ] CSRF protection
- [ ] Input sanitization
- [ ] Rate limiting
- [ ] Comprehensive error messages to users
- [ ] Validation for negative amounts, invalid dates, etc.

### API & Integration
- [ ] JSON API endpoints (optional, for future mobile/frontend frameworks)
- [ ] Data export to Excel format
- [ ] Data import from Excel/CSV
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Cross-origin resource sharing (CORS) if needed

### Frontend Polish
- [ ] Responsive design improvements
- [ ] Form validation feedback (client + server)
- [ ] Loading states and progress indicators
- [ ] Success/error notifications
- [ ] Data entry help text and tooltips
- [ ] Print-friendly views
- [ ] Dark mode support

### Testing
- [ ] Unit tests for domain model (currently some exist)
- [ ] Integration tests for handlers
- [ ] Database tests
- [ ] End-to-end tests
- [ ] Test fixtures and factories

### DevOps & Deployment
- [ ] Docker containerization
- [ ] Environment configuration (.env, config files)
- [ ] Database migrations automation
- [ ] CI/CD pipeline
- [ ] Logging and monitoring
- [ ] Error tracking (Sentry, etc.)
- [ ] Performance monitoring

### Documentation
- [ ] Architecture decision records (ADRs)
- [ ] API documentation
- [ ] Deployment guide
- [ ] Developer setup guide
- [ ] User guide/help documentation

### Security
- [ ] SQL injection prevention (use prepared statements)
- [ ] XSS protection in templates
- [ ] HTTPS enforcement
- [ ] Security headers (CSP, X-Frame-Options, etc.)
- [ ] Rate limiting and DDoS protection
- [ ] Input validation and sanitization
- [ ] Secure file uploads (if applicable)

## High-Priority Items (Do First)

1. ~~**User Authorization**~~ ✅ **COMPLETED** - Users can only access plans they own or have access to via Owner/Editor/Viewer roles
2. ~~**Form POST handlers**~~ ✅ **COMPLETED** - Payroll, Sales Forecast, Operating Expenses, Cash Flow, and Delete Plan all persist via POST
3. ~~**Financial calculations**~~ ✅ **COMPLETED** - Income Statement, Balance Sheet, and Analytics (breakeven/ratios/amortization) now show real computed 3-year projections
4. **Testing** - Add handler-level tests for the new report pages (GetIncomeStatement/GetBalanceSheet/GetAnalytics); domain-level projection engine already has test coverage in projection_test.go
5. **Form Validation** - Client and server-side validation for all inputs
6. **Error handling** - Comprehensive error messages and recovery flows
7. **Income tax modeling** - No tax-rate input exists anywhere in the app; Net Income is currently pre-tax

## Project Structure Reference

```
.
├── cmd/cli/
│   ├── main.go                  # CLI entry point
│   ├── serve.go                 # Server setup and initialization
│   ├── migrate.go               # Database migrations
│   └── ...                      # Other CLI commands
├── internal/
│   ├── domain/
│   │   ├── plan.go              # Plan aggregate root, business logic
│   │   ├── user.go              # User domain model
│   │   ├── session.go           # Session domain model
│   │   └── plan_test.go         # Domain tests
│   ├── handlers/
│   │   ├── plan.go              # Plan HTTP handlers
│   │   ├── auth.go              # Authentication handlers (login, signup, logout, profile)
│   │   └── routes.go            # Route registration (RegisterRoutes method)
│   ├── middleware/
│   │   └── logger.go            # Request logging middleware
│   ├── store/
│   │   ├── store.go             # Store interface
│   │   ├── sqlite.go            # SQLite implementation
│   │   ├── migrations.go        # Database migrations
│   │   └── sqlite_test.go       # Store tests
│   ├── templates/               # Go HTML templates
│   │   ├── base.html            # Base layout template
│   │   ├── pages/               # Page templates
│   │   ├── components/          # Reusable components
│   │   └── embed.go             # Embedded FS
│   ├── views/
│   │   ├── builders.go          # Page view builders
│   │   ├── types.go             # View data types
│   │   └── embed.go             # Embedded template resources
│   └── static/                  # CSS, JS, images
│       └── embed.go             # Embedded FS
├── go.mod
└── go.sum
```

## Key Files to Understand

- **plan.go** (domain) - Core business logic, validation, calculations
- **main.go** - Route definitions and server setup
- **memory.go** - Current data storage implementation (needs replacement)
- **plan.go** (handlers) - HTTP request/response handling
- Templates in `internal/templates/` - UI layer

## Important Notes for Future Agents

### When Updating This Document

⚠️ **CRITICAL**: Keep this AGENTS.md file up-to-date as you work. After each significant feature or milestone:

1. Move completed items from "Remaining Work" to "Completed Features"
2. Add new discovered remaining work to the appropriate section
3. Update the High-Priority Items if priorities shift
4. Record any tech decisions or pivots in a brief note below

### Tech Decisions Log

- **In-Memory Storage (Initial)**: Started with in-memory store for rapid prototyping. Switched to SQLite for persistent storage.
- **Go SSR (Server-Side Rendering)**: Using Go templates instead of a frontend framework for simplicity and rapid MVP development. Can be refactored to API + frontend framework later.
- **No Framework Overhead**: Intentionally using standard library to keep dependencies minimal during exploration phase.
- **Route Registration Encapsulation (2026-07-29)**: Moved route registration from CLI into handlers/routes.go via RegisterRoutes() method for better separation of concerns.
- **Middleware Package Separation (2026-07-29)**: Extracted Logger middleware to internal/middleware/logger.go to organize cross-cutting concerns separately.
- **Auth Handler Organization (2026-07-29)**: Consolidated authentication logic (login, signup, logout, profile) in handlers/auth.go rather than splitting across files.
- **Session-Based Auth (2026-07-29)**: Implemented cookie-based session management instead of JWT for simpler server-side state management in early stages.
- **POST for Logout (2026-07-29)**: Changed logout from GET to POST to follow HTTP semantics (state-modifying operations should use POST).
- **Authorization Middleware (2026-07-29)**: Applied RequireAccess middleware to plan-specific routes in RegisterRoutes() to enforce Owner/Editor/Viewer access control at the HTTP route level rather than within handlers.
- **Domain Model Extended for Form Handlers (2026-07-29)**: Added SalaryRole/Benefit/PayrollTaxRates (payroll), Product/SalesGrowthCurve (sales forecast — a single global unit-growth curve with Y1 quarterly rates + one rate per future year, applied across all product lines), and InventoryPurchase/Distribution (cash flow) to the Plan aggregate, each with Clear/Add methods following the existing StartingPoint pattern (clear-then-rebuild on every save so edits don't duplicate rows).
- **RequireAccess Owner Level Fixed (2026-07-29)**: `RequireAccess(domain.Owner)` previously had no matching branch and let any access level through; added an explicit Owner check so the new Delete Plan endpoint is actually owner-gated.
- **PlanStore.Delete Added (2026-07-29)**: Added `Delete(id)` to the PlanStore interface and SQLiteStore, transactionally removing the plan row and its plan_access rows.
- **Financial Projection Engine (2026-07-29)**: Added `internal/domain/projection.go`, a month-by-month projection engine (`ProjectMonths`) that produces Income Statement, Balance Sheet, ratio, and breakeven data. Key modeling decisions, all documented in the file's top comment and enforced by `TestBalanceSheetSnapshots_Balance`:
  - **No income tax modeled.** There's no tax-rate input anywhere in the app; NetIncome is pre-tax. Showing a fabricated rate would be worse than clearly showing $0.
  - **Debt vs. equity funding sources**: a FundingSource is an amortizing loan only if it has both `InterestRate > 0` and `TermMonths > 0`; otherwise its full Amount is treated as a one-time equity contribution with no debt service.
  - **AR/Prepaid/AP/AccruedExpenses held static** at their Starting Point values for the whole projection (no revenue-driven AR/AP growth).
  - **"Additional Inventory" cash-flow purchases** are an asset-for-asset swap (cash → inventory) with no depletion modeled — they never hit the Income Statement.
  - **Starting Point balances (Cash/AR/Prepaid/AP/AccruedExpenses) represent an existing business's pre-existing balance sheet.** Their net value is folded into the opening Retained Earnings baseline as implied pre-existing equity — this was a real bug caught by the balance-sheet identity test (initially off by exactly that net amount) before the fix.
  - **Sales growth curve is global** across all product lines: Year 1 uses quarterly monthly-compounding rates, subsequent years use one monthly-compounding rate per year (per the form's own "Monthly Growth" label).
  - The balance sheet is guaranteed to balance (Assets = Liabilities + Equity) by construction, up to a few dollars of whole-dollar (`Money` is `int64`) rounding drift across a 36-month projection.
- **Template Currency/Math Helpers Added (2026-07-29)**: Go's `html/template` has no built-in arithmetic or currency formatting. Added `internal/views/funcs.go` (`formatMoney`, `addMoney`, `formatPercent`, `formatRatio`) registered via `template.Funcs()` in `cmd/cli/serve.go`'s `loadTemplates()`.
- **Balance Sheet Categories Simplified (2026-07-29)**: The original balance-sheet.html mockup had separate "Line of Credit," "Commercial Loans & Mortgages," and "Other Bank & Credit Card Debt" rows, and separate "Real Estate/Vehicles/Equipment" vs "Other Fixed Assets" rows — but the domain model doesn't categorize funding sources or capital assets that finely (just one pool of amortizing loans, one pool of capital assets). Rewrote those sections to a single "Loans Payable" and "Fixed Assets (at cost)" line each rather than fabricating a category split the data doesn't support.

### Known Limitations

- **No Income Tax Modeling** - Net Income throughout the Income Statement, Balance Sheet, and Analytics pages is pre-tax; the "Income Tax Expense" line always shows $0
- **No Export Functionality** - Can't export plans as Excel or PDF yet (browser print-to-PDF only)
- **Money Truncates Cents** - `domain.Money` is `int64`; form amounts are parsed as float64 then truncated, so fractional cents are dropped (pre-existing pattern from Starting Point, kept for consistency). Over a 36-month projection this can drift the balance sheet by a few dollars — `BuildBalanceSheetPage`/`BuildAnalyticsPage` tolerate up to $25 before flagging "does not balance"
- **Current/Quick Ratio show 0.00 when undefined** - if a plan has no Accounts Payable/Accrued Expenses, current liabilities are $0 and these ratios are mathematically undefined; the UI currently can't distinguish that from an actually-computed 0.0
- **Loan/Fixed-Asset categories are not sub-typed** - the Balance Sheet and Amortization tab show one combined "Loans Payable" and "Fixed Assets" figure rather than breaking out Line-of-Credit vs. Commercial Loan vs. Equipment vs. Real Estate, since the domain model doesn't track those distinctions

### Questions to Ask If Stuck

1. **"Should I add a new route?"** → Check if it's in the route list in main.go. If not, add both GET and POST if it's a form.
2. **"How do I save data?"** → Use `app.Store.Save(plan)` after mutating the plan. Currently in-memory; will need DB migration.
3. **"Where do I put validation?"** → In the domain model (plan.go domain package) for business rules. Use handlers for HTTP-level validation.
4. **"Should I create a new handler?"** → Create a method on the App struct that returns an http.HandlerFunc.

---

**Last Updated**: 2026-07-29  
**Session Focus**: Financial projection engine (Income Statement, Balance Sheet, Analytics — breakeven/ratios/amortization) built on top of the Payroll/Sales/OpEx/CashFlow data captured in the prior session  
**Total Remaining Items**: ~35 (across all categories)  
**Critical Path**: ~~User Authorization~~ ✅ → ~~Form POST Handlers~~ ✅ → ~~Financial Calculations~~ ✅ → Testing → Form Validation

## Session Summary (2026-07-29)

### Authorization Implementation ✅
1. **Access Control Middleware** - Applied RequireAccess middleware to all plan-specific routes
   - GET endpoints require Viewer access
   - POST endpoints require Editor access
   - Owner, Editor, and Viewer roles properly enforced
2. **Route Middleware Integration** - Updated RegisterRoutes() to wrap plan handlers with authorization checks
3. **Comprehensive Testing** - Added auth_test.go with tests covering:
   - User can access their own plans ✓
   - Users cannot access other users' plans (returns 403) ✓
   - Unauthenticated users cannot access plans (returns 401) ✓
   - Users with Viewer access cannot edit (returns 403) ✓
   - Users with Editor access can edit ✓

### Previous Refactoring (from earlier session)
1. **Route Registration** - Moved route definitions from cmd/cli/serve.go to internal/handlers/routes.go via RegisterRoutes() method
2. **Middleware Organization** - Created internal/middleware package and moved Logger to internal/middleware/logger.go
3. **Handler Organization** - Consolidated auth handlers (login, signup, logout, profile) in internal/handlers/auth.go
4. **HTTP Semantics** - Changed logout endpoint from GET to POST (state-modifying operations require POST)

### Bug Fixes
1. **Bootstrap Sourcemap Errors** - Removed sourcemap references from bootstrap.min.js, bootstrap.bundle.min.js, and bootstrap.min.css (404 errors resolved)
2. **Dropdown Functionality** - Fixed "createPopper is not a function" error by removing duplicate bootstrap.min.js script tags (kept only bootstrap.bundle.min.js which includes Popper.js)
3. **Template Error** - Fixed cash-flow.html template error by removing reference to non-existent .Plan.AdditionalInventory field

### Code Quality
- Better separation of concerns (routes, middleware, auth in dedicated modules)
- Route-level authorization enforcement (defense in depth)
- Comprehensive test coverage for authorization flows
- Clean CLI entry point (cmd/cli/serve.go focused and minimal)
- All tests passing (handlers and store)

## Session Summary (2026-07-29, later): Form POST Handlers

### What Shipped
1. **Domain model extensions** (`internal/domain/plan.go`) - Added `SalaryRole`,
   `Benefit`, `PayrollTaxRates` (payroll); `Product`, `SalesGrowthCurve`
   (sales forecast); `InventoryPurchase`, `Distribution` (cash flow) — each
   with `Add*`/`Clear*` methods and full JSON marshal/unmarshal support so
   they persist through the existing SQLite JSON-blob store unchanged.
2. **New POST handlers** (`internal/handlers/plan.go`) - `PostPayroll`,
   `PostSalesForecast`, `PostOpExpenses`, `PostCashFlow`, `PostDeletePlan`,
   all following the existing `PostStartingPoint` clear-then-rebuild pattern
   (wipe existing rows, re-parse the submitted `name[]`/`amount[]` arrays,
   save, redirect back to the same URL so the browser does a GET after POST).
3. **Store & auth fixes** - Added `PlanStore.Delete(id)` (transactional:
   removes `plan_access` rows then the `plans` row) and fixed a real gap in
   `RequireAccess`: it previously had no branch for `domain.Owner`, so
   `RequireAccess(domain.Owner)` silently let any access level through. Now
   the new delete endpoint is properly owner-gated.
4. **Templates** - Wired real `form action` URLs (payroll.html was still
   posting to a placeholder `/test-submit`), added server-rendered prefill
   rows for salaries/benefits/products/opex/inventory so editing a plan
   round-trips saved data instead of always starting from an empty form.
5. **Verified end-to-end in a real browser** (signup → create plan → submit
   each of the four forms → confirm saved values reappear on reload → delete
   plan) against a scratch SQLite DB, not just `go build`/`go test`.

### Deliberately Deferred
- The Sales Forecast domain model captures exactly what the form submits
  (quarterly Y1 growth + per-year future rates, applied globally across
  product lines) but nothing yet *uses* that curve to project revenue —
  `RevenueStream.ProjectedAmount` is unchanged. Wiring that up is part of
  the still-open **Financial Calculations** work.
- No dedicated `/cogs` route was added — the sales-forecast form already
  captures per-unit COGS per product (`prod_cogs[]`); a separate COGS page
  was never built alongside it, so `AGENTS.md`'s original checklist item
  was stale.

## Session Summary (2026-07-29, later still): Financial Calculations

### What Shipped
1. **Projection engine** (`internal/domain/projection.go`) - A month-by-month
   engine (`ProjectMonths`) that turns Products+SalesGrowthCurve, Payroll,
   OpEx, CapitalAssets, FundingSources, and discretionary CashFlow items into
   full financial statements: `AnnualSummaries` (income statement),
   `BalanceSheetSnapshots`, `FinancialRatiosSeries`, `Breakeven`, plus
   `ProductFinancialsSeries`/`OpExAnnualBreakdown`/`AssetDepreciationBreakdown`/
   `LoanAmortizationSummary` for the per-line-item report tables.
2. **Report pages wired to real data** - `IncomeStatementPage`,
   `BalanceSheetPage`, and `AnalyticsPage` (in `internal/views/types.go` /
   `builders.go`) now carry computed projection data instead of nothing;
   `income-statement.html`, `balance-sheet.html`, and `analytics.html`
   (breakeven, ratios, diagnostic, and amortization tabs — COGS Calculator
   tab intentionally left alone, it's a client-side scratchpad) render real
   per-year and per-line-item numbers instead of static "$0.00" placeholders.
3. **Template math/currency helpers** (`internal/views/funcs.go`) -
   `formatMoney`, `addMoney`, `formatPercent`, `formatRatio`, since
   `html/template` has none of this built in.
4. **A real bug caught by testing, not eyeballing**: the balance sheet
   didn't balance by ~$3,400–$5,400 in early test runs. Root cause: Starting
   Point's Cash/AR/Prepaid/AP/AccruedExpenses balances (meant for an
   already-operating business) reduced/increased assets and liabilities with
   no offsetting equity entry. Fixed by folding their net value into the
   opening Retained Earnings baseline as implied pre-existing equity —
   verified by `TestBalanceSheetSnapshots_Balance`, which asserts Assets =
   Liabilities + Equity within a small rounding tolerance.
5. **Verified end-to-end in a real browser**: signed up, created a plan,
   entered a capital asset + two funding sources (one loan, one equity) +
   payroll + a product + an OpEx line, then confirmed the Income Statement,
   Balance Sheet, and all four Analytics tabs render sane, cross-consistent
   numbers (e.g. the loan's annual interest+principal summed to a constant
   payment across all 3 years, confirming the amortization math).

### Deliberately Deferred / Documented Simplifications
See the Tech Decisions Log entry "Financial Projection Engine" above for the
full list (no income tax, debt-vs-equity funding source rule, static
AR/Prepaid/AP, inventory-as-asset-swap). These are simplifying assumptions
appropriate for an MVP financial planning tool, not oversights — each is
documented in `projection.go`'s file-level comment so future work can
tighten them deliberately rather than rediscover them by surprise.
