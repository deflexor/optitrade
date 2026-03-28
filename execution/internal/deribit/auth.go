package deribit

// Authentication uses Deribit JSON-RPC public/auth with grant_type client_credentials
// (see https://docs.deribit.com/). Access tokens are cached in memory with mutex;
// do not log client_secret or access_token values.

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/dfr/optitrade/execution/internal/deribit/rpc"
)

// Credentials holds Deribit API application keys (never log).
type Credentials struct {
	ClientID     string
	ClientSecret string
}

type tokenManager struct {
	mu sync.Mutex

	baseURL string
	creds   Credentials

	rpcNoAuth *rpc.Client

	token  string
	expiry time.Time
}

func newTokenManager(baseURL string, creds Credentials) *tokenManager {
	return &tokenManager{
		baseURL:   baseURL,
		creds:     creds,
		rpcNoAuth: rpc.NewClient(baseURL, nil),
	}
}

// authResponse matches public/auth result (subset).
type authResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (tm *tokenManager) authorize(ctx context.Context) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// Refresh a minute before expiry to reduce 401 races.
	if tm.token != "" && time.Until(tm.expiry) > 60*time.Second {
		return tm.token, nil
	}
	return tm.refreshLocked(ctx)
}

func (tm *tokenManager) refreshLocked(ctx context.Context) (string, error) {
	params := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     tm.creds.ClientID,
		"client_secret": tm.creds.ClientSecret,
	}
	raw, err := tm.rpcNoAuth.Call(ctx, "public/auth", params)
	if err != nil {
		return "", err
	}
	var ar authResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return "", err
	}
	if ar.AccessToken == "" {
		return "", errAuthEmpty
	}
	tm.token = ar.AccessToken
	if ar.ExpiresIn <= 0 {
		ar.ExpiresIn = 3600
	}
	tm.expiry = time.Now().Add(time.Duration(ar.ExpiresIn) * time.Second)
	return tm.token, nil
}
