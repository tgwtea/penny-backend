package api_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tgwtea/penny-backend/internal/api"
	"github.com/tgwtea/penny-backend/internal/platform/buildinfo"
)

func TestLiveDoesNotCheckDatabase(t *testing.T) {
	checker := &fakeChecker{err: errors.New("database detail")}
	handler := newHandler(checker)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if checker.calls != 0 {
		t.Fatalf("Ping() calls = %d, want 0", checker.calls)
	}
}

func TestReadyChecksDatabase(t *testing.T) {
	checker := &fakeChecker{}
	handler := newHandler(checker)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if checker.calls != 1 {
		t.Fatalf("Ping() calls = %d, want 1", checker.calls)
	}
}

func TestReadyReturnsGenericProblem(t *testing.T) {
	checker := &fakeChecker{err: errors.New("password=must-not-leak")}
	handler := newHandler(checker)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	if response.Header().Get("Content-Type") != "application/problem+json" {
		t.Errorf("Content-Type = %q", response.Header().Get("Content-Type"))
	}
	if bytes.Contains(response.Body.Bytes(), []byte("password")) {
		t.Fatal("response leaks database error")
	}
}

func TestEmbeddedOpenAPIContractIsValid(t *testing.T) {
	contract, err := api.GetSwagger()
	if err != nil {
		t.Fatalf("GetSwagger() error = %v", err)
	}
	if err := contract.Validate(context.Background()); err != nil {
		t.Fatalf("OpenAPI validation error = %v", err)
	}
}

func newHandler(checker *fakeChecker) http.Handler {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
	server := api.NewServer(checker, buildinfo.Info{
		Version:          "test",
		Commit:           "abc123",
		ArtifactDigest:   "sha256:test",
		ContractVersion:  "1.0.0",
		ProductMilestone: "test",
	}, time.Second)
	return api.NewHandler(server, logger)
}

type fakeChecker struct {
	err   error
	calls int
}

func (f *fakeChecker) Ping(context.Context) error {
	f.calls++
	return f.err
}
