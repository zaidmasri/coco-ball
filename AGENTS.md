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

### Financial Calculations & Reporting
- [ ] Implement payroll tax calculations
- [ ] Implement COGS calculations
- [ ] Implement depreciation calculations
- [ ] Implement cash flow projections (monthly)
- [ ] Implement income statement generation
- [ ] Implement balance sheet generation
- [ ] Implement financial ratios calculations
- [ ] Implement breakeven analysis
- [ ] Export functionality (PDF, Excel)

### Form Handlers (POST Endpoints)
- [ ] POST `/plan/{id}/payroll` - Save payroll data
- [ ] POST `/plan/{id}/sales-forecast` - Save sales data
- [ ] POST `/plan/{id}/operating-expenses` - Save OpEx data
- [ ] POST `/plan/{id}/cogs` - Save COGS data
- [ ] POST `/plan/{id}/cash-flow` - Save cash flow assumptions
- [ ] POST `/plan/{id}/income-statement` - Generate/save income statement
- [ ] POST `/plan/{id}/balance-sheet` - Generate/save balance sheet
- [ ] POST `/plan/{id}/delete` - Delete a plan (with authorization)

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
2. **Form POST handlers** - Implement POST endpoints for all plan data input (payroll, sales, expenses, etc.)
3. **Financial calculations** - Implement core projection calculations (cash flow, income statement, balance sheet)
4. **Testing** - Add test coverage for handlers and business logic
5. **Form Validation** - Client and server-side validation for all inputs
6. **Error handling** - Comprehensive error messages and recovery flows

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

### Known Limitations

- **No Multi-User Support Yet** - All plans are in a shared store; needs authentication and per-user isolation
- **No Calculations Generated** - Pages exist but don't perform actual financial projections
- **No Data Persistence** - Everything is lost when server restarts
- **No Export Functionality** - Can't export plans as Excel or PDF yet

### Questions to Ask If Stuck

1. **"Should I add a new route?"** → Check if it's in the route list in main.go. If not, add both GET and POST if it's a form.
2. **"How do I save data?"** → Use `app.Store.Save(plan)` after mutating the plan. Currently in-memory; will need DB migration.
3. **"Where do I put validation?"** → In the domain model (plan.go domain package) for business rules. Use handlers for HTTP-level validation.
4. **"Should I create a new handler?"** → Create a method on the App struct that returns an http.HandlerFunc.

---

**Last Updated**: 2026-07-29  
**Session Focus**: User Authorization enforcement, middleware application to routes, comprehensive testing  
**Total Remaining Items**: ~45+ (across all categories)  
**Critical Path**: ~~User Authorization~~ ✅ → Form POST Handlers → Financial Calculations → Testing

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
