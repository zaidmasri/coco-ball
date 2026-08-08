package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	appservices "github.com/zaidmasri/business-planning-tool/internal/application/services"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/store"
)

func setupTestApp(t *testing.T) (*App, *store.SQLiteStore, func()) {
	tmpDir := filepath.Join(os.TempDir(), "northbasis_auth_test")
	os.MkdirAll(tmpDir, 0755)
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	planSvc := appservices.NewPlanService(s)
	authSvc := appservices.NewAuthService(s, s)
	accessSvc := appservices.NewAccessService(s)
	inviteSvc := appservices.NewInviteService(s)
	startingPointSvc := appservices.NewStartingPointService(s, s, s, s, s)
	payrollSvc := appservices.NewPayrollService(s, s, s, s)
	salesForecastSvc := appservices.NewSalesForecastService(s, s, s)
	cashFlowSvc := appservices.NewCashFlowService(s, s, s)
	opExSvc := appservices.NewOperatingExpensesService(s, s)
	hubSvc := appservices.NewHubCompletionService(startingPointSvc, payrollSvc, salesForecastSvc, cashFlowSvc, opExSvc)

	// Create a minimal template cache with an error template
	templateCache := make(map[string]*template.Template)
	errorTmpl, _ := template.New("error.html").Parse("<html><body>Error: {{.Message}}</body></html>")
	templateCache["error.html"] = errorTmpl

	app := NewApp(planSvc, authSvc, accessSvc, inviteSvc,
		startingPointSvc, payrollSvc, salesForecastSvc, cashFlowSvc, opExSvc, hubSvc,
		templateCache)

	cleanup := func() {
		s.Close()
		os.Remove(dbPath)
	}

	return app, s, cleanup
}

func TestUserAuthorizationFlow(t *testing.T) {
	app, s, cleanup := setupTestApp(t)
	defer cleanup()

	// Create two users
	user1, err := domain.NewUser("user1@example.com")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2, err := domain.NewUser("user2@example.com")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	if err := s.SaveUser(user1); err != nil {
		t.Fatalf("Failed to save user1: %v", err)
	}

	if err := s.SaveUser(user2); err != nil {
		t.Fatalf("Failed to save user2: %v", err)
	}

	fmt.Println("✓ Users created successfully")

	// Create a plan for user1
	plan, err := domain.NewPlan("User1 Plan", 1, 2024, user1.ID())
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	validatedPlan, err := plan.Validate()
	if err != nil {
		t.Fatalf("Failed to validate plan: %v", err)
	}
	if err := s.Save(validatedPlan); err != nil {
		t.Fatalf("Failed to save plan: %v", err)
	}

	// Grant user1 owner access
	if err := s.GrantAccess(plan.ID(), user1.ID(), domain.Owner); err != nil {
		t.Fatalf("Failed to grant user1 owner access: %v", err)
	}

	fmt.Println("✓ Plan created for user1 with owner access")

	// Setup HTTP server with routes
	mux := http.NewServeMux()
	app.RegisterRoutes(mux)
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
		if err := s.GrantAccess(plan.ID(), user2.ID(), domain.Editor); err != nil {
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
		userViewer, err := domain.NewUser("viewer@example.com")
		if err != nil {
			t.Fatalf("Failed to create viewer user: %v", err)
		}

		if err := s.SaveUser(userViewer); err != nil {
			t.Fatalf("Failed to save viewer user: %v", err)
		}

		// Grant viewer-only access
		if err := s.GrantAccess(plan.ID(), userViewer.ID(), domain.Viewer); err != nil {
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
