package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Policy is the typed in-memory representation of a validated policy file.
// Monetary and size limits use decimal strings (no float64 for money).
type Policy struct {
	Version        string          `json:"version"`
	Environment    string          `json:"environment,omitempty"`
	Limits         Limits          `json:"limits"`
	Liquidity      Liquidity       `json:"liquidity"`
	CostModel      *CostModel      `json:"cost_model,omitempty"`
	Regime         *Regime         `json:"regime,omitempty"`
	Playbooks      Playbooks       `json:"playbooks"`
	ProtectiveMode *ProtectiveMode `json:"protective_mode,omitempty"`
}

type Limits struct {
	MaxLossPerTrade           string `json:"max_loss_per_trade"`
	MaxDailyLoss              string `json:"max_daily_loss"`
	MaxOpenPremiumAtRisk      string `json:"max_open_premium_at_risk"`
	MaxPortfolioDelta         string `json:"max_portfolio_delta"`
	MaxPortfolioVega          string `json:"max_portfolio_vega"`
	MaxOpenOrdersPerInstrument int    `json:"max_open_orders_per_instrument"`
	MaxTimeInTradeSeconds     int    `json:"max_time_in_trade_seconds"`
}

type Liquidity struct {
	MinTopSize   string `json:"min_top_size"`
	MaxSpreadBps int    `json:"max_spread_bps"`
}

type CostModel struct {
	TakerFeeBps            *int `json:"taker_fee_bps,omitempty"`
	MakerFeeBps            *int `json:"maker_fee_bps,omitempty"`
	SlippageBps            *int `json:"slippage_bps,omitempty"`
	AdverseSelectionBpsLow *int `json:"adverse_selection_bps_low,omitempty"`
	AdverseSelectionBpsHigh *int `json:"adverse_selection_bps_high,omitempty"`
}

type Regime struct {
	Classifier             string `json:"classifier,omitempty"`
	LowVolThresholdIndex   string `json:"low_vol_threshold_index,omitempty"`
	HighVolThresholdIndex  string `json:"high_vol_threshold_index,omitempty"`
}

type Playbooks struct {
	Low    Playbook `json:"low"`
	Normal Playbook `json:"normal"`
	High   Playbook `json:"high"`
}

type Playbook struct {
	AllowedStructures      []string `json:"allowed_structures"`
	MaxNewPositionsPerDay  *int    `json:"max_new_positions_per_day,omitempty"`
}

type ProtectiveMode struct {
	BookGapSpreadBps *int `json:"book_gap_spread_bps,omitempty"`
	FeedStaleMs      *int `json:"feed_stale_ms,omitempty"`
}

const envPolicyPath = "OPTITRADE_POLICY_PATH"

// PolicyPathFromEnv returns OPTITRADE_POLICY_PATH if set and non-empty after trim.
func PolicyPathFromEnv() string {
	return strings.TrimSpace(os.Getenv(envPolicyPath))
}

// LoadFile reads path, validates JSON against the embedded schema, and unmarshals into Policy.
func LoadFile(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy file %q: %w", path, err)
	}
	return LoadBytes(data)
}

// LoadBytes validates raw policy JSON and unmarshals it.
func LoadBytes(data []byte) (*Policy, error) {
	if err := validatePolicyJSON(data); err != nil {
		return nil, err
	}
	var p Policy
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("policy unmarshal after schema validation: %w", err)
	}
	return &p, nil
}
