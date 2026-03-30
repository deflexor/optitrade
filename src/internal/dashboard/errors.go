package dashboard

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

const jsonContentType = "application/json; charset=utf-8"

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeAPIError writes a stable JSON error object. message is safe for clients (no secrets).
func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

func logHandlerError(log *slog.Logger, msg string, err error) {
	if err == nil {
		return
	}
	log.Error(msg, "err", err)
}
