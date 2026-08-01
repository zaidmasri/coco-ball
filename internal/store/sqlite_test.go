package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

func TestSQLiteStore(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "northbasis_test")
	os.MkdirAll(tmpDir, 0755)
	dbPath := filepath.Join(tmpDir, "test.db")
	defer os.Remove(dbPath)

	// Test SQLite store initialization
	fmt.Println("Testing SQLite store initialization...")
	s, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	defer s.Close()

	fmt.Println("✓ Database initialized successfully")

	// Test user creation
	fmt.Println("\nTesting user operations...")
	user, err := domain.NewUser("test@example.com")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if err := s.SaveUser(user); err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}
	fmt.Println("✓ User saved successfully")

	retrieved, err := s.GetUser(user.ID())
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}
	if retrieved.Email() != user.Email() {
		t.Fatalf("Email mismatch: %s != %s", retrieved.Email(), user.Email())
	}
	fmt.Println("✓ User retrieved successfully")

	// Test user with password
	fmt.Println("\nTesting user credentials...")
	userWithPwd, err := domain.NewUserWithPassword("auth@example.com", "Auth", "User", "securepassword123")
	if err != nil {
		t.Fatalf("Failed to create user with password: %v", err)
	}

	if err := s.SaveUserWithPassword(userWithPwd); err != nil {
		t.Fatalf("Failed to save user with password: %v", err)
	}
	fmt.Println("✓ User credentials saved successfully")

	retrievedCreds, err := s.GetUserWithPassword("auth@example.com")
	if err != nil {
		t.Fatalf("Failed to retrieve user credentials: %v", err)
	}
	if !retrievedCreds.VerifyPassword("securepassword123") {
		t.Fatalf("Password verification failed")
	}
	fmt.Println("✓ User credentials verified successfully")

	// Test plan creation and retrieval
	fmt.Println("\nTesting plan operations...")
	plan, err := domain.NewPlan(uuid.New(), "Test Plan", 1, 2025, user.ID())
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	if err := s.Save(plan); err != nil {
		t.Fatalf("Failed to save plan: %v", err)
	}
	fmt.Println("✓ Plan saved successfully")

	retrievedPlan, err := s.Get(plan.ID())
	if err != nil {
		t.Fatalf("Failed to retrieve plan: %v", err)
	}
	if retrievedPlan.ID() != plan.ID() {
		t.Fatalf("Plan ID mismatch")
	}
	fmt.Println("✓ Plan retrieved successfully")

	// Test access control
	fmt.Println("\nTesting access control...")
	if err := s.GrantAccess(plan.ID(), user.ID(), domain.Owner); err != nil {
		t.Fatalf("Failed to grant access: %v", err)
	}
	fmt.Println("✓ Access granted successfully")

	access, err := s.GetAccess(plan.ID(), user.ID())
	if err != nil {
		t.Fatalf("Failed to retrieve access: %v", err)
	}
	if access.AccessLevel != domain.Owner {
		t.Fatalf("Access level mismatch: %s != %s", access.AccessLevel, domain.Owner)
	}
	fmt.Println("✓ Access verified successfully")

	// Test session
	fmt.Println("\nTesting sessions...")
	session, err := domain.NewSession(user.ID(), 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if err := s.SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}
	fmt.Println("✓ Session saved successfully")

	retrievedSession, err := s.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}
	if retrievedSession.UserID != session.UserID {
		t.Fatalf("Session user ID mismatch")
	}
	fmt.Println("✓ Session retrieved successfully")

	fmt.Println("\n✅ All tests passed! SQLite store is working correctly.")
}
