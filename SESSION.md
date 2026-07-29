# Session Management & Authentication

## Overview

The Business Planning Tool now includes a complete session management system with login/signup functionality. Users can create accounts and maintain persistent sessions across browser sessions.

## Architecture

### Session Models

#### UserWithPassword
Located in `internal/domain/session.go`, extends the basic `User` model with password hashing:
- **ID**: UUID identifier (inherited from User)
- **Email**: User's email (inherited from User)
- **PasswordHash**: Bcrypt hashed password

```go
user, _ := domain.NewUserWithPassword("alice@example.com", "secure_password_123")
```

#### Session
Represents an authenticated session:
- **ID**: Random 64-character hex token (256 bits)
- **UserID**: The user this session belongs to
- **CreatedAt**: When the session was created
- **ExpiresAt**: When the session expires (7 days from creation)

```go
session, _ := domain.NewSession(userID, 7*24*time.Hour)
```

### Password Management

#### Hashing
Uses bcrypt with default cost factor (12):
```go
hash, err := domain.HashPassword("mypassword")
```

#### Verification
Safely compares password against hash:
```go
valid := domain.VerifyPassword(hash, "mypassword") // true
valid := domain.VerifyPassword(hash, "wrong")      // false
```

#### Validation Rules
- Minimum 8 characters
- Email must be valid
- Cannot be empty

## Authentication Flow

### Signup
1. User navigates to `/signup`
2. Enters email, password, and confirms password
3. Form validates:
   - Passwords match
   - Password is 8+ characters
   - Email is not already registered
4. System creates `UserWithPassword` and saves to store
5. Session is created automatically
6. User is logged in and redirected to home

### Login
1. User navigates to `/login`
2. Enters email and password
3. System retrieves user credentials by email
4. Bcrypt verifies password against hash
5. If valid, creates new session
6. Session cookie is set (HttpOnly, SameSite=Lax)
7. User is redirected to home

### Logout
1. User navigates to `/logout`
2. System retrieves session from cookie
3. Session is deleted from store
4. Session cookie is cleared
5. User is redirected to home

### Session Persistence
1. On every request, `AuthMiddleware` checks for session cookie
2. If found, validates session hasn't expired
3. Retrieves user and attaches to request context
4. Handlers access user via `GetUserFromContext(r)`

## Session Storage

### MemoryStore Implementation

The `MemoryStore` tracks sessions in memory:

```go
type MemoryStore struct {
    // ... other fields ...
    userCreds map[string]*domain.UserWithPassword  // email -> UserWithPassword
    sessions  map[string]*domain.Session           // sessionID -> Session
}
```

#### Key Operations

**SaveUserWithPassword**: Stores user with hashed password
```go
err := store.SaveUserWithPassword(user)
```

**GetUserWithPassword**: Retrieves user for login
```go
user, err := store.GetUserWithPassword("alice@example.com")
if err == nil && user.VerifyPassword(password) {
    // Login successful
}
```

**SaveSession**: Creates new session
```go
err := store.SaveSession(session)
```

**GetSession**: Validates session
```go
session, err := store.GetSession(sessionID)
if err == nil && session.IsValid() {
    // Session is valid
}
```

**DeleteSession**: Logs user out
```go
err := store.DeleteSession(sessionID)
```

## Routes

### Authentication Routes

| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| GET | `/signup` | GetSignup | Display signup form |
| POST | `/signup` | PostSignup | Process signup |
| GET | `/login` | GetLogin | Display login form |
| POST | `/login` | PostLogin | Process login |
| GET | `/logout` | GetLogout | Log out user |

### Cookies

**Session Cookie**
- Name: `session_id`
- Value: Random 64-char hex token
- Path: `/` (all routes)
- MaxAge: 7 days (604800 seconds)
- HttpOnly: True (JavaScript cannot access)
- SameSite: Lax (sent with cross-site top-level navigations)
- Secure: False (set to True in production with HTTPS)

## Security Considerations

### Current Implementation (Development)

The current setup is suitable for development:
- ✅ Passwords hashed with bcrypt (industry standard)
- ✅ Sessions stored server-side in memory
- ✅ HttpOnly cookies prevent XSS theft
- ✅ SameSite=Lax protects against CSRF
- ✅ 7-day session expiry

### Production Recommendations

Before going to production, implement:

1. **HTTPS Only**
   - Set `Secure: true` on cookies
   - Redirect HTTP to HTTPS

2. **Database Persistence**
   - Move sessions to database (SQLite, PostgreSQL)
   - Persist user credentials in database
   - Add session cleanup jobs (expired sessions)

3. **Rate Limiting**
   - Limit login attempts (prevent brute force)
   - Implement exponential backoff
   - Log failed attempts

4. **Password Requirements**
   - Enforce stronger passwords in production
   - Consider password complexity rules
   - Add optional MFA/2FA

5. **Session Management**
   - Implement "remember me" functionality
   - Add session invalidation on password change
   - Track active sessions per user

6. **Monitoring**
   - Log authentication events
   - Alert on suspicious activity
   - Monitor for account takeovers

## Usage Examples

### Creating an Account

```go
// From handler
email := "alice@example.com"
password := "secure_pass_123"

userCreds, err := domain.NewUserWithPassword(email, password)
if err != nil {
    // Invalid password or email
    return
}

if err := app.Store.SaveUserWithPassword(userCreds); err != nil {
    // Failed to save
    return
}

// Create session
session, _ := domain.NewSession(userCreds.ID(), 7*24*time.Hour)
app.Store.SaveSession(session)
setSessionCookie(w, session.ID)
```

### Logging In

```go
email := "alice@example.com"
password := "secure_pass_123"

userCreds, err := app.Store.GetUserWithPassword(email)
if err != nil {
    // User not found
    return
}

if !userCreds.VerifyPassword(password) {
    // Wrong password
    return
}

// Create session
session, _ := domain.NewSession(userCreds.ID(), 7*24*time.Hour)
app.Store.SaveSession(session)
setSessionCookie(w, session.ID)
```

### Getting Current User

```go
func (app *App) SomeHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := GetUserFromContext(r)
        if user == nil {
            // User not logged in
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        
        // User is logged in
        fmt.Printf("Welcome %s\n", user.Email())
    }
}
```

## Testing

### Manual Testing

1. **Signup Flow**
   - Navigate to `/signup`
   - Enter email and password
   - Verify account created and logged in

2. **Login Flow**
   - Logout from `/logout`
   - Navigate to `/login`
   - Enter credentials
   - Verify logged back in

3. **Session Persistence**
   - Login
   - Close browser
   - Reopen and navigate to site
   - Verify still logged in (session cookie restored)

4. **Session Expiry** (manual test)
   - Create session with 1-second expiry
   - Wait for expiry
   - Attempt to use session
   - Verify session invalid

### Automated Testing

```go
func TestSessionExpiry(t *testing.T) {
    session, _ := domain.NewSession(uuid.New(), 1*time.Millisecond)
    time.Sleep(10 * time.Millisecond)
    
    if session.IsValid() {
        t.Fatal("expired session should be invalid")
    }
}
```

## Debugging

### Common Issues

**Issue: "Invalid email or password" on correct credentials**
- Verify user exists: `store.GetUserWithPassword(email)`
- Check password matches: `user.VerifyPassword(password)`
- Ensure email is lowercase normalized

**Issue: User logged out unexpectedly**
- Check session expiry: `session.ExpiresAt`
- Verify session exists: `store.GetSession(sessionID)`
- Check cookie settings

**Issue: Cookie not persisting**
- Verify HttpOnly is set (should not be accessible from JS)
- Check SameSite mode
- Ensure Path is `/`

## Future Enhancements

1. **Email Verification**
   - Send verification email on signup
   - Require email confirmation before login

2. **Password Reset**
   - Implement forgot password flow
   - Send reset links via email

3. **Session Management UI**
   - Show active sessions to user
   - Allow logging out from other devices
   - Display last login time

4. **OAuth/SSO**
   - Add Google/GitHub login
   - Replace password auth with OAuth

5. **Rate Limiting**
   - Limit login attempts
   - Throttle signup requests
