// Package ws is a Deribit WebSocket JSON-RPC client with reconnect and resubscribe.
// Auth uses the same public/auth client_credentials flow as HTTPS (Deribit API v2).
package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/dfr/optitrade/execution/internal/deribit/rpc"
)

// Credentials are API keys for public/auth over the socket (never log).
type Credentials struct {
	ClientID     string
	ClientSecret string
}

// NotificationHandler is invoked for subscription pushes (method "subscription").
type NotificationHandler func(channel string, data json.RawMessage)

// ConnHealth reports logical socket connectivity for protective mode (WP12 / FR-009).
type ConnHealth int

const (
	// ConnDown indicates the reader stopped or dial failed (reconnect pending or closed).
	ConnDown ConnHealth = iota
	// ConnUp indicates a successful connect or resubscribe after (re)dial.
	ConnUp
)

// HealthHandler is optional; invoked on disconnect before reconnect and after a successful (re)dial.
type HealthHandler func(ConnHealth)

// Client manages one JSON-RPC WebSocket with reconnect and resubscribed channels.
type Client struct {
	URL    string
	Creds  *Credentials
	OnNote NotificationHandler
	// OnHealth notifies session FSM of connectivity (staleness / disconnect duration triggers).
	OnHealth HealthHandler
	Dialer   *websocket.Dialer

	minBackoff time.Duration
	maxBackoff time.Duration

	mu   sync.RWMutex
	conn *websocket.Conn

	closed atomic.Bool

	writeMu sync.Mutex

	nextID atomic.Uint64
	pendMu sync.Mutex
	pend   map[uint64]chan rpcResult

	subsMu sync.Mutex
	subs   []string
}

type rpcResult struct {
	data json.RawMessage
	err  error
}

// NewClient returns a client with default exponential backoff (1s to 60s cap).
func NewClient(url string, creds *Credentials, on NotificationHandler) *Client {
	return &Client{
		URL:        url,
		Creds:      creds,
		OnNote:     on,
		Dialer:     websocket.DefaultDialer,
		minBackoff: time.Second,
		maxBackoff: 60 * time.Second,
		pend:       make(map[uint64]chan rpcResult),
	}
}

// Connect dials the WebSocket, starts the reader, then authenticates when Creds is set.
func (c *Client) Connect(ctx context.Context) error {
	if c.closed.Load() {
		return errClosed
	}
	conn, _, err := c.Dialer.DialContext(ctx, c.URL, http.Header{})
	if err != nil {
		return err
	}
	conn.SetReadLimit(8 << 20)
	c.armReadDeadline(conn)

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	go c.readLoop()

	if c.Creds != nil {
		params := map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     c.Creds.ClientID,
			"client_secret": c.Creds.ClientSecret,
		}
		if _, err := c.Call(ctx, "public/auth", params); err != nil {
			_ = c.Close()
			return fmt.Errorf("ws auth: %w", err)
		}
	}
	c.notifyHealth(ConnUp)
	return nil
}

func (c *Client) armReadDeadline(conn *websocket.Conn) {
	_ = conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		return nil
	})
}

// Close stops reconnects and closes the socket.
func (c *Client) Close() error {
	c.closed.Store(true)
	c.notifyHealth(ConnDown)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *Client) notifyHealth(h ConnHealth) {
	if c == nil || c.OnHealth == nil {
		return
	}
	c.OnHealth(h)
}

// Subscribe remembers channels and sends public/subscribe.
func (c *Client) Subscribe(ctx context.Context, channels ...string) error {
	if len(channels) == 0 {
		return nil
	}
	c.subsMu.Lock()
	c.subs = append(c.subs, channels...)
	c.subsMu.Unlock()
	_, err := c.Call(ctx, "public/subscribe", map[string]any{"channels": channels})
	return err
}

// Call performs JSON-RPC over the active WebSocket.
func (c *Client) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if params == nil {
		params = map[string]any{}
	}
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return nil, errNoConn
	}
	id := c.nextID.Add(1)
	req := rpc.Request{JSONRPC: "2.0", ID: id, Method: method, Params: params}
	wait := make(chan rpcResult, 1)
	c.pendMu.Lock()
	c.pend[id] = wait
	c.pendMu.Unlock()
	defer func() {
		c.pendMu.Lock()
		delete(c.pend, id)
		c.pendMu.Unlock()
	}()

	c.writeMu.Lock()
	err := conn.WriteJSON(req)
	c.writeMu.Unlock()
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-wait:
		return res.data, res.err
	}
}

func (c *Client) readLoop() {
	backoff := c.minBackoff
	for !c.closed.Load() {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()
		if conn == nil {
			return
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if c.closed.Load() {
				return
			}
			c.notifyHealth(ConnDown)
			_ = conn.Close()
			c.reconnectWithBackoff(&backoff)
			continue
		}
		backoff = c.minBackoff
		c.dispatch(msg)
	}
}

func (c *Client) dispatch(msg []byte) {
	var head struct {
		Method string           `json:"method"`
		ID     uint64           `json:"id"`
		Result json.RawMessage  `json:"result"`
		Error  *json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(msg, &head); err != nil {
		return
	}
	if head.Method == "subscription" {
		c.handleSubscription(msg)
		return
	}
	if head.ID == 0 {
		return
	}
	c.pendMu.Lock()
	ch, ok := c.pend[head.ID]
	c.pendMu.Unlock()
	if !ok {
		return
	}
	if head.Error != nil && len(*head.Error) > 0 && string(*head.Error) != "null" {
		var re rpc.RPCError
		if err := json.Unmarshal(*head.Error, &re); err != nil {
			ch <- rpcResult{err: fmt.Errorf("rpc error decode: %w", err)}
			return
		}
		ch <- rpcResult{err: &re}
		return
	}
	ch <- rpcResult{data: head.Result}
}

func (c *Client) handleSubscription(raw []byte) {
	if c.OnNote == nil {
		return
	}
	var env struct {
		Params struct {
			Channel string          `json:"channel"`
			Data    json.RawMessage `json:"data"`
		} `json:"params"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return
	}
	c.OnNote(env.Params.Channel, env.Params.Data)
}

func (c *Client) reconnectWithBackoff(backoff *time.Duration) {
	for !c.closed.Load() {
		time.Sleep(*backoff)
		next := *backoff * 2
		if next > c.maxBackoff {
			next = c.maxBackoff
		}
		*backoff = next

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := c.redial(ctx)
		cancel()
		if err == nil {
			return
		}
	}
}

// redial replaces the socket, re-authenticates, and resubscribes (no new readLoop).
func (c *Client) redial(ctx context.Context) error {
	if c.closed.Load() {
		return errClosed
	}
	conn, _, err := c.Dialer.DialContext(ctx, c.URL, http.Header{})
	if err != nil {
		return err
	}
	conn.SetReadLimit(8 << 20)
	c.armReadDeadline(conn)

	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = conn
	c.mu.Unlock()

	if c.Creds != nil {
		params := map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     c.Creds.ClientID,
			"client_secret": c.Creds.ClientSecret,
		}
		if _, err := c.Call(ctx, "public/auth", params); err != nil {
			return err
		}
	}

	c.subsMu.Lock()
	channels := append([]string(nil), c.subs...)
	c.subsMu.Unlock()
	if len(channels) > 0 {
		if _, err := c.Call(ctx, "public/subscribe", map[string]any{"channels": channels}); err != nil {
			return err
		}
	}
	c.notifyHealth(ConnUp)
	return nil
}

var (
	errClosed = errors.New("ws: client closed")
	errNoConn = errors.New("ws: not connected")
)
