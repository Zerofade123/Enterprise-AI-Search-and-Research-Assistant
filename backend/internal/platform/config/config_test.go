package config

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("JWT_SIGNING_KEY", "test-secret")
	t.Setenv("JWT_ISSUER", "enterprise-ai-auth")
	t.Setenv("JWT_KEY_ID", "test-key")
	t.Setenv("JWT_ACCESS_TOKEN_TTL", "1h")
	t.Setenv("JWT_REFRESH_TOKEN_TTL", "168h")
	t.Setenv("JWT_MFA_TOKEN_TTL", "5m")

	cfg, err := Load("test-service")
	if err != nil { t.Fatalf("expected no error, got %v", err) }
	if cfg.HTTP.Port != 8080 { t.Fatalf("expected default http port, got %d", cfg.HTTP.Port) }
	if cfg.JWT.AccessTokenTTL != time.Hour { t.Fatalf("expected access token ttl 1h, got %v", cfg.JWT.AccessTokenTTL) }
}
