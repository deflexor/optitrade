// Package rpc implements JSON-RPC 2.0 over HTTP for Deribit API v2.
package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// AuthHeader resolves a bearer access token for private methods. If nil, no Authorization header is sent.
type AuthHeader func(ctx context.Context) (bearerToken string, err error)

// Client is a minimal Deribit JSON-RPC HTTP client.
type Client struct {
	baseURL string
	http    *http.Client
	id      atomic.Uint64
	auth    AuthHeader
}

// NewClient returns a client posting to baseURL (e.g. https://test.deribit.com/api/v2).
func NewClient(baseURL string, auth AuthHeader) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		auth: auth,
	}
}

// Call performs one JSON-RPC request and returns the raw result JSON or an error.
func (c *Client) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if params == nil {
		params = map[string]any{}
	}
	id := c.id.Add(1)
	req := Request{JSONRPC: "2.0", ID: id, Method: method, Params: params}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.auth != nil {
		tok, err := c.auth(ctx)
		if err != nil {
			return nil, err
		}
		if tok != "" {
			httpReq.Header.Set("Authorization", "bearer "+tok)
		}
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("deribit http %s: %s", resp.Status, redactBody(respBody))
	}
	var env Response
	if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("decode json-rpc: %w (body %s)", err, redactBody(respBody))
	}
	if env.Error != nil && len(*env.Error) > 0 && string(*env.Error) != "null" {
		var re RPCError
		if err := json.Unmarshal(*env.Error, &re); err != nil {
			return nil, fmt.Errorf("rpc error (raw): %s", redactBody(*env.Error))
		}
		return nil, &re
	}
	if env.ID != id {
		return nil, fmt.Errorf("json-rpc id mismatch: sent %d got %d", id, env.ID)
	}
	return env.Result, nil
}

// redactBody replaces patterns that might contain secrets in debug paths (no logging by default).
func redactBody(b []byte) string {
	if len(b) > 512 {
		return string(b[:512]) + "...(truncated)"
	}
	return string(b)
}
