package okx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicClient_OrderBookToDeribitShape(t *testing.T) {
	const raw = `{"code":"0","data":[{"asks":[["0.05","1","0","1"]],"bids":[["0.04","2","0","2"]]}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(raw))
	}))
	defer srv.Close()

	pc := &PublicClient{BaseURL: srv.URL, HTTP: srv.Client()}
	book, err := pc.GetOrderBookDeribit(context.Background(), "BTC-USD-260327-95000-P", 5)
	if err != nil {
		t.Fatal(err)
	}
	if book.InstrumentName != "BTC-USD-260327-95000-P" {
		t.Fatalf("inst %q", book.InstrumentName)
	}
	if len(book.Bids) != 1 || len(book.Asks) != 1 {
		t.Fatalf("depth %+v", book)
	}
	if book.Bids[0].Price != 0.04 || book.Asks[0].Price != 0.05 {
		t.Fatalf("prices %+v %+v", book.Bids[0], book.Asks[0])
	}
	if book.Bids[0].Amount != 2 || book.Asks[0].Amount != 1 {
		t.Fatalf("amounts %+v %+v", book.Bids[0], book.Asks[0])
	}
}
