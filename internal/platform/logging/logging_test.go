package logging_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/tgwtea/penny-backend/internal/platform/logging"
)

func TestNewWritesJSON(t *testing.T) {
	var output bytes.Buffer
	logger, err := logging.New(&output, "info")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("started", "service", "api")
	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("log is not JSON: %v", err)
	}
	if entry["service"] != "api" {
		t.Errorf("service = %v, want api", entry["service"])
	}
}

func TestNewRejectsUnknownLevel(t *testing.T) {
	if _, err := logging.New(&bytes.Buffer{}, "verbose"); err == nil {
		t.Fatal("New() error = nil, want unsupported level")
	}
}
