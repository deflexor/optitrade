package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPolicySearchParents(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	srcDir := filepath.Join(repoRoot, "src")
	t.Chdir(srcDir)
	rel := filepath.Join("config", "examples", "policy.example.json")
	p, err := LoadFile(rel)
	if err != nil {
		t.Fatal(err)
	}
	if p.Version != "1.0.0" {
		t.Fatalf("version: got %q", p.Version)
	}
}

func TestLoadExamplePolicy(t *testing.T) {
	root := filepath.Join("..", "..", "..", "config", "examples", "policy.example.json")
	p, err := LoadFile(root)
	if err != nil {
		t.Fatal(err)
	}
	if p.Version != "1.0.0" {
		t.Fatalf("version: got %q", p.Version)
	}
	if p.Environment != "testnet" {
		t.Fatalf("environment: got %q", p.Environment)
	}
	if p.Limits.MaxDailyLoss != "1500" {
		t.Fatalf("max_daily_loss: got %q", p.Limits.MaxDailyLoss)
	}
	if p.Regime == nil || p.Regime.Classifier != "rules_v1" {
		t.Fatalf("regime: %+v", p.Regime)
	}
}

func TestLoadBytesMissingMaxDailyLoss(t *testing.T) {
	raw := `{
		"version": "1.0.0",
		"limits": {
			"max_loss_per_trade": "500",
			"max_open_premium_at_risk": "8000",
			"max_portfolio_delta": "0.15",
			"max_portfolio_vega": "0.02",
			"max_open_orders_per_instrument": 2,
			"max_time_in_trade_seconds": 28800
		},
		"liquidity": { "min_top_size": "1", "max_spread_bps": 10 },
		"playbooks": {
			"low": { "allowed_structures": ["credit_spread"] },
			"normal": { "allowed_structures": ["credit_spread"] },
			"high": { "allowed_structures": ["credit_spread"] }
		}
	}`
	_, err := LoadBytes([]byte(raw))
	if err == nil {
		t.Fatal("expected error for missing max_daily_loss")
	}
	if !strings.Contains(err.Error(), "policy violates schema") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadBytesInvalidPlaybookStructureEnum(t *testing.T) {
	raw := `{
		"version": "1.0.0",
		"limits": {
			"max_loss_per_trade": "500",
			"max_daily_loss": "1500",
			"max_open_premium_at_risk": "8000",
			"max_portfolio_delta": "0.15",
			"max_portfolio_vega": "0.02",
			"max_open_orders_per_instrument": 2,
			"max_time_in_trade_seconds": 28800
		},
		"liquidity": { "min_top_size": "1", "max_spread_bps": 10 },
		"playbooks": {
			"low": { "allowed_structures": ["naked_short"] },
			"normal": { "allowed_structures": ["credit_spread"] },
			"high": { "allowed_structures": ["credit_spread"] }
		}
	}`
	_, err := LoadBytes([]byte(raw))
	if err == nil {
		t.Fatal("expected error for invalid playbook structure enum")
	}
	if !strings.Contains(err.Error(), "policy violates schema") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadBytesExtraFieldInLimits(t *testing.T) {
	raw := `{
		"version": "1.0.0",
		"limits": {
			"max_loss_per_trade": "500",
			"max_daily_loss": "1500",
			"max_open_premium_at_risk": "8000",
			"max_portfolio_delta": "0.15",
			"max_portfolio_vega": "0.02",
			"max_open_orders_per_instrument": 2,
			"max_time_in_trade_seconds": 28800,
			"surprise_field": "not allowed"
		},
		"liquidity": { "min_top_size": "1", "max_spread_bps": 10 },
		"playbooks": {
			"low": { "allowed_structures": ["credit_spread"] },
			"normal": { "allowed_structures": ["credit_spread"] },
			"high": { "allowed_structures": ["credit_spread"] }
		}
	}`
	_, err := LoadBytes([]byte(raw))
	if err == nil {
		t.Fatal("expected error for additionalProperties in limits")
	}
	if !strings.Contains(err.Error(), "policy violates schema") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadBytesInvalidVersionPattern(t *testing.T) {
	raw := `{
		"version": "v1.0.0",
		"limits": {
			"max_loss_per_trade": "500",
			"max_daily_loss": "1500",
			"max_open_premium_at_risk": "8000",
			"max_portfolio_delta": "0.15",
			"max_portfolio_vega": "0.02",
			"max_open_orders_per_instrument": 2,
			"max_time_in_trade_seconds": 28800
		},
		"liquidity": { "min_top_size": "1", "max_spread_bps": 10 },
		"playbooks": {
			"low": { "allowed_structures": ["credit_spread"] },
			"normal": { "allowed_structures": ["credit_spread"] },
			"high": { "allowed_structures": ["credit_spread"] }
		}
	}`
	_, err := LoadBytes([]byte(raw))
	if err == nil {
		t.Fatal("expected error for invalid semver pattern on version")
	}
}

func TestPolicyPathFromEnvUnset(t *testing.T) {
	t.Setenv(envPolicyPath, "")
	if got := PolicyPathFromEnv(); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestPolicyPathFromEnvSet(t *testing.T) {
	t.Setenv(envPolicyPath, "  /tmp/policy.json  ")
	if got := PolicyPathFromEnv(); got != "/tmp/policy.json" {
		t.Fatalf("got %q", got)
	}
}
