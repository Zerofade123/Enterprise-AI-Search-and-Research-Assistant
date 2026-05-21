package service

import (
	"context"
	"testing"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/domain"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type memoryUserRepo struct {
	users map[string]*domain.User
}

func (m *memoryUserRepo) Create(ctx context.Context, user *domain.User) error {
	m.users[user.Email] = user
	return nil
}
func (m *memoryUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.users[email], nil
}
func (m *memoryUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}
func (m *memoryUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return nil
}

type memorySessionRepo struct {
	sessions map[string]*domain.Session
}

func (m *memorySessionRepo) Create(ctx context.Context, session *domain.Session) error {
	m.sessions[session.RefreshTokenHash] = session
	return nil
}
func (m *memorySessionRepo) GetByRefreshHash(ctx context.Context, hash string) (*domain.Session, error) {
	return m.sessions[hash], nil
}
func (m *memorySessionRepo) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	for _, s := range m.sessions {
		if s.ID == sessionID {
			now := s.CreatedAt
			s.RevokedAt = &now
		}
	}
	return nil
}
func (m *memorySessionRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}

type memoryKeyRepo struct{}

func (m *memoryKeyRepo) Create(ctx context.Context, key *domain.APIKey) error { return nil }
func (m *memoryKeyRepo) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	return nil, nil
}
func (m *memoryKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error { return nil }
func (m *memoryKeyRepo) Revoke(ctx context.Context, id uuid.UUID) error { return nil }

func TestAuthRegisterLoginRefresh(t *testing.T) {
	cfg := &config.AppConfig{JWT: config.JWTConfig{
		SigningKey:      "test-secret",
		Issuer:          "enterprise-ai-auth",
		AccessTokenTTL:  3600,
		RefreshTokenTTL: 3600,
		MfaTokenTTL:     300,
	}}
	cfg.HTTP.Port = 8080

	logger := zap.NewNop()

	users := &memoryUserRepo{users: map[string]*domain.User{}}
	sessions := &memorySessionRepo{sessions: map[string]*domain.Session{}}
	keys := &memoryKeyRepo{}

	svc := NewAuthService(users, sessions, keys, cfg, logger)

	user, tokens, err := svc.Register(context.Background(), RegisterInput{
		Email:     "user@example.com",
		Password:  "StrongPassw0rd!",
		FirstName: "Jane",
		LastName:  "Doe",
	}, "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	if user.ID == uuid.Nil || tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected valid user and tokens")
	}

	_, _, err = svc.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "StrongPassw0rd!",
	}, "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("login error: %v", err)
	}

	newPair, err := svc.Refresh(context.Background(), tokens.RefreshToken, "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("refresh error: %v", err)
	}
	if newPair.AccessToken == "" || newPair.RefreshToken == "" {
		t.Fatalf("expected new tokens")
	}
}
