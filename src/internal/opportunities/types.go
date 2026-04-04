package opportunities

// RowStatus is the lifecycle of an opportunity row (spec 2026-04-04).
type RowStatus string

const (
	StatusCandidate RowStatus = "candidate"
	StatusOpening   RowStatus = "opening"
	StatusActive    RowStatus = "active"
	StatusPartial   RowStatus = "partial"
)

// LegQuote is one leg with touch prices for the UI.
type LegQuote struct {
	Instrument string  `json:"instrument"`
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
}

// Row is one ranked opportunity (candidate) for API and runner snapshots.
type Row struct {
	ID           string     `json:"id"`
	StrategyName string     `json:"strategy_name"`
	Status       RowStatus  `json:"status"`
	Legs         []LegQuote `json:"legs"`
	GreeksNote   string     `json:"greeks_note,omitempty"`
	MaxProfit    string     `json:"max_profit"`
	MaxLoss      string     `json:"max_loss"`
	Recommend    string     `json:"recommendation"`
	Rationale    string     `json:"rationale"`
	ExpectedEdge string     `json:"expected_edge"`
	EdgeAfter    float64    `json:"edge_after_costs"`
}

// Snapshot is the selector output for one tick.
type Snapshot struct {
	UpdatedAtMs int64 `json:"updated_at_ms"`
	Rows        []Row `json:"rows"`
}
