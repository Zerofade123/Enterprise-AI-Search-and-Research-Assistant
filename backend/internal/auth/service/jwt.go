package service

import (
	"time"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	cfg config.JWTConfig
}

func NewJWTManager(cfg config.JWTConfig) *JWTManager {
	return &JWTManager{cfg: cfg}
}

func (m *JWTManager) GenerateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iss": m.cfg.Issuer,
		"aud": "enterprise-ai-api",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(m.cfg.AccessTokenTTL).Unix(),
		"type": "access",
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.cfg.SigningKey))
}

func (m *JWTManager) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iss": m.cfg.Issuer,
		"aud": "enterprise-ai-api",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(m.cfg.RefreshTokenTTL).Unix(),
		"type": "refresh",
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.cfg.SigningKey))
}
