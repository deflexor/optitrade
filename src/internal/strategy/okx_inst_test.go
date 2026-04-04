package strategy

import "testing"

func TestOKXOptionInstID_put(t *testing.T) {
	got := OKXOptionInstID("BTC", "20260327", 95000, false)
	want := "BTC-USD-260327-95000-P"
	if got != want {
		t.Fatalf("OKXOptionInstID(...) = %q, want %q", got, want)
	}
}

func TestVerticalPutCreditOKX(t *testing.T) {
	legs, err := VerticalPutCreditOKX("BTC", "20260327", 95000, 500)
	if err != nil {
		t.Fatalf("VerticalPutCreditOKX: %v", err)
	}
	if len(legs) != 2 {
		t.Fatalf("len(legs) = %d, want 2", len(legs))
	}
	if legs[0].Side != LegSell || legs[0].Instrument != "BTC-USD-260327-95000-P" {
		t.Fatalf("short leg: got %+v, want sell BTC-USD-260327-95000-P", legs[0])
	}
	if legs[1].Side != LegBuy || legs[1].Instrument != "BTC-USD-260327-94500-P" {
		t.Fatalf("long leg: got %+v, want buy BTC-USD-260327-94500-P", legs[1])
	}
}
