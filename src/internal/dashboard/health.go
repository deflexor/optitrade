package dashboard

import (
	"net/http"
	"runtime"
	"time"
)

func (s *Server) handleHealthAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	uptime := int64(time.Since(s.started).Seconds())
	collected := time.Now().UTC().Format(time.RFC3339)

	mode := "unknown"
	reachable := false
	detail := "exchange_not_configured"

	var xchg exchangeReader
	if s.testXchg != nil {
		xchg = s.testXchg
		mode = "test"
		detail = ""
	} else if s.sessions != nil && s.settings != nil && s.settingsCrypto != nil {
		if c, err := r.Cookie(sessionCookieName); err == nil && c != nil && c.Value != "" {
			if user, err := s.sessions.LookupUsername(r.Context(), c.Value); err == nil {
				s.authMu.RLock()
				auth := s.auth
				s.authMu.RUnlock()
				if auth != nil && auth.Lookup(user) != nil {
					if m, ok := s.tradingModeForUsername(r.Context(), user); ok {
						mode = m
					}
					var err error
					xchg, err = s.resolveExchange(r.Context(), user)
					if err != nil {
						detail = "exchange_error"
					} else if xchg == nil {
						detail = "exchange_not_configured"
					} else {
						detail = ""
					}
				}
			}
		}
	}

	if xchg != nil {
		ctx, cancel := rpcTimeout(r.Context())
		_, err := xchg.GetServerTime(ctx)
		cancel()
		if err == nil {
			reachable = true
		} else {
			reachable = false
			if detail == "" {
				detail = "exchange_unreachable"
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"health": map[string]any{
			"uptime_seconds":            uptime,
			"memory_heap_alloc_bytes": int64(ms.Alloc),
			"memory_sys_bytes":        int64(ms.Sys),
			"collected_at":            collected,
		},
		"trading": map[string]any{
			"mode":               mode,
			"exchange_reachable": reachable,
			"detail":             detail,
		},
	})
}
