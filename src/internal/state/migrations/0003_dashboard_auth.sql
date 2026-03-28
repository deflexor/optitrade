-- Dashboard operator auth (feature 002-operator-trading-dashboard WP02).
CREATE TABLE IF NOT EXISTS dashboard_user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS dashboard_session (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  token_hash TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL,
  user_agent TEXT,
  FOREIGN KEY (user_id) REFERENCES dashboard_user(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_dashboard_session_token_hash ON dashboard_session(token_hash);
