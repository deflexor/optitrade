// Package okx implements OKX API v5 REST signing and requests.
package okx

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://www.okx.com"

// Client is a minimal signed REST client for private endpoints.
type Client struct {
	BaseURL    string
	Key        string
	Secret     string
	Passphrase string
	Simulated  bool // sets x-simulated-trading: 1
	HTTP       *http.Client
}

// okxTimeFormat is the OK-ACCESS-TIMESTAMP layout (UTC with milliseconds).
const okxTimeFormat = "2006-01-02T15:04:05.000Z"

// Do signs and runs a request. path is like "/api/v5/account/balance".
func (c *Client) Do(ctx context.Context, method, path, body string) (json.RawMessage, error) {
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 30 * time.Second}
	}
	base := strings.TrimSuffix(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		base = defaultBaseURL
	}
	ts := time.Now().UTC().Format(okxTimeFormat)
	prehash := ts + strings.ToUpper(method) + path + body
	mac := hmac.New(sha256.New, []byte(c.Secret))
	_, _ = mac.Write([]byte(prehash))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, method, base+path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("OK-ACCESS-KEY", c.Key)
	req.Header.Set("OK-ACCESS-SIGN", sign)
	req.Header.Set("OK-ACCESS-TIMESTAMP", ts)
	req.Header.Set("OK-ACCESS-PASSPHRASE", c.Passphrase)
	req.Header.Set("Content-Type", "application/json")
	if c.Simulated {
		req.Header.Set("x-simulated-trading", "1")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<22))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("okx http %d: %s", resp.StatusCode, bytes.TrimSpace(raw))
	}
	var env struct {
		Code string          `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("okx decode: %w", err)
	}
	if env.Code != "" && env.Code != "0" {
		return nil, fmt.Errorf("okx error %s: %s", env.Code, env.Msg)
	}
	return env.Data, nil
}

// GetServerTime calls GET /api/v5/public/time (no auth).
func (c *Client) GetServerTime(ctx context.Context) (int64, error) {
	base := strings.TrimSuffix(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		base = defaultBaseURL
	}
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/v5/public/time", nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("okx public time http %d", resp.StatusCode)
	}
	var env struct {
		Code string `json:"code"`
		Data []struct {
			TS string `json:"ts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return 0, err
	}
	if env.Code != "" && env.Code != "0" || len(env.Data) == 0 {
		return 0, fmt.Errorf("okx time: code=%s", env.Code)
	}
	// ts is string millis
	var ms int64
	_, _ = fmt.Sscan(env.Data[0].TS, &ms)
	return ms, nil
}
