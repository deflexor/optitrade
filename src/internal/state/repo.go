package state

import (
	"context"
	"errors"
)

// ErrNoRegimeState is returned when no regime row exists yet.
var ErrNoRegimeState = errors.New("no regime_state rows")

// OrderRepository persists and loads orders (parameterized SQL only in implementations).
type OrderRepository interface {
	InsertOrder(ctx context.Context, o *OrderRecord) error
	GetOrder(ctx context.Context, internalID string) (*OrderRecord, error)
	UpdateOrder(ctx context.Context, o *OrderRecord) error
	ListOrdersByStates(ctx context.Context, states []string) ([]OrderRecord, error)
}

// FillRepository persists executed trades linked to internal orders.
type FillRepository interface {
	// InsertFill inserts a fill unless trade_id already exists (unique index); inserted is false on duplicate.
	InsertFill(ctx context.Context, f *FillRecord) (inserted bool, err error)
}

// AuditRepository persists and loads audit decisions.
type AuditRepository interface {
	InsertAudit(ctx context.Context, a *AuditDecision) error
	GetAudit(ctx context.Context, id string) (*AuditDecision, error)
}

// RegimeRepository persists regime transitions for audit (SC-005).
type RegimeRepository interface {
	InsertRegimeState(ctx context.Context, r *RegimeState) error
	LatestRegimeState(ctx context.Context) (*RegimeState, error)
}
