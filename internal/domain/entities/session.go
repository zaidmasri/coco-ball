package entities

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"uuid"

	domevents "github.com/zaidmasri/business-planning-tool/internal/domain/events"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrWeakPassword    = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong = errors.New("password cannot exceed 72 characters")
)

// Session represents an authenticated user session
type Session struct {
	ID        string    // Random session token
	UserID    uuid.UUID // User this session belongs to
	ExpiresAt time.Time // When the session expires
	CreatedAt time.Time // When the session was created
}

// NewSession creates a new session for a user
func NewSession(userID uuid.UUID, duration time.Duration) (*Session, error) {
	// Generate random session token
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}

	return &Session{
		ID:        hex.EncodeToString(token),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}, nil
}

// IsValid checks if session is still valid
func (s *Session) IsValid() bool {
	return time.Now().Before(s.ExpiresAt)
}

// UserCredentials represents login credentials
type UserCredentials struct {
	Email    string
	Password string
}

// Validate ensures credentials meet minimum requirements
func (c *UserCredentials) Validate() error {
	if c.Email == "" {
		return ErrInvalidEmail
	}
	if err := validateEmailFormat(c.Email); err != nil {
		return err
	}
	if c.Password == "" {
		return ErrInvalidPassword
	}
	if len(c.Password) < 8 {
		return ErrWeakPassword
	}
	if len(c.Password) > 72 {
		return ErrPasswordTooLong
	}
	return nil
}

// HashPassword creates a bcrypt hash of the password. The 72-character cap
// matches bcrypt's real input limit — bytes beyond it are silently ignored
// by the algorithm itself, so rejecting longer passwords up front avoids a
// user believing their full password is part of the hash when it isn't.
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", ErrWeakPassword
	}
	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyPassword checks if a password matches its hash
func VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// UserWithPassword extends User with password hash for auth
type UserWithPassword struct {
	*User
	PasswordHash string
}

// NewUserWithPassword creates a user with a hashed password. First and last
// name are required at registration.
func NewUserWithPassword(email, firstName, lastName, password string) (*UserWithPassword, error) {
	if err := (&UserCredentials{Email: email, Password: password}).Validate(); err != nil {
		return nil, err
	}

	cleanFirst := strings.TrimSpace(firstName)
	cleanLast := strings.TrimSpace(lastName)
	if cleanFirst == "" || cleanLast == "" {
		return nil, ErrInvalidUserName
	}
	if len(cleanFirst) > maxNameLength || len(cleanLast) > maxNameLength {
		return nil, ErrNameTooLong
	}

	user, err := NewUser(email)
	if err != nil {
		return nil, err
	}
	user.SetFirstName(cleanFirst)
	user.SetLastName(cleanLast)

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user.recordEvent(domevents.NewUserRegistered(user.ID(), user.Email()))

	return &UserWithPassword{
		User:         user,
		PasswordHash: hash,
	}, nil
}

// VerifyPassword checks if the provided password matches this user's hash
func (u *UserWithPassword) VerifyPassword(password string) bool {
	return VerifyPassword(u.PasswordHash, password)
}
