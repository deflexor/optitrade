package market

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dfr/optitrade/execution/internal/deribit"
)

func fixturePath(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	dir := filepath.Dir(file)
	// execution/internal/market -> repo root is three levels up
	root := filepath.Join(dir, "..", "..", "..")
	p := filepath.Join(root, "tests", "fixtures", "deribit", name)
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(fixturePath(t, name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

type stubInstruments struct {
	all []deribit.Instrument
}

func (s *stubInstruments) GetInstruments(ctx context.Context, p *deribit.GetInstrumentsParams) ([]deribit.Instrument, error) {
	if p == nil || p.Currency == nil {
		return append([]deribit.Instrument(nil), s.all...), nil
	}
	cur := *p.Currency
	var out []deribit.Instrument
	for i := range s.all {
		if baseCurrency(s.all[i]) == cur {
			out = append(out, s.all[i])
		}
	}
	return out, nil
}

func TestDiscoverActiveOptions_fixtureNoNetwork(t *testing.T) {
	var all []deribit.Instrument
	if err := json.Unmarshal(readFixture(t, "get_instruments.json"), &all); err != nil {
		t.Fatal(err)
	}
	api := &stubInstruments{all: all}
	ctx := context.Background()
	got, err := DiscoverActiveOptions(ctx, api, InstrumentFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 active options BTC+ETH, got %d", len(got))
	}
	for _, i := range got {
		if i.Kind == nil || *i.Kind != "option" {
			t.Fatalf("non-option leaked: %#v", i)
		}
		if i.IsActive != nil && !*i.IsActive {
			t.Fatalf("inactive leaked: %s", i.InstrumentName)
		}
		if i.InstrumentName == "BTC-PERPETUAL" {
			t.Fatal("future leaked")
		}
	}
}

func TestDiscoverActiveOptions_whitelist(t *testing.T) {
	var all []deribit.Instrument
	if err := json.Unmarshal(readFixture(t, "get_instruments.json"), &all); err != nil {
		t.Fatal(err)
	}
	api := &stubInstruments{all: all}
	ctx := context.Background()
	got, err := DiscoverActiveOptions(ctx, api, InstrumentFilter{
		BaseCurrencies: []string{"BTC"},
		NameWhitelist:  map[string]struct{}{"BTC-31DEC99-50000-C": {}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].InstrumentName != "BTC-31DEC99-50000-C" {
		t.Fatalf("got %#v", got)
	}
}

type stubVol struct {
	res *deribit.VolatilityIndexData
}

func (s *stubVol) GetVolatilityIndexData(ctx context.Context, params deribit.GetVolatilityIndexDataParams) (*deribit.VolatilityIndexData, error) {
	return s.res, nil
}

func TestFetchLatestVolIndex_fixtureNoNetwork(t *testing.T) {
	var data deribit.VolatilityIndexData
	if err := json.Unmarshal(readFixture(t, "get_volatility_index_data.json"), &data); err != nil {
		t.Fatal(err)
	}
	api := &stubVol{res: &data}
	closePx, ts, err := FetchLatestVolIndex(context.Background(), api, "BTC", 1700000123000, time.Hour, "60")
	if err != nil {
		t.Fatal(err)
	}
	if ts != 1700000060000 {
		t.Fatalf("candle ts: got %d", ts)
	}
	if closePx != 0.558 {
		t.Fatalf("close: got %v", closePx)
	}
}

func TestOrderBookCache_depth(t *testing.T) {
	var ob deribit.OrderBook
	if err := json.Unmarshal(readFixture(t, "get_order_book.json"), &ob); err != nil {
		t.Fatal(err)
	}
	c := NewOrderBookCache(1)
	c.Update(&ob)
	bids, asks := c.Depth(0)
	if len(bids) != 1 || len(asks) != 1 {
		t.Fatalf("depth trim: bids=%d asks=%d", len(bids), len(asks))
	}
	bid, ask, ok := c.BestBidAsk()
	if !ok || bid != 50000 || ask != 50002 {
		t.Fatalf("tob: %v %v %v", bid, ask, ok)
	}
}

func TestBuildSnapshotFromBook_flags(t *testing.T) {
	var ob deribit.OrderBook
	if err := json.Unmarshal(readFixture(t, "get_order_book.json"), &ob); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_100, 0)
	bookTS := now.Add(-time.Minute * 5)
	s := BuildSnapshotFromBook(ob.InstrumentName, ob, bookTS, now, time.Minute, 0, 0, 0.55, 1700)
	if !s.HasFlag(StaleBook) {
		t.Fatal("expected stale book")
	}

	now2 := time.Unix(1_700_000_100, 0)
	bookTS2 := now2.Add(-time.Second)
	spreadWide := deribit.OrderBook{
		InstrumentName: ob.InstrumentName,
		Bids:           []deribit.PriceLevel{{Price: 100, Amount: 1}},
		Asks:           []deribit.PriceLevel{{Price: 200, Amount: 1}},
	}
	s2 := BuildSnapshotFromBook(ob.InstrumentName, spreadWide, bookTS2, now2, time.Minute, 50, 0, 0, 0)
	if !s2.HasFlag(WideSpread) {
		t.Fatal("expected wide spread (abs)")
	}
}
