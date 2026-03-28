package audit

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/dfr/optitrade/execution/internal/state"
)

// DecisionRecord is the in-memory audit payload (T052) before persistence and logging.
type DecisionRecord struct {
	CorrelationID string
	DecisionType  string
	CandidateID   *string

	RegimeLabel      string
	CostModelVersion string
	GateResults      map[string]bool
	Reason           DecisionReason

	// EnvelopeEventType optionally overrides mapping from DecisionType for JSONL (T054).
	EnvelopeEventType string

	TsMs int64
	ID   string
}

var errCorrelationID = errors.New("audit: correlation_id must be at least 8 characters")

// ToAuditDecision maps the record into a row for audit_decision.
func (r DecisionRecord) ToAuditDecision() (*state.AuditDecision, error) {
	if len(r.CorrelationID) < 8 {
		return nil, errCorrelationID
	}
	id := r.ID
	if id == "" {
		id = uuid.NewString()
	}
	ts := r.TsMs
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}
	gateJSON := []byte("{}")
	if r.GateResults != nil {
		var err error
		gateJSON, err = json.Marshal(r.GateResults)
		if err != nil {
			return nil, err
		}
	}
	return &state.AuditDecision{
		ID:               id,
		Ts:               ts,
		DecisionType:     r.DecisionType,
		CandidateID:      r.CandidateID,
		RegimeLabel:      r.RegimeLabel,
		CostModelVersion: r.CostModelVersion,
		RiskGateResults:  string(gateJSON),
		Reason:           string(r.Reason),
		CorrelationID:    r.CorrelationID,
	}, nil
}
