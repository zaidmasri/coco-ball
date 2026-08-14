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
- [x] In-memory plan store with GetAll(), GetByID(), Save() methods (superseded by SQLite; interface still supports swapping implementations)
- [x] Plan store interface for future database implementation
- [x] SQLite database implementation with persistent storage
- [x] All form data persists to SQLite: Payroll, Sales Forecast/Products, Operating Expenses, Cash Flow (inventory + distributions), Starting Point (startup costs, funding sources, capital assets, initial balances)
- [x] Migration system (`internal/store/migrations.go` — ordered list of SQL statements run automatically on store init; 49 migrations as of 2026-08-07, covering plans, users, sessions, plan_access, first/last name, plan_invites, `deleted_at` columns on all business entity tables, `outbox_events` table, and `idempotency_keys` table)

### Authentication & Authorization
- [x] User authentication system (login/logout/signup)
- [x] Password hashing and verification
- [x] Session management and middleware
- [x] User registration/signup flow (now requires first and last name in addition to email/password)
- [x] Session-based auth with cookies
- [x] User profile page (GET `/profile`)
- [x] Logout as POST (proper state-modifying HTTP semantics)
- [x] User Authorization - Ensure users can only access their own plans (access control middleware enforces Owner/Editor/Viewer roles)

### DDD / CQRS Architecture (2026-08-07)
- [x] `internal/domain/entities/` package — all domain types live here: Money value object, AggregateRoot marker interface, ErrValidation sentinel, Plan and User aggregate roots, and all companion files (payroll, cash flow, sales forecast, starting point, events, projection, session, invite)
- [x] `Money` is a proper value object (`entities.Money`, type `int64` backing so arithmetic operators work; JSON-compatible with existing blob data)
- [x] `AggregateRoot` marker interface in `entities/aggregate_root.go`; compile-time assertions `var _ AggregateRoot = (*Plan)(nil)` and `(*User)(nil)` enforce it
- [x] Plan and User documented as the two aggregate roots; aggregate boundary (wizard items belong to Plan) documented with `AggregateID()` godoc and in `repositories/plan.go` package comment
- [x] Sub-entity vs aggregate-root repository distinction documented in `repositories/plan.go` package comment and `repositories/wizard_item.go`
- [x] Cross-aggregate references confirmed ID-only throughout: Plan.ownerID, PlanInvite.PlanID/InvitedBy, Session.UserID, PlanAccess.PlanID/UserID — all `uuid.UUID`
- [x] Repository interfaces defined in `internal/domain/repositories/` (one file per aggregate boundary)
- [x] Application service interfaces in `internal/application/interfaces/` (one per domain hub)
- [x] Application service implementations in `internal/application/services/` (delegate to repository interfaces; domain policy lives here, not in the store)
- [x] `Find*` vs `Get*` naming enforced across all repository interfaces, implementations, and callers
- [x] Soft deletion on all business entity tables (`deleted_at INTEGER`); hard delete only for join tables (plan_access) and singleton tables
- [x] Business logic moved out of `internal/store/`: session expiry check and access-level validation now live in `auth_service.go` and `access_service.go`
- [x] UUIDv7 (`uuid.NewV7()`) in all domain constructors — time-ordered, never generated by the DB
- [x] Validated-Entity pattern: `ValidatedPlan` opaque token; `Plan.Validate()` produces one; `PlanRepository.Save` requires it. Extended (2026-08-12) to ownership: `VerifiedUser` opaque token (`user.go`) proves a `uuid.UUID` names a real User; `Plan.NewPlan` requires one instead of a bare `uuid.UUID`, and `PlanService.CreatePlan` is the only code that produces one (via `UserRepository.GetUser`) — see "Application Layer: go-ddd Comparison"
- [x] Domain events: `internal/domain/entities/events/` package following go-ddd pattern — `base_event.go` (`BaseEvent` + `DomainEvent` interface), `plan_events.go` (`PlanCreated`/`PlanUpdated`/`PlanDeleted`), `invite_events.go` (`UserInvitedToPlan`), `user_events.go` (`UserRegistered`); **only aggregate roots emit events** — `Plan` and `User` carry event buffers; `PlanInvite` (a non-root entity) does not; `Plan.RecordUserInvited()` emits `UserInvitedToPlan` on behalf of the invite
- [x] Transactional outbox: `outbox_events` table; `internal/infrastructure/sqlite/outbox.go`'s `insertOutboxEvents` writes each repository's entity row + outbox events atomically in one transaction (see `plan_repository.go`, `user_repository.go`, `invite_repository.go`)
- [x] `processed_at` column on `outbox_events` — **the outbox relay is implemented** (`internal/infrastructure/sqlite/outbox_relay.go`, log-only publish stub) and sets this on delivery; `idx_outbox_unprocessed` partial index filters `processed_at IS NULL` for efficient relay queries
- [x] Idempotency table: `idempotency_keys` table (migration 49) — schema in place; middleware wrapper not yet wired
- [x] CQRS command types wired for Plan's lifecycle: `internal/application/commands/plan.go` (CreatePlan, UpdatePlan, DeletePlan + Result wrappers), `internal/application/common/plan_result.go` (PlanResult), `internal/application/mapper/plan_result.go` — actually called from `PlanController`/`PlanService`, not dead structs (see "Application Layer: go-ddd Comparison" above for the full before/after and why `queries/plan.go` was deleted rather than wired the same way)
- [x] Application-layer tests: `internal/application/services/plan_service_test.go` (save/get, CreatePlan/UpdatePlan/DeletePlan commands, event emission, delete, not-found) — exercises a real temp-file `sqlite.PlanRepository`, not a fake

### Teams & Collaboration (2026-08-01)
- [x] Plan invites - Owner can invite a collaborator by email at a chosen access level (Editor/Viewer) via `POST /plan/{id}/invites`
- [x] Invite inbox - Pending invites addressed to a user's email surface in an onboarding section on the home dashboard (`GetRoot`/`pendingInvitesForUser`), not via email delivery — no outbound email is sent anywhere in the app
- [x] Accept/reject flow - `POST /invites/{id}/accept` grants access (`Store.GrantAccess`) and marks the invite accepted; `POST /invites/{id}/reject` marks it rejected; both verify the invite is addressed to the logged-in user's email and still pending
- [x] User display names - `User` now carries `FirstName`/`LastName`/`FullName()` (falls back to email if unset, for accounts created before names were required), used to show "invited by" on the setup page and invite inbox
- [x] Setup page collaborator management - Owners see pending invites and can send new ones from the plan's Setup page (`internal/views/templates/pages/setup.html`)

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
- [x] POST `/plan/{id}/invites` - Create a collaborator invite (requires Owner access)
- [x] POST `/invites/{id}/accept` - Accept an invite addressed to the logged-in user
- [x] POST `/invites/{id}/reject` - Decline an invite addressed to the logged-in user

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
- [x] User model and database table
- [x] User role-based access control (Owner, Editor, Viewer — per-plan via `plan_access`/invites, not a global Admin role)
- [ ] User profile management (profile page is currently display-only — no edit form)
- [ ] User preferences/settings
- [ ] Workspace/organization support (still no multi-plan "team" grouping above individual plan invites — see Teams & Collaboration above for what exists today)
- [ ] Audit logging for plan changes

### Database Implementation
- [x] Migration system (`internal/store/migrations.go`, run automatically on store init)
- [x] Persistent plan storage (SQLite, replacing the in-memory store)
- [x] User data persistence
- [ ] PostgreSQL/MySQL schema design (SQLite is still the only backend — fine for current single-instance Docker deployment, would need revisiting for horizontal scaling)
- [ ] Connection pooling and health checks (beyond the Docker Compose HTTP healthcheck — no DB-level pooling since SQLite is a single file)
- [ ] Plan versioning/history (optional)

### Data Persistence
- [x] Implement database store interface (`store.PlanStore`)
- [x] Save revenue streams/products to database
- [x] Save operating expenses to database
- [x] Save COGS to database (captured per-unit on the Sales Forecast form, persists with products)
- [x] Save capital assets/depreciation to database
- [x] Save funding sources and startup costs
- [x] Plan update/edit persistence (clear-then-rebuild pattern across all form handlers)

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
- [x] Docker containerization - Multi-stage production `Dockerfile` (Alpine, cgo-enabled build for `go-sqlite3`) plus a separate `Dockerfile.dev` running `air` for hot reload; `docker-compose.yml` (production, with HTTP healthcheck and a named volume for the SQLite file) and `docker-compose.dev.yml` (source bind-mounted, Go module/build caches as volumes)
- [x] Database migrations automation - `migrate` CLI command runs the migration list; also runs automatically on every store init
- [x] `Makefile` with `build`/`serve`/`migrate`/`reset`/`test` (local Go) and `docker-build`/`up`/`down`/`logs`/`dev`/`dev-down`/`dev-logs` (Docker) targets
- [ ] Environment configuration (.env, config files) - currently just CLI flags (`--db`, `--port`); no `.env`/secrets management
- [ ] CI/CD pipeline
- [ ] Logging and monitoring (beyond the existing request logger middleware)
- [ ] Error tracking (Sentry, etc.)
- [ ] Performance monitoring

Note: the Makefile comments indicate the production image (`docker-compose.yml`) is intended to be deployed via Coolify; actual live-deployment status is not tracked in this repo.

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
4. ~~**Data persistence**~~ ✅ **COMPLETED** - All plan sub-entities persist to SQLite via the clear-then-rebuild pattern
5. ~~**Docker containerization**~~ ✅ **COMPLETED** - Production + dev Dockerfiles, docker-compose files, Makefile targets
6. **Testing** - `internal/handlers/` has exactly one test file (`auth_test.go`); no handler-level tests exist yet for the report pages (GetIncomeStatement/GetBalanceSheet/GetAnalytics), the form POST handlers (Payroll/SalesForecast/OpExpenses/CashFlow), or the new invite flow (PostCreateInvite/PostAcceptInvite/PostRejectInvite). Domain-level projection engine already has coverage in `projection_test.go`
7. **Form Validation** - Client and server-side validation for all inputs
8. **Error handling** - Comprehensive error messages and recovery flows
9. **Income tax modeling** - No tax-rate input exists anywhere in the app; Net Income is currently pre-tax

## Project Structure Reference

```
.
├── cmd/cli/
│   ├── main.go                  # CLI entry point, flag parsing, command dispatch (serve/migrate/reset)
│   ├── serve.go                 # Server setup: DB connection, template loading, controller wiring, middleware chain
│   └── migrate.go               # `migrate` command (runs the migration list)
├── internal/
│   ├── domain/
│   │   ├── entities/            # All domain types (aggregates, value objects, projection)
│   │   │   ├── aggregate_root.go # AggregateRoot marker interface + aggregate boundary docs
│   │   │   ├── money.go         # Money value object (USD cents, int64-backed, JSON-compatible)
│   │   │   ├── errors.go        # ErrValidation sentinel
│   │   │   ├── plan.go          # Plan aggregate root, business logic, ValidatedPlan
│   │   │   ├── plan_*.go        # Plan companion files (payroll, cash flow, sales forecast, starting point, json)
│   │   │   ├── user.go          # User aggregate root (id, email, first/last name, access levels) + VerifiedUser opaque token
│   │   │   ├── session.go       # Session domain model + NewUserWithPassword
│   │   │   ├── invite.go        # PlanInvite domain model (teams & collaboration)
│   │   │   ├── projection.go    # Month-by-month financial projection engine
│   │   │   └── *_test.go        # Domain and projection tests
│   │   ├── events/               # DomainEvent interface + concrete events, one file per concern
│   │   │   ├── base_event.go    # DomainEvent interface, BaseEvent
│   │   │   ├── plan_events.go   # PlanCreated, PlanUpdated, PlanDeleted (unused so far)
│   │   │   ├── invite_events.go # UserInvitedToPlan
│   │   │   └── user_events.go   # UserRegistered
│   │   └── repositories/        # Repository interfaces (one file per aggregate/sub-entity boundary)
│   │       ├── plan.go          # PlanRepository
│   │       ├── access.go        # AccessRepository
│   │       ├── user.go          # UserRepository
│   │       ├── session.go       # SessionRepository
│   │       ├── invite.go        # InviteRepository
│   │       ├── wizard_progress.go # WizardProgressRepository
│   │       ├── wizard_item.go   # Shared ItemStatus + *Item/*Row DTOs for the wizard-item repos below
│   │       └── capital_asset.go, startup_cost.go, funding_source.go, salary_role.go,
│   │           benefit.go, product.go, inventory_purchase.go, distribution.go,
│   │           operating_expense.go, sales_growth_curve.go, payroll_tax_rates.go,
│   │           starting_balances.go  # Wizard sub-entity repos — Plan-aggregate-boundary
│   │           CRUD-draft-step interfaces, not separate aggregate roots
│   ├── application/
│   │   ├── interfaces/          # Application service interfaces (one per domain hub)
│   │   │   └── plan_service.go, auth_service.go, access_service.go, invite_service.go,
│   │   │       hub_completion_service.go, starting_point_service.go, payroll_service.go,
│   │   │       sales_forecast_service.go, cash_flow_service.go, operating_expenses_service.go
│   │   ├── commands/            # Write-side command DTOs — Plan and Operating Expenses so far,
│   │   │   │                      one file per command (mirrors go-ddd's
│   │   │   │                      create_product_command.go layout)
│   │   │   ├── create_plan_command.go, update_plan_command.go, delete_plan_command.go
│   │   │   │                      # CreatePlan/UpdatePlan/DeletePlan + Result wrappers
│   │   │   └── create_operating_expense_command.go, update_operating_expense_command.go,
│   │   │       delete_operating_expense_command.go
│   │   │                          # CreateOperatingExpense/UpdateOperatingExpense/
│   │   │                          # DeleteOperatingExpense + Result wrappers
│   │   │                          Wired through PlanController/PlanService and
│   │   │                          OperatingExpensesController/OperatingExpensesService (see
│   │   │                          "Application Layer: go-ddd Comparison" below)
│   │   ├── common/               # Output DTOs shared between command/query sides
│   │   │   ├── plan_result.go   # PlanResult — thin write-ack shape, not the Plan entity
│   │   │   └── operating_expense_result.go # OperatingExpenseResult — same shape, for Cost
│   │   ├── mapper/                # Entity → Result DTO conversion
│   │   │   ├── plan_result.go   # NewPlanResultFromEntity
│   │   │   └── operating_expense_result.go # NewOperatingExpenseResultFromEntity
│   │   └── services/            # Application service implementations (delegate to repositories)
│   │       ├── plan_service.go  # Domain construction/validation + commands for Plan
│   │       ├── plan_service_test.go # App-layer tests (real temp-file SQLite)
│   │       ├── auth_service.go  # Session expiry check lives here (not in infrastructure)
│   │       ├── access_service.go # Access level validation lives here (not in infrastructure)
│   │       ├── operating_expenses_service.go # CreateOperatingExpense/UpdateOperatingExpense
│   │       │   are the only callers of domain.NewCost; hub-summary/step-draft methods remain
│   │       │   thin pass-throughs (see roadmap below for why)
│   │       └── invite_service.go, hub_completion_service.go, starting_point_service.go,
│   │           payroll_service.go, sales_forecast_service.go, cash_flow_service.go
│   │           # Still thin CRUD pass-throughs over wizard sub-entities — no command/query
│   │           # DTOs yet, see roadmap below
│   ├── interface/web/           # HTTP controllers — one struct per domain, registers its own routes
│   │   ├── plan.go              # PlanController: plan lifecycle (setup/edit/delete) + report pages
│   │   ├── auth.go              # AuthController: login, signup, logout, profile
│   │   ├── invites.go           # InviteController: create/accept/reject
│   │   ├── access_middleware.go # PlanAccessMiddleware: route-guards by Owner/Editor/Viewer
│   │   ├── payroll.go, cash_flow.go, sales_forecast.go, starting_point.go,
│   │   │   operating_expenses.go  # One controller per wizard hub
│   │   ├── errors.go, wizard_shared.go  # Shared helpers (no controller struct)
│   │   └── auth_test.go         # Authorization middleware tests (only handler test file so far)
│   ├── middleware/
│   │   └── logger.go            # Request logging middleware
│   ├── infrastructure/
│   │   ├── sqlite/               # Repository implementations (one file per aggregate/sub-entity)
│   │   │   ├── connection.go, migrate.go  # DB connection + migration runner
│   │   │   ├── plan_repository.go, access_repository.go, user_repository.go,
│   │   │   │   session_repository.go, invite_repository.go, wizard_progress_repository.go,
│   │   │   │   capital_asset_repository.go, startup_cost_repository.go, ... (one per
│   │   │   │   domain/repositories/*.go interface)
│   │   │   ├── outbox.go        # insertOutboxEvents — writes drained domain events in the
│   │   │   │                      same transaction as the aggregate save
│   │   │   └── outbox_relay.go  # OutboxRelay — polls outbox_events, publishes (log-only
│   │   │                          stub), marks processed
│   │   ├── db/sqlc/               # sqlc-generated query code (compiled from sql/queries/*.sql)
│   │   └── config/config.go       # CLI flag / env config
│   └── views/                   # Templates, static assets, and view-layer glue
│       ├── builders.go          # Page view builders (Build*Page functions)
│       ├── types.go             # View data types (per-page structs, InviteSummary, etc.)
│       ├── funcs.go             # html/template helper funcs (formatMoney, addMoney, formatPercent, formatRatio)
│       ├── templates_embed.go   # Embedded template FS
│       ├── static_embed.go      # Embedded static FS
│       ├── templates/
│       │   ├── base.html        # Base layout template
│       │   ├── pages/           # Page templates
│       │   └── components/      # Reusable components
│       └── static/               # CSS, JS, icons
├── sql/
│   ├── queries/                  # Hand-written SQL that sqlc compiles into internal/infrastructure/db/sqlc
│   └── migrations/               # Ordered SQL migration files
├── Dockerfile / Dockerfile.dev   # Production (multi-stage, Alpine) and dev (air hot-reload) images
├── docker-compose.yml / docker-compose.dev.yml
├── .air.toml                    # air hot-reload config (used by Dockerfile.dev)
├── sqlc.yaml                    # sqlc config (engine: sqlite)
├── Makefile                     # build/serve/migrate/reset/test + docker-build/up/down/logs/dev targets
├── go.mod
└── go.sum
```

## Key Files to Understand

- **plan.go** (domain) - Core business logic, validation, calculations
- **projection.go** (domain) - Financial projection engine; its file-level comment documents every modeling simplification
- **serve.go** (cmd/cli) - Route definitions and server setup
- **plan.go** (handlers) - HTTP request/response handling
- **invites.go** (handlers) - Collaboration/invite flow
- Templates in `internal/views/templates/` - UI layer

## Architectural Guidelines (DDD / CQRS)

These rules govern how code should be written and reviewed in this project. They are adapted from an opinionated DDD template and reflect the intended direction for ongoing development.

### Layering Rules (Onion Architecture)

Dependencies must only point **inward** toward the domain layer.

| Layer | Path | Allowed dependencies |
|---|---|---|
| Domain | `internal/domain/` | None — zero imports of DB, HTTP, or third-party libs (`github.com/google/uuid` is the one permitted exception — see Tech Decisions Log) |
| Application | `internal/application/` | Domain only |
| Infrastructure | `internal/infrastructure/sqlite/` | Domain interfaces (`internal/domain/repositories/`) |
| Interface | `internal/interface/web/`, `internal/views/` | Application interfaces (`internal/application/interfaces/`) |

(As of the `refactor/web-controllers` branch: `internal/store/` was split into
`internal/infrastructure/{sqlite,db,config}/`, and `internal/handlers/` was reorganized into
`internal/interface/web/` as one controller struct per domain — see "Application Layer: go-ddd
Comparison" below for the fuller picture and its sources.)

### CQRS Separation

- **Commands** (write path): `Post*` handlers mutate state, then redirect (POST–redirect–GET).
  For `Plan`'s own lifecycle (create/update-core-details/delete), this now goes through real
  command DTOs — `PlanController` builds a `commands.CreatePlan`/`UpdatePlan`/`DeletePlan` and
  calls the matching `PlanService` method, which is the only place allowed to call
  `domain.NewPlan`/`Plan.ChangeCoreDetails`. Every other `Post*` handler (payroll, cash flow,
  sales forecast, starting point, operating expenses, invites) still fetches the `Plan` aggregate
  via `planSvc.Get`, mutates it directly via a domain method (`AddX`/`ClearX`/
  `RecordUserInvited`), and calls `planSvc.Save` — see "Application Layer: go-ddd Comparison" for
  why that's an intentional, scoped-down first step rather than an inconsistency to "fix" by
  reflex.
- **Queries** (read path): `Get*` handlers only read — never mutate. If a handler both reads and writes, split it.

### Repository Rules

1. **No business logic** in `internal/infrastructure/sqlite/` — only translate between domain structs and SQL rows.
2. **Dependency inversion** — repository implementations in `internal/infrastructure/sqlite/` implement interfaces defined in `internal/domain/repositories/`.
3. **`Find*` vs `Get*`** — `Find*` may return nil/empty; `Get*` must return a value or a descriptive error.
4. **Read-after-write** — after inserting or updating a row, re-read it from the DB before returning to the caller; never return the struct you wrote directly.
5. **Soft deletion only** — set `deleted_at` instead of hard-deleting rows.
6. **No DB-level defaults for IDs or timestamps** — generate UUIDs and `created_at` values in the domain layer factory (`New*`), not in the SQL schema.

### Domain Entity Rules

- Business rules and validation live in domain constructors (`NewX`) and update methods, not in handlers or the store.
- Do **not** validate entities on read — historical data may not satisfy today's rules; enforcement is write-side only.
- Money is stored as `int64` cents, never `float64`.

### Idempotency

The `idempotency_keys` table (migration 49) is the intended backing store. The key must be reserved atomically (`INSERT … ON CONFLICT DO NOTHING`) before the command executes. The HTTP middleware wrapper that reads the `Idempotency-Key` header and checks/stores it is not yet wired — this is the remaining gap.

### Domain Events

Aggregates emit events via `recordEvent()`; repositories drain them with `PullEvents()` inside the same DB transaction that saves the entity (`internal/infrastructure/sqlite/outbox.go`'s `insertOutboxEvents`, called from `plan_repository.go` and `user_repository.go`). Event types live in `internal/domain/events/` (`base_event.go`, `plan_events.go`, `invite_events.go`, `user_events.go`), one file per aggregate/concern, not a single `events.go`. The outbox relay (`internal/infrastructure/sqlite/outbox_relay.go`) **is implemented**: a goroutine polls `outbox_events` for `processed_at IS NULL` rows on a ticker and marks them processed after a (currently log-only, matching the go-ddd template's own stub) publish step — see "Application Layer: go-ddd Comparison" for the source this mirrors.

### Application Layer: go-ddd Comparison

This project is deliberately mirroring [`sklinkert/go-ddd`](https://github.com/sklinkert/go-ddd)
(pinned at commit `08a11fc1a278977ad6a240e1399ceaa61556c01a`), an opinionated Go DDD template with
a 9-chapter tutorial (`docs/tutorial/01-the-domain.md` through `09-testing.md`). This section
records what that template's `internal/application/` actually contains, how this repo's
`internal/application/` compares today, and what's still gap vs. deliberate divergence — so a
future session doesn't re-derive this or re-introduce dead code.

**The reference shape** (`internal/application/{command,query,common,interfaces,mapper,services}`):
a `command` package holds one `CreateXCommand`/`UpdateXCommand`/`DeleteXCommand` struct per
write operation (raw primitive fields — the command is a request, not a validated object) plus a
`*CommandResult` wrapper; a `query` package holds `GetXByIdQuery`/`GetAllXQueryResult` structs; a
`common` package holds the `XResult` DTO both sides return (never the entity — "the entity...
never crosses into the interface layer," chapter 6); `mapper` converts entity → `XResult`;
`interfaces` declares one `XService` per aggregate whose methods take/return only these DTOs; and
`services` implements them — the *only* place allowed to call the domain's `NewX` constructor and
mutation methods, wrapped in an idempotency decorator (chapter 8) where the operation has side
effects. Chapter 6 is explicit about what it deliberately skips: no command/query bus (direct
interface calls), no separate read store, no event sourcing — "show the pattern, not the
framework."

**Where this repo already matches, without any application-layer change** (chapters 1–5, 7):
`Plan`/`User` aggregate roots with unexported fields and `NewX` constructors that validate
invariants (chapter 2); `Money` as a real value object (chapter 3); ID-only cross-aggregate
references, wizard sub-entities inside the `Plan` boundary (chapter 4); repository interfaces
owned by `internal/domain/repositories/`, implementations in `internal/infrastructure/sqlite/`
(chapter 5); a working transactional outbox with a log-only relay publisher, same as the template
(chapter 7). These were done in the 2026-08-07 session and didn't need touching for this pass.

**What was dead code, and what changed for it (this session):**
`internal/application/commands/plan.go` and `internal/application/queries/plan.go` existed but
were never imported anywhere (confirmed by repo-wide grep) — and `PlanController` called
`domain.NewPlan(...)` and `plan.ChangeCoreDetails(...)` directly from the HTTP layer, which is
exactly the layering violation chapter 6 is about (the interface layer touching domain
construction/mutation instead of the application service doing it).
- `commands/plan.go` was split into one file per command — `create_plan_command.go`,
  `update_plan_command.go`, `delete_plan_command.go` — matching go-ddd's own layout exactly
  (`create_product_command.go`, `update_product_command.go`, `delete_product_command.go`, each
  holding its `XCommand` + `XCommandResult` pair). Each now has real `CreatePlan`/`UpdatePlan`/
  `DeletePlan` structs plus `*Result` wrappers (`internal/application/common/plan_result.go`'s
  `PlanResult`, built by `internal/application/mapper/plan_result.go`).
  `PlanService.CreatePlan`/`UpdatePlan`/`DeletePlan` are now the only code that calls
  `domain.NewPlan`/`Plan.ChangeCoreDetails`; `PlanController` builds a command from the parsed
  form and calls the service. This is the actual fix — verified end-to-end in a browser (signup →
  create plan → edit core details → delete; `outbox_events` shows `plan.created` then
  `plan.updated` in order) and by `TestPlanService_CreatePlan(_RejectsInvalid)`/
  `_RejectsUnknownOwner`/`UpdatePlan`/`DeletePlan` in `plan_service_test.go`.
- `queries/plan.go` was **deleted**, not wired up. Reasoning: `PlanService.Get`/`GetAll`/
  `GetUserPlans` already satisfy the *substance* of the read-side rule — no validation, no
  construction, no side effects, pure reads — which is what chapter 6 actually requires of the
  read path. Wrapping them in `GetPlanByIdQuery{PlanID}`-style single-field structs that still
  return the raw `*domain.Plan` would add indirection without adding the actual protection a
  `PlanResult` DTO buys (hiding mutation methods from callers), because every caller — the report
  pages (`views.BuildIncomeStatementPage`/`BuildBalanceSheetPage`/`BuildAnalyticsPage`, which need
  `projection.go`'s calculation methods) and every wizard controller (which mutates the fetched
  `Plan` and calls `Save`) — structurally needs the full aggregate, not a display shape. Unlike
  go-ddd's `Product`, `Plan` doubles as both the write-side aggregate and the read-side
  calculation engine, so go-ddd's "reads are thin display DTOs" premise doesn't hold for it.
  **Trigger to reverse this**: if a genuinely display-only read model shows up (e.g. a lightweight
  "plan summary for the dashboard list" that doesn't need projections), that's the moment to add a
  real `common.PlanSummary` + mapper + query struct — not before.

**`VerifiedUser` — mirroring go-ddd's `ValidatedSeller` pattern for Plan ownership.** Chapter 4
requires `NewProduct(name, price, seller ValidatedSeller)` — "the constructor takes a
`ValidatedSeller`, not a `Seller`... you literally cannot call this function with an unvalidated
one." Before this session, `Plan.NewPlan` took a bare `ownerID uuid.UUID`: any well-formed UUID
was accepted as the owner, with nothing checking it actually named an existing User. Now:
- `entities.VerifiedUser` (`internal/domain/entities/user.go`, next to `User` — same file
  placement as `ValidatedPlan` living in `plan.go`) is an opaque wrapper: `NewVerifiedUser(u *User)
  (VerifiedUser, error)` only accepts an actual `*User`, and the domain layer has no database
  access to check existence itself — so a `VerifiedUser` can only come from application-layer code
  that already looked one up.
- `Plan.NewPlan`'s signature is now `NewPlan(name string, startingMonth, startingYear int, owner
  VerifiedUser) (*Plan, error)` — a bare `uuid.UUID` doesn't type-check anymore, the same
  "invalid states unrepresentable" guarantee `ValidatedSeller`/`ValidatedPlan` give elsewhere.
- `PlanService.CreatePlan` does the actual verification: `s.verifyOwner(cmd.OwnerID)` calls
  `UserRepository.GetUser(ownerID)` (returns `domain.ErrUserNotFound` if no such row exists) and
  wraps the result in `NewVerifiedUser`. `NewPlanService` now takes a `repositories.UserRepository`
  as a second constructor argument (`cmd/cli/serve.go` already had `userRepo` in scope for
  `authSvc`; wiring it into `planSvc` too was a one-line change).
- **Why this matters even though the only current caller (the logged-in web session) already
  proves the user exists**: the guarantee needs to live in the aggregate/service boundary, not in
  caller discipline — chapter 6's own test for "is this in the right layer" is "could a second
  delivery mechanism (a CLI, a queue consumer) get different business behavior by calling things
  differently?" Before this change, a CLI script or queue consumer calling `PlanService.CreatePlan`
  (or `domain.NewPlan` directly) with a fabricated UUID would silently create an orphaned plan
  owned by nobody. Now it can't — proven by
  `TestPlanService_CreatePlan_RejectsUnknownOwner` in `plan_service_test.go`, and verified live
  (created a plan through the real signup → session → `POST /plan/setup` flow against a scratch
  DB; `plan_access` row shows the real user's id as `owner`).
- Domain-layer tests (`plan_test.go`, `projection_test.go`) that call `NewPlan` directly build a
  `VerifiedUser` via the shared `newVerifiedOwner(t)` helper (wraps a plain in-memory `NewUser(...)`
  — no repository needed, since `NewVerifiedUser` itself doesn't do I/O). Only
  `plan_service_test.go`'s `CreatePlan`-based tests need an actually-persisted user
  (`newPersistedOwner(t, users)`), since that's the path that exercises the real lookup.

**Deliberate divergences from the reference, and why:**
- `PlanService.Get`/`GetAll`/`GetUserPlans`/`Save`/`Delete` still take/return `*domain.Plan`
  directly and remain on the interface unchanged — see above. `Save` in particular stays as the
  general-purpose "persist an already-mutated aggregate" entry point every non-`Plan`-lifecycle
  controller depends on.
- No `context.Context` parameter on `CreatePlan`/`UpdatePlan`/`DeletePlan`, unlike go-ddd's
  service methods. Nothing in this codebase's `application`/`domain` layers uses
  `context.Context` today (checked — zero hits); adding it only to these three methods would be
  an inconsistency the rest of the interface doesn't share. This will need revisiting the moment
  idempotency (below) is wired, since chapter 8's cleanup-on-failure path specifically needs
  `context.WithoutCancel`.
- The web controller still imports `internal/domain/entities` (as `domain`) — go-ddd's REST
  controllers never do, since they only see command/query/result types serialized to/from JSON.
  This repo is server-side-rendered (`html/template`), so read pages must pass the real entity
  (with its getter methods) to view builders, and `domain.Viewer/Editor/Owner` access-level
  constants are used directly by route middleware. That's a structural consequence of SSR vs. a
  JSON API, not a violation of the rule chapter 6 actually cares about (don't *construct or
  mutate* domain objects at the edge) — `PostSetup`/`PostUpdateSetup`/`PostDeletePlan` no longer
  do that; the read-only handlers never did.

**Roadmap: extending the command pattern to the other 9 domains.** `access`, `auth`, `invite`,
`hub_completion`, `starting_point`, `payroll`, `sales_forecast`, `cash_flow` services still take
plain scalar args and pre-built domain value-object structs (`SalaryRole`, `Product`,
`CapitalAsset`, etc.), with **no domain-level constructors or validation for those wizard
sub-entities** — they're populated and mutated as part of `Plan`'s `Add*`/`Clear*` methods, not
independently constructed like `Plan`/`User`/`Money` are. Retrofitting `CreateXCommand`-style DTOs
onto them now would be cargo-culting go-ddd's shape without its actual payoff (nothing to validate
independently — the invariant-guarding is what the pattern is *for*). The real prerequisite,
flagged in Known Limitations below ("Wizard items have no domain-level ID"), is giving each wizard
sub-entity its own `uuid.UUID` and constructor first. Once that lands, the same three-step recipe
used for `Plan` applies: (1) command DTO + Result wrapper, (2) service method that's the only
caller of the new constructor, (3) controller builds the command instead of touching the entity.

**Operating Expenses (2026-08-12, continued) — the recipe proven on a second domain.** Operating
Expenses was picked deliberately as the smallest wizard hub (a single repeatable `Cost` list, per
`SectionOperatingExpenses`'s doc comment in `plan.go`) to prove the `Plan` recipe generalizes
before tackling the larger hubs. What changed:
- **Prerequisite closed for `Cost`**: `entities.Cost` (`plan_growth.go`, shared by Operating
  Expenses and the still-unwired COGS field) gained an unexported `id uuid.UUID`, an `ID()`
  getter, and `NewCost(name, baseAmount, growth) (Cost, error)` — same shape as `NewPlan`:
  generates a UUIDv7, validates invariants (empty name / negative amount / invalid growth type),
  refuses to return an invalid value. The old free function `ValidateOpExpense` was deleted (dead
  code once `AddOpEx`/`AddCOGS` routed through `NewCost` instead of hand-rolling the same three
  checks). `Cost` also gained `SetID(id)` (pointer receiver, no validation) for the *reconstruction*
  path — mirrors `User.SetID`'s existing convention (`NewUser` then override the ID) — because rows
  read back from SQLite are sometimes incomplete wizard drafts (empty name, no growth chosen yet)
  that would fail `NewCost`'s validation. `operatingExpenseFromRow` (sqlite) and the OpEx
  reconstruction loop in `plan_repository.go`'s `LoadPlanChildren`-equivalent both now thread the
  row's own ID through `SetID` so `item.Cost.ID()` always equals the wizard row's ID — closing the
  domain-level-ID gap for this one sub-entity.
- **Commands**: `commands/create_operating_expense_command.go` / `update_...` / `delete_...`, one
  file each, matching `Plan`'s layout exactly. `common.OperatingExpenseResult` +
  `mapper.NewOperatingExpenseResultFromEntity` mirror `PlanResult`/`NewPlanResultFromEntity`.
  Command fields use `domain.Money`/`domain.GrowthStrategy` directly rather than further-decomposed
  primitives — consistent with how this codebase already treats `Money` as a pervasive value type
  at every layer (unlike `Plan`'s command, which only ever had primitive fields to begin with).
- **Service**: `OperatingExpensesService.CreateOperatingExpense`/`UpdateOperatingExpense` are now
  the only code that calls `domain.NewCost`; `DeleteOperatingExpense`'s signature changed from a
  bare `(uuid.UUID) error` pass-through to `(*commands.DeleteOperatingExpense)
  (*commands.DeleteOperatingExpenseResult, error)`, matching `DeletePlan`.
- **Controller — the one deliberate divergence from a literal Create/Update/Delete mapping**:
  Operating Expenses is a 3-step wizard (name → amount → growth), not a single-shot form like
  `Plan`'s setup page. The first two steps only ever hold a *partial* `Cost` (no growth strategy
  yet), which would fail `NewCost`'s validation — so they still go through the pre-existing raw
  `SaveOperatingExpenseStep(itemID, cost, step, StatusDraft)` pass-through unchanged, since there is
  no valid domain entity to construct until the wizard finishes. Only the final "growth" step —
  where a full name/amount/growth triple first exists — now builds a `CreateOperatingExpense` or
  `UpdateOperatingExpense` command (chosen by whether the item was already complete) instead of
  calling the deleted `domain.ValidateOpExpense` directly. `PostOperatingExpenseDelete` builds a
  `DeleteOperatingExpense` command. This is the shape future wizard-hub work should expect: the
  command pattern applies at the point a full valid entity exists, not to every intermediate
  step-save in a multi-step form.
- **Verified**: `go build ./...` / `go test ./...` pass (new `TestNewCost`/`TestCost_SetID`/
  `TestPlan_AddOpEx_AssignsID` in `plan_growth_test.go`), plus live in a browser against a scratch
  DB — created an expense through all 3 steps (draft→complete via `CreateOperatingExpense`),
  edited it back through the wizard (same row updated via `UpdateOperatingExpense`, no duplicate
  row, ID stable across the edit), and deleted it (soft-deleted, `deleted_at` set, disappears from
  the list).
- **Next hub**: Starting Point (`internal/interface/web/starting_point.go`) is next in line, and is
  a bigger step up in shape, not just size — it has **three** independent sub-entity types
  (`CapitalAsset`, `StartupCost`, `FundingSource`), each needing its own `NewX` constructor plus its
  own three command files/service methods (so nine new command files total, not three), and a
  fourth singleton section (`StartingBalances`, cash-on-hand) that has no wizard-item ID at all
  today and doesn't obviously need one (it's a per-plan singleton row, not a repeatable list — worth
  confirming that reasoning before assuming it needs the same treatment). Expect to reuse the same
  "only the step that completes a full entity constructs it" pattern established here for any of
  the three that are also multi-step wizards.

**Idempotency (chapter 8) is the natural next increment if commands expand.** The
`idempotency_keys` table exists (migration 49) but nothing reserves/checks it yet (see the
Idempotency section above). `CreatePlan` is exactly the shape that would carry an
`IdempotencyKey string` field and go through a `withIdempotency[T any]`-style decorator — but that
needs `context.Context` threaded through first (previous bullet), so do that as one deliberate
step, not bolted onto this pass.

---

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
- **Templates/Static Merged into `views` Package (2026-07-29)**: Moved `internal/templates/` and `internal/static/` under `internal/views/` (as `templates/`, `static/`, `templates_embed.go`, `static_embed.go`) so view-layer concerns (builders, types, embedded assets) live in one package instead of three.
- **All Form Data Now Persists to SQLite (2026-07-30)**: Wired Starting Point, Payroll, Sales Forecast/Products, Operating Expenses, and Cash Flow handlers to actually save/reload from the SQLite store (previously some of these round-tripped through in-memory state only within a session). Closes out the "Data Persistence" checklist that had been open since the store was first built.
- **Plan-Level Invites, Not Org/Workspace Teams (2026-08-01)**: "Teams and collab" was implemented as per-plan email invites (`domain.PlanInvite`, `plan_invites` table) at a chosen Owner/Editor/Viewer access level, surfaced as a pending-invite inbox on the home dashboard — deliberately not a workspace/organization model. No email is actually sent; the invitee only sees the invite if they're logged into the app with a matching email. Choose this scope if extending collaboration further, rather than assuming an org layer already exists.
- **First/Last Name Now Required at Signup (2026-08-01)**: `NewUserWithPassword` gained `firstName`/`lastName` params (migrations 7–8 add the columns, default `''` for existing rows). `User.FullName()` falls back to email when both are blank, so legacy accounts created before this change still render sensibly in the invite inbox and setup page's "invited by" text.
- **Docker Deployment Split into Prod/Dev Images (2026-08-01/02)**: Production `Dockerfile` is a multi-stage Alpine build (cgo enabled for `go-sqlite3`, `CGO_CFLAGS=-D_LARGEFILE64_SOURCE` needed for musl+sqlite compatibility) producing a small runtime image; `Dockerfile.dev` instead installs `air` and hot-reloads from a bind-mounted source tree via `docker-compose.dev.yml`. `docker-compose.yml` (prod) adds an HTTP healthcheck and a named volume for the SQLite file so data survives container recreation. Makefile centralizes both local-Go and Docker workflows.

- **DDD / CQRS Architecture Implemented (2026-08-07)**: Added `internal/domain/repositories/` (repository interfaces), `internal/application/interfaces/` (service interfaces), and `internal/application/services/` (implementations). All wizard-item Delete methods now use soft-delete (`deleted_at`). Business logic (session expiry, access-level validation) moved from `internal/store/` into the service layer. `Find*` vs `Get*` naming convention enforced: nine methods that returned `nil, nil` on not-found renamed from `Get*Draft` to `Find*Draft` across all four layers. UUIDv7 adopted in all domain constructors. ValidatedPlan opaque token pattern applied to PlanRepository.Save. Domain events (`DomainEvent`, `PlanCreated`/`PlanUpdated`/`PlanDeleted`) in `internal/domain/events.go`. Transactional outbox via `outbox_events` table (migration 48). Idempotency schema via `idempotency_keys` table (migration 49). CQRS command/query types in `internal/application/commands/` and `internal/application/queries/`. Application-layer tests in `internal/application/services/plan_service_test.go` (real temp-file `SQLiteStore`).
- **`github.com/google/uuid` Permitted in Domain Layer (2026-08-07)**: The "zero third-party" rule for the domain layer has one explicit carve-out: `github.com/google/uuid`. It is a pure value type (no DB driver, no HTTP framework coupling), stable, and pervasive across all layers; banning it from the domain would force awkward string conversions at every boundary. All other third-party imports remain prohibited in `internal/domain/`.
- **Read-After-Write Deferred (2026-08-07)**: Rule 4 (re-read from DB after insert/update before returning to caller) was not implemented. All write methods still return only `error`. The application uses POST-redirect-GET everywhere — the next `Get*` handler re-reads from the DB — so the intent is satisfied at the HTTP level. Cascading the change through 19 repository interfaces, 10 service implementations, and all callers is out of scope for this session. This is a known gap; document it rather than silently violate it.

### Known Limitations

- **No Income Tax Modeling** - Net Income throughout the Income Statement, Balance Sheet, and Analytics pages is pre-tax; the "Income Tax Expense" line always shows $0
- **No Export Functionality** - Can't export plans as Excel or PDF yet (browser print-to-PDF only)
- **Money Truncates Cents** - `domain.Money` is `int64`; form amounts are parsed as float64 then truncated, so fractional cents are dropped (pre-existing pattern from Starting Point, kept for consistency). Over a 36-month projection this can drift the balance sheet by a few dollars — `BuildBalanceSheetPage`/`BuildAnalyticsPage` tolerate up to $25 before flagging "does not balance"
- **Current/Quick Ratio show 0.00 when undefined** - if a plan has no Accounts Payable/Accrued Expenses, current liabilities are $0 and these ratios are mathematically undefined; the UI currently can't distinguish that from an actually-computed 0.0
- **Loan/Fixed-Asset categories are not sub-typed** - the Balance Sheet and Amortization tab show one combined "Loans Payable" and "Fixed Assets" figure rather than breaking out Line-of-Credit vs. Commercial Loan vs. Equipment vs. Real Estate, since the domain model doesn't track those distinctions
- **No email delivery** - Plan invites are stored in the database and only visible in-app to a logged-in user with the matching email; nothing is sent externally, so an invitee who doesn't already have an account (or doesn't think to log in) will never learn they were invited
- **Collaboration is per-plan, not org-wide** - there is no workspace/team entity above individual plans; each plan has its own independent set of invites/access grants
- **SQLite only** - single-file database; fine for the current single-instance Docker deployment but no connection pooling or horizontal-scaling story if that becomes necessary
- **No handler-level test coverage for report/form pages or invites** - `internal/handlers/auth_test.go` is the only handler test file; the report pages, all form POST handlers, and the invite accept/reject flow are exercised only by manual browser testing during development, not by automated tests
- **Most wizard items still have no domain-level ID** - `CapitalAsset`, `SalaryRole`, `Product`, etc. carry no `uuid.UUID` in the domain struct; the ID lives only in the `repositories.*Item` wrapper. `Cost` (used for Operating Expenses and COGS) is now the one exception — it carries its own `id` + `NewCost` constructor + `SetID` reconstruction hook, see "Application Layer: go-ddd Comparison" above. The rest are entities within the Plan aggregate boundary and should eventually carry their own ID in the domain (so `Plan.futurePurchases []CapitalAsset` holds identifiable entities). Doing so requires updating all store serialization, repository interfaces, and handler code. Tracked in `repositories/wizard_item.go`'s package comment.

### Questions to Ask If Stuck

1. **"Should I add a new route?"** → Check if it's in the route list in main.go. If not, add both GET and POST if it's a form.
2. **"How do I save data?"** → Use `app.Store.Save(plan)` after mutating the plan. Currently in-memory; will need DB migration.
3. **"Where do I put validation?"** → In the domain model (plan.go domain package) for business rules. Use handlers for HTTP-level validation.
4. **"Should I create a new handler?"** → Create a method on the App struct that returns an http.HandlerFunc.

---

**Last Updated**: 2026-08-12 (continued further)
**Session Focus**: Extended the go-ddd command pattern from `Plan` to Operating Expenses — the smallest wizard hub — to prove the recipe generalizes before tackling the larger hubs. Closed the actual prerequisite first (`entities.Cost` gained its own `uuid.UUID` + validating `NewCost` constructor, mirroring `Plan.validate()`/`NewPlan`), then applied the three-step recipe: `commands/create_operating_expense_command.go`/`update_...`/`delete_...` + `common.OperatingExpenseResult` + mapper; `OperatingExpensesService.CreateOperatingExpense`/`UpdateOperatingExpense`/`DeleteOperatingExpense` as the only code allowed to construct/mutate the entity; `OperatingExpensesController` builds commands from parsed form data at the wizard step that actually completes a valid `Cost`, rather than validating a hand-mutated struct directly. See "Application Layer: go-ddd Comparison" above for the full writeup, including the one deliberate divergence (multi-step wizard drafts still save partial fields via the pre-existing raw pass-through, since there's no valid entity to construct until the final step) and what's different about Starting Point, next in line.
**Total Remaining Items**: ~28 (across all categories)
**Critical Path**: ~~User Authorization~~ ✅ → ~~Form POST Handlers~~ ✅ → ~~Financial Calculations~~ ✅ → ~~Data Persistence~~ ✅ → ~~Docker Containerization~~ ✅ → ~~DDD/CQRS Architecture~~ ✅ → Testing → Form Validation → Income Tax Modeling

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

## Session Summary (2026-07-30): Data Persistence

### What Shipped
- Wired Starting Point, Payroll, Sales Forecast/Products, Operating Expenses,
  and Cash Flow handlers so every form actually saves to and reloads from the
  SQLite store, closing out the "Data Persistence" checklist that had been
  open since the store interface was first built.

## Session Summary (2026-08-01): Teams & Collaboration

### What Shipped
1. **Invite domain model** (`internal/domain/invite.go`) - `PlanInvite` with
   `Pending`/`Accepted`/`Rejected` status, tied to a plan, an email, an
   `AccessLevel`, and the inviting user.
2. **Invite handlers** (`internal/handlers/invites.go`) - `PostCreateInvite`
   (Owner-only, `POST /plan/{id}/invites`), `PostAcceptInvite` and
   `PostRejectInvite` (`POST /invites/{id}/accept|reject`, keyed by invite ID
   so authorization is checked inside the handler against the logged-in
   user's email rather than via the plan-scoped `RequireAccess` middleware).
3. **Store support** - `plan_invites` table (migration 9) plus an index on
   email (migration 10) for the pending-invite inbox lookup; new
   `PlanStore` methods `CreateInvite`/`GetInvite`/`GetInvitesForPlan`/
   `GetPendingInvitesForEmail`/`UpdateInviteStatus`.
4. **User names** - `User` gained `FirstName`/`LastName`/`FullName()`
   (migrations 7–8); signup now requires both, `NewUserWithPassword` gained
   the params. Used to show "invited by {name}" instead of a bare email.
5. **UI** - Home dashboard (`index.html`) shows a pending-invites onboarding
   section for the logged-in user's email; the Setup page shows an owner-only
   collaborator invite form and the plan's outstanding invites.

### Deliberately Deferred
- No email is sent for invites — the invitee must already have an account
  and log in to see it in their dashboard's pending-invites section.
- No workspace/organization layer — invites and access are strictly
  per-plan; there's still no way to group multiple plans under one team.

## Session Summary (2026-08-01 to 2026-08-02): Docker Deployment

### What Shipped
1. **Production image** (`Dockerfile`) - Multi-stage Alpine build; builder
   stage installs gcc/musl-dev for `go-sqlite3`'s cgo dependency and sets
   `CGO_CFLAGS=-D_LARGEFILE64_SOURCE` (needed for musl+sqlite largefile
   compatibility), runtime stage is a slim Alpine image with just the binary
   and a `/app/data` mount point for the SQLite file.
2. **Dev image** (`Dockerfile.dev`) - Installs `air` and hot-reloads Go/HTML
   changes from a bind-mounted source tree per `.air.toml`
   (`full_bin = "./tmp/northbasis-cli serve --db ./northbasis.db --port :8080"`).
3. **Compose files** - `docker-compose.yml` (production: HTTP healthcheck
   against `/`, named volume `northbasis-data` for the SQLite file) and
   `docker-compose.dev.yml` (bind-mounts the repo plus separate volumes for
   the Go module and build caches so rebuilds inside the container stay fast).
4. **Makefile** - Centralizes both local-Go (`build`/`serve`/`migrate`/
   `reset`/`test`/`clean`) and Docker (`docker-build`/`up`/`down`/`logs`/
   `dev`/`dev-down`/`dev-logs`) workflows; comments note production is
   intended to deploy via Coolify using `docker-compose.yml`.
5. **CLI flag hardening** (`cmd/cli/main.go`, 2026-08-02) - Minor cleanup to
   flag parsing/validation for the `serve`/`migrate`/`reset` commands.

### Deliberately Deferred
- No CI/CD pipeline wired up yet — Docker builds are local/manual only.
- No `.env`/secrets management — configuration is CLI flags only (`--db`,
  `--port`).

## Session Summary (2026-08-07): DDD / CQRS Architecture

### What Shipped
1. **Repository interfaces** (`internal/domain/repositories/`) — One file per
   aggregate boundary: `capital_asset.go`, `startup_cost.go`,
   `funding_source.go`, `salary_role.go`, `benefit.go`, `product.go`,
   `inventory_purchase.go`, `distribution.go`, `operating_expense.go`, plus
   `plan.go` and `user.go` for the core aggregates. These define the
   contracts that `SQLiteStore` implements.
2. **Application service layer** (`internal/application/interfaces/` +
   `internal/application/services/`) — Eight service pairs (plan, auth,
   access, starting_point, payroll, sales_forecast, cash_flow,
   operating_expenses). Service implementations delegate to repository
   interfaces; domain policy (session expiry, access-level validation) was
   moved here from the store.
3. **`Find*` vs `Get*` naming** — Nine `Get*Draft` methods renamed to
   `Find*Draft` across all four layers (repositories, store, services,
   handlers). `Get*` methods now consistently return a value or an error;
   `Find*` may return `nil, nil`.
4. **Soft deletion** — `deleted_at INTEGER` columns added to 10 tables via
   migrations 38–47 (`plans`, `capital_assets`, `startup_costs`,
   `funding_sources`, `salary_roles`, `benefits`, `products`,
   `inventory_purchases`, `distributions`, `operating_expenses`). All
   `Delete*` methods in the store now set `deleted_at` instead of issuing
   `DELETE FROM`. All `SELECT` queries on those tables now filter
   `AND deleted_at IS NULL`. `plan_access` and singleton tables
   (starting_balances, payroll_tax_rates, sales_growth_curve, wizard_sections)
   keep hard delete.
5. **Business logic out of store** — Removed session-expiry check and
   access-level validation from `sqlite.go`; added them to `auth_service.go`
   and `access_service.go` respectively.
6. **AGENTS.md updated** — Layering table corrected (application layer no
   longer marked "future"), project structure reference updated, new tech
   decisions logged.

### Deliberately Deferred
- **Read-after-write (Rule 4)**: All write methods still return only `error`.
  POST-redirect-GET at the HTTP level provides the equivalent guarantee.
  Full implementation would cascade through 19 repository interfaces, 10
  services, and all callers — out of scope.
- **Outbox relay goroutine**: `outbox_events` rows are written atomically but
  never delivered. A relay goroutine (or background job) that polls for
  `processed_at IS NULL` rows and publishes them to a broker, then sets
  `processed_at = now()` on success, is the remaining gap.
- **Idempotency middleware**: `idempotency_keys` schema is in place (migration
  49) but the HTTP middleware that reads the `Idempotency-Key` header and
  executes the check-and-reserve logic is not yet wired.

## Session Summary (2026-08-07, continued): go-ddd Rules Implementation

### What Shipped
1. **UUIDv7 in all domain constructors** — `NewPlan`, `NewUser`, `NewPlanInvite`, and all wizard-item `Create*Draft` methods now call `uuid.NewV7()` instead of `uuid.New()`. Time-ordered IDs; errors propagated (not panicked).
2. **ValidatedPlan pattern** — `Plan.validate()` private method checks domain invariants; `Plan.Validate()` returns a `ValidatedPlan` opaque token on success. `PlanRepository.Save` requires `ValidatedPlan`. The `PlanService.Save(*Plan)` bridge validates internally before delegating, so all callers keep the ergonomic signature.
3. **Domain events** — `DomainEvent` interface + `PlanCreated`/`PlanUpdated`/`PlanDeleted` structs in `internal/domain/events.go`. `NewPlan` emits `PlanCreated`; `ChangeCoreDetails` emits `PlanUpdated`. Aggregates buffer events via `recordEvent()`; `PullEvents()` drains and clears the buffer (destructive — call once per save).
4. **Transactional outbox** — `SQLiteStore.Save` opens a transaction, writes the plan row, drains `plan.PullEvents()`, inserts each into `outbox_events`, then commits. Schema via migration 48 with a partial index on `processed_at IS NULL` for efficient relay queries.
5. **Idempotency schema** — `idempotency_keys` table via migration 49 (`key`, `status`, `response`, `created_at`, `updated_at`). Intended for `INSERT … ON CONFLICT DO NOTHING` atomic reservation.
6. **CQRS formal types** — `internal/application/commands/plan.go` (CreatePlan, UpdatePlan, DeletePlan) and `internal/application/queries/plan.go` (GetPlan, GetUserPlans, GetAllPlans). These are the intended command/query envelopes for future handler refactors; handlers currently pass arguments directly to service methods.
7. **Application-layer tests** — `internal/application/services/plan_service_test.go` uses a real temp-file `store.SQLiteStore` (migrations applied, cleaned up after each test); covers save/get round-trip, event emission on create (queried directly from the `outbox_events` table), delete, and not-found error.
8. **Test files updated** — `plan_test.go`, `projection_test.go`, `sqlite_test.go`, `auth_test.go` all updated for the new `NewPlan(name, month, year, ownerID)` signature (removed UUID arg), and `s.Save` call sites updated to validate first (`plan.Validate()` → `s.Save(validated)`).
9. **Full suite green** — `go build ./...` and `go test ./...` pass with no errors.

## Session Summary (2026-08-12): Wire Up Dead Commands/Queries, go-ddd Comparison

### What Prompted It
The commands/queries files from the 2026-08-07 session (item 6 above) had never actually been
wired to anything — a repo-wide grep found zero imports of `application/commands` or
`application/queries` outside those two files. Read the full 9-chapter
[`sklinkert/go-ddd`](https://github.com/sklinkert/go-ddd) tutorial (pinned at commit `08a11fc`)
and its actual `internal/application/{command,query,common,interfaces,mapper,services}` source to
ground what "wired up correctly" should mean, then compared it against this repo's domain and
application layers in full (three parallel Explore passes: domain layer, infrastructure/web
layer, existing docs/other services).

### What Shipped
See the new "Application Layer: go-ddd Comparison" section above (in Architectural Guidelines)
for the full reasoning — summary here:
1. **`commands/plan.go` rewritten and wired** — `CreatePlan`/`UpdatePlan`/`DeletePlan` (+ `*Result`
   wrappers) are now real, used types. Added `internal/application/common/plan_result.go`
   (`PlanResult`) and `internal/application/mapper/plan_result.go`. `PlanService` gained
   `CreatePlan`/`UpdatePlan`/`DeletePlan` methods — the only code now allowed to call
   `domain.NewPlan`/`Plan.ChangeCoreDetails`. `PlanController.PostSetup`/`PostUpdateSetup`/
   `PostDeletePlan` build commands instead of touching domain constructors/mutation methods
   directly, closing the actual layering violation (interface layer → domain layer, skipping the
   application service).
2. **`queries/plan.go` deleted**, not wired — `PlanService.Get`/`GetAll`/`GetUserPlans` already
   satisfy CQRS's read-side substance (no validation/construction/side effects) and every caller
   (report pages via the projection engine, every wizard controller) needs the full `*domain.Plan`
   aggregate, not a display DTO — go-ddd's thin-Result-DTO premise for reads doesn't hold for
   `Plan`, which is both the write aggregate and the read/calculation model. Documented the
   trigger condition for reintroducing a query layer if that ever changes.
3. **Scope decision**: extending this pattern to the other 9 domain services (payroll, sales
   forecast, cash flow, operating expenses, starting point, access, auth, invite,
   hub_completion) was explicitly deferred — they're thin CRUD pass-throughs over wizard
   sub-entities that don't have domain-level constructors/IDs yet (pre-existing gap, see Known
   Limitations), so wrapping them in command DTOs now would add ceremony without the
   invariant-guarding payoff the pattern exists for. Documented as a roadmap gated on that
   prerequisite.
4. **AGENTS.md corrected for drift**: layering table and Repository Rules referenced
   `internal/store/`/`internal/handlers/`, which no longer exist (now
   `internal/infrastructure/sqlite/`/`internal/interface/web/` post the
   `refactor/web-controllers` branch's per-domain-controller reorganization); Domain Events
   section claimed the outbox relay "is not yet implemented" when `outbox_relay.go` already
   exists and runs; the Project Structure Reference tree was rebuilt from the actual current
   tree (`find`-verified) rather than left stale.
5. **Tests added**: `TestPlanService_CreatePlan`, `TestPlanService_CreatePlan_RejectsInvalid`,
   `TestPlanService_UpdatePlan`, `TestPlanService_DeletePlan` in `plan_service_test.go`, using the
   existing temp-file SQLite fixture pattern.
6. **Verified end-to-end in a real browser** against a scratch SQLite DB (not the dev DB): signed
   up, created a plan via `POST /plan/setup` (redirect + Owner access granted, confirmed via
   `CreatePlan`), edited its core details via `POST /plan/{id}/setup` (confirmed persisted on
   reload, confirmed via `UpdatePlan`), deleted it via `POST /plan/{id}/delete` (confirmed gone
   from the list and soft-deleted in the DB). Queried `outbox_events` directly and confirmed
   `plan.created` then `plan.updated` were recorded in order.

### Deliberately Deferred
- The other 9 domain services' command/query wiring — see "Roadmap" in the new AGENTS.md section.
- Idempotency middleware — still schema-only; flagged as the natural next increment once/if
  `context.Context` is threaded through the application layer (needed for chapter 8's
  cancellation-safe cleanup path).

## Session Summary (2026-08-12, continued): Per-File Commands + VerifiedUser Ownership

### What Prompted It
Two follow-up asks: (1) go-ddd puts each command in its own file
(`create_product_command.go`/`update_product_command.go`/`delete_product_command.go`), not one
file per aggregate — `commands/plan.go` didn't match that. (2) The `Plan.NewPlan`/`CreatePlan`
owner parameter was a bare `uuid.UUID` with nothing confirming it actually named a real User —
go-ddd's `ValidatedSeller` pattern (chapter 4: `NewProduct` requires a `ValidatedSeller`, not a
`Seller`) has no analogue here for plan ownership.

### What Shipped
1. **Commands split one-per-file**: `commands/plan.go` → `create_plan_command.go` /
   `update_plan_command.go` / `delete_plan_command.go`, each holding its `XCommand` +
   `XCommandResult` pair, matching go-ddd's file layout exactly. `DeletePlanResult` also gained a
   `Success bool` field (go-ddd's `DeleteProductCommandResult` has one; ours was an empty struct).
2. **`entities.VerifiedUser`** added to `internal/domain/entities/user.go` — an opaque wrapper
   (`NewVerifiedUser(u *User) (VerifiedUser, error)`) that only accepts a real `*User`, mirroring
   `ValidatedPlan`'s "invalid states unrepresentable" shape. `Plan.NewPlan`'s owner parameter
   changed from `uuid.UUID` to `VerifiedUser`.
3. **`PlanService.CreatePlan` now verifies ownership for real**: added a `verifyOwner(ownerID)`
   helper that calls `UserRepository.GetUser` (returns `domain.ErrUserNotFound` if no such row
   exists) before constructing the plan. `NewPlanService` gained a second constructor parameter,
   `repositories.UserRepository` — updated its two call sites (`cmd/cli/serve.go`,
   `internal/interface/web/auth_test.go`).
4. **New test**: `TestPlanService_CreatePlan_RejectsUnknownOwner` — the actual proof the guarantee
   holds (a syntactically valid but never-persisted `uuid.UUID` is refused). All domain-layer call
   sites of `NewPlan` (`plan_test.go`, `projection_test.go`, `auth_test.go`) updated to build a
   `VerifiedUser` via a shared `newVerifiedOwner(t)` helper (in-memory `NewUser` + wrap — no I/O
   needed, since verification-the-type-check and verification-the-DB-lookup are different
   concerns living in different layers).
5. **Verified end-to-end in a real browser** against a fresh scratch DB: signed up, created a plan
   via `POST /plan/setup` (skipped clicking through the UI, drove it via `fetch` to exercise the
   real session → owner-lookup path faster), confirmed `plan_access` recorded the real user's id
   as owner and `outbox_events` recorded `user.registered` then `plan.created`.
6. **AGENTS.md updated**: new `VerifiedUser` subsection under "Application Layer: go-ddd
   Comparison", Project Structure Reference tree updated for the per-file command split and the
   `VerifiedUser` addition to `user.go`, Completed Features checklist updated.

### Deliberately Deferred
- `VerifiedUser` is scoped to plan ownership only. Other places a bare `uuid.UUID` is trusted as
  "this is a real user" without verification (e.g. `AccessService.GrantAccess`'s `userID` param,
  invite acceptance) were not touched — same reasoning as the broader roadmap above: extend the
  pattern deliberately, one call site at a time, when it's actually load-bearing, not everywhere
  a `uuid.UUID` appears.
