package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

// New returns an HTTP server with bounded transport timeouts.
func New(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// Run serves until cancellation, then drains requests within shutdownTimeout.
func Run(ctx context.Context, server *http.Server, shutdownTimeout time.Duration) error {
	result := make(chan error, 1)
	go func() {
		result <- server.ListenAndServe()
	}()

	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-result
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
