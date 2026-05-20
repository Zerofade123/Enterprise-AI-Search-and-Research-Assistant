# API Structure

## Overview

The platform provides both REST and gRPC APIs for different use cases. REST APIs serve external clients, while gRPC is used for internal service-to-service communication.

## API Gateway

### Base URL
```
Production:  https://api.enterprise-ai-search.com/api/v1
Staging:     https://staging-api.enterprise-ai-search.com/api/v1
Development: http://localhost:8000/api/v1
```

### Authentication

All requests require authentication via one of:

1. **JWT Token** (Bearer Token)
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

2. **API Key** (Header)
```http
X-API-Key: sk_live_1234567890abcdef
```

3. **OAuth2** (Client Credentials)
```http
Authorization: Bearer <oauth2_token>
```

### Common Headers
```http
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
X-Workspace-ID: workspace-uuid
X-Version: 1
Accept: application/json
Content-Type: application/json
```

### Response Format

All responses follow a consistent format:

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2024-05-20T10:30:00Z",
    "version": "1"
  }
}
```

Error Response:
```json
{
  "success": false,
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Document not found",
    "details": {
      "resource_type": "document",
      "resource_id": "doc-123"
    }
  },
  "meta": { ... }
}
```

## REST API Endpoints

### Authentication Endpoints

#### POST /auth/login
User login with email and password
```json
Request:
{
  "email": "user@example.com",
  "password": "secure_password",
  "remember_me": true
}

Response (200):
{
  "success": true,
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": { ... }
  }
}
```

#### POST /auth/register
New user registration
```json
Request:
{
  "email": "user@example.com",
  "password": "secure_password",
  "first_name": "John",
  "last_name": "Doe",
  "workspace_name": "My Organization"
}

Response (201):
{
  "success": true,
  "data": {
    "user": { ... },
    "workspace": { ... },
    "access_token": "..."
  }
}
```

#### POST /auth/refresh
Refresh access token
```json
Request:
{
  "refresh_token": "..."
}

Response (200):
{
  "success": true,
  "data": {
    "access_token": "...",
    "expires_in": 3600
  }
}
```

### Document Endpoints

#### POST /documents
Upload a new document
```http
Content-Type: multipart/form-data

Request:
file: <binary>
metadata: {"tags": ["report", "2024"]}

Response (201):
{
  "success": true,
  "data": {
    "id": "doc-123",
    "name": "report.pdf",
    "status": "uploading",
    "size": 1024000,
    "created_at": "2024-05-20T10:00:00Z"
  }
}
```

#### GET /documents
List documents
```http
GET /documents?page=1&limit=20&sort=-created_at&filter[status]=ready

Response (200):
{
  "success": true,
  "data": [
    {
      "id": "doc-123",
      "name": "report.pdf",
      "status": "ready",
      "size": 1024000,
      "created_at": "2024-05-20T10:00:00Z"
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "pages": 8
    }
  }
}
```

#### GET /documents/{id}
Get document details
```http
Response (200):
{
  "success": true,
  "data": {
    "id": "doc-123",
    "name": "report.pdf",
    "status": "ready",
    "size": 1024000,
    "mime_type": "application/pdf",
    "metadata": {
      "author": "John Doe",
      "created_date": "2024-05-15",
      "pages": 42
    },
    "chunks_count": 15,
    "created_at": "2024-05-20T10:00:00Z",
    "updated_at": "2024-05-20T10:05:00Z"
  }
}
```

#### GET /documents/{id}/download
Download document
```http
Response (200):
Content-Type: application/pdf
Content-Disposition: attachment; filename="report.pdf"
<binary data>
```

#### PUT /documents/{id}
Update document metadata
```json
Request:
{
  "name": "new-name.pdf",
  "metadata": {
    "tags": ["updated", "2024"],
    "department": "Finance"
  }
}

Response (200):
{
  "success": true,
  "data": { ... }
}
```

#### DELETE /documents/{id}
Delete document
```http
Response (204): No Content
```

### Search Endpoints

#### POST /search
Execute semantic search
```json
Request:
{
  "query": "quarterly financial results",
  "limit": 10,
  "filters": {
    "document_types": ["pdf"],
    "date_range": {
      "start": "2024-01-01",
      "end": "2024-05-31"
    }
  }
}

Response (200):
{
  "success": true,
  "data": {
    "results": [
      {
        "id": "chunk-123",
        "document_id": "doc-123",
        "document_name": "report.pdf",
        "content_preview": "The quarterly financial results show...",
        "score": 0.92,
        "chunk_position": 5,
        "metadata": { ... }
      }
    ],
    "total": 42,
    "query_time_ms": 234
  }
}
```

#### GET /search/autocomplete
Get search suggestions
```http
GET /search/autocomplete?q=financial&limit=5

Response (200):
{
  "success": true,
  "data": [
    {"text": "financial results", "frequency": 45},
    {"text": "financial planning", "frequency": 32},
    {"text": "financial statements", "frequency": 28}
  ]
}
```

### Workspace Endpoints

#### GET /workspaces
List user's workspaces
```http
Response (200):
{
  "success": true,
  "data": [
    {
      "id": "ws-123",
      "name": "Finance Team",
      "slug": "finance-team",
      "role": "admin",
      "tier": "professional",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

#### POST /workspaces
Create new workspace
```json
Request:
{
  "name": "Marketing Team",
  "slug": "marketing-team"
}

Response (201):
{
  "success": true,
  "data": {
    "id": "ws-124",
    "name": "Marketing Team",
    "slug": "marketing-team",
    "tier": "free",
    "created_at": "2024-05-20T10:00:00Z"
  }
}
```

#### GET /workspaces/{id}/members
List workspace members
```http
Response (200):
{
  "success": true,
  "data": [
    {
      "id": "mem-123",
      "user_id": "user-456",
      "email": "john@example.com",
      "name": "John Doe",
      "role": "admin",
      "joined_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

#### POST /workspaces/{id}/members
Invite member to workspace
```json
Request:
{
  "email": "jane@example.com",
  "role": "editor"
}

Response (201):
{
  "success": true,
  "data": {
    "id": "inv-123",
    "email": "jane@example.com",
    "role": "editor",
    "status": "pending",
    "expires_at": "2024-05-27T10:00:00Z"
  }
}
```

## gRPC APIs

### Service Definitions

All gRPC services defined in Protocol Buffers:

#### Auth Service
```proto
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (TokenInfo);
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateAPIKey(CreateAPIKeyRequest) returns (APIKey);
  rpc RefreshToken(RefreshTokenRequest) returns (TokenResponse);
}
```

#### Document Service
```proto
service DocumentService {
  rpc UploadDocument(stream UploadRequest) returns (UploadResponse);
  rpc GetDocument(GetDocumentRequest) returns (Document);
  rpc ListDocuments(ListDocumentsRequest) returns (DocumentList);
  rpc DeleteDocument(DeleteDocumentRequest) returns (Empty);
  rpc GetChunks(GetChunksRequest) returns (ChunkList);
}
```

#### Search Service
```proto
service SearchService {
  rpc Search(SearchRequest) returns (SearchResponse);
  rpc Autocomplete(AutocompleteRequest) returns (AutocompleteResponse);
  rpc GetRecommendations(RecommendationRequest) returns (RecommendationResponse);
}
```

## Error Codes

### Standard Error Codes

| Code | HTTP | Description |
|------|------|-------------|
| SUCCESS | 200 | Request successful |
| BAD_REQUEST | 400 | Invalid request |
| UNAUTHORIZED | 401 | Missing or invalid auth |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| CONFLICT | 409 | Resource conflict |
| RATE_LIMITED | 429 | Rate limit exceeded |
| INTERNAL_ERROR | 500 | Server error |
| SERVICE_UNAVAILABLE | 503 | Service down |

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "You don't have permission to access this resource",
    "request_id": "uuid",
    "details": {
      "required_role": "admin",
      "user_role": "editor"
    }
  }
}
```

## Rate Limiting

### Limits by Plan

| Plan | Requests/min | Documents/month | Search/day |
|------|--------------|-----------------|------------|
| Free | 60 | 10 | 100 |
| Pro | 1000 | 1000 | 10000 |
| Enterprise | Custom | Unlimited | Unlimited |

### Rate Limit Headers
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 987
X-RateLimit-Reset: 1716194400
```

## Pagination

### Query Parameters
```http
GET /documents?page=2&limit=50&sort=-created_at

page:   Page number (1-indexed)
limit:  Items per page (default: 20, max: 100)
sort:   Sort field with +/- prefix (default: -created_at)
```

### Response
```json
{
  "data": [ ... ],
  "meta": {
    "pagination": {
      "page": 2,
      "limit": 50,
      "total": 5000,
      "pages": 100,
      "has_next": true,
      "has_prev": true
    }
  }
}
```

## Versioning

API versioning through URL path:
```
GET /api/v1/documents   # Current version
GET /api/v2/documents   # Next major version
```

Deprecation timeline:
- Announced: 6 months advance notice
- Deprecated: Available for 12 months
- Sunset: Removed after 12 months

## Webhooks

### Supported Events
```
document.uploaded
document.processed
document.failed
search.completed
task.started
task.completed
task.failed
```

### Webhook Payload
```json
{
  "event": "document.processed",
  "timestamp": "2024-05-20T10:00:00Z",
  "data": {
    "document_id": "doc-123",
    "status": "ready",
    "chunks_created": 42
  },
  "signature": "sha256=..."
}
```