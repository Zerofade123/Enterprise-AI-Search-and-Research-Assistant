# Auth Service

Production-grade authentication service with JWT, refresh tokens, and RBAC hooks.

## Features

- User registration and login
- JWT access tokens + refresh token rotation
- Session management
- Password hashing (bcrypt)
- API key storage hooks
- Clean architecture structure

## Structure

```
cmd/auth-service/main.go
internal/auth/
  domain/        # Entities
  ports/         # Interfaces
  service/       # Use cases
  transport/http # HTTP handlers
  infra/postgres # Postgres repositories
```
