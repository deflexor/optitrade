-- Trading prefs: admin account_status, operator bot_mode, max loss % of equity (spec 2026-04-04).
ALTER TABLE dashboard_operator_settings ADD COLUMN account_status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE dashboard_operator_settings ADD COLUMN bot_mode TEXT NOT NULL DEFAULT 'manual';
ALTER TABLE dashboard_operator_settings ADD COLUMN max_loss_equity_pct INTEGER NOT NULL DEFAULT 10;
