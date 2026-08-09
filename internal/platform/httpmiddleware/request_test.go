package httpmiddleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tgwtea/penny-backend/internal/platform/httpmiddleware"
)

func TestChainAcceptsSafeRequestIDAndDoesNotLogQuery(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := httpmiddleware.Chain(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/health/live?token=secret", nil)
	request.Header.Set("X-Request-ID", "safe-request-123")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if got := response.Header().Get("X-Request-ID"); got != "safe-request-123" {
		t.Errorf("X-Request-ID = %q, want safe-request-123", got)
	}
	if bytes.Contains(logs.Bytes(), []byte("secret")) {
		t.Fatal("access log contains query-string secret")
	}
}

func TestChainRecoversWithoutExposingPanic(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
	handler := httpmiddleware.Chain(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("sensitive detail")
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	var problem map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if problem["code"] != "INTERNAL_ERROR" {
		t.Errorf("code = %v, want INTERNAL_ERROR", problem["code"])
	}
	if bytes.Contains(response.Body.Bytes(), []byte("sensitive")) {
		t.Fatal("response exposes panic detail")
	}
}
