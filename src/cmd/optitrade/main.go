package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/dashboard"
	"github.com/dfr/optitrade/src/internal/deribit"
	"github.com/dfr/optitrade/src/internal/opportunities"
	"github.com/dfr/optitrade/src/internal/observe"
	"github.com/dfr/optitrade/src/internal/state/sqlite"
)

const version = "0.0.1-dev"

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if addr, ok := dashboardAddrFromLeadingFlags(os.Args[1:]); ok {
		if err := runDashboardShortcut(log, addr); err != nil {
			log.Error("dashboard", "err", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version", "-v", "--version":
		fmt.Printf("optitrade %s\n", version)
	case "observe":
		if err := runObserve(log, os.Args[2:]); err != nil {
			log.Error("observe", "err", err)
			os.Exit(1)
		}
	case "smoke-order":
		if err := runSmokeOrder(log, os.Args[2:]); err != nil {
			log.Error("smoke-order", "err", err)
			os.Exit(1)
		}
	case "dashboard":
		if err := runDashboardCmd(log, os.Args[2:]); err != nil {
			log.Error("dashboard", "err", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `optitrade — Deribit options execution helper (testnet-first).

Binary: build with: (cd src && go build -o optitrade ./cmd/optitrade)

Commands:
  version              Print version
  observe              Read-only loop: positions + order book (needs DERIBIT_CLIENT_ID/SECRET)
  smoke-order          Tiny post-only test order (needs %s=1 and testnet policy)
  dashboard            Operator HTTP UI + API (set -listen or OPTITRADE_DASHBOARD_LISTEN)

Shortcut:
  optitrade --dashboard-listen=:8080   Same as dashboard with that listen address

Env:
  DERIBIT_CLIENT_ID, DERIBIT_CLIENT_SECRET   API keys (testnet keys for testnet URL)
  DERIBIT_BASE_URL                           Default %s
  OPTITRADE_POLICY_PATH                      Policy JSON for smoke-order gate
  OPTITRADE_DASHBOARD_LISTEN                 Dashboard listen addr (e.g. 127.0.0.1:8080)
  OPTITRADE_DASHBOARD_AUTH_PATH              Dashboard allowlist JSON (username + password_hash)
  OPTITRADE_DASHBOARD_SESSION_PATH           SQLite file for dashboard sessions
  OPTITRADE_SETTINGS_SECRET                  32-byte key: base64, hex, or raw (or use OPTITRADE_SETTINGS_KEY_FILE)
  OPTITRADE_SETTINGS_KEY_FILE                File with raw 32-byte key for encrypting operator API settings

`, observe.EnvAllowTestnetOrders, deribit.TestnetRPCBaseURL)
}

func dashboardAddrFromLeadingFlags(args []string) (string, bool) {
	if len(args) == 0 || !strings.HasPrefix(args[0], "-") {
		return "", false
	}
	fs := flag.NewFlagSet("optitrade", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	listen := fs.String("listen", "", "HTTP listen address (dashboard shortcut)")
	dashListen := fs.String("dashboard-listen", "", "HTTP listen address (dashboard shortcut)")
	if err := fs.Parse(args); err != nil {
		return "", false
	}
	if fs.NArg() != 0 {
		return "", false
	}
	addr := strings.TrimSpace(*dashListen)
	if addr == "" {
		addr = strings.TrimSpace(*listen)
	}
	if addr == "" {
		addr = strings.TrimSpace(os.Getenv("OPTITRADE_DASHBOARD_LISTEN"))
	}
	if addr == "" {
		return "", false
	}
	return addr, true
}

func runDashboardCmd(log *slog.Logger, args []string) error {
	fs := flag.NewFlagSet("dashboard", flag.ContinueOnError)
	listen := fs.String(
		"listen",
		strings.TrimSpace(os.Getenv("OPTITRADE_DASHBOARD_LISTEN")),
		"HTTP listen address (e.g. 127.0.0.1:8080)",
	)
	dashListen := fs.String("dashboard-listen", "", "alias for -listen")
	authPath := fs.String(
		"auth",
		strings.TrimSpace(os.Getenv("OPTITRADE_DASHBOARD_AUTH_PATH")),
		"path to dashboard allowlist JSON (OPTITRADE_DASHBOARD_AUTH_PATH)",
	)
	sessionPath := fs.String(
		"session-db",
		strings.TrimSpace(os.Getenv("OPTITRADE_DASHBOARD_SESSION_PATH")),
		"SQLite path for dashboard sessions (OPTITRADE_DASHBOARD_SESSION_PATH)",
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("dashboard: unexpected arguments: %q", fs.Args())
	}
	addr := strings.TrimSpace(*dashListen)
	if addr == "" {
		addr = strings.TrimSpace(*listen)
	}
	if addr == "" {
		return fmt.Errorf("dashboard: set -listen or OPTITRADE_DASHBOARD_LISTEN")
	}
	return runDashboardCmdFull(log, addr, *authPath, *sessionPath)
}

func runDashboardShortcut(log *slog.Logger, addr string) error {
	return runDashboardCmdFull(log, addr, "", "")
}

func runDashboardCmdFull(log *slog.Logger, addr, authPath, sessionPath string) error {
	loadDashboardEnv(log)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var auth *dashboard.DashboardAuthFile
	sessPath := strings.TrimSpace(sessionPath)
	if sessPath == "" {
		sessPath = filepath.Clean("optitrade-dashboard.sqlite")
	}
	db, err := sqlite.Open(sessPath)
	if err != nil {
		return fmt.Errorf("dashboard session db: %w", err)
	}
	defer db.Close()

	if p := strings.TrimSpace(authPath); p != "" {
		var loadErr error
		auth, loadErr = dashboard.LoadDashboardAuthFile(p)
		if loadErr != nil {
			return fmt.Errorf("dashboard auth: %w", loadErr)
		}
	} else {
		var err error
		auth, err = dashboard.DefaultEmbeddedAuth()
		if err != nil {
			return fmt.Errorf("dashboard default auth: %w", err)
		}
		log.Info("dashboard using embedded dev allowlist (user opti / password opti); set OPTITRADE_DASHBOARD_AUTH_PATH to use a file")
	}

	crypto, err := dashboard.LoadSettingsCrypto()
	if err != nil {
		return fmt.Errorf("dashboard: %w", err)
	}
	settingsStore := dashboard.NewOperatorSettingsStore(db)
	var policy *config.Policy
	if p := config.PolicyPathFromEnv(); p != "" {
		var lerr error
		policy, lerr = config.LoadFile(p)
		if lerr != nil {
			return fmt.Errorf("dashboard policy: %w", lerr)
		}
	} else {
		log.Warn("OPTITRADE_POLICY_PATH unset; trading runner will not publish opportunity candidates")
	}
	oppStore := dashboard.NewOpportunityStore(db)
	runnerMgr := dashboard.NewRunnerManager(log, settingsStore, crypto, policy, oppStore)
	h := dashboard.NewServer(dashboard.Options{
		Logger:         log,
		Auth:           auth,
		Sessions:       dashboard.NewSessionStore(db),
		SettingsCrypto: crypto,
		Settings:       settingsStore,
		RunnerManager:  runnerMgr,
		Opportunities:  oppStore,
		OpportunitySnapshot: func(u string) opportunities.Snapshot {
			return runnerMgr.Snapshot(u)
		},
	})
	go func() {
		rctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		runnerMgr.Reconcile(rctx)
	}()

	if p := strings.TrimSpace(authPath); p != "" {
		hup := make(chan os.Signal, 1)
		signal.Notify(hup, syscall.SIGHUP)
		go func() {
			for range hup {
				f, err := dashboard.LoadDashboardAuthFile(p)
				if err != nil {
					log.Warn("reload dashboard auth failed", "err", err)
					continue
				}
				h.ReloadAuth(f)
				log.Info("reloaded dashboard auth file", "path", p)
			}
		}()
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: h.Handler(),
	}

	go func() {
		log.Info("dashboard listening", "addr", addr, "session_db", sessPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("dashboard server", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Info("dashboard shutting down")
	return srv.Shutdown(shutdownCtx)
}

func runObserve(log *slog.Logger, args []string) error {
	fs := flag.NewFlagSet("observe", flag.ContinueOnError)
	dur := fs.Duration("duration", 30*time.Second, "how long to run before exit")
	interval := fs.Duration("interval", 3*time.Second, "poll interval")
	inst := fs.String("instrument", "", "instrument for get_order_book (e.g. BTC-PERPETUAL)")
	_ = fs.Parse(args)

	if strings.TrimSpace(*inst) == "" {
		return fmt.Errorf("-instrument is required")
	}
	base := strings.TrimSpace(os.Getenv("DERIBIT_BASE_URL"))
	if base == "" {
		base = deribit.TestnetRPCBaseURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), *dur)
	defer cancel()

	return observe.RunReadOnly(ctx, observe.Config{
		BaseURL:    base,
		ClientID:   os.Getenv("DERIBIT_CLIENT_ID"),
		Secret:     os.Getenv("DERIBIT_CLIENT_SECRET"),
		Instrument: *inst,
		Interval:   *interval,
	}, log)
}

func runSmokeOrder(log *slog.Logger, args []string) error {
	fs := flag.NewFlagSet("smoke-order", flag.ContinueOnError)
	inst := fs.String("instrument", "BTC-PERPETUAL", "instrument")
	amt := fs.Float64("amount", 10, "contracts / size (small; testnet only)")
	policy := fs.String("policy", os.Getenv("OPTITRADE_POLICY_PATH"), "policy JSON path")
	_ = fs.Parse(args)

	p := strings.TrimSpace(*policy)
	if p == "" {
		return fmt.Errorf("set -policy or OPTITRADE_POLICY_PATH")
	}
	base := strings.TrimSpace(os.Getenv("DERIBIT_BASE_URL"))
	if base == "" {
		base = deribit.TestnetRPCBaseURL
	}
	log.Info("smoke-order: placing deep post-only bid then cancel-all", "instrument", *inst)
	return observe.RunSmokeOrder(context.Background(), base, os.Getenv("DERIBIT_CLIENT_ID"), os.Getenv("DERIBIT_CLIENT_SECRET"), p, *inst, *amt)
}
