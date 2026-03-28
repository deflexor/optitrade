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
