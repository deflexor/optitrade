package dashboard

import (
	"fmt"
	"testing"
	"time"

	"github.com/dfr/optitrade/src/internal/opportunities"
	"github.com/dfr/optitrade/src/internal/okx"
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
