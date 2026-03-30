package dashboard

import (
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/dfr/optitrade/src/internal/deribit"
)

func tradingModeFromEnv() string {
	base := strings.TrimSpace(os.Getenv("DERIBIT_BASE_URL"))
	if base == "" {
		base = deribit.TestnetRPCBaseURL
	}
	if strings.Contains(strings.ToLower(base), "test.") {
		return "test"
	}
	return "live"
}

func (s *Server) handleHealthAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	uptime := int64(time.Since(s.started).Seconds())
	collected := time.Now().UTC().Format(time.RFC3339)

	mode := tradingModeFromEnv()
	reachable := false
	detail := ""
	if s.xchg != nil {
		ctx, cancel := rpcTimeout(r.Context())
		_, err := s.xchg.GetServerTime(ctx)
		cancel()
		if err == nil {
			reachable = true
		} else {
			detail = "exchange_unreachable"
		}
	} else {
		detail = "exchange_not_configured"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"health": map[string]any{
			"uptime_seconds":           uptime,
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
