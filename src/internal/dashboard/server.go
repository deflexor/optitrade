package dashboard

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// Server is the dashboard BFF (API + static assets).
type Server struct {
	log *slog.Logger
	mux *http.ServeMux

	authMu sync.RWMutex
	auth   *DashboardAuthFile

	sessions *SessionStore

	settingsCrypto *SettingsCrypto
	settings       *OperatorSettingsStore

	// testXchg, when set, is used for all routes (unit tests); production leaves it nil.
	testXchg exchangeReader

	// runnerRunning reports whether the per-user trading runner loop is active (nil = always false).
	runnerRunning func(username string) bool

	xchgMu    sync.Mutex
	xchgCache map[string]cachedExchange

	started time.Time

	previews *previewStore
}

// Options configures the dashboard server.
type Options struct {
	Logger *slog.Logger
	Auth   *DashboardAuthFile
	// Sessions and operator settings share one migrated SQLite file opened by the caller.
	Sessions *SessionStore
	// SettingsCrypto and Settings enable per-operator venue credentials (required in production).
	SettingsCrypto *SettingsCrypto
	Settings       *OperatorSettingsStore
	// TestExchange bypasses DB resolution (tests only).
	TestExchange exchangeReader
	// RunnerRunning is optional; when nil, GET /trading/status reports runner_running: false.
	RunnerRunning func(username string) bool
}

// NewServer builds a dashboard HTTP handler tree.
func NewServer(opts Options) *Server {
	log := opts.Logger
	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	s := &Server{
		log:            log,
		mux:            http.NewServeMux(),
		auth:           opts.Auth,
		sessions:       opts.Sessions,
		settingsCrypto: opts.SettingsCrypto,
		settings:       opts.Settings,
		testXchg:       opts.TestExchange,
		runnerRunning:  opts.RunnerRunning,
		xchgCache:      map[string]cachedExchange{},
		started:        time.Now(),
		previews:       newPreviewStore(),
	}

	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	api := http.NewServeMux()
	api.Handle("POST /auth/login", http.HandlerFunc(s.handleLogin))

	api.Handle("GET /health", http.HandlerFunc(s.handleHealthAPI))

	protected := http.NewServeMux()
	protected.Handle("POST /auth/logout", http.HandlerFunc(s.handleLogout))
	protected.Handle("GET /auth/me", http.HandlerFunc(s.handleMe))
	protected.Handle("GET /overview", http.HandlerFunc(s.handleOverview))
	protected.Handle("GET /positions/open", http.HandlerFunc(s.handleOpenPositions))
	protected.Handle("GET /positions/closed", http.HandlerFunc(s.handleClosedPositions))
	protected.Handle("GET /positions/{id}", http.HandlerFunc(s.handlePositionDetail))
	protected.Handle("POST /positions/{id}/close/preview", http.HandlerFunc(s.handleClosePreview))
	protected.Handle("POST /positions/{id}/close/confirm", http.HandlerFunc(s.handleCloseConfirm))
	protected.Handle("POST /rebalance/preview", http.HandlerFunc(s.handleRebalancePreview))
	protected.Handle("POST /rebalance/confirm", http.HandlerFunc(s.handleRebalanceConfirm))
	protected.Handle("GET /settings", http.HandlerFunc(s.handleSettingsGet))
	protected.Handle("PUT /settings", http.HandlerFunc(s.handleSettingsPut))
	protected.Handle("GET /trading/status", http.HandlerFunc(s.handleTradingStatus))
	protected.Handle("PUT /trading/mode", http.HandlerFunc(s.handleTradingModePut))

	api.Handle("/", s.requireAuth(protected))

	s.mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	s.mountStatic()
	return s
}

// ReloadAuth replaces the allowlist in memory (e.g. after SIGHUP).
func (s *Server) ReloadAuth(auth *DashboardAuthFile) {
	s.authMu.Lock()
	defer s.authMu.Unlock()
	s.auth = auth
}

// Handler returns the root HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) mountStatic() {
	assets := embeddedAssets()
	if assets == nil {
		s.mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error":   "not_found",
				"message": "static assets not available",
			})
		}))
		return
	}

	s.mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" || name == "." {
			name = "index.html"
		}
		if _, err := fs.Stat(assets, name); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error":   "not_found",
				"message": "no such asset",
			})
			return
		}
		http.ServeFileFS(w, r, assets, name)
	}))
}
