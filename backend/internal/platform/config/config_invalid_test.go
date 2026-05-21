package config

import (
	"testing"
	"time"
)

func TestInvalidTTLRejected(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-secret")
	t.Setenv("JWT_KEY_ID", "kid")
	t.Setenv("JWT_ACCESS_TOKEN_TTL", "0s")
	t.Setenv("JWT_REFRESH_TOKEN_TTL", "168h")
	t.Setenv("JWT_MFA_TOKEN_TTL", "5m")
	if _, err := Load("test-service"); err == nil { t.Fatalf("expected error for invalid ttl") }
	_ = time.Second
}
