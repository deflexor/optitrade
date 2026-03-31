-- Per-operator dashboard settings (encrypted venue credentials). Feature 004.
CREATE TABLE IF NOT EXISTS dashboard_operator_settings (
  username TEXT PRIMARY KEY,
  provider TEXT NOT NULL DEFAULT 'deribit',
  deribit_use_mainnet INTEGER NOT NULL DEFAULT 0,
  okx_demo INTEGER NOT NULL DEFAULT 0,
  currencies TEXT,
  secrets_blob BLOB,
  updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dashboard_operator_settings_updated ON dashboard_operator_settings(updated_at);
