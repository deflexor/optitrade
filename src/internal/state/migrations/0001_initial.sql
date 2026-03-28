CREATE TABLE IF NOT EXISTS instrument (
  instrument_name TEXT PRIMARY KEY NOT NULL,
  kind TEXT NOT NULL,
  base TEXT NOT NULL,
  expiry TEXT NOT NULL,
  strike TEXT NOT NULL,
  option_type TEXT NOT NULL,
  tick_size TEXT NOT NULL,
  min_trade_amount TEXT NOT NULL,
  is_active INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS regime_state (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  effective_at INTEGER NOT NULL,
  label TEXT NOT NULL,
  classifier_version TEXT NOT NULL,
  inputs_digest TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS trade_candidate (
  id TEXT PRIMARY KEY NOT NULL,
  created_at INTEGER NOT NULL,
  regime_label TEXT NOT NULL,
  playbook_id TEXT NOT NULL,
  structure_type TEXT NOT NULL,
  legs_json TEXT NOT NULL,
  expected_edge TEXT NOT NULL,
  cost_breakdown_json TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS order_record (
  internal_order_id TEXT PRIMARY KEY NOT NULL,
  exchange_order_id TEXT,
  instrument_name TEXT NOT NULL,
  label TEXT NOT NULL,
  side TEXT NOT NULL,
  order_type TEXT NOT NULL,
  price TEXT,
  amount TEXT NOT NULL,
  post_only INTEGER NOT NULL,
  reduce_only INTEGER NOT NULL,
  state TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL,
  candidate_id TEXT,
  FOREIGN KEY (candidate_id) REFERENCES trade_candidate(id)
);

CREATE INDEX IF NOT EXISTS idx_order_record_instrument_name ON order_record(instrument_name);

CREATE TABLE IF NOT EXISTS fill_record (
  id TEXT PRIMARY KEY NOT NULL,
  order_id TEXT NOT NULL,
  trade_id TEXT NOT NULL,
  instrument_name TEXT NOT NULL,
  qty TEXT NOT NULL,
  price TEXT NOT NULL,
  fee TEXT NOT NULL,
  filled_at INTEGER NOT NULL,
  FOREIGN KEY (order_id) REFERENCES order_record(internal_order_id)
);

CREATE TABLE IF NOT EXISTS position_snapshot (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  captured_at INTEGER NOT NULL,
  net_delta TEXT NOT NULL,
  net_vega TEXT NOT NULL,
  net_gamma TEXT,
  premium_at_risk TEXT NOT NULL,
  unrealized_pnl TEXT NOT NULL,
  raw_json TEXT
);

CREATE TABLE IF NOT EXISTS risk_policy (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  version TEXT NOT NULL,
  policy_json TEXT NOT NULL,
  active INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_decision (
  id TEXT PRIMARY KEY NOT NULL,
  ts INTEGER NOT NULL,
  decision_type TEXT NOT NULL,
  candidate_id TEXT,
  regime_label TEXT NOT NULL,
  cost_model_version TEXT NOT NULL,
  risk_gate_results TEXT NOT NULL,
  reason TEXT NOT NULL,
  correlation_id TEXT NOT NULL,
  FOREIGN KEY (candidate_id) REFERENCES trade_candidate(id)
);

CREATE INDEX IF NOT EXISTS idx_audit_decision_ts ON audit_decision(ts);
