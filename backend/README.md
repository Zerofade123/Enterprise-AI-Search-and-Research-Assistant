# Backend Services (Go)

This directory contains Go services implemented using clean architecture principles.

Planned services:
- API Gateway
- Authentication Service ✅
- Workspace Service
- Document Service
- Search Service
- Audit Service

Each service will have:
- `cmd/<service>/main.go` entrypoint
- `internal/<service>` domain + use cases
- `internal/platform` shared infrastructure
