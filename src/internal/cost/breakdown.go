package cost

// CostBreakdown captures inputs and the net edge for audit logs and cost_breakdown_json (WP08 T038).
// Numeric fields use float64 for the MVP scorer; money in production remains decimal strings at storage boundaries.
type CostBreakdown struct {
	CostModelVersion string  `json:"cost_model_version"`
	QuoteCurrency    string  `json:"quote_currency,omitempty"`

	ExpectedEdgeRaw string  `json:"expected_edge_raw"`
	ExpectedEdge    float64 `json:"expected_edge_parsed"`

	FeeBps          int     `json:"fee_bps"`
	HalfSpreadBps   float64 `json:"half_spread_bps"`
	SlippageBps     int     `json:"slippage_bps"`
	AdverseBps      int     `json:"adverse_selection_bps"`
	TotalHaircutBps float64 `json:"total_haircut_bps"`

	HaircutAmount  float64 `json:"haircut_amount"`
	EdgeAfterCosts float64 `json:"edge_after_costs"`
}
