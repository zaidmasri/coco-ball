package common

import "uuid"

// UserResult is the write-side acknowledgement shape for
// CreateUser/UpdateUser. It deliberately does not carry the password hash
// or any other credential — the entity never crosses into the interface
// layer. See AGENTS.md's "Application Layer: go-ddd Comparison" section for
// why PlanResult/OperatingExpenseResult follow the same split.
type UserResult struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
}
