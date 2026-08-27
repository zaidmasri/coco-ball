package entities

import (
	"errors"
	"strings"
	"testing"
)

func TestUserCredentials_Validate(t *testing.T) {
	tests := []struct {
		testName    string
		email       string
		password    string
		expectedErr error
	}{
		{testName: "Valid credentials", email: "jane@example.com", password: "supersecret1", expectedErr: nil},
		{testName: "Fails on empty email", email: "", password: "supersecret1", expectedErr: ErrInvalidEmail},
		{testName: "Fails on malformed email", email: "not-an-email", password: "supersecret1", expectedErr: ErrInvalidEmailFormat},
		{testName: "Fails on empty password", email: "jane@example.com", password: "", expectedErr: ErrInvalidPassword},
		{testName: "Fails on short password", email: "jane@example.com", password: "short", expectedErr: ErrWeakPassword},
		{testName: "Fails on over-length password", email: "jane@example.com", password: strings.Repeat("a", 73), expectedErr: ErrPasswordTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			creds := &UserCredentials{Email: tt.email, Password: tt.password}
			if err := creds.Validate(); !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestNewUserWithPassword_RejectsOverLengthName(t *testing.T) {
	_, err := NewUserWithPassword("jane@example.com", strings.Repeat("a", maxNameLength+1), "Doe", "supersecret1")
	if !errors.Is(err, ErrNameTooLong) {
		t.Errorf("expected ErrNameTooLong, got %v", err)
	}
}

func TestHashPassword_RejectsOverLengthPassword(t *testing.T) {
	if _, err := HashPassword(strings.Repeat("a", 73)); !errors.Is(err, ErrPasswordTooLong) {
		t.Errorf("expected ErrPasswordTooLong, got %v", err)
	}
}
