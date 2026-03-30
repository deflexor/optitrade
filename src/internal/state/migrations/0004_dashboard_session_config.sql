-- dashboard_session only: operator allowlist is DashboardAuthFile JSON (feature 002).
DROP TABLE IF EXISTS dashboard_session;
DROP TABLE IF EXISTS dashboard_user;

CREATE TABLE IF NOT EXISTS dashboard_session (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL,
  token_hash TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  user_agent TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dashboard_session_token_hash ON dashboard_session(token_hash);
CREATE INDEX IF NOT EXISTS idx_dashboard_session_username ON dashboard_session(username);
