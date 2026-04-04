package dashboard

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/dfr/optitrade/src/internal/state/sqlite"
)

func TestRunnerEligible(t *testing.T) {
	t.Parallel()
	okx := &OperatorSettingsRow{
		Username:         "a",
		Provider:         ProviderOKX,
		AccountStatus:    "active",
		BotMode:          "manual",
		MaxLossEquityPct: 10,
		Secrets: OperatorSecrets{
			OKXAPIKey: "k", OKXSecretKey: "s", OKXPassphrase: "p",
		},
	}
	if !RunnerEligible(okx) {
		t.Fatal("want eligible OKX")
	}
	paused := *okx
	paused.BotMode = "paused"
	if RunnerEligible(&paused) {
		t.Fatal("paused ineligible")
	}
	disabled := *okx
	disabled.AccountStatus = "disabled"
	if RunnerEligible(&disabled) {
		t.Fatal("disabled ineligible")
	}
	deribit := *okx
	deribit.Provider = ProviderDeribit
	deribit.Secrets = OperatorSecrets{DeribitClientID: "x", DeribitClientSecret: "y"}
	deribit.MaxLossEquityPct = 10
	if RunnerEligible(&deribit) {
		t.Fatal("deribit ineligible for M1 runner")
	}
}

func TestRunnerManager_Reconcile_OKXLifecycle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "r.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()

	prov := "okx"
	_, err = st.Put(ctx, "opx", crypto, OperatorSettingsPatch{
		Provider: &prov,
		Secrets: map[string]string{
			"okx_api_key":        "k",
			"okx_secret_key":     "s",
			"okx_passphrase":     "p",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	rm := NewRunnerManager(log, st, crypto)
	rm.Reconcile(ctx)
	if !rm.IsRunning("opx") {
		t.Fatal("expected runner for OKX user")
	}

	paused := "paused"
	if _, err := st.Put(ctx, "opx", crypto, OperatorSettingsPatch{BotMode: &paused}); err != nil {
		t.Fatal(err)
	}
	rm.Reconcile(ctx)
	if rm.IsRunning("opx") {
		t.Fatal("expected runner stopped when paused")
	}
}
