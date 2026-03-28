package audit

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestRedactingHandlerScrubsSensitiveKeys(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	h := NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	log := slog.New(h)

	log.Info("auth",
		slog.String("client_secret", "super-secret"),
		slog.String("api_token", "abc"),
		slog.String("safe", "ok"),
	)

	var m map[string]any
	line := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatal(err)
	}
	if m["client_secret"] != "[REDACTED]" {
		t.Fatalf("client_secret: %v", m["client_secret"])
	}
	if m["api_token"] != "[REDACTED]" {
		t.Fatalf("api_token: %v", m["api_token"])
	}
	if m["safe"] != "ok" {
		t.Fatalf("safe: %v", m["safe"])
	}
}

func TestRedactingHandlerNestedGroup(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slog.New(NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	log.Info("nested", slog.Group("creds",
		slog.String("password", "hunter2"),
		slog.String("label", "x"),
	))

	raw := buf.String()
	if strings.Contains(raw, "hunter2") {
		t.Fatalf("password leaked: %s", raw)
	}
	if !strings.Contains(raw, "[REDACTED]") {
		t.Fatalf("expected redaction in: %s", raw)
	}
}
