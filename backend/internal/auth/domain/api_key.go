package domain

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	KeyHash   string
	CreatedAt time.Time
	LastUsedAt *time.Time
	ExpiresAt *time.Time
	RevokedAt *time.Time
}
