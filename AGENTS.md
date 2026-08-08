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
- [x] Validated-Entity pattern: `ValidatedPlan` opaque token; `Plan.Validate()` produces one; `PlanRepository.Save` requires it
- [x] Domain events: `internal/domain/entities/events/` package following go-ddd pattern — `base_event.go` (`BaseEvent` + `DomainEvent` interface), `plan_events.go` (`PlanCreated`/`PlanUpdated`/`PlanDeleted`), `invite_events.go` (`UserInvitedToPlan`), `user_events.go` (`UserRegistered`); **only aggregate roots emit events** — `Plan` and `User` carry event buffers; `PlanInvite` (a non-root entity) does not; `Plan.RecordUserInvited()` emits `UserInvitedToPlan` on behalf of the invite
- [x] Transactional outbox: `outbox_events` table (migration 48); `SQLiteStore.Save` (plan), `SQLiteStore.CreateInvite`, and `SQLiteStore.SaveUserWithPassword` each write their entity row + outbox events atomically in one transaction via the shared `writeOutboxEvents` helper
- [x] `processed_at` column on `outbox_events` (migration 48) — the outbox relay goroutine (not yet implemented) will set this when it delivers an event to the message broker; `idx_outbox_unprocessed` partial index filters `processed_at IS NULL` for efficient relay queries
- [x] Idempotency table: `idempotency_keys` table (migration 49) — schema in place; middleware wrapper not yet wired
- [x] CQRS formal types: `internal/application/commands/plan.go` (CreatePlan, UpdatePlan, DeletePlan) and `internal/application/queries/plan.go` (GetPlan, GetUserPlans, GetAllPlans)
- [x] In-memory fakes: `internal/application/fakes/plan_repository.go` implements `repositories.PlanRepository` for application-layer tests; includes `DrainEvents()` for event-emission assertions
- [x] Application-layer tests using fakes: `internal/application/services/plan_service_test.go` (save/get, event emission, delete, not-found)

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
│   ├── serve.go                 # Server setup: store init, template loading, route registration, middleware chain
│   └── migrate.go               # `migrate` command (runs the migration list via store init)
├── internal/
│   ├── domain/
│   │   ├── entities/            # All domain types (aggregates, value objects, events, projection)
│   │   │   ├── aggregate_root.go # AggregateRoot marker interface + aggregate boundary docs
│   │   │   ├── money.go         # Money value object (USD cents, int64-backed, JSON-compatible)
│   │   │   ├── errors.go        # ErrValidation sentinel
│   │   │   ├── plan.go          # Plan aggregate root, business logic, ValidatedPlan
│   │   │   ├── plan_*.go        # Plan companion files (payroll, cash flow, sales forecast, starting point, growth, json)
│   │   │   ├── user.go          # User aggregate root (id, email, first/last name, access levels)
│   │   │   ├── session.go       # Session domain model + NewUserWithPassword
│   │   │   ├── invite.go        # PlanInvite domain model (teams & collaboration)
│   │   │   ├── events.go        # DomainEvent interface + PlanCreated/PlanUpdated/PlanDeleted
│   │   │   ├── projection.go    # Month-by-month financial projection engine
│   │   │   └── *_test.go        # Domain and projection tests
│   │   └── repositories/        # Repository interfaces (one file per aggregate boundary)
│   │       ├── plan.go          # PlanRepository, AccessRepository
│   │       ├── user.go          # UserRepository, SessionRepository
│   │       ├── capital_asset.go # CapitalAssetRepository (Find/Get/Save/List/Delete)
│   │       ├── startup_cost.go  # StartupCostRepository
│   │       ├── funding_source.go
│   │       ├── salary_role.go
│   │       ├── benefit.go
│   │       ├── product.go
│   │       ├── inventory_purchase.go
│   │       ├── distribution.go
│   │       └── operating_expense.go
│   ├── application/
│   │   ├── interfaces/          # Application service interfaces
│   │   │   ├── plan_service.go
│   │   │   ├── auth_service.go
│   │   │   ├── access_service.go
│   │   │   ├── starting_point_service.go
│   │   │   ├── payroll_service.go
│   │   │   ├── sales_forecast_service.go
│   │   │   ├── cash_flow_service.go
│   │   │   └── operating_expenses_service.go
│   │   ├── commands/            # CQRS write-side command types
│   │   │   └── plan.go          # CreatePlan, UpdatePlan, DeletePlan
│   │   ├── queries/             # CQRS read-side query types
│   │   │   └── plan.go          # GetPlan, GetUserPlans, GetAllPlans
│   │   ├── fakes/               # In-memory fakes for application-layer tests
│   │   │   └── plan_repository.go # Implements PlanRepository; DrainEvents() for test assertions
│   │   └── services/            # Application service implementations (delegate to repositories)
│   │       ├── plan_service.go
│   │       ├── plan_service_test.go # App-layer tests (uses fakes, no DB)
│   │       ├── auth_service.go  # Session expiry check lives here (not in store)
│   │       ├── access_service.go # Access level validation lives here (not in store)
│   │       ├── starting_point_service.go
│   │       ├── payroll_service.go
│   │       ├── sales_forecast_service.go
│   │       ├── cash_flow_service.go
│   │       └── operating_expenses_service.go
│   ├── handlers/
│   │   ├── plan.go              # Plan HTTP handlers (setup, all report/form pages, delete)
│   │   ├── auth.go              # Authentication handlers (login, signup, logout, profile)
│   │   ├── invites.go           # Invite handlers (create/accept/reject)
│   │   ├── routes.go            # Route registration (RegisterRoutes method)
│   │   └── auth_test.go         # Authorization middleware tests (only handler test file so far)
│   ├── middleware/
│   │   └── logger.go            # Request logging middleware
│   ├── store/
│   │   ├── store.go             # PlanStore interface (plans, access, invites)
│   │   ├── sqlite.go            # SQLite implementation (plans, users, sessions, access, invites)
│   │   ├── sqlite_starting_point.go  # CapitalAsset, StartupCost, FundingSource, StartingBalances, wizard sections
│   │   ├── sqlite_payroll.go    # SalaryRole, Benefit, PayrollTaxRates
│   │   ├── sqlite_sales_forecast.go  # Product, SalesGrowthCurve
│   │   ├── sqlite_cash_flow.go  # InventoryPurchase, Distribution
│   │   ├── sqlite_operating_expenses.go
│   │   ├── migrations.go        # Ordered SQL migration list, run automatically on store init
│   │   └── sqlite_test.go       # Store tests
│   └── views/                   # Templates, static assets, and view-layer glue (merged into one package 2026-07-29)
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
├── Dockerfile / Dockerfile.dev   # Production (multi-stage, Alpine) and dev (air hot-reload) images
├── docker-compose.yml / docker-compose.dev.yml
├── .air.toml                    # air hot-reload config (used by Dockerfile.dev)
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
| Infrastructure | `internal/store/` | Domain interfaces (`internal/domain/repositories/`) |
| Interface | `internal/handlers/`, `internal/views/` | Application interfaces (`internal/application/interfaces/`) |

### CQRS Separation

- **Commands** (write path): `Post*` handlers mutate state, call `Store.Save`, enforce domain invariants, then redirect (POST–redirect–GET).
- **Queries** (read path): `Get*` handlers only read — never mutate. If a handler both reads and writes, split it.

### Repository Rules

1. **No business logic** in `internal/store/` — only translate between domain structs and SQL rows.
2. **Dependency inversion** — store implementations in `internal/store/` implement interfaces defined in `internal/domain/` (or alongside them).
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

Aggregates emit events via `recordEvent()`; repositories drain them with `PullEvents()` inside the same DB transaction that saves the entity. Event types live in `internal/domain/events.go`. The outbox relay goroutine that picks up unprocessed rows from `outbox_events` and publishes them to a broker is not yet implemented.

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

- **DDD / CQRS Architecture Implemented (2026-08-07)**: Added `internal/domain/repositories/` (repository interfaces), `internal/application/interfaces/` (service interfaces), and `internal/application/services/` (implementations). All wizard-item Delete methods now use soft-delete (`deleted_at`). Business logic (session expiry, access-level validation) moved from `internal/store/` into the service layer. `Find*` vs `Get*` naming convention enforced: nine methods that returned `nil, nil` on not-found renamed from `Get*Draft` to `Find*Draft` across all four layers. UUIDv7 adopted in all domain constructors. ValidatedPlan opaque token pattern applied to PlanRepository.Save. Domain events (`DomainEvent`, `PlanCreated`/`PlanUpdated`/`PlanDeleted`) in `internal/domain/events.go`. Transactional outbox via `outbox_events` table (migration 48). Idempotency schema via `idempotency_keys` table (migration 49). CQRS command/query types in `internal/application/commands/` and `internal/application/queries/`. In-memory fake `PlanRepository` in `internal/application/fakes/` with application-layer tests in `internal/application/services/plan_service_test.go`.
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
- **Wizard items have no domain-level ID** - `CapitalAsset`, `SalaryRole`, `Product`, etc. carry no `uuid.UUID` in the domain struct; the ID lives only in the `repositories.*Item` wrapper. These are entities within the Plan aggregate boundary and should eventually carry their own ID in the domain (so `Plan.futurePurchases []CapitalAsset` holds identifiable entities). Doing so requires updating all store serialization, repository interfaces, and handler code. Tracked in `repositories/wizard_item.go`'s package comment.

### Questions to Ask If Stuck

1. **"Should I add a new route?"** → Check if it's in the route list in main.go. If not, add both GET and POST if it's a form.
2. **"How do I save data?"** → Use `app.Store.Save(plan)` after mutating the plan. Currently in-memory; will need DB migration.
3. **"Where do I put validation?"** → In the domain model (plan.go domain package) for business rules. Use handlers for HTTP-level validation.
4. **"Should I create a new handler?"** → Create a method on the App struct that returns an http.HandlerFunc.

---

**Last Updated**: 2026-08-07 (continued)
**Session Focus**: DDD/CQRS architecture — UUIDv7, ValidatedPlan, domain events, outbox, idempotency schema, CQRS command/query types, in-memory fakes, application-layer tests
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
7. **In-memory fakes** — `internal/application/fakes/plan_repository.go` is a thread-safe `PlanRepository` backed by a `map`. Implements `Save(ValidatedPlan)`, `Get`, `GetAll`, `Delete`, `GetUserPlans`, and `DrainEvents()` for post-save event assertions.
8. **Application-layer tests** — `internal/application/services/plan_service_test.go` uses `fakes.PlanRepository`; covers save/get round-trip, event emission on create, delete, and not-found error. No DB required.
9. **Test files updated** — `plan_test.go`, `projection_test.go`, `sqlite_test.go`, `auth_test.go` all updated for the new `NewPlan(name, month, year, ownerID)` signature (removed UUID arg), and `s.Save` call sites updated to validate first (`plan.Validate()` → `s.Save(validated)`).
10. **Full suite green** — `go build ./...` and `go test ./...` pass with no errors.
