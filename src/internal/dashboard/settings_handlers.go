package dashboard

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type settingsFieldDef struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Help     string `json:"help"`
	Kind     string `json:"kind"`
	Required bool   `json:"required"`
	Options  []struct {
		Value string `json:"value"`
		Label string `json:"label"`
	} `json:"options,omitempty"`
}

type settingsSecretView struct {
	Masked     string `json:"masked"`
	Configured bool   `json:"configured"`
}

func settingsFieldCatalog() []settingsFieldDef {
	return []settingsFieldDef{
		{
			ID: "provider", Label: "Exchange", Kind: "select", Required: true,
			Help: "Venue for this operator account.",
			Options: []struct {
				Value string `json:"value"`
				Label string `json:"label"`
			}{
				{Value: "deribit", Label: "Deribit"},
				{Value: "okx", Label: "OKX"},
			},
		},
		{
			ID: "deribit_use_mainnet", Label: "Deribit mainnet", Kind: "bool", Required: false,
			Help: "When enabled, use Deribit production. When disabled, use Deribit testnet (default).",
		},
		{
			ID: "deribit_client_id", Label: "Deribit client ID", Kind: "password", Required: false,
			Help: "API key from Deribit (Account → API). Use testnet keys when mainnet is off.",
		},
		{
			ID: "deribit_client_secret", Label: "Deribit client secret", Kind: "password", Required: false,
			Help: "API secret paired with the client ID. Stored encrypted on the server.",
		},
		{
			ID: "okx_demo", Label: "OKX demo trading", Kind: "bool", Required: false,
			Help: "Use demo keys from OKX Demo Trading → Demo Trading API (REST header x-simulated-trading: 1).",
		},
		{
			ID: "okx_api_key", Label: "OKX API key", Kind: "password", Required: false,
			Help: "OKX API key (live or demo, matching the demo trading toggle).",
		},
		{
			ID: "okx_secret_key", Label: "OKX secret key", Kind: "password", Required: false,
			Help: "OKX secret key for signing REST requests.",
		},
		{
			ID: "okx_passphrase", Label: "OKX passphrase", Kind: "password", Required: false,
			Help: "Passphrase set when creating the OKX API key.",
		},
		{
			ID: "currencies", Label: "Overview currencies", Kind: "string", Required: false,
			Help: "Comma-separated list (e.g. BTC, ETH) for P/L rollup. Blank uses server default or OPTITRADE_DASHBOARD_CURRENCIES.",
		},
	}
}

func maskSecret(s string) settingsSecretView {
	s = strings.TrimSpace(s)
	if s == "" {
		return settingsSecretView{Masked: "", Configured: false}
	}
	if len(s) <= 4 {
		return settingsSecretView{Masked: "****", Configured: true}
	}
	return settingsSecretView{Masked: "****" + s[len(s)-4:], Configured: true}
}

func (s *Server) handleSettingsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := requestUser(r.Context())
	if !ok || s.settings == nil || s.settingsCrypto == nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "settings unavailable")
		return
	}
	s.writeSettingsJSON(w, r.Context(), user)
}

func (s *Server) writeSettingsJSON(w http.ResponseWriter, ctx context.Context, user string) {
	row, err := s.settings.GetDecrypting(ctx, user, s.settingsCrypto)
	if err != nil {
		logHandlerError(s.log, "settings_get", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load settings")
		return
	}

	prov := string(ProviderDeribit)
	dm := false
	od := false
	cur := ""
	if row != nil {
		prov = string(row.Provider)
		dm = row.DeribitUseMainnet
		od = row.OKXDemo
		cur = row.Currencies
	}

	sec := rowSecrets(row)
	values := map[string]any{
		"provider":              prov,
		"deribit_use_mainnet":   dm,
		"okx_demo":              od,
		"currencies":            cur,
		"deribit_client_id":     maskSecret(sec.DeribitClientID),
		"deribit_client_secret": maskSecret(sec.DeribitClientSecret),
		"okx_api_key":           maskSecret(sec.OKXAPIKey),
		"okx_secret_key":        maskSecret(sec.OKXSecretKey),
		"okx_passphrase":        maskSecret(sec.OKXPassphrase),
	}

	var warnings []string
	switch OperatorProvider(prov) {
	case ProviderDeribit:
		if !maskSecret(sec.DeribitClientID).Configured || !maskSecret(sec.DeribitClientSecret).Configured {
			warnings = append(warnings, "Deribit API credentials are required to load balances and positions.")
		}
	case ProviderOKX:
		if !maskSecret(sec.OKXAPIKey).Configured || !maskSecret(sec.OKXSecretKey).Configured || !maskSecret(sec.OKXPassphrase).Configured {
			warnings = append(warnings, "OKX API key, secret, and passphrase are required.")
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"fields":   settingsFieldCatalog(),
		"values":   values,
		"warnings": warnings,
	})
}

func rowSecrets(row *OperatorSettingsRow) OperatorSecrets {
	if row == nil {
		return OperatorSecrets{}
	}
	return row.Secrets
}

func (s *Server) handleSettingsPut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := requestUser(r.Context())
	if !ok || s.settings == nil || s.settingsCrypto == nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "settings unavailable")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_body", "could not read body")
		return
	}
	var body struct {
		Provider          *string           `json:"provider"`
		DeribitUseMainnet *bool             `json:"deribit_use_mainnet"`
		OKXDemo           *bool             `json:"okx_demo"`
		Currencies        *string           `json:"currencies"`
		Secrets           map[string]string `json:"secrets"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_body", "expected JSON")
		return
	}

	patch := OperatorSettingsPatch{
		Provider:          body.Provider,
		DeribitUseMainnet: body.DeribitUseMainnet,
		OKXDemo:           body.OKXDemo,
		Currencies:        body.Currencies,
		Secrets:           body.Secrets,
	}
	if _, err := s.settings.Put(r.Context(), user, s.settingsCrypto, patch); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_settings", err.Error())
		return
	}
	s.invalidateExchangeCache(user)
	s.writeSettingsJSON(w, r.Context(), user)
}
