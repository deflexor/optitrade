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
SELECT username, provider, deribit_use_mainnet, okx_demo, currencies, secrets_blob, updated_at
FROM dashboard_operator_settings WHERE username = ?`, username)

	var prov string
	var mainnet, okxDemo int
	var currencies sql.NullString
	var blob []byte
	var updated int64
	var u string
	err := row.Scan(&u, &prov, &mainnet, &okxDemo, &currencies, &blob, &updated)
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
		existing = &OperatorSettingsRow{Username: username, Provider: ProviderDeribit}
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
INSERT INTO dashboard_operator_settings(username, provider, deribit_use_mainnet, okx_demo, currencies, secrets_blob, updated_at)
VALUES(?,?,?,?,?,?,?)
ON CONFLICT(username) DO UPDATE SET
  provider = excluded.provider,
  deribit_use_mainnet = excluded.deribit_use_mainnet,
  okx_demo = excluded.okx_demo,
  currencies = excluded.currencies,
  secrets_blob = excluded.secrets_blob,
  updated_at = excluded.updated_at`,
		username,
		string(existing.Provider),
		boolAsInt(existing.DeribitUseMainnet),
		boolAsInt(existing.OKXDemo),
		nullStr(existing.Currencies),
		blob,
		now,
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
