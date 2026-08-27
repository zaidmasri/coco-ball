package entities

import (
	"strings"
	"testing"
)

func TestUser_ChangeName(t *testing.T) {
	tests := []struct {
		testName    string
		firstName   string
		lastName    string
		expectedErr error
	}{
		{testName: "Valid name change", firstName: "Jane", lastName: "Doe", expectedErr: nil},
		{testName: "Fails on empty first name", firstName: "  ", lastName: "Doe", expectedErr: ErrInvalidUserName},
		{testName: "Fails on empty last name", firstName: "Jane", lastName: "  ", expectedErr: ErrInvalidUserName},
		{testName: "Fails on over-length first name", firstName: strings.Repeat("a", maxNameLength+1), lastName: "Doe", expectedErr: ErrNameTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			user, err := NewUser("jane@example.com")
			if err != nil {
				t.Fatalf("NewUser: %v", err)
			}

			err = user.ChangeName(tt.firstName, tt.lastName)
			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr != nil {
				return
			}

			if user.FirstName() != "Jane" || user.LastName() != "Doe" {
				t.Fatalf("expected name Jane Doe, got %s %s", user.FirstName(), user.LastName())
			}

			events := user.PullEvents()
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].EventName() != "user.updated" {
				t.Fatalf("expected user.updated event, got %s", events[0].EventName())
			}
		})
	}
}
