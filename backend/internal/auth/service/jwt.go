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

func (m *JWTManager) GenerateAccessToken(userID, workspaceID uuid.UUID, role string, permissions []string) (string, error) {
	claims := jwt.MapClaims{
		"sub":         userID.String(),
		"workspace_id": workspaceID.String(),
		"role":        role,
		"permissions": permissions,
		"iss":         m.cfg.Issuer,
		"aud":         "enterprise-ai-api",
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(m.cfg.AccessTokenTTL).Unix(),
		"nbf":         time.Now().Unix(),
		"jti":         uuid.NewString(),
		"kid":         m.cfg.KeyID,
		"type":        "access",
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.cfg.SigningKey))
}

func (m *JWTManager) GenerateRefreshToken(userID, sessionID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"sid":   sessionID.String(),
		"iss":   m.cfg.Issuer,
		"aud":   "enterprise-ai-api",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(m.cfg.RefreshTokenTTL).Unix(),
		"jti":   uuid.NewString(),
		"kid":   m.cfg.KeyID,
		"type":  "refresh",
		"nbf":   time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.cfg.SigningKey))
}

type TokenClaims struct {
	UserID      string
	WorkspaceID string
	Role        string
	Permissions []string
	SessionID   string
	TokenType   string
	JTI         string
	KID         string
}

func (m *JWTManager) ParseAndValidate(tokenString string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(m.cfg.SigningKey), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	out := &TokenClaims{}
	if v, ok := claims["sub"].(string); ok { out.UserID = v }
	if v, ok := claims["workspace_id"].(string); ok { out.WorkspaceID = v }
	if v, ok := claims["role"].(string); ok { out.Role = v }
	if v, ok := claims["sid"].(string); ok { out.SessionID = v }
	if v, ok := claims["type"].(string); ok { out.TokenType = v }
	if v, ok := claims["jti"].(string); ok { out.JTI = v }
	if v, ok := claims["kid"].(string); ok { out.KID = v }
	if raw, ok := claims["permissions"].([]interface{}); ok {
		out.Permissions = make([]string, 0, len(raw))
		for _, p := range raw {
			if s, ok := p.(string); ok {
				out.Permissions = append(out.Permissions, s)
			}
		}
	}
	return out, nil
}
