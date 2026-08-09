package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/tgwtea/penny-backend/internal/platform/buildinfo"
	"github.com/tgwtea/penny-backend/internal/platform/httpmiddleware"
)

// ReadinessChecker is implemented by PostgreSQL and remains narrow enough for
// deterministic handler tests.
type ReadinessChecker interface {
	Ping(context.Context) error
}

// Server implements generated operational handlers only.
type Server struct {
	checker ReadinessChecker
	info    buildinfo.Info
	timeout time.Duration
}

func NewServer(checker ReadinessChecker, info buildinfo.Info, timeout time.Duration) *Server {
	return &Server{checker: checker, info: info, timeout: timeout}
}

// NewHandler assembles the generated router with shared safe middleware.
func NewHandler(server *Server, logger *slog.Logger) http.Handler {
	return httpmiddleware.Chain(logger, Handler(server))
}

func (s *Server) GetHealthLive(w http.ResponseWriter, _ *http.Request) {
	if err := writeJSON(w, http.StatusOK, s.healthResponse("alive")); err != nil {
		return
	}
}

func (s *Server) GetHealthReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()
	if err := s.checker.Ping(ctx); err != nil {
		if writeErr := writeProblem(w, http.StatusServiceUnavailable, Problem{
			Type:   "about:blank",
			Title:  "Service Unavailable",
			Status: http.StatusServiceUnavailable,
			Code:   "SERVICE_UNAVAILABLE",
		}); writeErr != nil {
			return
		}
		return
	}
	if err := writeJSON(w, http.StatusOK, s.healthResponse("ready")); err != nil {
		return
	}
}

func (s *Server) healthResponse(status HealthResponseStatus) HealthResponse {
	return HealthResponse{
		Status:           status,
		Service:          Api,
		Version:          s.info.Version,
		Commit:           s.info.Commit,
		ArtifactDigest:   s.info.ArtifactDigest,
		ContractVersion:  s.info.ContractVersion,
		ProductMilestone: s.info.ProductMilestone,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}

func writeProblem(w http.ResponseWriter, status int, problem Problem) error {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(problem)
}
