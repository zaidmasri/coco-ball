package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zaidmasri/business-planning-tool/internal/domain"
)

type contextKey string

const (
	userContextKey    contextKey = "user"
	sessionCookieName string     = "session_id"
	sessionDuration            = 7 * 24 * time.Hour // 7 days
)

// GetUserFromContext extracts the current user from the request context
func GetUserFromContext(r *http.Request) *domain.User {
	user, ok := r.Context().Value(userContextKey).(*domain.User)
	if !ok {
		return nil
	}
	return user
}

// WithUserContext creates a new context with the user
func WithUserContext(r *http.Request, user *domain.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// setSessionCookie sets a session cookie in the response
func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production with HTTPS
	})
}

// clearSessionCookie removes the session cookie
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// AuthMiddleware checks for valid session and attaches user to context
func (app *App) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user != nil {
			// User already in context
			next.ServeHTTP(w, r)
			return
		}

		// Check for session cookie
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil {
			// Try to retrieve session
			session, err := app.Store.GetSession(cookie.Value)
			if err == nil {
				// Session is valid, get user
				user, err := app.Store.GetUser(session.UserID)
				if err == nil {
					r = WithUserContext(r, user)
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		// No valid session, proceed without user
		next.ServeHTTP(w, r)
	})
}

// RequireAccess is a middleware that checks if the user has access to a plan
func (app *App) RequireAccess(requiredLevel domain.AccessLevel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r)
			if user == nil {
				app.renderErrorPage(w, r, http.StatusUnauthorized, "You must be logged in to access this resource.")
				return
			}

			planIDStr := r.PathValue("id")
			if planIDStr == "" {
				// No plan ID in path, allow through (e.g., home page)
				next.ServeHTTP(w, r)
				return
			}

			planID, err := uuid.Parse(planIDStr)
			if err != nil {
				app.renderErrorPage(w, r, http.StatusBadRequest, "Invalid plan ID.")
				return
			}

			// Check if user has access to the plan
			access, err := app.Store.GetAccess(planID, user.ID())
			if err != nil {
				app.renderErrorPage(w, r, http.StatusForbidden, "You don't have access to this plan.")
				return
			}

			// Check if user has the required access level
			if requiredLevel == domain.Editor && !access.CanEdit() {
				app.renderErrorPage(w, r, http.StatusForbidden, "You don't have permission to edit this plan.")
				return
			}
			if requiredLevel == domain.Viewer && !access.CanView() {
				app.renderErrorPage(w, r, http.StatusForbidden, "You don't have permission to view this plan.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetLogin displays the login form
func (app *App) GetLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user != nil {
			// Already logged in, redirect to home
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		app.renderPage(w, "login.html", "login.html", map[string]any{
			"Title": "Login | Business Planning Tool",
		})
	}
}

// PostLogin handles login form submission
func (app *App) PostLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		// Get user with password
		userCreds, err := app.Store.GetUserWithPassword(email)
		if err != nil {
			// User not found or password error
			app.renderPage(w, "login.html", "login.html", map[string]any{
				"Title":       "Login | Business Planning Tool",
				"ErrorMsg":    "Invalid email or password",
				"Email":       email,
			})
			return
		}

		// Verify password
		if !userCreds.VerifyPassword(password) {
			app.renderPage(w, "login.html", "login.html", map[string]any{
				"Title":       "Login | Business Planning Tool",
				"ErrorMsg":    "Invalid email or password",
				"Email":       email,
			})
			return
		}

		// Create session
		session, err := domain.NewSession(userCreds.ID(), sessionDuration)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			http.Error(w, "Session creation failed", http.StatusInternalServerError)
			return
		}

		if err := app.Store.SaveSession(session); err != nil {
			log.Printf("Failed to save session: %v", err)
			http.Error(w, "Session save failed", http.StatusInternalServerError)
			return
		}

		// Set cookie and redirect
		setSessionCookie(w, session.ID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// GetSignup displays the signup form
func (app *App) GetSignup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user != nil {
			// Already logged in, redirect to home
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		app.renderPage(w, "signup.html", "signup.html", map[string]any{
			"Title": "Sign Up | Business Planning Tool",
		})
	}
}

// PostSignup handles signup form submission
func (app *App) PostSignup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")
		confirmPassword := r.PostForm.Get("confirmPassword")

		// Validate form
		if password != confirmPassword {
			app.renderPage(w, "signup.html", "signup.html", map[string]any{
				"Title":    "Sign Up | Business Planning Tool",
				"ErrorMsg": "Passwords do not match",
				"Email":    email,
			})
			return
		}

		// Check if user already exists
		_, err := app.Store.GetUserWithPassword(email)
		if err == nil {
			// User exists
			app.renderPage(w, "signup.html", "signup.html", map[string]any{
				"Title":    "Sign Up | Business Planning Tool",
				"ErrorMsg": "Email already registered",
				"Email":    email,
			})
			return
		}

		// Create user with password
		userCreds, err := domain.NewUserWithPassword(email, password)
		if err != nil {
			app.renderPage(w, "signup.html", "signup.html", map[string]any{
				"Title":    "Sign Up | Business Planning Tool",
				"ErrorMsg": err.Error(),
				"Email":    email,
			})
			return
		}

		// Save user
		if err := app.Store.SaveUserWithPassword(userCreds); err != nil {
			log.Printf("Failed to save user: %v", err)
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}

		// Create and save session
		session, err := domain.NewSession(userCreds.ID(), sessionDuration)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			http.Error(w, "Session creation failed", http.StatusInternalServerError)
			return
		}

		if err := app.Store.SaveSession(session); err != nil {
			log.Printf("Failed to save session: %v", err)
			http.Error(w, "Session save failed", http.StatusInternalServerError)
			return
		}

		// Set cookie and redirect
		setSessionCookie(w, session.ID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// GetLogout logs out the user
func (app *App) GetLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil {
			// Delete session from store
			_ = app.Store.DeleteSession(cookie.Value)
		}

		// Clear cookie
		clearSessionCookie(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
