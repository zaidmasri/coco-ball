// Package ports holds outbound interfaces owned by the domain layer that
// don't fit domain/repositories (not persistence for an aggregate) but
// still need the same dependency-inversion treatment: the domain declares
// the interface, infrastructure implements it.
package ports

// Mailer sends a single transactional email. Implementations live in
// internal/infrastructure/email/.
type Mailer interface {
	Send(to, subject, body string) error
}
