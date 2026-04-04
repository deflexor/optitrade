package dashboard

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/dfr/optitrade/src/internal/opportunities"
	"github.com/dfr/optitrade/src/internal/okx"
	"github.com/dfr/optitrade/src/internal/state/sqlite"
)

func TestBuildPutCreditSpecs_nearDatedChain(t *testing.T) {
	t.Parallel()
	exp := time.Now().UTC().Add(7 * 24 * time.Hour)
	yyyymmdd := exp.Format("20060102")
	yyMMdd := yyyymmdd[2:]
	short := int64(90000)
	long := short - 1000
	insts := []okx.InstrumentSummary{
		{InstId: fmt.Sprintf("BTC-USD-%s-%d-P", yyMMdd, short)},
		{InstId: fmt.Sprintf("BTC-USD-%s-%d-P", yyMMdd, long)},
	}
	specs := buildPutCreditSpecs(insts, 95000, 1000, 10, "BTC")
	if len(specs) != 1 {
		t.Fatalf("specs: %+v", specs)
	}
	if specs[0].ShortStrike != short || specs[0].Width != 1000 || specs[0].Expiry != yyyymmdd {
		t.Fatalf("spec: %+v", specs[0])
	}
}

func TestApplyMaxLossEquityGate_opensOnly(t *testing.T) {
	t.Parallel()
	rows := []opportunities.Row{
		{Recommend: "open", MaxLoss: "100"},
		{Recommend: "open", MaxLoss: "5000"},
		{Recommend: "pass", MaxLoss: "99999"},
	}
	applyMaxLossEquityGate(rows, 10000, 10) // limit 1000 USD
	if rows[0].Recommend != "open" {
		t.Fatalf("row0: %+v", rows[0])
	}
	if rows[1].Recommend != "pass" {
		t.Fatalf("row1: %+v", rows[1])
	}
	if rows[2].Recommend != "pass" {
		t.Fatalf("row2 should stay pass: %+v", rows[2])
	}
}

func TestMaybeAutoOpenAfterTick_hookOnceWhenNoDBRow(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	db, err := sqlite.Open(filepath.Join(dir, "auto.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	opp := NewOpportunityStore(db)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	rm := NewRunnerManager(log, nil, nil, nil, opp)

	var n int
	rm.AutoOpenHook = func(ctx context.Context, user, id string, cand opportunities.Row) error {
		n++
		return nil
	}

	ctx := context.Background()
	settings := &OperatorSettingsRow{BotMode: "auto", AccountStatus: "active"}
	rows := []opportunities.Row{{
		ID:           "cand-top",
		StrategyName: "credit_spread",
		Status:       opportunities.StatusCandidate,
		Recommend:    "open",
		Rationale:    "ok",
		Legs: []opportunities.LegQuote{
			{Instrument: "BTC-USD-260327-90000-P"},
			{Instrument: "BTC-USD-260327-89000-P"},
		},
		MaxProfit: "1", MaxLoss: "2", ExpectedEdge: "0.5", EdgeAfter: 0.4,
	}}

	rm.maybeAutoOpenAfterTick(ctx, "u1", settings, rows)
	if n != 1 {
		t.Fatalf("first tick: want hook once, got %d", n)
	}

	row := rows[0]
	legsJ, metaJ, err := encodeOpportunityRow(row)
	if err != nil {
		t.Fatal(err)
	}
	if err := opp.Upsert(ctx, &OpportunityRecord{
		ID:           row.ID,
		Username:     "u1",
		Status:       string(opportunities.StatusOpening),
		StrategyName: row.StrategyName,
		LegsJSON:     legsJ,
		MetaJSON:     metaJ,
	}); err != nil {
		t.Fatal(err)
	}

	rm.maybeAutoOpenAfterTick(ctx, "u1", settings, rows)
	if n != 1 {
		t.Fatalf("second tick with DB row: want still 1 hook call, got %d", n)
	}
}
