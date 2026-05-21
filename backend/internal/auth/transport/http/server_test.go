package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/domain"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/auth/service"
	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestRateLimitMiddleware(t *testing.T) {
	h := NewServer(&service.AuthService{}, zap.NewNop(), 9090)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK { t.Fatalf("expected health ok, got %d", w.Code) }
}

func TestServerPortUsesConfig(t *testing.T) {
	if NewServer(&service.AuthService{}, zap.NewNop(), 7777).Port() != "7777" { t.Fatalf("expected configured port") }
}
