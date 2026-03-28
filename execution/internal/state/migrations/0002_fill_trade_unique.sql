-- One row per exchange trade_id for idempotent fill ingestion (T049).
CREATE UNIQUE INDEX IF NOT EXISTS idx_fill_record_trade_id ON fill_record(trade_id);
