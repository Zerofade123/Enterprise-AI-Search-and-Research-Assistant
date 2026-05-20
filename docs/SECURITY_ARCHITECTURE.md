# Security Architecture

## Overview

Security is implemented at every layer: network, application, data, and operations. The architecture follows zero-trust principles and defense-in-depth strategies.

## Security Layers

### 1. Network Security

#### Network Perimeter

```
                    ┌─────────────────┐
                    │  CloudFlare CDN │
                    │  (DDoS Protection)
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   AWS WAF       │
                    │ • SQL Injection  │
                    │ • XSS           │
                    │ • Rate Limit    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Load Balancer  │
                    │  (TLS 1.3)      │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   NGINX         │
                    │   Ingress       │
                    └────────┬────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
      API GW        Auth Svc        Document   Search
```

#### TLS Configuration

```yaml
TLS Version: 1.3+ only
Minimum Cipher Suites:
  - TLS_AES_256_GCM_SHA384
  - TLS_CHACHA20_POLY1305_SHA256
  - TLS_AES_128_GCM_SHA256

Certificate:
  - Issuer: Let's Encrypt
  - Duration: 90 days
  - Auto-renewal: 30 days before expiry
  - OCSP Stapling: Enabled
  - Certificate Pinning: For API clients

HTTPSHeader:
  - Strict-Transport-Security: max-age=31536000; includeSubDomains
  - Content-Security-Policy: default-src 'self'
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - X-XSS-Protection: 1; mode=block
```

#### mTLS (Service-to-Service)

```go
// Enable mTLS for internal service communication
type mTLSConfig struct {
  CertFile string // /etc/tls/certs/server.crt
  KeyFile  string // /etc/tls/certs/server.key
  CAFile   string // /etc/tls/certs/ca.crt
}

// Enforce mutual certificate validation
tlsConfig := &tls.Config{
  ClientAuth: tls.RequireAndVerifyClientCert,
  ClientCAs:  caCertPool,
  Certificates: []tls.Certificate{serverCert},
}
```

#### Network Policies

```yaml
Kubernetes Network Policies:
  Default: Deny all ingress/egress
  
  Allow List:
    - API Gateway: Accept traffic from ingress-nginx
    - Services: Accept traffic only from API Gateway
    - Database: Accept traffic only from services
    - Redis: Accept traffic only from services
    
  Egress:
    - Allow DNS queries to coredns
    - Allow external APIs (egress to 443)
    - Deny internal to internet
```

### 2. Application Security

#### Authentication (OAuth2 + OpenID Connect)

```go
type AuthFlow struct {
  Method         string // password, oauth2, oidc, saml
  MFARequired    bool
  SessionTimeout time.Duration // 1 hour
}

// OAuth2 with PKCE (Proof Key for Code Exchange)
type PKCEConfig struct {
  CodeChallenge       string // base64(sha256(codeVerifier))
  CodeChallengeMethod string // S256
  CodeVerifier        string // Random 43-128 chars
}
```

#### Authorization (RBAC + ABAC)

```go
type Permission struct {
  Resource string // document, workspace, user
  Action   string // read, write, delete, admin
  Scope    string // workspace-123 (workspace isolation)
  Condition map[string]interface{} // Custom conditions
}

// Example RBAC Policy
var Policies = map[string][]Permission{
  "admin": {
    {Resource: "*", Action: "*", Scope: "workspace"},
  },
  "editor": {
    {Resource: "document", Action: "read", Scope: "workspace"},
    {Resource: "document", Action: "write", Scope: "workspace"},
    {Resource: "document", Action: "delete", Scope: "owned"},
  },
  "viewer": {
    {Resource: "document", Action: "read", Scope: "shared"},
  },
}
```

#### Input Validation

```go
type ValidationRules struct {
  Email    *regexp.Regexp // RFC 5322
  Password string         // min 12 chars, complexity
  UUID     *regexp.Regexp // UUID v4
  Filename string         // max 255 bytes, safe chars
}

// Validate all inputs
if !isValidEmail(email) {
  return errors.New("invalid email")
}

if !isStrongPassword(password) {
  return errors.New("password too weak")
}
```

#### Output Encoding

```go
// HTML encoding for web output
html.EscapeString(userInput) // <script> becomes &lt;script&gt;

// URL encoding
url.QueryEscape(userInput)

// JSON encoding (automatic with json.Marshal)

// SQL query parameterization (prevent SQL injection)
db.Query("SELECT * FROM users WHERE id = $1", userID) // NOT string concatenation
```

#### CSRF Protection

```go
// Generate CSRF token
token := generateRandomToken(32)
session.SetCSRFToken(token)

// Verify token on state-changing requests
func csrfMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
      token := r.FormValue("csrf_token")
      if !verifyCSRFToken(token) {
        http.Error(w, "Invalid CSRF token", http.StatusForbidden)
        return
      }
    }
    next.ServeHTTP(w, r)
  })
}
```

### 3. Data Security

#### Encryption at Rest

```yaml
Database (PostgreSQL):
  - Algorithm: AES-256
  - Managed by: AWS RDS
  - Key Management: AWS KMS
  - Key Rotation: Automatic (annual)
  
Object Storage (S3):
  - Algorithm: AES-256 or KMS
  - Default encryption: Enabled
  - Bucket policies: Deny unencrypted uploads
  
Redis:
  - Algorithm: AES-256
  - Managed by: AWS ElastiCache
  - Key Management: AWS KMS
```

#### Encryption in Transit

```yaml
TLS 1.3: All network traffic
Data Classification:
  - Public: CDN cacheable
  - Internal: Service-to-service (mTLS)
  - Sensitive: User data (TLS + encryption)
  - Confidential: API keys, passwords (TLS + application encryption)
```

#### Sensitive Data Handling

```python
# Never log sensitive data
BANNED_FIELDS = [
  'password', 'api_key', 'secret', 'token',
  'credit_card', 'ssn', 'private_key'
]

def should_log(field_name: str, value: Any) -> bool:
  if field_name.lower() in BANNED_FIELDS:
    return False  # Don't log
  
  # For emails/IDs, log only hash
  if field_name in ['email', 'user_id']:
    return False  # Handle separately
  
  return True

# Mask sensitive data in logs
def mask_sensitive(value: str) -> str:
  if len(value) <= 4:
    return '****'
  return value[:2] + '*' * (len(value) - 4) + value[-2:]
```

#### Data Retention & Deletion

```yaml
Data Retention Policies:
  - Active user data: Indefinite (until deleted)
  - Deleted user data: 30 days (hard delete)
  - Audit logs: 90 days (configurable)
  - Search history: 30 days (user can extend)
  - Backup data: 30 days (post-deletion)
  
Data Deletion:
  - Soft delete: Mark as deleted (24h reversible)
  - Hard delete: Cryptographic erasure (immediate)
  - Backup cleanup: Automated 30 days post-delete
```

### 4. Infrastructure Security

#### Secrets Management

```yaml
HashiCorp Vault:
  Deployment: HA cluster (3 nodes)
  Storage: PostgreSQL encrypted
  Audit Logging: All access logged
  
  Secret Types:
    - Database credentials
    - API keys
    - OAuth secrets
    - Private encryption keys
    - SSL certificates
  
  Rotation Policy:
    - Database passwords: Monthly
    - API keys: Quarterly
    - Certificates: Auto 30 days before expiry
    - OAuth secrets: Annual or on compromise
```

#### Key Management

```yaml
AWS KMS:
  Master Key: Customer-managed
  Key Policy: Least privilege
  Audit: CloudTrail logs all operations
  
  Encryption Key Hierarchy:
    Master Key (KMS)
      ├── Database key
      ├── S3 key
      └── Application key
```

#### Container Security

```yaml
Image Scanning:
  - Scan on build: Trivy/Clair
  - Scan on push: ECR image scanning
  - Scan on run: Falco runtime detection
  - Vulnerability threshold: Block high/critical
  
Runtime Security:
  - Read-only root filesystem
  - No privileged containers
  - Resource limits enforced
  - Non-root user (UID 1000+)
  - No capability elevation
```

#### RBAC & Access Control

```yaml
Kubernetes RBAC:
  Service Accounts: One per service
  Role Bindings: Least privilege
  Example:
    - API Gateway: Can access Auth Service
    - Document Service: Can access PostgreSQL secret
    - Search Service: Cannot access Auth secrets
  
AWS IAM:
  - Service roles for K8s nodes
  - Cross-account access: Via assumed roles
  - MFA: Required for production changes
  - Password policy: 20+ chars, auto-expire 90 days
```

### 5. Compliance & Audit

#### Audit Logging

```go
type AuditEvent struct {
  ID             string    // UUID
  Timestamp      time.Time
  UserID         string
  WorkspaceID    string
  Action         string    // create, read, update, delete
  ResourceType   string    // user, document, workspace
  ResourceID     string
  Status         string    // success, failure
  StatusCode     int
  IPAddress      net.IP
  UserAgent      string
  Changes        map[string]interface{} // before/after
  ErrorMessage   string
}

// Immutable audit log
func LogAuditEvent(ctx context.Context, event AuditEvent) error {
  // Write to append-only table
  // Encrypt sensitive fields
  // Send to SIEM
  // No updates/deletes allowed
  return auditService.Log(ctx, event)
}
```

#### Compliance Frameworks

```yaml
SOC2 Type II:
  - CC (Common Criteria) controls
  - Annual audit
  - Test of controls
  
GDPR:
  - Data minimization
  - Right to deletion
  - Data portability
  - Privacy by design
  
HIPAA:
  - PHI encryption
  - Access logging
  - Business Associate Agreements
  - Breach notification
  
ISO 27001:
  - Information security policy
  - Asset management
  - Access control
  - Incident management
```

#### Vulnerability Management

```yaml
Scanning Schedule:
  - Code: On every commit (SAST)
  - Dependencies: Weekly (OWASP Dependency-Check)
  - Container images: On build (Trivy)
  - Infrastructure: Monthly (Trivy/Nessus)
  
Response Process:
  - Critical: Patch within 24 hours
  - High: Patch within 1 week
  - Medium: Patch within 2 weeks
  - Low: Patch within 30 days
  
Testing:
  - Penetration testing: Quarterly (third-party)
  - Bug bounty: Ongoing (HackerOne)
  - Vulnerability disclosure: Responsible disclosure policy
```

### 6. Incident Response

#### Incident Categories

```yaml
Severity 1 (Critical):
  - Data breach
  - Unauthorized access
  - Service unavailable (complete)
  - Security control failure
  
Severity 2 (High):
  - Elevated privileges granted
  - Malware detected
  - Service degradation
  - Unauthorized attempted access
  
Severity 3 (Medium):
  - Configuration drift
  - Failed access attempt
  - Suspicious activity
  
Severity 4 (Low):
  - Policy violation
  - Outdated software notice
```

#### Incident Response Plan

```yaml
Detection:
  - Automated alerts (SIEM)
  - User reports
  - Third-party notifications
  
Containment:
  - Isolate affected systems
  - Revoke compromised credentials
  - Block suspicious IPs
  - Preserve evidence
  
Investigation:
  - Forensic analysis
  - Timeline reconstruction
  - Root cause analysis
  - Impact assessment
  
Recovery:
  - Restore from clean backup
  - Patch vulnerabilities
  - Verify system integrity
  - Restore services
  
Post-Incident:
  - Post-mortem (24h)
  - Process improvements
  - Security enhancements
  - Stakeholder communication
```

### 7. Security Testing

#### Static Analysis (SAST)

```bash
# Go
golangci-lint run ./...

# Python
bandit -r .
pylint --security-check-enabled

# JavaScript
npm audit
eslint --plugin security
```

#### Dependency Scanning (SCA)

```bash
# Go
go list -json ./... | nancy sleuth

# Python
safety check
bandit

# JavaScript
snyk test
```

#### Dynamic Analysis (DAST)

```bash
# OWASP ZAP
zap-cli quick-scan --self-contained \
  -r zap-report.html \
  https://staging-api.example.com
```

## Security Checklist

- [ ] TLS 1.3 enabled on all endpoints
- [ ] Secrets not in source code (Vault)
- [ ] Input validation on all endpoints
- [ ] Output encoding for all responses
- [ ] CSRF tokens on state-changing requests
- [ ] CORS properly configured
- [ ] Rate limiting enabled
- [ ] Audit logging for all operations
- [ ] Encryption at rest for all data
- [ ] Encryption in transit (TLS)
- [ ] Database credentials rotated (monthly)
- [ ] API keys rotated (quarterly)
- [ ] Security headers enabled
- [ ] WAF rules configured
- [ ] DDoS protection enabled
- [ ] Multi-factor authentication available
- [ ] Session timeouts enforced
- [ ] Password policy enforced
- [ ] Backup encryption verified
- [ ] Disaster recovery tested
- [ ] Incident response plan documented
- [ ] Security training completed
- [ ] Vulnerability scanning automated
- [ ] Penetration testing scheduled
- [ ] Bug bounty program active