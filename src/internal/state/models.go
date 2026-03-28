package state

// SessionState mirrors the session FSM in data-model.md (state machine session).
// Runtime state lives in internal/session.FSM; persist to SQLite in a
// follow-up if multi-process or restarts require continuity.
type SessionState struct {
	Label     string
	UpdatedAt int64
	Reason    string
}

// OrderRecord matches order_record (data-model.md).
type OrderRecord struct {
	InternalOrderID string
	ExchangeOrderID *string
	InstrumentName  string
	Label           string
	Side            string
	OrderType       string
	Price           *string
	Amount          string
	PostOnly        bool
	ReduceOnly      bool
	State           string
	CreatedAt       int64
	UpdatedAt       int64
	CandidateID     *string
}

// RegimeState matches regime_state (data-model.md).
type RegimeState struct {
	ID                int64
	EffectiveAt       int64
	Label             string
	ClassifierVersion string
	InputsDigest      string
}

// FillRecord matches fill_record (data-model.md).
type FillRecord struct {
	ID             string
	OrderID        string
	TradeID        string
	InstrumentName string
	Qty            string
	Price          string
	Fee            string
	FilledAt       int64
}

// AuditDecision matches audit_decision (data-model.md).
type AuditDecision struct {
	ID               string
	Ts               int64
	DecisionType     string
	CandidateID      *string
	RegimeLabel      string
	CostModelVersion string
	RiskGateResults  string
	Reason           string
	CorrelationID    string
}
