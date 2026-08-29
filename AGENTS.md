# northbasis.com

## Project Overview

This project converts the **SCORE Financial Projections Template** (an Excel-based business
planning tool) into a web application. It helps entrepreneurs and small business owners build
detailed financial projections:

- **Starting Point** — initial balance sheet, fixed assets, startup costs, funding sources
- **Payroll Planning** — employee/contractor salary projections with payroll taxes and benefits
- **Sales Forecasting** — revenue stream projections with growth scenarios
- **Operating Expenses** — OpEx tracking with growth strategies
- **Cost of Goods Sold (COGS)** — product cost projections
- **Cash Flow Analysis** — monthly cash flow projections
- **Income Statement / Balance Sheet** — P&L and balance sheet projections
- **Financial Ratios & Breakeven Analysis** — key metrics and analysis tools

## Tech Stack

**Backend** — Go 1.27.0, standard library `net/http` with a custom router (Go 1.22+ routing),
SQLite (via `go-sqlite3`, cgo) with a hand-written migration runner, sqlc for generated query code.

**Frontend** — server-side rendering via Go `html/template`; static assets (CSS/JS, Bootstrap)
served from an embedded filesystem (`embed.FS`).

**Dependency philosophy** — standard library first. The domain layer (`internal/domain/`) has
**zero third-party imports**; UUIDs use the Go 1.27 standard-library `uuid` package (RFC 9562),
not `github.com/google/uuid` — see [Software Dependencies](#software-dependencies).

## Architecture: Onion / DDD Layering

Dependencies point **inward only**, toward the domain layer.

| Layer | Path | May depend on |
|---|---|---|
| Domain | `internal/domain/` | Nothing — no DB, no HTTP, no third-party libs (stdlib only) |
| Application | `internal/application/` | Domain only |
| Infrastructure | `internal/infrastructure/sqlite/` | Domain interfaces (`internal/domain/repositories/`) |
| Interface | `internal/interface/web/`, `internal/views/` | Application interfaces (`internal/application/interfaces/`) |

**CQRS separation:**
- **Commands** (write path): `Post*` handlers build a command DTO, call an application service,
  then redirect (POST–redirect–GET). The service method is the *only* code allowed to call the
  matching domain constructor/mutator.
- **Queries** (read path): `Get*` handlers only read, never mutate. A handler that does both should
  be split.

## Design Patterns in Use

- **Aggregate Root** — `Plan` and `User` are the two aggregate roots (`entities.AggregateRoot`
  marker interface, enforced by compile-time assertions `var _ AggregateRoot = (*Plan)(nil)`).
  Everything else (payroll, sales forecast, cash flow, starting point items, `PlanInvite`) is a
  sub-entity inside `Plan`'s aggregate boundary — no independent repository, no cross-aggregate FK,
  only ID references (`uuid.UUID`).
- **Value Object** — `entities.Money` (`int64` cents, never `float64`) is the model for all
  monetary fields.
- **Validated-Entity / opaque token** — a constructor that proves an invariant is enforced by the
  type system, not by caller discipline: `Plan.Validate()` returns a `ValidatedPlan` that
  `PlanRepository.Save` requires; `entities.VerifiedUser` (`NewVerifiedUser(*User)`) proves a
  `uuid.UUID` names a real, looked-up `User` before `Plan.NewPlan` will accept it as an owner.
  Apply this pattern when an invariant must hold at the aggregate/service boundary, not just at
  the call site that happens to be careful today.
- **Repository pattern with dependency inversion** — interfaces live in `internal/domain/repositories/`;
  implementations in `internal/infrastructure/sqlite/` translate rows ↔ domain structs and contain
  **no business logic**. `Find*` may return `nil, nil`; `Get*` must return a value or a descriptive
  error.
- **CQRS-lite (command/service/result)** — one `Create*`/`Update*`/`Delete*` command struct per
  write operation (`internal/application/commands/`), a `*Result` DTO returned via a mapper
  (`internal/application/common/`, `internal/application/mapper/`) — the entity itself never
  crosses into the interface layer. No command bus, no separate read store, no event sourcing —
  this deliberately mirrors [`sklinkert/go-ddd`](https://github.com/sklinkert/go-ddd)'s scope
  ("show the pattern, not the framework"). The read side stays as plain `Get*`/`GetAll*` methods
  returning the aggregate directly — that's a deliberate simplification, not a gap, because most
  read paths here (report pages, wizard edit forms) need the full aggregate's calculation methods,
  not a display DTO. Reach for a real query DTO only when a genuinely display-only read model shows
  up (e.g. a lightweight plan-summary list that doesn't need projections).
- **Domain Events + Transactional Outbox** — aggregates buffer events via `recordEvent()`
  (`internal/domain/events/`, one file per concern); a repository's write method drains them with
  `PullEvents()` and inserts them into `outbox_events` inside the **same transaction** as the
  aggregate row (`internal/infrastructure/sqlite/outbox.go`'s `insertOutboxEvents`). A background
  relay (`outbox_relay.go`) polls for unprocessed rows and publishes them (currently a log-only
  stub — swap in a real broker call when one exists).
- **Soft deletion** — every business-entity table has `deleted_at`; `Delete*` sets it, every
  `SELECT` filters `deleted_at IS NULL`. Hard delete is reserved for join tables (`plan_access`)
  and singleton rows.
- **Read-after-write via POST-redirect-GET** — repository write methods return only `error` (no
  read-after-write re-fetch inside the repository); the HTTP layer's POST-redirect-GET pattern
  satisfies the same guarantee at a higher level. This is a known, accepted simplification, not an
  oversight — see [Known Limitations](#known-limitations).

## Software Dependencies

- **Domain layer (`internal/domain/`): zero third-party imports.** If a domain rule needs
  something a library would normally provide (UUIDs, email format validation), find the standard
  library equivalent first (`uuid`, `net/mail`) before reaching for a package. This is a hard rule,
  not a guideline — a domain-layer PR that imports a third-party package should be rejected.
- **Application/infrastructure/interface layers**: prefer the standard library; add a dependency
  only when it removes real complexity (e.g. `go-sqlite3` for cgo SQLite bindings, `sqlc` as a
  build-time code generator, not a runtime dependency).
- **Before adding any new dependency**, ask: does the standard library already cover 90% of this?
  Will it need to be imported from the domain layer (if so, it's very likely disqualified)? Is it
  actively maintained? Prefer generation-time tools (like `sqlc`) over runtime frameworks.
- **Frontend**: Bootstrap is a static asset (CSS/JS served from the embedded FS), not a package
  dependency — there is no npm/node toolchain in this repo.

---

## Steps to Ship a Full Vertical Feature

This is the recipe this codebase actually follows, proven across `Plan`, Operating Expenses,
`User`, and `PlanInvite`. Follow it top-to-bottom for a new domain entity/feature; skip steps that
don't apply (e.g. a read-only report page skips the command/repository-write steps).

1. **Domain (`internal/domain/entities/`)**
   - Add or extend the entity with unexported fields and a `NewX(...) (X, error)` constructor that
     validates every invariant and generates a UUIDv7 ID (`uuid.NewV7()`) — never accept an ID from
     outside, never let an invalid value construct successfully.
   - Add validated update methods (`ChangeX(...)`) that validate before mutating — never expose a
     public setter that skips validation, except a `SetX` explicitly reserved for row
     *reconstruction* from the database (document it as such).
   - If this is a new aggregate root, implement the `AggregateRoot` marker and add the compile-time
     assertion. If it's a sub-entity, confirm which aggregate owns it — it does not get its own
     repository.
   - Route bounds-checking through the shared helpers in `entities/validation.go`
     (`validateRequiredName`, `validateMoneyAmount`, `validateGrowthRate`, `validatePercentRate`,
     `validateEmailFormat`) rather than hand-rolling checks per entity.
   - If the aggregate root should emit a domain event, add it to `internal/domain/events/` (one
     file per concern) and call `recordEvent()` from the constructor/mutator.
   - Write domain tests for the constructor and every invariant it enforces.

2. **Repository interface (`internal/domain/repositories/`)**
   - One file per aggregate/sub-entity boundary. Signatures use domain types only — no SQL, no
     driver types. Name read methods `Find*` (may return `nil, nil`) or `Get*` (must return a value
     or a descriptive error) per the existing convention.

3. **Migration + generated queries**
   - Add an ordered SQL file under `sql/migrations/` if the schema changes; migrations run
     automatically on store init and via `make migrate`.
   - Add hand-written SQL under `sql/queries/*.sql`, then regenerate `internal/infrastructure/db/sqlc/`
     (`sqlc generate`, per `sqlc.yaml`).

4. **Infrastructure (`internal/infrastructure/sqlite/`)**
   - Implement the repository interface: translate rows ↔ domain structs, **no business logic**.
   - If the write touches multiple tables and/or drains domain events, wrap it in one transaction
     (`tx.Begin()` / `Commit()`), following `outbox.go`'s `insertOutboxEvents` pattern for the
     event-drain step.
   - Soft-delete only; never issue a hard `DELETE` against a business-entity table.

5. **Application layer**
   - Command DTOs in `internal/application/commands/` — one file per `Create*`/`Update*`/`Delete*`
     operation, each with a `*Result` wrapper.
   - Result DTO in `internal/application/common/`, built by a mapper in `internal/application/mapper/`
     (entity → DTO — the entity itself must never reach the interface layer).
   - Service interface in `internal/application/interfaces/`; implementation in
     `internal/application/services/`. This service is the *only* code allowed to call the domain
     constructor/mutator you wrote in step 1.
   - Write an application-layer test against a real temp-file SQLite repository (this codebase does
     not mock repositories in these tests) covering the happy path and at least one validation
     rejection.

6. **Interface/web layer (`internal/interface/web/`)**
   - Add or extend a controller: `Get*` handlers read only; `Post*` handlers parse the form, build
     a command, call the service, and redirect on success (POST–redirect–GET).
   - Never call `domain.NewX` or a mutation method directly from a handler — that's the layering
     violation this recipe exists to prevent.
   - Apply `PlanAccessMiddleware` (Owner/Editor/Viewer) to any plan-scoped route.
   - Register the route(s).

7. **Views (`internal/views/`)**
   - Add a page-view struct in `types.go` and a `Build*Page` builder in `builders.go`.
   - Add or edit the template under `templates/pages/` (or a shared piece under `templates/components/`).
   - If the template needs formatting/arithmetic `html/template` doesn't provide, add a helper to
     `funcs.go` and register it in `cmd/cli/serve.go`'s `loadTemplates()`.

8. **Wire it up** — construct and inject any new repository/service in `cmd/cli/serve.go`.

9. **Verify**
   - `go build ./...`, `go vet ./...`, `go test ./...`.
   - Drive the feature end-to-end in a real browser against a scratch DB (sign up/log in, exercise
     the happy path and at least one validation failure) — tests verify correctness, not that the
     feature actually works through the UI. Use the Browser preview tools rather than asking the
     user to check manually.
   - Confirm migrations apply cleanly from empty (`make reset && make migrate`, or the Docker dev
     compose flow) if you added one.

10. **Update this document** — move the item from "Remaining Work" to "Completed Features," and add
    a one-line entry to the Tech Decisions Log if you made a non-obvious call.

---

## Project Structure Reference

```
.
├── cmd/cli/
│   ├── main.go                  # CLI entry point, flag parsing, command dispatch (serve/migrate/reset)
│   ├── serve.go                 # DB connection, template loading, controller wiring, middleware chain
│   └── migrate.go               # `migrate` command (runs the migration list)
├── internal/
│   ├── domain/
│   │   ├── entities/            # All domain types: aggregates, value objects, projection engine
│   │   │   ├── aggregate_root.go  # AggregateRoot marker interface + aggregate boundary docs
│   │   │   ├── money.go           # Money value object (USD cents, int64-backed)
│   │   │   ├── errors.go          # ErrValidation sentinel, IsUserFacing(err) classifier
│   │   │   ├── validation.go      # Shared bound-checking helpers used by every entity's validator
│   │   │   ├── plan.go            # Plan aggregate root; HubSections (hub → section membership)
│   │   │   ├── plan_growth.go     # Cost (Operating Expenses/COGS line item)
│   │   │   ├── plan_*.go          # Other Plan companion files (payroll, cash flow, sales forecast,
│   │   │   │                        starting point, json)
│   │   │   ├── user.go            # User aggregate root + VerifiedUser opaque token
│   │   │   ├── session.go         # Session domain model
│   │   │   ├── invite.go          # PlanInvite domain model
│   │   │   ├── projection.go      # Month-by-month financial projection engine (see its file-level
│   │   │   │                        comment for every modeling simplification)
│   │   │   └── *_test.go
│   │   ├── events/                # DomainEvent interface + concrete events, one file per concern
│   │   └── repositories/          # Repository interfaces, one file per aggregate/sub-entity boundary
│   ├── application/
│   │   ├── interfaces/            # One service interface per domain hub
│   │   ├── commands/               # Write-side command DTOs, one file per Create/Update/Delete
│   │   ├── common/                 # Result DTOs shared between command/query sides
│   │   ├── mapper/                 # Entity → Result DTO conversion
│   │   └── services/                # Service implementations — the only callers of domain New*/mutators
│   ├── interface/web/             # HTTP controllers, one struct per domain, registers its own routes
│   │   ├── plan.go, auth.go, invites.go, payroll.go, cash_flow.go, sales_forecast.go,
│   │   │   starting_point.go, operating_expenses.go
│   │   ├── access_middleware.go   # PlanAccessMiddleware: route-guards by Owner/Editor/Viewer
│   │   ├── recover_middleware.go  # Panic-recovery middleware (outermost layer)
│   │   └── errors.go              # renderCommandError/renderInternalError/safeErrorMessage
│   ├── middleware/                # Cross-cutting HTTP middleware (request logger, etc.)
│   ├── infrastructure/
│   │   ├── sqlite/                 # Repository implementations, one file per aggregate/sub-entity
│   │   │   ├── connection.go, migrate.go
│   │   │   ├── outbox.go           # insertOutboxEvents — drains events in the same tx as the save
│   │   │   └── outbox_relay.go     # Polls outbox_events, publishes (log-only stub), marks processed
│   │   ├── db/sqlc/                 # sqlc-generated query code (compiled from sql/queries/*.sql)
│   │   └── config/config.go         # CLI flag / env config
│   └── views/
│       ├── builders.go            # Page view builders (Build*Page functions)
│       ├── types.go               # View data types (per-page structs)
│       ├── funcs.go               # html/template helper funcs (formatMoney, addMoney, ...)
│       ├── templates_embed.go / static_embed.go
│       └── templates/{base.html, pages/, components/}, static/
├── sql/
│   ├── queries/                   # Hand-written SQL that sqlc compiles into internal/infrastructure/db/sqlc
│   └── migrations/                # Ordered SQL migration files
├── Dockerfile / Dockerfile.dev, docker-compose.yml / docker-compose.dev.yml, .air.toml
├── sqlc.yaml, Makefile, go.mod, go.sum
```

## Key Files to Understand

- [plan.go](internal/domain/entities/plan.go) (domain) — core business logic, validation, aggregate methods
- [projection.go](internal/domain/entities/projection.go) (domain) — financial projection engine; read its
  file-level comment before touching any calculation
- [serve.go](cmd/cli/serve.go) — route wiring and server setup
- [plan.go](internal/interface/web/plan.go) (interface/web) — plan lifecycle + report page handlers
- [invites.go](internal/interface/web/invites.go) — collaboration/invite flow
- `internal/views/templates/` — UI layer

---

## Current Status

### Completed

- **Core infrastructure** — project layout, HTTP server with timeouts/middleware, embedded static
  FS, 404/error pages, panic-recovery middleware, template caching pipeline, route registration
  encapsulation.
- **Domain model** — `Plan`/`User` aggregate roots; revenue streams, OpEx, COGS, capital-asset
  depreciation (straight-line, double-declining), financing terms, startup costs, starting balance
  sheet; comprehensive validation.
- **Data storage** — SQLite (replacing the original in-memory store) with an ordered migration
  system (`internal/store/migrations.go`, migrations run automatically on init); every wizard
  domain persists (Starting Point, Payroll, Sales Forecast/Products, Operating Expenses, Cash Flow).
- **Auth & authorization** — login/logout/signup, password hashing, session middleware, profile
  view/edit (first/last name), self-service account deletion, Owner/Editor/Viewer per-plan access
  control enforced by route middleware.
- **DDD/CQRS architecture** (2026-08-07, remediated 2026-08-19) — see
  [Architecture](#architecture-onion--dds-layering) and [Design Patterns](#design-patterns-in-use)
  above for the current, settled shape. A same-day full-application audit (application services,
  infrastructure, interface/web) found and fixed 10 concrete layering violations — invite-creation
  atomicity, plan-setup atomicity, missing wizard-hub validation gates, a freely-mutable `Cost`,
  existence checks living in the wrong layer, hub-section membership duplicated in two places, dead
  repository methods, and construction bypasses for `PlanInvite`/`Session`. Command-pattern coverage
  now spans `Plan`, Operating Expenses, `User`, and `PlanInvite`; the other wizard hubs (`payroll`,
  `sales_forecast`, `cash_flow`, `starting_point`) remain thin CRUD pass-throughs pending
  domain-level IDs for their sub-entities (see [Known Limitations](#known-limitations)).
- **Teams & collaboration** — per-plan email invites at Owner/Editor/Viewer level; pending-invite
  inbox on the home dashboard; accept/reject flow; no outbound email (in-app only).
- **Routes & pages** — home, login/signup, profile (+edit/delete), plan setup, all 7 wizard hub
  pages, income statement, balance sheet, analytics — all with matching `Post*` handlers where
  applicable.
- **Financial calculations** — payroll taxes (SS/Medicare/FUTA/SUTA, contractors exempt),
  product-based COGS, depreciation, monthly cash flow, income statement, balance sheet, financial
  ratios (current/quick/debt-to-equity/DSCR/growth/margins/ROE), Year-1 breakeven analysis.
- **Testing** — domain/projection unit tests; application-layer tests against real temp-file SQLite
  repositories (no mocks); handler-level tests in `internal/interface/web/` for report pages, the
  invite flow, and all four wizard domains' POST handlers.
- **Form validation** — domain-layer bounds (name length, money magnitude, growth/tax rate ranges,
  email format) across every wizard domain including the three singleton sections that previously
  had none; matching client-side HTML5 attributes.
- **Error handling** (2026-08-28) — panic-recovery middleware renders the shared error page instead
  of killing the connection; every `http.Error`/leaked-`err.Error()` call site in
  `internal/interface/web/` now routes through `renderCommandError`/`renderInternalError`, which
  uses `entities.IsUserFacing(err)` to show real validation messages while keeping infrastructure
  errors (SQL, etc.) off the response.
- **DevOps** — multi-stage production `Dockerfile` (Alpine, cgo) + `Dockerfile.dev` (air hot
  reload); `docker-compose.yml`/`docker-compose.dev.yml`; `Makefile` with local-Go and Docker
  targets; migrations runnable via CLI or automatically on store init.

### Remaining

**Auth & security**
- [ ] Password reset, email verification, optional 2FA
- [ ] CSRF protection
- [ ] Rate limiting / DDoS protection
- [ ] Dedicated input-sanitization layer (currently relying on `html/template`'s auto-escaping)
- [ ] Security headers (CSP, X-Frame-Options, etc.), HTTPS enforcement

**User management**
- [ ] User preferences/settings
- [ ] Workspace/organization grouping above per-plan invites
- [ ] Audit logging for plan changes

**Database**
- [ ] Postgres/MySQL option for horizontal scaling (SQLite is fine for the current single-instance
  deployment)
- [ ] Plan versioning/history (optional)

**Financial calculations**
- [ ] Income tax modeling (currently deliberately unmodeled — Net Income is pre-tax)
- [ ] Server-side export (PDF/Excel) — currently browser print-to-PDF only
- [ ] `POST /plan/{id}/income-statement` and `/balance-sheet` (blocked on tax modeling)

**API & integration**
- [ ] JSON API, Excel import/export, OpenAPI docs, CORS (all optional/future)

**Frontend polish**
- [ ] Responsive design pass, loading states, success/error toasts, help text/tooltips,
  print-friendly views, dark mode

**Testing**
- [ ] End-to-end browser tests spanning multiple pages
- [ ] Shared test fixtures/factories (each test still constructs users/plans inline)

**DevOps**
- [ ] `.env`/secrets management (currently CLI flags only: `--db`, `--port`)
- [ ] CI/CD pipeline
- [ ] Logging/monitoring and error tracking beyond the request-logger middleware

**Documentation**
- [ ] ADRs, API docs, deployment guide, developer setup guide, user guide

## High-Priority / Next Up

The critical path through Authorization → Form Handlers → Financial Calculations → Data
Persistence → Docker → DDD/CQRS → DDD Remediation → Testing → Form Validation → Error Handling is
complete. **Income tax modeling** is next: there is no tax-rate input anywhere in the app, so Net
Income is currently pre-tax everywhere it's shown.

## Known Limitations

- **No income tax modeling** — Income Tax Expense always shows $0.
- **No export functionality** — browser print-to-PDF only.
- **Money truncates cents** — `Money` is `int64`; form amounts are parsed as `float64` then
  truncated. Over a 36-month projection this can drift the balance sheet by a few dollars; the
  report pages tolerate up to $25 before flagging "does not balance."
- **Current/quick ratio show 0.00 when undefined** — no AP/accrued expenses means $0 current
  liabilities, which is mathematically undefined, not distinguishable from an actual 0.0 today.
- **Loan/fixed-asset categories aren't sub-typed** — one combined "Loans Payable" and "Fixed
  Assets" figure; the domain model doesn't track finer categories.
- **No email delivery** — plan invites are only visible in-app to a logged-in user with a matching
  email.
- **Collaboration is per-plan, not org-wide** — no workspace/team entity above individual plans.
- **SQLite only** — single file, fine for the current single-instance deployment; no pooling or
  horizontal-scaling story.
- **Handler test coverage is representative, not exhaustive** — one repeatable sub-section per
  wizard domain is covered end-to-end; siblings (Benefits, Payroll Tax Rates, Sales Growth Curve,
  Distributions) follow the identical pattern but aren't independently tested.
- **Most wizard sub-entities have no domain-level ID** — `CapitalAsset`, `SalaryRole`, `Product`,
  etc. carry no `uuid.UUID` in the domain struct; the ID lives only in the repository wrapper.
  `Cost` (Operating Expenses/COGS) is the one exception (`NewCost` + `SetID` reconstruction hook).
  This is the actual prerequisite blocking the command pattern from extending to the remaining
  wizard hubs — see step 1 of the [vertical feature recipe](#steps-to-ship-a-full-vertical-feature).
- **Read-after-write is HTTP-level, not repository-level** — see
  [Design Patterns](#design-patterns-in-use).
- **Form validation deliberately leaves some cross-field/business-rule checks open** — no upper
  bound on headcount, no `price >= cost` check (a loss-leader is a legitimate scenario), no
  duplicate-name checking within a wizard section. These are product decisions, not gaps.

## Tech Decisions Log

Durable decisions worth knowing before you touch related code. (Full narrative history lives in
git log / commit messages, not here.)

- **SQLite over in-memory** — switched early for persistence; migration system runs automatically
  on store init and via `make migrate`.
- **Server-side rendering over a frontend framework** — `html/template`, chosen for MVP speed; can
  be replaced with an API + frontend framework later if needed.
- **Session-based auth over JWT** — cookie sessions, simpler server-side state for this stage.
- **Standard-library `uuid` over `github.com/google/uuid`** (2026-08-20, after the Go 1.27 upgrade)
  — RFC 9562, `UUID` is still `[16]byte`. API differences from the third-party package:
  `uuid.NewV7()` no longer returns an `error`; `uuid.Nil` is a function, not a value; there's no
  `uuid.NewString()` (use `uuid.New().String()`).
- **`users.email` uniqueness is a partial index, not a column constraint** — soft-deleting users
  exposed a bug where a deleted account permanently squatted on its email. Fixed via
  `CREATE UNIQUE INDEX idx_users_email_active ON users(email) WHERE deleted_at IS NULL`.
- **No `context.Context` in the application layer** — nothing in `application`/`domain` uses it
  today; adding it selectively would be an inconsistency. Revisit once idempotency (below) is
  wired, since a cleanup-on-failure path needs `context.WithoutCancel`.
- **Idempotency table exists, middleware doesn't yet** — `idempotency_keys` (migration 49) is the
  intended backing store for an `Idempotency-Key` header middleware (`INSERT … ON CONFLICT DO
  NOTHING` to reserve atomically before the command runs). Not wired yet.
- **Financial projection engine simplifications** (`internal/domain/entities/projection.go`, see its
  file-level comment for the authoritative list) — no income tax; a funding source is debt only if
  it has both a positive interest rate and term, otherwise it's treated as a one-time equity
  contribution; AR/Prepaid/AP/Accrued Expenses are held static at their Starting Point values;
  "Additional Inventory" cash-flow purchases are an asset-for-asset swap with no depletion modeled;
  Starting Point balances fold into opening Retained Earnings as implied pre-existing equity. The
  balance sheet is guaranteed to balance by construction, up to whole-dollar rounding drift.
- **Docker split into prod/dev images** — production `Dockerfile` is multi-stage Alpine (cgo,
  `CGO_CFLAGS=-D_LARGEFILE64_SOURCE` for musl+sqlite compatibility); `Dockerfile.dev` runs `air` for
  hot reload against a bind-mounted source tree. Comments in the Makefile indicate the production
  image is intended to deploy via Coolify; live-deployment status isn't tracked in this repo.

## Questions to Ask If Stuck

1. **Should I add a new route?** Check `cmd/cli/serve.go`'s route registration. Add both GET and
   POST if it's a form.
2. **How do I save data?** Mutate the aggregate, then call the matching application service's
   save/command method — never write to the store directly from a handler.
3. **Where does validation go?** Business rules live in the domain layer (constructors/`Change*`
   methods). HTTP-level parsing/format checks live in the handler.
4. **Should I create a new handler?** Add a method on the relevant controller struct in
   `internal/interface/web/` and register its route.
5. **Am I about to violate layering?** If a handler is about to call `domain.NewX` or a mutation
   method directly, stop — that belongs in an application service. See
   [Steps to Ship a Full Vertical Feature](#steps-to-ship-a-full-vertical-feature).

---

## Maintaining This Document

Keep this file current as you work:

1. Move completed items from **Remaining** to **Completed** in [Current Status](#current-status).
2. Add newly discovered work to **Remaining**.
3. Update **High-Priority / Next Up** if priorities shift.
4. Log a decision in the [Tech Decisions Log](#tech-decisions-log) only if it's non-obvious and
   durable — not a play-by-play of the session. Session-by-session narrative belongs in commit
   messages and PR descriptions, not here.

**Last updated**: 2026-08-28 — cleaned up and reorganized; added the "Steps to Ship a Full Vertical
Feature" recipe; condensed ~900 lines of chronological session-log narrative (now recoverable from
git history) into the Tech Decisions Log and Current Status sections above.
