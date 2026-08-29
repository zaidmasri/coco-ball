package entities

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsUserFacing(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"ErrValidation itself", ErrValidation, true},
		{"wrapped ErrValidation", fmt.Errorf("%w: bad currency", ErrValidation), true},
		{"plan.go sentinel", ErrInvalidName, true},
		{"session.go sentinel", ErrWeakPassword, true},
		{"user.go sentinel", ErrInvalidEmailFormat, true},
		{"invite.go sentinel", ErrSelfInvite, true},
		{"invite.go sentinel wrapped by a caller", fmt.Errorf("create invite: %w", ErrDuplicateInvite), true},
		{"unrelated stdlib error", errors.New("sql: database is locked"), false},
		{"unrelated wrapped error", fmt.Errorf("save plan: %w", errors.New("disk full")), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUserFacing(tt.err); got != tt.want {
				t.Errorf("IsUserFacing(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
