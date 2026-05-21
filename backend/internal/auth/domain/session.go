package auth

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash  string
	RefreshFamilyID   string
	IPAddress        string
	UserAgent        string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	RevokedAt        *time.Time
	RotatedAt        *time.Time
}
