package okx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBatchPlaceOrders_parsesOrdIDs(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v5/trade/batch-orders" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		var orders []BatchPlaceOrderItem
		if err := json.Unmarshal(b, &orders); err != nil {
			t.Errorf("body: %s", b)
			http.Error(w, err.Error(), 400)
			return
		}
		if len(orders) != 2 || orders[0].InstID != "BTC-USD-260327-90000-P" {
			t.Errorf("orders %+v", orders)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"0","data":[{"ordId":"oid1","sCode":"0","sMsg":""},{"ordId":"oid2","sCode":"0","sMsg":""}]}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		BaseURL:    strings.TrimSuffix(srv.URL, "/"),
		Key:        "k",
		Secret:     strings.Repeat("a", 32),
		Passphrase: "p",
		HTTP:       srv.Client(),
	}
	ids, err := c.BatchPlaceOrders(context.Background(), []BatchPlaceOrderItem{
		{InstID: "BTC-USD-260327-90000-P", Side: "sell"},
		{InstID: "BTC-USD-260327-91000-P", Side: "buy"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != "oid1" || ids[1] != "oid2" {
		t.Fatalf("got %v", ids)
	}
}

func TestBatchCancelOrders_checksSCode(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v5/trade/cancel-batch-orders" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"0","data":[{"ordId":"x","sCode":"51400","sMsg":"fail"}]}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		BaseURL:    strings.TrimSuffix(srv.URL, "/"),
		Key:        "k",
		Secret:     strings.Repeat("b", 32),
		Passphrase: "p",
		HTTP:       srv.Client(),
	}
	err := c.BatchCancelOrders(context.Background(), []BatchCancelItem{{InstID: "BTC-USD-1", OrdID: "x"}})
	if err == nil || !strings.Contains(err.Error(), "51400") {
		t.Fatalf("want sCode error, got %v", err)
	}
}
