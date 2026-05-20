# Authentication Architecture

## Overview

The authentication system implements industry-standard OAuth2 and OpenID Connect protocols, combined with JWT tokens for stateless API authentication. The architecture supports multiple authentication methods while maintaining security and usability.

## Authentication Methods

### 1. Email & Password

**Flow:**
```
User Input → Validation → Hash Verification → JWT Generation → Client
```

**Process:**
```go
// User login
POST /auth/login
{
  "email": "user@example.com",
  "password": "user_password"
}

// Server validates:
1. Check email exists
2. Verify password hash (bcrypt)
3. Check MFA if enabled
4. Generate JWT tokens
5. Return tokens + user info
```

**Password Requirements:**
- Minimum 12 characters
- Must contain: uppercase, lowercase, number, special character
- Cannot contain email or username
- No common passwords (checked against HIBP)

**Password Security:**
```go
// Hashing
hashedPassword, err := bcrypt.GenerateFromPassword(
  []byte(password), 
  bcrypt.DefaultCost, // cost=12
)

// Verification
err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
```

### 2. OAuth2 / OIDC

**Supported Providers:**
- Google
- GitHub
- Microsoft
- Custom OIDC providers

**Authorization Code Flow:**
```
┌─────────────────────────────────────────────────────────┐
│ 1. User clicks "Sign in with Google"                   │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 2. Redirect to OAuth provider                          │
│ https://accounts.google.com/o/oauth2/v2/auth?          │
│   client_id=...&scope=...&redirect_uri=...             │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 3. User grants permission at Google                    │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 4. Google redirects with authorization code            │
│ https://api.enterprise-ai.com/auth/oauth/callback?     │
│   code=...&state=...                                   │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 5. Backend exchanges code for tokens                   │
│ POST https://oauth2.googleapis.com/token               │
│ {                                                       │
│   "code": "...",                                      │
│   "client_id": "...",                                 │
│   "client_secret": "...",                             │
│   "redirect_uri": "...",                              │
│   "grant_type": "authorization_code"                  │
│ }                                                       │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 6. Backend creates/updates user & workspace            │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ 7. Return JWT tokens to client                         │
└─────────────────────────────────────────────────────────┘
```

**Implementation:**
```go
type OAuthConfig struct {
  Provider     string // google, github, microsoft
  ClientID     string
  ClientSecret string
  RedirectURI  string
  Scopes       []string
}

func (s *AuthService) HandleOAuthCallback(code, state string) (*TokenResponse, error) {
  // Verify state to prevent CSRF
  if !s.VerifyState(state) {
    return nil, errors.New("invalid state")
  }
  
  // Exchange code for tokens
  tokens, err := s.ExchangeCodeForTokens(code)
  if err != nil {
    return nil, err
  }
  
  // Get user info from provider
  userInfo, err := s.GetUserInfo(tokens.AccessToken)
  if err != nil {
    return nil, err
  }
  
  // Create or update user
  user, err := s.UpsertUser(userInfo)
  if err != nil {
    return nil, err
  }
  
  // Generate JWT tokens
  return s.GenerateTokens(user)
}
```

### 3. Multi-Factor Authentication (MFA)

**Supported Methods:**
- TOTP (Time-based One-Time Password)
- SMS (Short Message Service)
- Email codes
- Backup codes

**TOTP Setup Flow:**
```
1. User requests MFA setup
2. Server generates secret (RFC 4226)
3. Client displays QR code
4. User scans with authenticator app
5. User provides verification code
6. Server validates and stores secret
```

**TOTP Implementation:**
```go
import "github.com/pquerna/otp/totp"

// Generate secret
secret, err := totp.GenerateSecret(totp.GenerateOpts{
  Issuer:      "Enterprise AI",
  AccountName: user.Email,
})

// Verify code
valid, err := totp.ValidateCode(
  code,
  secret.Secret(),
  time.Now(),
)
```

**MFA Flow During Login:**
```
1. User submits email + password
2. Credentials validated
3. If MFA enabled:
   - Send OTP (SMS/email)
   - Return partial token (limited scope)
4. User submits OTP
5. Validate OTP
6. Issue full JWT tokens
```

### 4. API Keys

**Key Generation:**
```go
type APIKey struct {
  ID        string
  KeyHash   string // sha256(key)
  PublicKey string // prefix for display
  Secret    string // only shown once
  Name      string
  ExpiresAt time.Time
  CreatedAt time.Time
  LastUsedAt time.Time
}

// Generate key
key := generateRandomKey(32) // 256 bits
hash := sha256(key)

// Return to user (only once)
// Store hash in database
```

**API Key Usage:**
```bash
curl -H "X-API-Key: sk_live_abcd1234..." \
     https://api.enterprise-ai.com/api/v1/documents
```

**Validation:**
```go
func (s *AuthService) ValidateAPIKey(keyString string) (*User, error) {
  hash := sha256(keyString)
  
  // Find key by hash
  key := s.repo.GetAPIKeyByHash(hash)
  if key == nil {
    return nil, errors.New("invalid key")
  }
  
  // Check expiration
  if key.ExpiresAt.Before(time.Now()) {
    return nil, errors.New("key expired")
  }
  
  // Update last used
  s.repo.UpdateAPIKeyLastUsed(key.ID)
  
  // Get user
  return s.repo.GetUserByID(key.UserID), nil
}
```

## JWT Token Structure

### Access Token
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
.
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "workspace_id": "workspace-uuid",
  "workspace_role": "editor",
  "iat": 1716194400,
  "exp": 1716198000,
  "aud": "enterprise-ai-api",
  "iss": "enterprise-ai-auth",
  "scope": "read:documents write:documents",
  "permissions": ["document:read", "document:write"]
}
.
<signature>
```

**Token Lifetimes:**
```
Access Token:   1 hour (3600 seconds)
Refresh Token:  7 days (604800 seconds)
MFA Token:      5 minutes (temporary, limited scope)
API Key:        Until revoked or expiration date
```

### Refresh Token
```json
{
  "sub": "user-uuid",
  "iat": 1716194400,
  "exp": 1716799200,
  "type": "refresh",
  "family_id": "refresh-family-uuid"
}
```

**Token Rotation:**
```
Client stores refresh_token
     ↓
Access token expires
     ↓
Client sends: POST /auth/refresh { "refresh_token": "..." }
     ↓
Server validates refresh token
     ↓
Server invalidates old refresh token
     ↓
Server issues new access token + new refresh token
     ↓
Client updates stored tokens
```

## Token Management

### Token Blacklist
```go
// On logout or token revocation
type TokenBlacklist struct {
  TokenID   string
  ExpiresAt time.Time
}

// Check before accepting token
func (s *AuthService) ValidateToken(token string) error {
  claims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
    return s.signingKey, nil
  })
  
  if err != nil {
    return err
  }
  
  jti := claims["jti"] // JWT ID
  
  // Check blacklist (Redis)
  if s.isBlacklisted(jti) {
    return errors.New("token revoked")
  }
  
  return nil
}

// On logout, add to blacklist
func (s *AuthService) RevokeToken(jti string, expiresAt time.Time) {
  ttl := time.Until(expiresAt)
  s.cache.Set(fmt.Sprintf("blacklist:%s", jti), true, ttl)
}
```

### Session Management
```go
type Session struct {
  ID           string
  UserID       string
  RefreshToken string
  IPAddress    string
  UserAgent    string
  CreatedAt    time.Time
  LastActivity time.Time
  ExpiresAt    time.Time
  Revoked      bool
}

// Store in Redis for fast access
key := fmt.Sprintf("session:%s", sessionID)
redis.Set(ctx, key, session, 7*24*time.Hour)
```

## Session Invalidation

### Logout
```go
func (s *AuthService) Logout(sessionID string) error {
  // Revoke refresh token
  if err := s.repo.RevokeSession(sessionID); err != nil {
    return err
  }
  
  // Remove from Redis
  s.cache.Del(ctx, fmt.Sprintf("session:%s", sessionID))
  
  // Add access token to blacklist (if JTI present)
  if jti := extractJTI(token); jti != "" {
    s.RevokeToken(jti, tokenExpiry)
  }
  
  return nil
}
```

### Concurrent Session Limit
```go
type SessionPolicy struct {
  MaxActiveSessions int // 5 per user default
  IdleTimeout        time.Duration // 30 minutes
}

// Enforce limit
func (s *AuthService) EnforceSessionLimit(userID string) error {
  sessions := s.repo.GetUserSessions(userID)
  
  if len(sessions) >= s.policy.MaxActiveSessions {
    // Revoke oldest session
    oldest := findOldest(sessions)
    s.repo.RevokeSession(oldest.ID)
  }
  
  return nil
}
```

## Authorization

### Role-Based Access Control (RBAC)

**Roles:**
```yaml
owner:
  - Manage workspace members
  - Change workspace settings
  - Delete workspace
  - Access all documents
  - Full audit logs

admin:
  - Manage members (except owner)
  - Change workspace settings
  - Access all documents
  - View audit logs

editor:
  - Upload documents
  - Search documents
  - Edit owned documents
  - View shared documents
  - Manage own API keys

viewer:
  - Search documents
  - Download documents
  - View shared documents
  - Cannot upload or edit
```

### Permission Checking
```go
func (s *AuthService) CheckPermission(
  userID, workspaceID, permission string,
) error {
  // Get user role
  member, err := s.repo.GetWorkspaceMember(userID, workspaceID)
  if err != nil {
    return err
  }
  
  // Check role permissions
  rolePerms := s.getRolePermissions(member.Role)
  if !contains(rolePerms, permission) {
    return errors.New("insufficient permissions")
  }
  
  return nil
}

// Middleware
func (s *AuthService) RequirePermission(permission string) gin.HandlerFunc {
  return func(c *gin.Context) {
    userID := c.GetString("user_id")
    workspaceID := c.GetString("workspace_id")
    
    if err := s.CheckPermission(userID, workspaceID, permission); err != nil {
      c.JSON(403, gin.H{"error": "forbidden"})
      c.Abort()
      return
    }
    
    c.Next()
  }
}
```

## Security Considerations

### Secret Storage
```go
// Never hardcode secrets
// Use environment variables or HashiCorp Vault

signingKey := os.Getenv("JWT_SIGNING_KEY")
oauthClientSecret := vault.GetSecret("oauth/google/client_secret")
```

### CSRF Protection
```go
// State parameter for OAuth flows
state := generateSecureRandom(32)
session.Set("oauth_state", state)

// Verify in callback
if requestState != session.Get("oauth_state") {
  return errors.New("invalid state")
}
```

### Rate Limiting
```go
// Prevent brute force attacks
type RateLimiter struct {
  LoginAttempts     int           // 5 per 15 minutes
  OTPAttempts       int           // 3 per 15 minutes
  KeyCreation       int           // 100 per day
  PasswordReset     int           // 3 per hour
}

// Store in Redis
key := fmt.Sprintf("ratelimit:login:%s", email)
attempts := redis.Incr(ctx, key)
redis.Expire(ctx, key, 15*time.Minute)

if attempts > 5 {
  return errors.New("too many login attempts")
}
```

### Password Reset
```
1. User requests password reset
2. Generate secure token (32 bytes, random)
3. Send via email with expiration (1 hour)
4. User clicks link, verifies token
5. User enters new password
6. Validate password strength
7. Hash and store
8. Invalidate all sessions
9. Force re-login
```

## Audit Logging

```go
type AuthEvent struct {
  EventType   string    // login, logout, mfa_setup, key_created
  UserID      string
  Status      string    // success, failed
  IPAddress   string
  UserAgent   string
  ErrorReason string    // if failed
  CreatedAt   time.Time
}

// Log all auth events
func (s *AuthService) LogAuthEvent(event AuthEvent) {
  s.auditService.Log(event)
}
```

## Integration with API Gateway

```go
// Middleware chain
router.Use(
  middleware.Logging(),
  middleware.CORS(),
  middleware.RateLimiting(),
  middleware.AuthenticationValidator(authService),
  middleware.AuthorizationEnforcer(authService),
  middleware.AuditLogger(auditService),
)
```