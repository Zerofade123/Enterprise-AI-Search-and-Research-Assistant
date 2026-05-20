# Service Boundaries

## Overview

This document defines the clear boundaries, responsibilities, and interfaces for each microservice in the Enterprise AI Search & Research Assistant platform.

## Core Principles

1. **Single Responsibility**: Each service owns one business capability
2. **Autonomous Deployment**: Services deploy independently
3. **Database per Service**: Each service owns its data
4. **Clear Contracts**: gRPC/REST APIs define integration points
5. **Minimal Coupling**: Services communicate asynchronously when possible

---

## Service Inventory

### 1. API Gateway Service

**Purpose**: Central entry point for all external requests

**Responsibilities**:
- Request routing to appropriate services
- Authentication token validation
- Rate limiting and quota enforcement
- Request/response transformation
- Load balancing
- Distributed tracing initiation
- API versioning management

**Interfaces**:
```
External REST API: /api/v1/*
Internal: Routes to all services via gRPC/REST
```

**Dependencies**:
- Auth Service (token validation)
- All internal services (routing)
- Redis (rate limit tracking)
- Prometheus (metrics)

**Scaling**: Horizontal (stateless)

**Technology**: Go + Gin/Echo

---

### 2. Authentication Service

**Purpose**: Manage all authentication and authorization concerns

**Responsibilities**:
- OAuth2/OIDC provider integration
- JWT token generation and validation
- JWT refresh token management
- Multi-factor authentication (TOTP/SMS)
- Session management
- API key generation and revocation
- Password hashing and validation
- SSO integration

**Data Models**:
```sql
users
├── id (UUID)
├── email (unique)
├── password_hash
├── mfa_enabled
├── created_at
└── updated_at

auth_sessions
├── id (UUID)
├── user_id (FK)
├── refresh_token
├── expires_at
└── created_at

api_keys
├── id (UUID)
├── user_id (FK)
├── key_hash
├── name
├── last_used_at
├── expires_at
└── created_at
```

**REST Endpoints**:
```
POST   /auth/login           - Username/password login
POST   /auth/register         - User registration
POST   /auth/refresh          - Refresh JWT token
POST   /auth/logout           - Logout and invalidate token
POST   /auth/oauth/callback   - OAuth2 callback
POST   /auth/mfa/setup        - Setup MFA
POST   /auth/mfa/verify       - Verify MFA code
GET    /auth/me               - Get current user
GET    /auth/keys             - List API keys
POST   /auth/keys             - Create API key
DELETE /auth/keys/{id}        - Delete API key
```

**gRPC Endpoints**:
```proto
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (TokenInfo);
  rpc GetUser(GetUserRequest) returns (User);
  rpc RefreshToken(RefreshTokenRequest) returns (TokenResponse);
}
```

**Dependencies**:
- PostgreSQL (user storage)
- Redis (token blacklist, session cache)
- External OAuth providers (Google, GitHub, Microsoft)

**Scaling**: Horizontal

**Technology**: Go + PostgreSQL

---

### 3. Document Service

**Purpose**: Manage document lifecycle and metadata

**Responsibilities**:
- Document upload and storage
- Document versioning
- Metadata extraction and storage
- File validation and virus scanning
- Document deletion and soft deletes
- Access control enforcement
- Document search indexing triggers

**Data Models**:
```sql
documents
├── id (UUID)
├── workspace_id (FK)
├── name
├── file_size
├── mime_type
├── status (uploading|processing|ready|failed)
├── s3_key
├── created_by (FK)
├── created_at
├── updated_at
└── deleted_at (soft delete)

document_versions
├── id (UUID)
├── document_id (FK)
├── version_number
├── s3_key
├── created_at
└── created_by (FK)

document_metadata
├── id (UUID)
├── document_id (FK)
├── author
├── created_date
├── page_count
├── language
├── summary
├── tags
└── custom_fields (JSON)
```

**REST Endpoints**:
```
POST   /documents              - Upload document
GET    /documents              - List documents (paginated)
GET    /documents/{id}         - Get document details
GET    /documents/{id}/download - Download document
PUT    /documents/{id}         - Update document metadata
DELETE /documents/{id}         - Delete document
GET    /documents/{id}/versions - Get document versions
GET    /documents/{id}/metadata - Get metadata
POST   /documents/batch-upload - Bulk upload
```

**Dependencies**:
- AWS S3/MinIO (file storage)
- PostgreSQL (metadata)
- Auth Service (user verification)
- RabbitMQ (publish document.uploaded event)
- VirusTotal API (optional virus scanning)

**Scaling**: Horizontal

**Technology**: Go + PostgreSQL + S3

---

### 4. Search Service

**Purpose**: Semantic search and retrieval

**Responsibilities**:
- Semantic similarity search
- Vector similarity queries
- Filter and faceted search
- Search result ranking and reranking
- Search analytics and tracking
- Auto-completion and suggestions
- Full-text search fallback

**Data Models**:
```sql
search_indexes
├── id (UUID)
├── workspace_id (FK)
├── name
├── vector_dimension
├── metric_type (L2|cosine|inner_product)
└── status (indexing|ready|failed)

search_queries
├── id (UUID)
├── workspace_id (FK)
├── user_id (FK)
├── query
├── result_count
├── execution_time_ms
├── created_at
└── custom_fields (JSON)

search_results_cache
├── id (UUID)
├── query_hash
├── workspace_id
├── results (JSON)
├── ttl_seconds
└── created_at
```

**REST Endpoints**:
```
POST   /search                  - Execute semantic search
GET    /search/autocomplete     - Get suggestions
GET    /search/history          - Get search history
GET    /search/trending         - Get trending queries
POST   /search/filter           - Advanced filtered search
GET    /search/{id}/details     - Get search result details
```

**gRPC Endpoints**:
```proto
service SearchService {
  rpc SemanticSearch(SearchRequest) returns (SearchResponse);
  rpc GetRecommendations(RecommendationRequest) returns (RecommendationResponse);
}
```

**Dependencies**:
- Milvus/Weaviate (vector database)
- PostgreSQL (metadata)
- Redis (query cache, rate limiting)
- Document Service (document metadata)

**Scaling**: Horizontal

**Technology**: Go + Milvus

---

### 5. Workspace Service

**Purpose**: Multi-tenant workspace and organization management

**Responsibilities**:
- Workspace creation and management
- Member invitation and management
- Role and permission management
- Workspace settings (branding, features)
- Member activity tracking
- Workspace billing integration
- Team collaboration features

**Data Models**:
```sql
workspaces
├── id (UUID)
├── name
├── slug (unique)
├── owner_id (FK)
├── status (active|archived|deleted)
├── tier (free|professional|enterprise)
├── settings (JSON)
├── created_at
└── updated_at

workspace_members
├── id (UUID)
├── workspace_id (FK)
├── user_id (FK)
├── role (owner|admin|editor|viewer)
├── invited_at
├── joined_at
├── permissions (JSON - specific overrides)
└── updated_at

workspace_invitations
├── id (UUID)
├── workspace_id (FK)
├── invited_email
├── role
├── token
├── expires_at
├── created_at
└── accepted_at
```

**REST Endpoints**:
```
GET    /workspaces              - List user's workspaces
POST   /workspaces              - Create workspace
GET    /workspaces/{id}         - Get workspace details
PUT    /workspaces/{id}         - Update workspace
DELETE /workspaces/{id}         - Delete workspace
GET    /workspaces/{id}/members - List members
POST   /workspaces/{id}/members - Add member
PUT    /workspaces/{id}/members/{uid} - Update member role
DELETE /workspaces/{id}/members/{uid} - Remove member
POST   /workspaces/{id}/invitations - Send invitation
GET    /workspaces/{id}/invitations - List invitations
PUT    /workspaces/{id}/settings - Update settings
```

**Dependencies**:
- PostgreSQL (workspace data)
- Auth Service (user info)
- Email Service (invitations)
- Audit Service (logging)

**Scaling**: Horizontal

**Technology**: Go + PostgreSQL

---

### 6. User Service

**Purpose**: User profile and preference management

**Responsibilities**:
- User profile management
- User preferences storage
- Activity tracking
- Notification preferences
- Account settings
- Profile picture management

**Data Models**:
```sql
user_profiles
├── id (UUID)
├── user_id (FK)
├── first_name
├── last_name
├── bio
├── avatar_url
├── phone
├── department
├── job_title
├── timezone
└── updated_at

user_preferences
├── id (UUID)
├── user_id (FK)
├── theme (light|dark|auto)
├── language
├── email_notifications (boolean)
├── slack_integration (JSON)
├── preferences (JSON)
└── updated_at

user_activity
├── id (UUID)
├── user_id (FK)
├── workspace_id (FK)
├── action (login|upload|search|download)
├── metadata (JSON)
├── created_at
└── ip_address
```

**REST Endpoints**:
```
GET    /users/me                 - Get current user profile
PUT    /users/me                 - Update profile
GET    /users/me/preferences     - Get preferences
PUT    /users/me/preferences     - Update preferences
GET    /users/me/activity        - Get activity log
POST   /users/me/avatar          - Upload avatar
DELETE /users/me                 - Delete account
GET    /users/{id}               - Get user (public profile)
```

**Dependencies**:
- PostgreSQL (user data)
- S3 (avatar storage)
- Auth Service (user info)
- Audit Service (activity logging)

**Scaling**: Horizontal

**Technology**: Go + PostgreSQL

---

### 7. Notification Service

**Purpose**: Real-time notifications and event delivery

**Responsibilities**:
- WebSocket connection management
- Real-time event broadcasting
- Notification persistence
- Email notification sending
- SMS notification sending
- Notification preference enforcement
- Notification history

**Data Models**:
```sql
notifications
├── id (UUID)
├── user_id (FK)
├── workspace_id (FK)
├── type (document_ready|search_complete|task_done|system)
├── title
├── message
├── data (JSON)
├── read
├── created_at
└── read_at

websocket_sessions
├── session_id
├── user_id
├── workspace_id
├── connected_at
└── last_heartbeat
```

**WebSocket Events**:
```javascript
// Client -> Server
{
  "type": "subscribe",
  "channels": ["workspace:123", "user:456"]
}

// Server -> Client
{
  "type": "notification",
  "id": "uuid",
  "title": "Document Ready",
  "message": "Your document has been processed",
  "data": { ... }
}
```

**REST Endpoints**:
```
GET    /notifications            - Get notifications
GET    /notifications/{id}       - Get notification
PUT    /notifications/{id}/read  - Mark as read
DELETE /notifications/{id}       - Delete notification
GET    /notifications/preferences - Get preferences
PUT    /notifications/preferences - Update preferences
```

**Dependencies**:
- Redis (WebSocket sessions, message queue)
- PostgreSQL (notification storage)
- SendGrid/Twilio (email/SMS)
- RabbitMQ (event subscription)

**Scaling**: Horizontal with sticky sessions

**Technology**: Go + Redis

---

### 8. Audit Service

**Purpose**: Compliance and security audit logging

**Responsibilities**:
- Event logging for all actions
- Access logs for sensitive data
- Change tracking and history
- Compliance reporting
- Security incident logging
- Data retention policy enforcement

**Data Models**:
```sql
audit_logs
├── id (UUID)
├── workspace_id (FK)
├── user_id (FK)
├── action (create|read|update|delete|export)
├── resource_type (document|user|workspace)
├── resource_id
├── status (success|failure)
├── ip_address
├── user_agent
├── changes (JSON - before/after)
├── error_message
├── created_at
└── indexed_at

access_logs
├── id (UUID)
├── workspace_id (FK)
├── user_id (FK)
├── resource_type
├── resource_id
├── access_type (read|write|delete)
├── granted
├── created_at
└── ip_address
```

**gRPC Endpoints** (internal only):
```proto
service AuditService {
  rpc LogEvent(AuditEvent) returns (LogResponse);
  rpc GetAuditLog(AuditLogRequest) returns (AuditLogResponse);
  rpc ExportAuditLog(ExportRequest) returns (stream AuditEvent);
}
```

**REST Endpoints** (admin only):
```
GET    /admin/audit-logs         - Get audit logs
GET    /admin/audit-logs/export  - Export audit logs
GET    /admin/audit-logs/report  - Compliance report
```

**Dependencies**:
- PostgreSQL (audit storage)
- Elasticsearch (log indexing and search)
- S3 (long-term storage)
- All services (log events)

**Scaling**: Horizontal

**Technology**: Go + PostgreSQL + Elasticsearch

---

## Communication Patterns

### Synchronous (gRPC)
- Auth Service ← → all services
- API Gateway ← → Search Service
- API Gateway ← → Document Service

### Asynchronous (RabbitMQ)
- Document Service → document.uploaded → Document Processor
- Document Processor → embedding.requested → Embedding Service
- Embedding Service → embedding.completed → Search Service (index update)
- Document Service → document.deleted → Search Service (index cleanup)
- All services → audit.* → Audit Service

### Real-time (WebSocket)
- Client ← → Notification Service
- Notification Service ← RabbitMQ events

---

## Deployment Units

Each service is independently deployable:

```
services/
├── api-gateway/
│   ├── Dockerfile
│   ├── docker-compose.yml (local)
│   └── k8s/ (Kubernetes manifests)
├── auth-service/
├── document-service/
├── search-service/
├── workspace-service/
├── user-service/
├── notification-service/
└── audit-service/
```

---

## API Versioning

All REST APIs follow semantic versioning:

```
GET /api/v1/documents              # Current version
GET /api/v2/documents              # Next major version
```

**Deprecation Policy**:
- Major versions maintained for 12 months
- Deprecation warnings in headers
- Migration guides provided

---

## Service Contracts

All service-to-service contracts defined in Protocol Buffers:

```
proto/
├── auth/auth.proto
├── document/document.proto
├── search/search.proto
├── workspace/workspace.proto
├── user/user.proto
├── notification/notification.proto
├── audit/audit.proto
└── common/common.proto
```

Contracts are versioned and changes require explicit migration.