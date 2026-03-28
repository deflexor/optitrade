package risk

import (
	"math/big"
	"time"

	"github.com/dfr/optitrade/execution/internal/deribit"
)

// CandidateRisk is trade metadata for per-trade max loss and instrument-level order counts.
type CandidateRisk struct {
	ID           string
	StrategyID   string
	MaxLossQuote string
	FeesQuote    string
	Instruments  []string
}

// PreTradeInput is the full context for a pre-trade risk check.
type PreTradeInput struct {
	CorrelationID    string
	CandidateID      *string
	RegimeLabel      string
	CostModelVersion string

	Positions  []deribit.Position
	OpenOrders []deribit.OpenOrder

	DeltaAdjustment *big.Rat
	VegaAdjustment  *big.Rat

	Candidate CandidateRisk

	StrategyOpenedAt map[string]time.Time
	Now              time.Time

	CumulativePnL *big.Rat
	DailyTracker  *DailyLossTracker
}
