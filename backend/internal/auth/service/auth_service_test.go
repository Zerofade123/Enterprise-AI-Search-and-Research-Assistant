package service

import (
	"context"
	"testing"
	"time"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/domain"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type memoryUserRepo struct{ users map[string]*domain.User }
func (m *memoryUserRepo) Create(ctx context.Context, user *domain.User) error { m.users[user.Email] = user; return nil }
func (m *memoryUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) { return m.users[email], nil }
func (m *memoryUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) { for _, u := range m.users { if u.ID == id { return u, nil } }; return nil, nil }
func (m *memoryUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error { return nil }

type memorySessionRepo struct{ sessions map[string]*domain.Session }
func (m *memorySessionRepo) Create(ctx context.Context, session *domain.Session) error { m.sessions[session.RefreshTokenHash] = session; return nil }
func (m *memorySessionRepo) GetByRefreshHash(ctx context.Context, hash string) (*domain.Session, error) { return m.sessions[hash], nil }
func (m *memorySessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) { for _, s := range m.sessions { if s.ID == id { return s, nil } }; return nil, nil }
func (m *memorySessionRepo) Revoke(ctx context.Context, sessionID uuid.UUID) error { for _, s := range m.sessions { if s.ID == sessionID { now := time.Now(); s.RevokedAt = &now } }; return nil }
func (m *memorySessionRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *memorySessionRepo) MarkRotated(ctx context.Context, sessionID uuid.UUID) error { for _, s := range m.sessions { if s.ID == sessionID { now := time.Now(); s.RotatedAt = &now } }; return nil }

type memoryKeyRepo struct{}
func (m *memoryKeyRepo) Create(ctx context.Context, key *domain.APIKey) error { return nil }
func (m *memoryKeyRepo) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) { return nil, nil }
func (m *memoryKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error { return nil }
func (m *memoryKeyRepo) Revoke(ctx context.Context, id uuid.UUID) error { return nil }

func TestAuthRegisterLoginRefresh(t *testing.T) {
	cfg := &config.AppConfig{JWT: config.JWTConfig{SigningKey: "test-secret", Issuer: "enterprise-ai-auth", KeyID: "test-key", AccessTokenTTL: time.Hour, RefreshTokenTTL: 24 * time.Hour, MfaTokenTTL: 5 * time.Minute}}
	logger := zap.NewNop()
	users := &memoryUserRepo{users: map[string]*domain.User{}}
	sessions := &memorySessionRepo{sessions: map[string]*domain.Session{}}
	keys := &memoryKeyRepo{}
	svc := NewAuthService(users, sessions, keys, cfg, logger)

	user, tokens, err := svc.Register(context.Background(), RegisterInput{Email: "  USER@example.com ", Password: "StrongPassw0rd!", FirstName: "Jane", LastName: "Doe"}, "127.0.0.1", "test")
	if err != nil { t.Fatalf("register error: %v", err) }
	if user.Email != "user@example.com" { t.Fatalf("expected normalized email") }
	if tokens.AccessToken == "" || tokens.RefreshToken == "" { t.Fatalf("expected tokens") }

	_, _, err = svc.Login(context.Background(), LoginInput{Email: "USER@example.com", Password: "StrongPassw0rd!"}, "127.0.0.1", "test")
	if err != nil { t.Fatalf("login error: %v", err) }

	newPair, err := svc.Refresh(context.Background(), tokens.RefreshToken, "127.0.0.1", "test")
	if err != nil { t.Fatalf("refresh error: %v", err) }
	if newPair.AccessToken == "" || newPair.RefreshToken == "" { t.Fatalf("expected new tokens") }
	if _, err := svc.Refresh(context.Background(), tokens.RefreshToken, "127.0.0.1", "test"); err == nil { t.Fatalf("expected replayed token rejection") }
}

func TestPasswordStrength(t *testing.T) {
	if err := ValidatePasswordStrength("weak"); err == nil { t.Fatalf("expected weak password error") }
	if err := ValidatePasswordStrength("StrongPassw0rd!"); err != nil { t.Fatalf("unexpected error: %v", err) }
}
