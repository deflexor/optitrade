package audit

// DecisionReason is a stable veto/approval label stored in audit_decision.reason.
type DecisionReason string

const (
	ReasonApproved   DecisionReason = "approved"
	ReasonOrderSubmit DecisionReason = "order_submit"
	ReasonCostVeto   DecisionReason = "cost_veto"
	ReasonRiskVeto   DecisionReason = "risk_veto"
	ReasonRegimeVeto DecisionReason = "regime_veto"
	ReasonDataStale  DecisionReason = "data_stale"
	ReasonIVStale    DecisionReason = "iv_stale"
	ReasonConfig     DecisionReason = "config_error"
)

func (r DecisionReason) String() string {
	return string(r)
}
