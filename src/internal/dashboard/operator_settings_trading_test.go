package dashboard

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"path/filepath"
	"testing"

	"github.com/dfr/optitrade/src/internal/state/sqlite"
)

func testOperatorSettingsCrypto(t *testing.T) *SettingsCrypto {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	return &SettingsCrypto{gcm: gcm}
}

func TestOperatorSettingsTradingPrefsRoundTrip(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	user := "op1"

	deribit := "deribit"
	mainnet := false
	disabled := "disabled"
	auto := "auto"
	maxLoss := 25
	_, err = st.Put(ctx, user, crypto, OperatorSettingsPatch{
		Provider:          &deribit,
		DeribitUseMainnet: &mainnet,
		Secrets: map[string]string{
			"deribit_client_id":     "id",
			"deribit_client_secret": "sec",
		},
		AccountStatus:    &disabled,
		BotMode:          &auto,
		MaxLossEquityPct: &maxLoss,
	})
	if err != nil {
		t.Fatal(err)
	}

	row, err := st.GetDecrypting(ctx, user, crypto)
	if err != nil {
		t.Fatal(err)
	}
	if row == nil {
		t.Fatal("expected row")
	}
	if row.AccountStatus != "disabled" {
		t.Fatalf("AccountStatus: got %q want disabled", row.AccountStatus)
	}
	if row.BotMode != "auto" {
		t.Fatalf("BotMode: got %q want auto", row.BotMode)
	}
	if row.MaxLossEquityPct != 25 {
		t.Fatalf("MaxLossEquityPct: got %d want 25", row.MaxLossEquityPct)
	}
}

func TestOperatorSettingsTradingPrefsDefaults(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	user := "op2"

	deribit := "deribit"
	mainnet := false
	_, err = st.Put(ctx, user, crypto, OperatorSettingsPatch{
		Provider:          &deribit,
		DeribitUseMainnet: &mainnet,
		Secrets: map[string]string{
			"deribit_client_id":     "id",
			"deribit_client_secret": "sec",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	row, err := st.GetDecrypting(ctx, user, crypto)
	if err != nil {
		t.Fatal(err)
	}
	if row.AccountStatus != "active" || row.BotMode != "manual" || row.MaxLossEquityPct != 10 {
		t.Fatalf("defaults: got status=%q mode=%q max=%d", row.AccountStatus, row.BotMode, row.MaxLossEquityPct)
	}
}

func TestOperatorSettingsPatchValidation(t *testing.T) {
	t.Parallel()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	crypto := testOperatorSettingsCrypto(t)
	st := NewOperatorSettingsStore(db)
	ctx := context.Background()
	user := "op3"

	deribit := "deribit"
	mainnet := false
	base := OperatorSettingsPatch{
		Provider:          &deribit,
		DeribitUseMainnet: &mainnet,
		Secrets: map[string]string{
			"deribit_client_id":     "id",
			"deribit_client_secret": "sec",
		},
	}

	badAcct := "nope"
	p := base
	p.AccountStatus = &badAcct
	if _, err := st.Put(ctx, user, crypto, p); err == nil {
		t.Fatal("expected error for invalid account_status")
	}

	badMode := "robot"
	p = base
	p.BotMode = &badMode
	if _, err := st.Put(ctx, user, crypto, p); err == nil {
		t.Fatal("expected error for invalid bot_mode")
	}

	zero := 0
	p = base
	p.MaxLossEquityPct = &zero
	if _, err := st.Put(ctx, user, crypto, p); err == nil {
		t.Fatal("expected error for max_loss_equity_pct 0")
	}

	high := 51
	p = base
	p.MaxLossEquityPct = &high
	if _, err := st.Put(ctx, user, crypto, p); err == nil {
		t.Fatal("expected error for max_loss_equity_pct 51")
	}
}
