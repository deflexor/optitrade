package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/opportunities"
)

// OpportunityStore persists bot-tracked opportunity rows (opening / active / partial).
type OpportunityStore struct {
	db *sql.DB
}

// NewOpportunityStore wraps the dashboard SQLite DB (migrated, same file as sessions).
func NewOpportunityStore(db *sql.DB) *OpportunityStore {
	if db == nil {
		return nil
	}
	return &OpportunityStore{db: db}
}

// OpportunityRecord is a row in dashboard_opportunity.
type OpportunityRecord struct {
	ID           string
	Username     string
	Status       string
	StrategyName string
	LegsJSON     string
	MetaJSON     string
	CreatedAtMs  int64
	UpdatedAtMs  int64
}

// ListByUser returns all opportunities for the operator (any status).
func (st *OpportunityStore) ListByUser(ctx context.Context, username string) ([]OpportunityRecord, error) {
	if st == nil || st.db == nil {
		return nil, nil
	}
	u := strings.TrimSpace(username)
	if u == "" {
		return nil, fmt.Errorf("opportunity: empty username")
	}
	rows, err := st.db.QueryContext(ctx, `
SELECT id, username, status, strategy_name, legs_json, meta_json, created_at, updated_at
FROM dashboard_opportunity WHERE username = ? ORDER BY updated_at DESC`, u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []OpportunityRecord
	for rows.Next() {
		var r OpportunityRecord
		if err := rows.Scan(&r.ID, &r.Username, &r.Status, &r.StrategyName, &r.LegsJSON, &r.MetaJSON, &r.CreatedAtMs, &r.UpdatedAtMs); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Upsert inserts or replaces an opportunity row by id.
func (st *OpportunityStore) Upsert(ctx context.Context, r *OpportunityRecord) error {
	if st == nil || st.db == nil {
		return fmt.Errorf("opportunity: nil store")
	}
	if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.Username) == "" {
		return fmt.Errorf("opportunity: id and username required")
	}
	now := time.Now().UnixMilli()
	if r.CreatedAtMs == 0 {
		r.CreatedAtMs = now
	}
	r.UpdatedAtMs = now
	_, err := st.db.ExecContext(ctx, `
INSERT INTO dashboard_opportunity (id, username, status, strategy_name, legs_json, meta_json, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  username = excluded.username,
  status = excluded.status,
  strategy_name = excluded.strategy_name,
  legs_json = excluded.legs_json,
  meta_json = excluded.meta_json,
  updated_at = excluded.updated_at`,
		r.ID, r.Username, r.Status, r.StrategyName, r.LegsJSON, r.MetaJSON, r.CreatedAtMs, r.UpdatedAtMs)
	return err
}

// Get returns one row if it exists and belongs to username.
func (st *OpportunityStore) Get(ctx context.Context, id, username string) (*OpportunityRecord, error) {
	if st == nil || st.db == nil {
		return nil, fmt.Errorf("opportunity: nil store")
	}
	var r OpportunityRecord
	err := st.db.QueryRowContext(ctx, `
SELECT id, username, status, strategy_name, legs_json, meta_json, created_at, updated_at
FROM dashboard_opportunity WHERE id = ? AND username = ?`, id, username).Scan(
		&r.ID, &r.Username, &r.Status, &r.StrategyName, &r.LegsJSON, &r.MetaJSON, &r.CreatedAtMs, &r.UpdatedAtMs)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateStatus sets status and replaces meta_json; bumps updated_at.
func (st *OpportunityStore) UpdateStatus(ctx context.Context, id, username, status, metaJSON string) error {
	if st == nil || st.db == nil {
		return fmt.Errorf("opportunity: nil store")
	}
	res, err := st.db.ExecContext(ctx, `
UPDATE dashboard_opportunity SET status = ?, meta_json = ?, updated_at = ?
WHERE id = ? AND username = ?`,
		status, metaJSON, time.Now().UnixMilli(), id, username)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete removes a row owned by username.
func (st *OpportunityStore) Delete(ctx context.Context, id, username string) error {
	if st == nil || st.db == nil {
		return fmt.Errorf("opportunity: nil store")
	}
	_, err := st.db.ExecContext(ctx, `DELETE FROM dashboard_opportunity WHERE id = ? AND username = ?`, id, username)
	return err
}

// legSideMeta records exchange side used when placing the spread.
type legSideMeta struct {
	Instrument string `json:"instrument"`
	Side       string `json:"side"` // buy | sell
}

// opportunityMetaJSON holds row fields stored beside legs_json.
type opportunityMetaJSON struct {
	GreeksNote   string  `json:"greeks_note,omitempty"`
	MaxProfit    string  `json:"max_profit"`
	MaxLoss      string  `json:"max_loss"`
	Recommend    string  `json:"recommendation"`
	Rationale    string  `json:"rationale"`
	ExpectedEdge string  `json:"expected_edge"`
	EdgeAfter    float64 `json:"edge_after_costs"`
	OrderIDs     []string       `json:"order_ids,omitempty"`
	LegSides     []legSideMeta  `json:"leg_sides,omitempty"`
}

func encodeOpportunityRow(row opportunities.Row) (legsJSON, metaJSON string, err error) {
	return encodeOpportunityRowPersist(row, nil)
}

// encodeOpportunityRowPersist extends meta with execution fields when exec is non-nil.
func encodeOpportunityRowPersist(row opportunities.Row, exec *opportunityExecPersist) (legsJSON, metaJSON string, err error) {
	legs, err := json.Marshal(row.Legs)
	if err != nil {
		return "", "", err
	}
	meta := opportunityMetaJSON{
		GreeksNote:   row.GreeksNote,
		MaxProfit:    row.MaxProfit,
		MaxLoss:      row.MaxLoss,
		Recommend:    row.Recommend,
		Rationale:    row.Rationale,
		ExpectedEdge: row.ExpectedEdge,
		EdgeAfter:    row.EdgeAfter,
	}
	if exec != nil {
		meta.OrderIDs = exec.OrderIDs
		meta.LegSides = exec.LegSides
	}
	mb, err := json.Marshal(meta)
	if err != nil {
		return "", "", err
	}
	return string(legs), string(mb), nil
}

// opportunityExecPersist is optional persistence for open/cancel/close.
type opportunityExecPersist struct {
	OrderIDs []string
	LegSides []legSideMeta
}

func decodeOpportunityRow(rec *OpportunityRecord) (opportunities.Row, error) {
	var row opportunities.Row
	row.ID = rec.ID
	row.StrategyName = rec.StrategyName
	row.Status = opportunities.RowStatus(rec.Status)
	if err := json.Unmarshal([]byte(rec.LegsJSON), &row.Legs); err != nil {
		return row, fmt.Errorf("legs_json: %w", err)
	}
	var meta opportunityMetaJSON
	if err := json.Unmarshal([]byte(rec.MetaJSON), &meta); err != nil {
		return row, fmt.Errorf("meta_json: %w", err)
	}
	row.GreeksNote = meta.GreeksNote
	row.MaxProfit = meta.MaxProfit
	row.MaxLoss = meta.MaxLoss
	row.Recommend = meta.Recommend
	row.Rationale = meta.Rationale
	row.ExpectedEdge = meta.ExpectedEdge
	row.EdgeAfter = meta.EdgeAfter
	return row, nil
}
