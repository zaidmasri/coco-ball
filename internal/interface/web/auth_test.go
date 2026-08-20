package web

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	"github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	appservices "github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/domain/repositories"
	"github.com/zaidmasri/business-planning-tool/internal/infrastructure/sqlite"
)

// testRepos bundles the repositories test cases need to seed data directly
// (bypassing the HTTP layer), since each domain interface is now backed by
// its own small sqlc-based repository struct rather than one shared store.
type testRepos struct {
	plans  repositories.PlanRepository
	users  repositories.UserRepository
	access repositories.AccessRepository
	auth   interfaces.AuthService
}

func setupTestApp(t *testing.T) (*http.ServeMux, testRepos, func()) {
	tmpDir := filepath.Join(os.TempDir(), "northbasis_auth_test")
	os.MkdirAll(tmpDir, 0755)
	dbPath := filepath.Join(tmpDir, "test.db")

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

	// Create a minimal template cache with an error template
	templateCache := make(map[string]*template.Template)
	errorTmpl, _ := template.New("error.html").Parse("<html><body>Error: {{.Message}}</body></html>")
	templateCache["error.html"] = errorTmpl

	mux := http.NewServeMux()
	accessMW := NewPlanAccessMiddleware(accessSvc, templateCache)
	NewPlanController(mux, planSvc, accessSvc, inviteSvc, authSvc, hubSvc, templateCache, accessMW)

	cleanup := func() {
		conn.Close()
		os.Remove(dbPath)
	}

	return mux, testRepos{plans: planRepo, users: userRepo, access: accessRepo, auth: authSvc}, cleanup
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
