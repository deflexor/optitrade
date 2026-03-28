package audit

import (
	"encoding/json"
	"errors"

	"github.com/dfr/optitrade/execution/internal/state"
)

// EventEnvelope matches contracts/event-envelope.schema.json (payload for JSONL, T054).
type EventEnvelope struct {
	SchemaVersion string         `json:"schema_version"`
	TsMs          int64          `json:"ts_ms"`
	CorrelationID string         `json:"correlation_id"`
	EventType     string         `json:"event_type"`
	Payload       map[string]any `json:"payload"`
}

const envelopeSchemaVersion = "1.0.0"

var allowedEventTypes = map[string]struct{}{
	"regime_changed":              {},
	"candidate_evaluated":         {},
	"risk_gate_result":            {},
	"order_submitted":             {},
	"order_terminal":              {},
	"protective_mode_entered":     {},
	"session_state_changed":       {},
}

func envelopeEventTypeFor(d DecisionRecord) string {
	if d.EnvelopeEventType != "" {
		return d.EnvelopeEventType
	}
	switch d.DecisionType {
	case "risk_veto", "veto_risk":
		return "risk_gate_result"
	case "cost_veto", "veto_cost", "approve", "approved", "":
		return "candidate_evaluated"
	default:
		return "candidate_evaluated"
	}
}

// MarshalEnvelopeJSON returns one JSON object line for JSONL sinks.
func MarshalEnvelopeJSON(d DecisionRecord, row *state.AuditDecision) ([]byte, error) {
	if row == nil {
		return nil, errors.New("audit: nil audit row for envelope")
	}
	et := envelopeEventTypeFor(d)
	if _, ok := allowedEventTypes[et]; !ok {
		return nil, errors.New("audit: unknown event_type for envelope: " + et)
	}
	gates := map[string]any{}
	for k, v := range d.GateResults {
		gates[k] = v
	}
	env := EventEnvelope{
		SchemaVersion: envelopeSchemaVersion,
		TsMs:          row.Ts,
		CorrelationID: row.CorrelationID,
		EventType:     et,
		Payload: map[string]any{
			"decision_type":      row.DecisionType,
			"regime_label":       row.RegimeLabel,
			"cost_model_version": row.CostModelVersion,
			"gate_results":       gates,
			"reason":             row.Reason,
		},
	}
	if row.CandidateID != nil {
		env.Payload["candidate_id"] = *row.CandidateID
	}
	return json.Marshal(env)
}
