// Package web has the http endpoints the application uses
package web

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/zaidmasri/business-planning-tool/internal/application/commands"
	ifaces "github.com/zaidmasri/business-planning-tool/internal/application/interfaces"
	domain "github.com/zaidmasri/business-planning-tool/internal/domain/entities"
	"github.com/zaidmasri/business-planning-tool/internal/views"
)

type contextKey string

const (
	userContextKey    contextKey = "user"
	sessionCookieName string     = "session_id"
	sessionDuration              = 7 * 24 * time.Hour // 7 days
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

// SessionMiddleware resolves the session cookie into a user and attaches it
// to the request context for every downstream handler.
type SessionMiddleware struct {
	authSvc ifaces.AuthService
}

func NewSessionMiddleware(authSvc ifaces.AuthService) *SessionMiddleware {
	return &SessionMiddleware{authSvc: authSvc}
}

// Authenticate checks for a valid session and attaches the user to context.
func (m *SessionMiddleware) Authenticate(next http.Handler) http.Handler {
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
			session, err := m.authSvc.GetSession(cookie.Value)
			if err == nil {
				// Session is valid, get user
				user, err := m.authSvc.GetUser(session.UserID)
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

// AuthController handles login, signup, logout, and the user's profile page.
type AuthController struct {
	authSvc       ifaces.AuthService
	templateCache map[string]*template.Template
}

func NewAuthController(mux *http.ServeMux, authSvc ifaces.AuthService, templateCache map[string]*template.Template) *AuthController {
	c := &AuthController{authSvc: authSvc, templateCache: templateCache}

	mux.HandleFunc("GET /login", c.GetLogin)
	mux.HandleFunc("POST /login", c.PostLogin)
	mux.HandleFunc("GET /signup", c.GetSignup)
	mux.HandleFunc("POST /signup", c.PostSignup)
	mux.HandleFunc("POST /logout", c.PostLogout)
	mux.HandleFunc("GET /profile", c.GetProfile)
	mux.HandleFunc("GET /profile/edit", c.GetProfileEdit)
	mux.HandleFunc("POST /profile/edit", c.PostProfileEdit)
	mux.HandleFunc("POST /profile/delete", c.PostProfileDelete)

	return c
}

// GetLogin displays the login form
func (c *AuthController) GetLogin(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user != nil {
		// Already logged in, redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	page := views.BuildLoginPage()
	views.RenderLoginPage(w, c.templateCache, page)
}

// PostLogin handles login form submission
func (c *AuthController) PostLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "We couldn't process that form submission. Please try again.")
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	// Get user with password
	userCreds, err := c.authSvc.GetUserWithPassword(email)
	if err != nil {
		// User not found or password error
		page := views.BuildLoginPageWithError(email, "Invalid email or password")
		views.RenderLoginPage(w, c.templateCache, page)
		return
	}

	// Verify password
	if !userCreds.VerifyPassword(password) {
		page := views.BuildLoginPageWithError(email, "Invalid email or password")
		views.RenderLoginPage(w, c.templateCache, page)
		return
	}

	// Create session
	session, err := c.authSvc.CreateSession(userCreds.ID(), sessionDuration)
	if err != nil {
		renderInternalError(w, r, c.templateCache, "AuthSvc CreateSession", err)
		return
	}

	// Set cookie and redirect
	setSessionCookie(w, session.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GetSignup displays the signup form
func (c *AuthController) GetSignup(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user != nil {
		// Already logged in, redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	page := views.BuildSignupPage()
	views.RenderSignupPage(w, c.templateCache, page)
}

// PostSignup handles signup form submission
func (c *AuthController) PostSignup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "We couldn't process that form submission. Please try again.")
		return
	}

	email := r.PostForm.Get("email")
	firstName := r.PostForm.Get("firstName")
	lastName := r.PostForm.Get("lastName")
	password := r.PostForm.Get("password")
	confirmPassword := r.PostForm.Get("confirmPassword")

	// Validate form
	if password != confirmPassword {
		page := views.BuildSignupPageWithError(email, firstName, lastName, "Passwords do not match")
		views.RenderSignupPage(w, c.templateCache, page)
		return
	}

	// Check if user already exists
	_, err := c.authSvc.GetUserWithPassword(email)
	if err == nil {
		// User exists
		page := views.BuildSignupPageWithError(email, firstName, lastName, "Email already registered")
		views.RenderSignupPage(w, c.templateCache, page)
		return
	}

	// Create user with password
	result, err := c.authSvc.CreateUser(&commands.CreateUser{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	})
	if err != nil {
		msg := safeErrorMessage("AuthSvc CreateUser", err)
		page := views.BuildSignupPageWithError(email, firstName, lastName, msg)
		views.RenderSignupPage(w, c.templateCache, page)
		return
	}

	// Create and save session
	session, err := c.authSvc.CreateSession(result.Result.ID, sessionDuration)
	if err != nil {
		renderInternalError(w, r, c.templateCache, "AuthSvc CreateSession", err)
		return
	}

	// Set cookie and redirect
	setSessionCookie(w, session.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// PostLogout logs out the user
func (c *AuthController) PostLogout(w http.ResponseWriter, r *http.Request) {
	// Get session cookie
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		// Delete session from store
		_ = c.authSvc.DeleteSession(cookie.Value)
	}

	// Clear cookie
	clearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GetProfile displays the user's profile page
func (c *AuthController) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	page := views.BuildProfilePage(user)
	views.RenderProfilePage(w, c.templateCache, page)
}

// GetProfileEdit displays the profile page with the name fields editable
func (c *AuthController) GetProfileEdit(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	page := views.BuildProfileEditPage(user, "")
	views.RenderProfilePage(w, c.templateCache, page)
}

// PostProfileEdit handles the profile name edit form submission
func (c *AuthController) PostProfileEdit(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		renderErrorPage(w, r, c.templateCache, http.StatusBadRequest, "We couldn't process that form submission. Please try again.")
		return
	}

	firstName := r.PostForm.Get("firstName")
	lastName := r.PostForm.Get("lastName")

	if _, err := c.authSvc.UpdateUser(&commands.UpdateUser{
		UserID:    user.ID(),
		FirstName: firstName,
		LastName:  lastName,
	}); err != nil {
		msg := safeErrorMessage("AuthSvc UpdateUser", err)
		page := views.BuildProfileEditPage(user, msg)
		page.FirstName = firstName
		page.LastName = lastName
		views.RenderProfilePage(w, c.templateCache, page)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// PostProfileDelete deletes the logged-in user's own account and clears
// their session. Plans they own are left in place - see AGENTS.md's
// "Application Layer: go-ddd Comparison" section for why deletion isn't
// blocked by plan ownership.
func (c *AuthController) PostProfileDelete(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if _, err := c.authSvc.DeleteUser(&commands.DeleteUser{UserID: user.ID()}); err != nil {
		msg := safeErrorMessage("AuthSvc DeleteUser", err)
		page := views.BuildProfilePageWithDeleteError(user, msg)
		views.RenderProfilePage(w, c.templateCache, page)
		return
	}

	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		_ = c.authSvc.DeleteSession(cookie.Value)
	}
	clearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
