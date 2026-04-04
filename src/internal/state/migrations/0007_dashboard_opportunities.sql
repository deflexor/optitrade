CREATE TABLE IF NOT EXISTS dashboard_opportunity (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  status TEXT NOT NULL,
  strategy_name TEXT NOT NULL,
  legs_json TEXT NOT NULL,
  meta_json TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_dashboard_opportunity_user ON dashboard_opportunity(username);
CREATE INDEX IF NOT EXISTS idx_dashboard_opportunity_updated ON dashboard_opportunity(updated_at);
