package dashboard

import (
	"testing"

	"github.com/dfr/optitrade/src/internal/deribit"
)

func TestPositionRowID_roundTrip(t *testing.T) {
	t.Parallel()
	dir := "buy"
	p := deribit.Position{
		InstrumentName: "BTC-PERPETUAL",
		Direction:      &dir,
	}
	raw := positionRowID(p)
	inst, d, ok := parsePositionRowID(raw)
	if !ok {
		t.Fatalf("parsePositionRowID(%q)", raw)
	}
	if inst != "BTC-PERPETUAL" || d != "buy" {
		t.Fatalf("got inst=%q dir=%q", inst, d)
	}
}

func TestPositionRowID_specialCharsEscaped(t *testing.T) {
	t.Parallel()
	dir := "sell"
	p := deribit.Position{
		InstrumentName: "OPTION A|B",
		Direction:      &dir,
	}
	raw := positionRowID(p)
	inst, d, ok := parsePositionRowID(raw)
	if !ok {
		t.Fatal(ok)
	}
	if inst != "OPTION A|B" || d != "sell" {
		t.Fatalf("got inst=%q dir=%q", inst, d)
	}
	if raw == "OPTION A|B|sell" {
		t.Fatal("expected escaped instrument")
	}
}

func TestParsePositionRowID_invalid(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"", "onlyone", "%ZZ|%ZZ"} {
		_, _, ok := parsePositionRowID(s)
		if ok {
			t.Fatalf("expected false for %q", s)
		}
	}
}
