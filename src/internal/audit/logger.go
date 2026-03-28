package audit

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/dfr/optitrade/src/internal/state"
)

// DecisionLogger writes audit rows and structured logs with a shared correlation_id (T051).
type DecisionLogger interface {
	LogDecision(ctx context.Context, rec DecisionRecord) error
}

// LoggerOptions configures JSONL and the structured logger sink.
type LoggerOptions struct {
	// JSONL, when non-nil, receives one JSON line per decision (T054).
	JSONL io.Writer
}

type decisionLogger struct {
	repo state.AuditRepository
	log  *slog.Logger
	opts LoggerOptions
}

// NewDecisionLogger returns a logger that persists to repo and emits slog records.
// Use [NewJSONHandler] for the slog handler so secrets are redacted (T053).
func NewDecisionLogger(repo state.AuditRepository, log *slog.Logger, opts LoggerOptions) DecisionLogger {
	if log == nil {
		log = slog.Default()
	}
	return &decisionLogger{repo: repo, log: log, opts: opts}
}

// LogDecision inserts into SQLite, logs at info, and optionally writes a JSONL envelope.
// If the DB insert fails, the decision is still logged and an error is returned (per WP11 risk note).
func (l *decisionLogger) LogDecision(ctx context.Context, rec DecisionRecord) error {
	row, err := rec.ToAuditDecision()
	if err != nil {
		return err
	}
	dbErr := l.repo.InsertAudit(ctx, row)

	l.log.InfoContext(ctx, "audit_decision",
		slog.String("correlation_id", row.CorrelationID),
		slog.String("audit_id", row.ID),
		slog.String("decision_type", row.DecisionType),
		slog.String("regime_label", row.RegimeLabel),
		slog.String("cost_model_version", row.CostModelVersion),
		slog.String("reason", row.Reason),
		slog.String("risk_gate_results", row.RiskGateResults),
		slog.Any("candidate_id", row.CandidateID),
	)

	if l.opts.JSONL != nil {
		line, encErr := MarshalEnvelopeJSON(rec, row)
		if encErr != nil {
			l.log.ErrorContext(ctx, "audit_jsonl_marshal_failed", slog.String("correlation_id", row.CorrelationID), slog.Any("err", encErr))
		} else {
			if _, wErr := fmt.Fprintf(l.opts.JSONL, "%s\n", line); wErr != nil {
				l.log.ErrorContext(ctx, "audit_jsonl_write_failed", slog.String("correlation_id", row.CorrelationID), slog.Any("err", wErr))
			}
		}
	}

	if dbErr != nil {
		l.log.ErrorContext(ctx, "audit_db_write_failed",
			slog.String("correlation_id", row.CorrelationID),
			slog.String("audit_id", row.ID),
			slog.Any("err", dbErr),
		)
		return fmt.Errorf("audit persist: %w", dbErr)
	}
	return nil
}
