package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	appservices "github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

// testRepos bundles the repositories and services test cases need to seed
// data directly (bypassing the HTTP layer) or verify persisted results,
// since each domain interface is now backed by its own small sqlc-based
// repository struct rather than one shared store.
type testRepos struct {
	plans             repositories.PlanRepository
	users             repositories.UserRepository
	access            repositories.AccessRepository
	invites           repositories.InviteRepository
	auth              interfaces.AuthService
	salaryRoles       repositories.SalaryRoleRepository
	benefits          repositories.BenefitRepository
	products          repositories.ProductRepository
	inventoryPurchase repositories.InventoryPurchaseRepository
	distributions     repositories.DistributionRepository
	operatingExpenses repositories.OperatingExpenseRepository
}

// setupTestApp wires every controller the same way cmd/cli/serve.go does,
// against a fresh temp-file SQLite database and the real embedded page
// templates, so handler tests exercise the full request path (routing,
// access middleware, and template rendering) rather than a stub.
func setupTestApp(t *testing.T) (*http.ServeMux, testRepos, func()) {
	tmpDir := filepath.Join(os.TempDir(), "northbasis_web_test")
	os.MkdirAll(tmpDir, 0755)
	dbFile, err := os.CreateTemp(tmpDir, "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp db file: %v", err)
	}
	dbPath := dbFile.Name()
	dbFile.Close()

	conn, err := sqlite.NewConnection(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	if err := sqlite.RunMigrations(conn); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	planRepo := sqlite.NewPlanRepository(conn)
	userRepo := sqlite.NewUserRepository(conn)
	sessionRepo := sqlite.NewSessionRepository(conn)
	accessRepo := sqlite.NewAccessRepository(conn)
	inviteRepo := sqlite.NewInviteRepository(conn)
	wizardProgressRepo := sqlite.NewWizardProgressRepository(conn)
	capitalAssetRepo := sqlite.NewCapitalAssetRepository(conn)
	startupCostRepo := sqlite.NewStartupCostRepository(conn)
	fundingSourceRepo := sqlite.NewFundingSourceRepository(conn)
	startingBalancesRepo := sqlite.NewStartingBalancesRepository(conn)
	salaryRoleRepo := sqlite.NewSalaryRoleRepository(conn)
	benefitRepo := sqlite.NewBenefitRepository(conn)
	payrollTaxRatesRepo := sqlite.NewPayrollTaxRatesRepository(conn)
	productRepo := sqlite.NewProductRepository(conn)
	salesGrowthCurveRepo := sqlite.NewSalesGrowthCurveRepository(conn)
	inventoryPurchaseRepo := sqlite.NewInventoryPurchaseRepository(conn)
	distributionRepo := sqlite.NewDistributionRepository(conn)
	operatingExpenseRepo := sqlite.NewOperatingExpenseRepository(conn)

	planSvc := appservices.NewPlanService(planRepo, userRepo)
	authSvc := appservices.NewAuthService(userRepo, sessionRepo)
	accessSvc := appservices.NewAccessService(accessRepo, planRepo, userRepo)
	inviteSvc := appservices.NewInviteService(inviteRepo, planRepo)
	startingPointSvc := appservices.NewStartingPointService(capitalAssetRepo, startupCostRepo, fundingSourceRepo, startingBalancesRepo, wizardProgressRepo)
	payrollSvc := appservices.NewPayrollService(salaryRoleRepo, benefitRepo, payrollTaxRatesRepo, wizardProgressRepo)
	salesForecastSvc := appservices.NewSalesForecastService(productRepo, salesGrowthCurveRepo, wizardProgressRepo)
	cashFlowSvc := appservices.NewCashFlowService(inventoryPurchaseRepo, distributionRepo, wizardProgressRepo)
	opExSvc := appservices.NewOperatingExpensesService(operatingExpenseRepo, wizardProgressRepo)
	hubSvc := appservices.NewHubCompletionService(startingPointSvc, payrollSvc, salesForecastSvc, cashFlowSvc, opExSvc)

	templateCache := views.LoadTemplates()

	mux := http.NewServeMux()
	accessMW := NewPlanAccessMiddleware(accessSvc, templateCache)
	NewAuthController(mux, authSvc, templateCache)
	NewPlanController(mux, planSvc, accessSvc, inviteSvc, authSvc, hubSvc, templateCache, accessMW)
	NewInviteController(mux, inviteSvc, planSvc, accessSvc, hubSvc, templateCache, accessMW)
	NewStartingPointController(mux, planSvc, startingPointSvc, templateCache, accessMW)
	NewPayrollController(mux, planSvc, payrollSvc, templateCache, accessMW)
	NewSalesForecastController(mux, planSvc, salesForecastSvc, templateCache, accessMW)
	NewCashFlowController(mux, planSvc, cashFlowSvc, templateCache, accessMW)
	NewOperatingExpensesController(mux, planSvc, opExSvc, templateCache, accessMW)

	cleanup := func() {
		conn.Close()
		os.Remove(dbPath)
	}

	return mux, testRepos{
		plans:             planRepo,
		users:             userRepo,
		access:            accessRepo,
		invites:           inviteRepo,
		auth:              authSvc,
		salaryRoles:       salaryRoleRepo,
		benefits:          benefitRepo,
		products:          productRepo,
		inventoryPurchase: inventoryPurchaseRepo,
		distributions:     distributionRepo,
		operatingExpenses: operatingExpenseRepo,
	}, cleanup
}

// createTestPlan creates and persists a plan owned by owner, grants role
// access to owner, and returns the validated plan. Tests that need a plan
// seeded with wizard data should save it through the appropriate service
// after calling this.
func createTestPlan(t *testing.T, s testRepos, owner *domain.User, role domain.AccessLevel) *domain.Plan {
	t.Helper()
	verifiedOwner, err := domain.NewVerifiedUser(owner)
	if err != nil {
		t.Fatalf("Failed to verify owner: %v", err)
	}
	plan, err := domain.NewPlan("Test Plan", 1, 2024, verifiedOwner)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}
	validatedPlan, err := plan.Validate()
	if err != nil {
		t.Fatalf("Failed to validate plan: %v", err)
	}
	if err := s.plans.Save(validatedPlan); err != nil {
		t.Fatalf("Failed to save plan: %v", err)
	}
	if err := s.access.GrantAccess(plan.ID(), owner.ID(), role); err != nil {
		t.Fatalf("Failed to grant %s access: %v", role, err)
	}
	return plan
}

// createTestUser registers a user through AuthService.CreateUser — the only
// validated entry point for constructing a User — and re-fetches the
// persisted entity via UserRepository.GetUser, for tests that need a real
// *domain.User (e.g. for WithUserContext or domain.NewVerifiedUser).
func createTestUser(t *testing.T, s testRepos, email string) *domain.User {
	t.Helper()
	result, err := s.auth.CreateUser(&commands.CreateUser{
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
		Password:  "supersecret1",
	})
	if err != nil {
		t.Fatalf("Failed to create user %s: %v", email, err)
	}
	user, err := s.users.GetUser(result.Result.ID)
	if err != nil {
		t.Fatalf("Failed to fetch created user %s: %v", email, err)
	}
	return user
}

// TestPostSignup_Validation exercises the new domain-level email-format and
// password-length checks through the full HTTP signup path. PostSignup
// re-renders the signup form with an inline error on every rejection
// (matching its existing "email already registered" behavior) rather than
// returning a 4xx status, so these assert on the rendered error message.
func TestPostSignup_Validation(t *testing.T) {
	mux, _, cleanup := setupTestApp(t)
	defer cleanup()

	tests := []struct {
		name     string
		form     url.Values
		wantText string
	}{
		{
			name: "malformed email",
			form: url.Values{
				"email": {"not-an-email"}, "firstName": {"Jane"}, "lastName": {"Doe"},
				"password": {"supersecret1"}, "confirmPassword": {"supersecret1"},
			},
			wantText: "is not valid",
		},
		{
			name: "weak password",
			form: url.Values{
				"email": {"jane@example.com"}, "firstName": {"Jane"}, "lastName": {"Doe"},
				"password": {"short"}, "confirmPassword": {"short"},
			},
			wantText: "8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postForm(mux, "/signup", nil, tt.form)
			if w.Code != http.StatusOK {
				t.Fatalf("expected 200 (form re-rendered with error), got %d: %s", w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), tt.wantText) {
				t.Errorf("expected response body to contain %q, got: %s", tt.wantText, w.Body.String())
			}
		})
	}
}

func TestUserAuthorizationFlow(t *testing.T) {
	mux, s, cleanup := setupTestApp(t)
	defer cleanup()

	// Create two users
	user1 := createTestUser(t, s, "user1@example.com")
	user2 := createTestUser(t, s, "user2@example.com")

	fmt.Println("✓ Users created successfully")

	// Create a plan for user1
	verifiedOwner, err := domain.NewVerifiedUser(user1)
	if err != nil {
		t.Fatalf("Failed to verify user1: %v", err)
	}
	plan, err := domain.NewPlan("User1 Plan", 1, 2024, verifiedOwner)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	validatedPlan, err := plan.Validate()
	if err != nil {
		t.Fatalf("Failed to validate plan: %v", err)
	}
	if err := s.plans.Save(validatedPlan); err != nil {
		t.Fatalf("Failed to save plan: %v", err)
	}

	// Grant user1 owner access
	if err := s.access.GrantAccess(plan.ID(), user1.ID(), domain.Owner); err != nil {
		t.Fatalf("Failed to grant user1 owner access: %v", err)
	}

	fmt.Println("✓ Plan created for user1 with owner access")

	// Setup HTTP server with routes
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test 1: User1 can access their own plan
	t.Run("User1 can access their own plan", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/plan/"+plan.ID().String()+"/setup", nil)
		req = WithUserContext(req, user1)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		// Should succeed or get template error (which is expected, we didn't set up templates)
		if w.Code == http.StatusForbidden || w.Code == http.StatusUnauthorized {
			t.Errorf("Expected successful access, got %d", w.Code)
		}
	})

	fmt.Println("✓ User1 successfully accessed their plan")

	// Test 2: User2 cannot access user1's plan
	t.Run("User2 cannot access user1's plan", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/plan/"+plan.ID().String()+"/setup", nil)
		req = WithUserContext(req, user2)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})

	fmt.Println("✓ User2 correctly denied access to user1's plan")

	// Test 3: Unauthenticated user cannot access the plan
	t.Run("Unauthenticated user cannot access plan", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/plan/"+plan.ID().String()+"/setup", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", w.Code)
		}
	})

	fmt.Println("✓ Unauthenticated user correctly denied access")

	// Test 4: User2 with Editor access can edit
	t.Run("User with Editor access can edit", func(t *testing.T) {
		// Grant user2 editor access
		if err := s.access.GrantAccess(plan.ID(), user2.ID(), domain.Editor); err != nil {
			t.Fatalf("Failed to grant user2 editor access: %v", err)
		}

		req := httptest.NewRequest("POST", "/plan/"+plan.ID().String()+"/setup", nil)
		req = WithUserContext(req, user2)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		// Should not be forbidden - 400 is expected due to missing form data
		if w.Code == http.StatusForbidden || w.Code == http.StatusUnauthorized {
			t.Errorf("Expected successful authorization, got %d", w.Code)
		}
	})

	fmt.Println("✓ User with Editor access can successfully edit")

	// Test 5: User with Viewer access cannot edit
	t.Run("User with Viewer access cannot edit", func(t *testing.T) {
		userViewer := createTestUser(t, s, "viewer@example.com")

		// Grant viewer-only access
		if err := s.access.GrantAccess(plan.ID(), userViewer.ID(), domain.Viewer); err != nil {
			t.Fatalf("Failed to grant viewer access: %v", err)
		}

		req := httptest.NewRequest("POST", "/plan/"+plan.ID().String()+"/setup", nil)
		req = WithUserContext(req, userViewer)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403, got %d", w.Code)
		}
	})

	fmt.Println("✓ User with Viewer access correctly denied edit access")

	fmt.Println("\n✅ All authorization tests passed!")
}
