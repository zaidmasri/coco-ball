package entities

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	domevents "github.com/zaidmasri/business-planning-tool/internal/domain/events"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword   = errors.New("invalid password")
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionExpired    = errors.New("session expired")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrWeakPassword      = errors.New("password must be at least 8 characters")
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
	if c.Password == "" {
		return ErrInvalidPassword
	}
	if len(c.Password) < 8 {
		return ErrWeakPassword
	}
	return nil
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", ErrWeakPassword
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

	if strings.TrimSpace(firstName) == "" || strings.TrimSpace(lastName) == "" {
		return nil, ErrInvalidUserName
	}

	user, err := NewUser(email)
	if err != nil {
		return nil, err
	}
	user.SetFirstName(strings.TrimSpace(firstName))
	user.SetLastName(strings.TrimSpace(lastName))

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
