package state

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
