package httpmiddleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"
)

const requestIDHeader = "X-Request-ID"

type requestIDKey struct{}

var safeRequestID = regexp.MustCompile(`^[A-Za-z0-9_-]{8,128}$`)

// Chain installs request identity, access logging, and panic recovery without
// logging query strings, bodies, credentials, or bearer tokens.
func Chain(logger *slog.Logger, next http.Handler) http.Handler {
	return withRequestID(withAccessLog(logger, withRecovery(logger, next)))
}

// RequestID returns the validated or generated request ID for a request.
func RequestID(ctx context.Context) string {
	value, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return ""
	}
	return value
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if !safeRequestID.MatchString(requestID) {
			requestID = newRequestID()
		}
		w.Header().Set(requestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, requestID)))
	})
}

func withAccessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		response := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(response, r)
		logger.InfoContext(r.Context(), "http_request",
			"request_id", RequestID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", response.status,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}

func withRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(r.Context(), "unexpected_http_panic",
					"request_id", RequestID(r.Context()),
					"stack", string(debug.Stack()),
				)
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(map[string]any{
					"type":      "about:blank",
					"title":     "Internal Server Error",
					"status":    http.StatusInternalServerError,
					"code":      "INTERNAL_ERROR",
					"requestId": RequestID(r.Context()),
				}); err != nil {
					logger.WarnContext(r.Context(), "problem_response_write_failed",
						"request_id", RequestID(r.Context()),
					)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func newRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		// crypto/rand failures are exceptional; the time component still avoids
		// accepting or echoing an unsafe caller-controlled identifier.
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(value[:])
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
