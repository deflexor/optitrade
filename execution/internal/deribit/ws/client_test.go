package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestReconnectResubscribes(t *testing.T) {
	var serverSubs atomic.Int32
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		readOne := false
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				return
			}
			var msg map[string]any
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			method, _ := msg["method"].(string)
			idf, _ := msg["id"].(float64)
			id := uint64(idf)
			switch method {
			case "public/subscribe":
				serverSubs.Add(1)
				readOne = true
				_ = c.WriteJSON(map[string]any{
					"jsonrpc": "2.0",
					"id":      id,
					"result":  []string{"book.TEST.1ms"},
				})
				if readOne && serverSubs.Load() == 1 {
					// Force disconnect so client reconnect logic runs.
					return
				}
			default:
				_ = c.WriteJSON(map[string]any{
					"jsonrpc": "2.0",
					"id":      id,
					"result":  json.RawMessage("0"),
				})
			}
		}
	}))
	defer srv.Close()

	u := "ws" + strings.TrimPrefix(srv.URL, "http")

	cli := NewClient(u, nil, nil)
	cli.minBackoff = 10 * time.Millisecond
	cli.maxBackoff = 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cli.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	if err := cli.Subscribe(ctx, "book.TEST.1ms"); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for serverSubs.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if serverSubs.Load() < 2 {
		t.Fatalf("expected 2 subscribe RPCs after reconnect, got %d", serverSubs.Load())
	}
}
