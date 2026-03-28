package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dfr/optitrade/src/internal/state/sqlite"
)

func TestDecisionLoggerPersistAndLog(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var logBuf bytes.Buffer
	log := slog.New(NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	var jsonlBuf bytes.Buffer
	dl := NewDecisionLogger(sqlite.NewStore(db), log, LoggerOptions{JSONL: &jsonlBuf})

	ctx := context.Background()
	rec := DecisionRecord{
		CorrelationID:    "corr-abcdef",
		DecisionType:     "risk_veto",
		RegimeLabel:      "volatile",
		CostModelVersion: "cm-v1",
		GateResults:      map[string]any{"delta_cap": false},
		Reason:           ReasonRiskVeto,
		TsMs:             time.UnixMilli(1700000000000).UnixMilli(),
		ID:               "fixed-audit-id",
	}

	if err := dl.LogDecision(ctx, rec); err != nil {
		t.Fatal(err)
	}

	store := sqlite.NewStore(db)
	got, err := store.GetAudit(ctx, "fixed-audit-id")
	if err != nil {
		t.Fatal(err)
	}
	if got.CorrelationID != rec.CorrelationID || got.Reason != string(ReasonRiskVeto) {
		t.Fatalf("audit row: %+v", got)
	}
	if !strings.Contains(got.RiskGateResults, "delta_cap") {
		t.Fatalf("gates json: %s", got.RiskGateResults)
	}

	logLine := strings.TrimSpace(logBuf.String())
	if !strings.Contains(logLine, rec.CorrelationID) {
		t.Fatalf("log missing correlation: %s", logLine)
	}

	envLine := strings.TrimSpace(jsonlBuf.String())
	var env map[string]any
	if err := json.Unmarshal([]byte(envLine), &env); err != nil {
		t.Fatal(err)
	}
	if env["schema_version"] != "1.0.0" {
		t.Fatalf("schema: %v", env["schema_version"])
	}
	if env["event_type"] != "risk_gate_result" {
		t.Fatalf("event_type: %v", env["event_type"])
	}
}

func TestLogDecisionShortCorrelationRejected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "short.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	dl := NewDecisionLogger(sqlite.NewStore(db), slog.Default(), LoggerOptions{})
	err = dl.LogDecision(context.Background(), DecisionRecord{
		CorrelationID: "short",
		DecisionType:  "approve",
		RegimeLabel:   "r",
		CostModelVersion: "1",
		Reason:        ReasonApproved,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLogDecisionDBFailureStillLogs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "fk.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var logBuf bytes.Buffer
	log := slog.New(NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	dl := NewDecisionLogger(sqlite.NewStore(db), log, LoggerOptions{})

	// Missing trade_candidate row for FK on candidate_id.
	badCand := "missing-cand"
	rec := DecisionRecord{
		CorrelationID:    "corr-xxxxxxxx",
		DecisionType:     "approve",
		CandidateID:      &badCand,
		RegimeLabel:      "n",
		CostModelVersion: "1",
		GateResults:      map[string]any{},
		Reason:           ReasonApproved,
	}

	err = dl.LogDecision(context.Background(), rec)
	if err == nil {
		t.Fatal("expected db error")
	}
	out := logBuf.String()
	if !strings.Contains(out, "audit_decision") {
		t.Fatalf("expected audit_decision log: %s", out)
	}
	if !strings.Contains(out, "audit_db_write_failed") {
		t.Fatalf("expected db failure log: %s", out)
	}
}
