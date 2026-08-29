package entities

import "errors"

// ErrValidation is the sentinel that all value-object and entity validation
// errors wrap. Callers can check with errors.Is(err, ErrValidation).
var ErrValidation = errors.New("validation error")

// userFacingErrors lists every domain sentinel whose message is safe to
// show an end user verbatim: plain English describing what was wrong with
// their input or the business rule that blocked them, never anything
// derived from infrastructure state (SQL errors, file paths, etc.).
//
// This list exists because ErrValidation above is only actually wrapped by
// one call site (money.go's currency check) - most validation/business-rule
// errors across the domain are plain errors.New sentinels declared next to
// the entity they belong to (plan.go, session.go, user.go, invite.go).
// IsUserFacing is the single place that knows the full set, so interface/web
// has one check to call rather than re-deriving this list per handler.
var userFacingErrors = []error{
	ErrValidation,
	// plan.go
	ErrInvalidName, ErrNegativeAmount, ErrInvalidStartingMonth,
	ErrInvalidDepreciationMethod, ErrInvalidUsefulLife,
	ErrPurchaseCostLessThanSalvageValue, ErrInvalidGrowthType,
	ErrInvalidStartingYear, ErrNameTooLong, ErrAmountTooLarge,
	ErrInvalidGrowthRate, ErrInvalidRate, ErrInvalidTerm,
	// session.go
	ErrInvalidPassword, ErrWeakPassword, ErrPasswordTooLong,
	ErrSessionNotFound, ErrSessionExpired,
	// user.go
	ErrInvalidEmail, ErrInvalidEmailFormat, ErrInvalidUserName,
	ErrUserNotFound, ErrAccessDenied,
	// invite.go
	ErrSelfInvite, ErrDuplicateInvite,
	ErrInviteNotFound, ErrInviteNotPending, ErrInviteForbidden,
}

// IsUserFacing reports whether err (or something in its chain) is a known
// domain sentinel safe to display to the end user verbatim. Anything else -
// infrastructure/repository failures in particular - is not: callers should
// log the real error and show a generic message instead.
func IsUserFacing(err error) bool {
	for _, sentinel := range userFacingErrors {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}
