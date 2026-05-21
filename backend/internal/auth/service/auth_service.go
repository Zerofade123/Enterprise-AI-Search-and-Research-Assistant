package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/domain"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/ports"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/config"
	platformErrors "github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/errors"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/validation"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthService struct {
	users    ports.UserRepository
	sessions ports.SessionRepository
	apiKeys  ports.APIKeyRepository
	cfg      *config.AppConfig
	logger   *zap.Logger
	jwt      *JWTManager
}

func NewAuthService(users ports.UserRepository, sessions ports.SessionRepository, apiKeys ports.APIKeyRepository, cfg *config.AppConfig, logger *zap.Logger) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		apiKeys:  apiKeys,
		cfg:      cfg,
		logger:   logger,
		jwt:      NewJWTManager(cfg.JWT),
	}
}

type RegisterInput struct {
	Email     string `validate:"required,email"`
	Password  string `validate:"required"`
	FirstName string `validate:"required"`
	LastName  string `validate:"required"`
}

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput, ip, userAgent string) (*domain.User, *TokenPair, error) {
	if err := validation.ValidateStruct(input); err != nil {
		return nil, nil, err
	}

	if err := ValidatePasswordStrength(input.Password); err != nil {
		return nil, nil, err
	}

	existing, _ := s.users.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, nil, platformErrors.Wrap("auth.Register", platformErrors.CodeConflict, "email already registered", nil)
	}

	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, nil, platformErrors.Wrap("auth.Register", platformErrors.CodeInternal, "failed to hash password", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: hash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		MFAEnabled:   false,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	pair, err := s.issueTokens(ctx, user.ID, ip, userAgent)
	if err != nil {
		return nil, nil, err
	}

	return user, pair, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput, ip, userAgent string) (*domain.User, *TokenPair, error) {
	if err := validation.ValidateStruct(input); err != nil {
		return nil, nil, err
	}

	user, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, platformErrors.Wrap("auth.Login", platformErrors.CodeUnauthorized, "invalid credentials", nil)
	}

	if err := ComparePassword(user.PasswordHash, input.Password); err != nil {
		return nil, nil, platformErrors.Wrap("auth.Login", platformErrors.CodeUnauthorized, "invalid credentials", nil)
	}

	if user.MFAEnabled {
		return nil, nil, platformErrors.Wrap("auth.Login", platformErrors.CodeForbidden, "mfa required", nil)
	}

	if err := s.users.UpdateLastLogin(ctx, user.ID); err != nil {
		s.logger.Warn("failed to update last login", zap.Error(err))
	}

	pair, err := s.issueTokens(ctx, user.ID, ip, userAgent)
	if err != nil {
		return nil, nil, err
	}

	return user, pair, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken, ip, userAgent string) (*TokenPair, error) {
	if refreshToken == "" {
		return nil, platformErrors.Wrap("auth.Refresh", platformErrors.CodeValidation, "refresh token required", nil)
	}

	hash := hashToken(refreshToken)
	session, err := s.sessions.GetByRefreshHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if session == nil || session.RevokedAt != nil || session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, platformErrors.Wrap("auth.Refresh", platformErrors.CodeUnauthorized, "invalid refresh token", nil)
	}

	if err := s.sessions.Revoke(ctx, session.ID); err != nil {
		return nil, err
	}

	pair, err := s.issueTokens(ctx, session.UserID, ip, userAgent)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return platformErrors.Wrap("auth.Logout", platformErrors.CodeValidation, "refresh token required", nil)
	}

	hash := hashToken(refreshToken)
	session, err := s.sessions.GetByRefreshHash(ctx, hash)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}

	return s.sessions.Revoke(ctx, session.ID)
}

func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID, ip, userAgent string) (*TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccessToken(userID)
	if err != nil {
		return nil, platformErrors.Wrap("auth.issueTokens", platformErrors.CodeInternal, "failed to generate access token", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(userID)
	if err != nil {
		return nil, platformErrors.Wrap("auth.issueTokens", platformErrors.CodeInternal, "failed to generate refresh token", err)
	}

	session := &domain.Session{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: hashToken(refreshToken),
		IPAddress:        ip,
		UserAgent:        userAgent,
		ExpiresAt:        time.Now().UTC().Add(s.cfg.JWT.RefreshTokenTTL),
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.cfg.JWT.AccessTokenTTL,
	}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
