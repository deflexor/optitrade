package state

import "context"

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
