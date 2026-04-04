package dashboard

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func (s *Server) handleTradingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := requestUser(r.Context())
	if !ok || s.settings == nil || s.settingsCrypto == nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "settings unavailable")
		return
	}
	row, err := s.settings.GetDecrypting(r.Context(), user, s.settingsCrypto)
	if err != nil {
		logHandlerError(s.log, "trading_status", err)
		writeAPIError(w, http.StatusInternalServerError, "server_error", "could not load settings")
		return
	}
	accountStatus := "active"
	botMode := "manual"
	if row != nil {
		accountStatus = row.AccountStatus
		botMode = row.BotMode
	}
	runnerRunning := false
	if s.runnerRunning != nil {
		runnerRunning = s.runnerRunning(user)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"account_status": accountStatus,
		"bot_mode":       botMode,
		"runner_running": runnerRunning,
	})
}

func (s *Server) handleTradingModePut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, ok := requestUser(r.Context())
	if !ok || s.settings == nil || s.settingsCrypto == nil {
		writeAPIError(w, http.StatusInternalServerError, "server_error", "settings unavailable")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_body", "could not read body")
		return
	}
	var body struct {
		BotMode string `json:"bot_mode"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_body", "expected JSON")
		return
	}
	mode := strings.ToLower(strings.TrimSpace(body.BotMode))
	if mode != "manual" && mode != "auto" && mode != "paused" {
		writeAPIError(w, http.StatusBadRequest, "invalid_body", "invalid bot_mode (want manual, auto, or paused)")
		return
	}
	if _, err := s.settings.Put(r.Context(), user, s.settingsCrypto, OperatorSettingsPatch{BotMode: &mode}); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_settings", err.Error())
		return
	}
	s.invalidateExchangeCache(user)
	s.reconcileRunnersAsync()
	writeJSON(w, http.StatusOK, map[string]string{"bot_mode": mode})
}
