// Package sqlite implements state repositories with SQLite using parameterized queries only.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dfr/optitrade/src/internal/state"

	_ "modernc.org/sqlite" // register "sqlite" driver
)

const busyTimeoutMs = 5000

// Open opens a SQLite database with WAL, busy_timeout (see const), and foreign keys,
// then applies [state.ApplyMigrations]. Use one open DB per bot process (MaxOpenConns=1).
func Open(path string) (*sql.DB, error) {
	// DSN parameters documented in code: single-writer bot expects WAL + busy wait.
	dsn := "file:" + filepath.ToSlash(path) +
		"?_journal_mode=WAL&_busy_timeout=" + fmt.Sprint(busyTimeoutMs)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pragma foreign_keys: %w", err)
	}

	if err := state.ApplyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// Store implements order and audit repositories.
type Store struct {
	db *sql.DB
}

// NewStore wraps a *sql.DB (already opened and migrated, e.g. via [Open]).
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

var (
	_ state.OrderRepository  = (*Store)(nil)
	_ state.FillRepository   = (*Store)(nil)
	_ state.AuditRepository  = (*Store)(nil)
	_ state.RegimeRepository = (*Store)(nil)
)

func (s *Store) InsertOrder(ctx context.Context, o *state.OrderRecord) error {
	if o == nil {
		return fmt.Errorf("nil order")
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO order_record (
  internal_order_id, exchange_order_id, instrument_name, label, side, order_type,
  price, amount, post_only, reduce_only, state, created_at, updated_at, candidate_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		o.InternalOrderID,
		nullString(o.ExchangeOrderID),
		o.InstrumentName,
		o.Label,
		o.Side,
		o.OrderType,
		nullString(o.Price),
		o.Amount,
		boolInt(o.PostOnly),
		boolInt(o.ReduceOnly),
		o.State,
		o.CreatedAt,
		o.UpdatedAt,
		nullString(o.CandidateID),
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	return nil
}

func (s *Store) GetOrder(ctx context.Context, internalID string) (*state.OrderRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT internal_order_id, exchange_order_id, instrument_name, label, side, order_type,
       price, amount, post_only, reduce_only, state, created_at, updated_at, candidate_id
FROM order_record WHERE internal_order_id = ?`, internalID)

	var o state.OrderRecord
	var exchangeID, price, candidateID sql.NullString
	var postOnly, reduceOnly int
	err := row.Scan(
		&o.InternalOrderID,
		&exchangeID,
		&o.InstrumentName,
		&o.Label,
		&o.Side,
		&o.OrderType,
		&price,
		&o.Amount,
		&postOnly,
		&reduceOnly,
		&o.State,
		&o.CreatedAt,
		&o.UpdatedAt,
		&candidateID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order %q: %w", internalID, err)
	}
	if err != nil {
		return nil, fmt.Errorf("scan order: %w", err)
	}
	o.ExchangeOrderID = ptrFromNullString(exchangeID)
	o.Price = ptrFromNullString(price)
	o.CandidateID = ptrFromNullString(candidateID)
	o.PostOnly = postOnly != 0
	o.ReduceOnly = reduceOnly != 0
	return &o, nil
}

func (s *Store) UpdateOrder(ctx context.Context, o *state.OrderRecord) error {
	if o == nil {
		return fmt.Errorf("nil order")
	}
	_, err := s.db.ExecContext(ctx, `
UPDATE order_record SET
  exchange_order_id = ?, instrument_name = ?, label = ?, side = ?, order_type = ?,
  price = ?, amount = ?, post_only = ?, reduce_only = ?, state = ?,
  created_at = ?, updated_at = ?, candidate_id = ?
WHERE internal_order_id = ?`,
		nullString(o.ExchangeOrderID),
		o.InstrumentName,
		o.Label,
		o.Side,
		o.OrderType,
		nullString(o.Price),
		o.Amount,
		boolInt(o.PostOnly),
		boolInt(o.ReduceOnly),
		o.State,
		o.CreatedAt,
		o.UpdatedAt,
		nullString(o.CandidateID),
		o.InternalOrderID,
	)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}
	return nil
}

func (s *Store) ListOrdersByStates(ctx context.Context, states []string) ([]state.OrderRecord, error) {
	if len(states) == 0 {
		return nil, nil
	}
	args := make([]any, len(states))
	placeholders := make([]string, len(states))
	for i, s := range states {
		args[i] = s
		placeholders[i] = "?"
	}
	q := `SELECT internal_order_id, exchange_order_id, instrument_name, label, side, order_type,
       price, amount, post_only, reduce_only, state, created_at, updated_at, candidate_id
FROM order_record WHERE state IN (` + strings.Join(placeholders, ",") + `) ORDER BY created_at ASC`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var out []state.OrderRecord
	for rows.Next() {
		var o state.OrderRecord
		var exchangeID, price, candidateID sql.NullString
		var postOnly, reduceOnly int
		if err := rows.Scan(
			&o.InternalOrderID,
			&exchangeID,
			&o.InstrumentName,
			&o.Label,
			&o.Side,
			&o.OrderType,
			&price,
			&o.Amount,
			&postOnly,
			&reduceOnly,
			&o.State,
			&o.CreatedAt,
			&o.UpdatedAt,
			&candidateID,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.ExchangeOrderID = ptrFromNullString(exchangeID)
		o.Price = ptrFromNullString(price)
		o.CandidateID = ptrFromNullString(candidateID)
		o.PostOnly = postOnly != 0
		o.ReduceOnly = reduceOnly != 0
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) InsertFill(ctx context.Context, f *state.FillRecord) (bool, error) {
	if f == nil {
		return false, fmt.Errorf("nil fill")
	}
	res, err := s.db.ExecContext(ctx, `
INSERT OR IGNORE INTO fill_record (id, order_id, trade_id, instrument_name, qty, price, fee, filled_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		f.ID,
		f.OrderID,
		f.TradeID,
		f.InstrumentName,
		f.Qty,
		f.Price,
		f.Fee,
		f.FilledAt,
	)
	if err != nil {
		return false, fmt.Errorf("insert fill: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Store) InsertAudit(ctx context.Context, a *state.AuditDecision) error {
	if a == nil {
		return fmt.Errorf("nil audit")
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO audit_decision (
  id, ts, decision_type, candidate_id, regime_label, cost_model_version,
  risk_gate_results, reason, correlation_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID,
		a.Ts,
		a.DecisionType,
		nullString(a.CandidateID),
		a.RegimeLabel,
		a.CostModelVersion,
		a.RiskGateResults,
		a.Reason,
		a.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	return nil
}

func (s *Store) InsertRegimeState(ctx context.Context, r *state.RegimeState) error {
	if r == nil {
		return fmt.Errorf("nil regime state")
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO regime_state (effective_at, label, classifier_version, inputs_digest)
VALUES (?, ?, ?, ?)`,
		r.EffectiveAt,
		r.Label,
		r.ClassifierVersion,
		r.InputsDigest,
	)
	if err != nil {
		return fmt.Errorf("insert regime_state: %w", err)
	}
	return nil
}

func (s *Store) LatestRegimeState(ctx context.Context) (*state.RegimeState, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, effective_at, label, classifier_version, inputs_digest
FROM regime_state ORDER BY effective_at DESC, id DESC LIMIT 1`)

	var r state.RegimeState
	err := row.Scan(&r.ID, &r.EffectiveAt, &r.Label, &r.ClassifierVersion, &r.InputsDigest)
	if err == sql.ErrNoRows {
		return nil, state.ErrNoRegimeState
	}
	if err != nil {
		return nil, fmt.Errorf("scan regime_state: %w", err)
	}
	return &r, nil
}

func (s *Store) GetAudit(ctx context.Context, id string) (*state.AuditDecision, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, ts, decision_type, candidate_id, regime_label, cost_model_version,
       risk_gate_results, reason, correlation_id
FROM audit_decision WHERE id = ?`, id)

	var a state.AuditDecision
	var cand sql.NullString
	err := row.Scan(
		&a.ID,
		&a.Ts,
		&a.DecisionType,
		&cand,
		&a.RegimeLabel,
		&a.CostModelVersion,
		&a.RiskGateResults,
		&a.Reason,
		&a.CorrelationID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("audit %q: %w", id, err)
	}
	if err != nil {
		return nil, fmt.Errorf("scan audit: %w", err)
	}
	a.CandidateID = ptrFromNullString(cand)
	return &a, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullString(p *string) any {
	if p == nil {
		return nil
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return nil
	}
	return *p
}

func ptrFromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}
