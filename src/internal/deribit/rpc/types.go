package rpc

import (
	"encoding/json"
	"fmt"
)

// Request is a JSON-RPC 2.0 request envelope.
type Request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      uint64 `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// Response is a JSON-RPC 2.0 response envelope (Deribit adds usIn/usOut test fields).
type Response struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      uint64           `json:"id"`
	Result  json.RawMessage  `json:"result"`
	Error   *json.RawMessage `json:"error"`
	Test    json.RawMessage  `json:"test,omitempty"`
	UsIn    int64            `json:"usIn,omitempty"`
	UsOut   int64            `json:"usOut,omitempty"`
}

// RPCError is Deribit's JSON-RPC error object.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("deribit rpc: %d %s", e.Code, e.Message)
}
