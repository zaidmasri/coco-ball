package entities

import "errors"

// ErrValidation is the sentinel that all value-object and entity validation
// errors wrap. Callers can check with errors.Is(err, ErrValidation).
var ErrValidation = errors.New("validation error")
