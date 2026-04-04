package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// OperatorProvider names a supported venue in dashboard settings.
type OperatorProvider string

const (
	ProviderDeribit OperatorProvider = "deribit"
	ProviderOKX     OperatorProvider = "okx"
)

// OperatorSecrets is stored encrypted in dashboard_operator_settings.secrets_blob.
type OperatorSecrets struct {
	DeribitClientID     string `json:"deribit_client_id,omitempty"`
	DeribitClientSecret string `json:"deribit_client_secret,omitempty"`
	OKXAPIKey           string `json:"okx_api_key,omitempty"`
	OKXSecretKey        string `json:"okx_secret_key,omitempty"`
	OKXPassphrase       string `json:"okx_passphrase,omitempty"`
}

// OperatorSettingsRow is the non-secret face of a settings row plus decrypted secrets in memory.
type OperatorSettingsRow struct {
	Username          string
	Provider          OperatorProvider
	DeribitUseMainnet bool
	OKXDemo           bool
	Currencies        string // comma-separated, empty = use server default
	AccountStatus     string // active|disabled
	BotMode           string // manual|auto|paused
	MaxLossEquityPct  int    // 1-50
	Secrets           OperatorSecrets
	UpdatedAtMs       int64
}

// OperatorSettingsStore persists encrypted operator settings.
type OperatorSettingsStore struct {
	db *sql.DB
}

// NewOperatorSettingsStore wraps the dashboard session DB (same sqlite file).
func NewOperatorSettingsStore(db *sql.DB) *OperatorSettingsStore {
	return &OperatorSettingsStore{db: db}
}

// GetDecrypting loads and decrypts secrets. Returns (nil, nil) if no row.
func (st *OperatorSettingsStore) GetDecrypting(ctx context.Context, username string, crypto *SettingsCrypto) (*OperatorSettingsRow, error) {
	row := st.db.QueryRowContext(ctx, `
SELECT username, provider, deribit_use_mainnet, okx_demo, currencies, secrets_blob, updated_at,
       account_status, bot_mode, max_loss_equity_pct
FROM dashboard_operator_settings WHERE username = ?`, username)

	var prov string
	var mainnet, okxDemo int
	var currencies sql.NullString
	var blob []byte
	var updated int64
	var u string
	var accountStatus, botMode string
	var maxLoss sql.NullInt64
	err := row.Scan(&u, &prov, &mainnet, &okxDemo, &currencies, &blob, &updated, &accountStatus, &botMode, &maxLoss)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out := &OperatorSettingsRow{
		Username:          u,
		Provider:          OperatorProvider(strings.ToLower(strings.TrimSpace(prov))),
		DeribitUseMainnet: mainnet != 0,
		OKXDemo:           okxDemo != 0,
		UpdatedAtMs:       updated,
	}
	if currencies.Valid {
		out.Currencies = currencies.String
	}
	out.AccountStatus = strings.ToLower(strings.TrimSpace(accountStatus))
	if out.AccountStatus == "" {
		out.AccountStatus = "active"
	}
	out.BotMode = strings.ToLower(strings.TrimSpace(botMode))
	if out.BotMode == "" {
		out.BotMode = "manual"
	}
	if maxLoss.Valid && maxLoss.Int64 > 0 {
		out.MaxLossEquityPct = int(maxLoss.Int64)
	} else {
		out.MaxLossEquityPct = 10
	}
	if len(blob) == 0 {
		return out, nil
	}
	if crypto == nil {
		return nil, fmt.Errorf("secrets present but decryption unavailable")
	}
	plain, err := crypto.Decrypt(blob)
	if err != nil {
		return nil, fmt.Errorf("decrypt settings: %w", err)
	}
	if err := json.Unmarshal(plain, &out.Secrets); err != nil {
		return nil, fmt.Errorf("decode secrets json: %w", err)
	}
	return out, nil
}

// ListUsernames returns every username with a dashboard_operator_settings row.
func (st *OperatorSettingsStore) ListUsernames(ctx context.Context) ([]string, error) {
	if st == nil || st.db == nil {
		return nil, fmt.Errorf("operator settings: nil store")
	}
	rows, err := st.db.QueryContext(ctx, `SELECT username FROM dashboard_operator_settings ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// Put merges secrets (empty string keeps previous value) and upserts the row.
func (st *OperatorSettingsStore) Put(ctx context.Context, username string, crypto *SettingsCrypto, patch OperatorSettingsPatch) (*OperatorSettingsRow, error) {
	if crypto == nil {
		return nil, fmt.Errorf("nil crypto")
	}
	existing, err := st.GetDecrypting(ctx, username, crypto)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		existing = &OperatorSettingsRow{
			Username:         username,
			Provider:         ProviderDeribit,
			AccountStatus:    "active",
			BotMode:          "manual",
			MaxLossEquityPct: 10,
		}
	}
	if patch.Provider != nil {
		p := OperatorProvider(strings.ToLower(strings.TrimSpace(*patch.Provider)))
		if p != ProviderDeribit && p != ProviderOKX {
			return nil, fmt.Errorf("invalid provider %q", p)
		}
		existing.Provider = p
	}
	if patch.DeribitUseMainnet != nil {
		existing.DeribitUseMainnet = *patch.DeribitUseMainnet
	}
	if patch.OKXDemo != nil {
		existing.OKXDemo = *patch.OKXDemo
	}
	if patch.Currencies != nil {
		existing.Currencies = strings.TrimSpace(*patch.Currencies)
	}
	if patch.AccountStatus != nil {
		s := strings.ToLower(strings.TrimSpace(*patch.AccountStatus))
		if s != "active" && s != "disabled" {
			return nil, fmt.Errorf("invalid account_status %q (want active or disabled)", *patch.AccountStatus)
		}
		existing.AccountStatus = s
	}
	if patch.BotMode != nil {
		s := strings.ToLower(strings.TrimSpace(*patch.BotMode))
		if s != "manual" && s != "auto" && s != "paused" {
			return nil, fmt.Errorf("invalid bot_mode %q (want manual, auto, or paused)", *patch.BotMode)
		}
		existing.BotMode = s
	}
	if patch.MaxLossEquityPct != nil {
		v := *patch.MaxLossEquityPct
		if v < 1 || v > 50 {
			return nil, fmt.Errorf("max_loss_equity_pct must be between 1 and 50 inclusive, got %d", v)
		}
		existing.MaxLossEquityPct = v
	}
	se := existing.Secrets
	if patch.Secrets != nil {
		if v, ok := patch.Secrets["deribit_client_id"]; ok && v != "" {
			se.DeribitClientID = v
		}
		if v, ok := patch.Secrets["deribit_client_secret"]; ok && v != "" {
			se.DeribitClientSecret = v
		}
		if v, ok := patch.Secrets["okx_api_key"]; ok && v != "" {
			se.OKXAPIKey = v
		}
		if v, ok := patch.Secrets["okx_secret_key"]; ok && v != "" {
			se.OKXSecretKey = v
		}
		if v, ok := patch.Secrets["okx_passphrase"]; ok && v != "" {
			se.OKXPassphrase = v
		}
	}
	existing.Secrets = se

	if err := validateOperatorSettings(existing); err != nil {
		return nil, err
	}

	plain, err := json.Marshal(existing.Secrets)
	if err != nil {
		return nil, err
	}
	blob, err := crypto.Encrypt(plain)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	_, err = st.db.ExecContext(ctx, `
INSERT INTO dashboard_operator_settings(username, provider, deribit_use_mainnet, okx_demo, currencies, secrets_blob, updated_at, account_status, bot_mode, max_loss_equity_pct)
VALUES(?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(username) DO UPDATE SET
  provider = excluded.provider,
  deribit_use_mainnet = excluded.deribit_use_mainnet,
  okx_demo = excluded.okx_demo,
  currencies = excluded.currencies,
  secrets_blob = excluded.secrets_blob,
  updated_at = excluded.updated_at,
  account_status = excluded.account_status,
  bot_mode = excluded.bot_mode,
  max_loss_equity_pct = excluded.max_loss_equity_pct`,
		username,
		string(existing.Provider),
		boolAsInt(existing.DeribitUseMainnet),
		boolAsInt(existing.OKXDemo),
		nullStr(existing.Currencies),
		blob,
		now,
		existing.AccountStatus,
		existing.BotMode,
		existing.MaxLossEquityPct,
	)
	if err != nil {
		return nil, err
	}
	existing.UpdatedAtMs = now
	return existing, nil
}

// OperatorSettingsPatch is a partial update from the API.
type OperatorSettingsPatch struct {
	Provider          *string
	DeribitUseMainnet *bool
	OKXDemo           *bool
	Currencies        *string
	AccountStatus     *string
	BotMode           *string
	MaxLossEquityPct  *int
	Secrets           map[string]string
}

func boolAsInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullStr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func validateOperatorSettings(r *OperatorSettingsRow) error {
	if err := validateTradingPrefsRow(r); err != nil {
		return err
	}
	switch r.Provider {
	case ProviderDeribit:
		if strings.TrimSpace(r.Secrets.DeribitClientID) == "" || strings.TrimSpace(r.Secrets.DeribitClientSecret) == "" {
			return fmt.Errorf("deribit requires client id and client secret")
		}
	case ProviderOKX:
		if strings.TrimSpace(r.Secrets.OKXAPIKey) == "" || strings.TrimSpace(r.Secrets.OKXSecretKey) == "" || strings.TrimSpace(r.Secrets.OKXPassphrase) == "" {
			return fmt.Errorf("okx requires api key, secret key, and passphrase")
		}
	default:
		return fmt.Errorf("unknown provider %q", r.Provider)
	}
	return nil
}

func validateTradingPrefsRow(r *OperatorSettingsRow) error {
	a := strings.ToLower(strings.TrimSpace(r.AccountStatus))
	if a != "active" && a != "disabled" {
		return fmt.Errorf("invalid account_status %q", r.AccountStatus)
	}
	b := strings.ToLower(strings.TrimSpace(r.BotMode))
	if b != "manual" && b != "auto" && b != "paused" {
		return fmt.Errorf("invalid bot_mode %q", r.BotMode)
	}
	if r.MaxLossEquityPct < 1 || r.MaxLossEquityPct > 50 {
		return fmt.Errorf("max_loss_equity_pct must be between 1 and 50 inclusive, got %d", r.MaxLossEquityPct)
	}
	return nil
}
