package dashboard

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/dfr/optitrade/src/internal/config"
	"github.com/dfr/optitrade/src/internal/opportunities"
)

// RunnerManager starts and stops per-user trading runner goroutines (spec 2026-04-04).
type RunnerManager struct {
	log      *slog.Logger
	settings *OperatorSettingsStore
	crypto   *SettingsCrypto
	policy   *config.Policy

	// oppStore persists opening/active rows; when nil, auto-open is skipped.
	oppStore *OpportunityStore
	// AutoOpenHook, when non-nil, replaces OKX batch place + DB upsert in auto mode (tests only).
	AutoOpenHook func(ctx context.Context, user, id string, cand opportunities.Row) error

	mu          sync.Mutex
	runners     map[string]context.CancelFunc
	snapMu      sync.RWMutex
	snapshots   map[string]opportunities.Snapshot
}

// NewRunnerManager builds a manager; log/settings/crypto must be non-nil for Reconcile.
// policy may be nil; the runner will not populate opportunity snapshots until a policy is loaded.
// opp may be nil; auto-open and DB-backed flows are skipped when nil.
func NewRunnerManager(log *slog.Logger, settings *OperatorSettingsStore, crypto *SettingsCrypto, policy *config.Policy, opp *OpportunityStore) *RunnerManager {
	if log == nil {
		log = slog.Default()
	}
	return &RunnerManager{
		log:           log,
		settings:      settings,
		crypto:        crypto,
		policy:        policy,
		oppStore:      opp,
		runners:       map[string]context.CancelFunc{},
		snapshots:     map[string]opportunities.Snapshot{},
	}
}

// IsRunning reports whether a runner goroutine is active for username.
func (rm *RunnerManager) IsRunning(username string) bool {
	if rm == nil {
		return false
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()
	_, ok := rm.runners[strings.TrimSpace(username)]
	return ok
}

// RunnerEligible is true when the operator may run the opportunities loop (M1: OKX only).
func RunnerEligible(row *OperatorSettingsRow) bool {
	if row == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(row.AccountStatus)) {
	case "", "active":
	default:
		return false
	}
	switch strings.ToLower(strings.TrimSpace(row.BotMode)) {
	case "manual", "auto":
	default:
		return false
	}
	if row.Provider != ProviderOKX {
		return false
	}
	return validateOperatorSettings(row) == nil
}

// Reconcile starts eligible runners and stops ineligible or unknown users.
func (rm *RunnerManager) Reconcile(ctx context.Context) {
	if rm == nil || rm.settings == nil || rm.crypto == nil {
		return
	}
	users, err := rm.settings.ListUsernames(ctx)
	if err != nil {
		rm.log.Error("runner reconcile list usernames", "err", err)
		return
	}
	want := make(map[string]struct{})
	for _, u := range users {
		row, err := rm.settings.GetDecrypting(ctx, u, rm.crypto)
		if err != nil {
			rm.log.Warn("runner reconcile get settings", "user", u, "err", err)
			continue
		}
		if RunnerEligible(row) {
			want[u] = struct{}{}
			rm.startUser(u)
		}
	}
	rm.mu.Lock()
	var stop []string
	for u := range rm.runners {
		if _, ok := want[u]; !ok {
			stop = append(stop, u)
		}
	}
	rm.mu.Unlock()
	for _, u := range stop {
		rm.stopUser(u)
	}
}

func (rm *RunnerManager) startUser(username string) {
	u := strings.TrimSpace(username)
	if u == "" {
		return
	}
	rm.mu.Lock()
	if _, exists := rm.runners[u]; exists {
		rm.mu.Unlock()
		return
	}
	runCtx, cancel := context.WithCancel(context.Background())
	rm.runners[u] = cancel
	rm.mu.Unlock()

	go rm.runLoop(runCtx, u)
}

func (rm *RunnerManager) stopUser(username string) {
	u := strings.TrimSpace(username)
	rm.mu.Lock()
	cancel, ok := rm.runners[u]
	if ok {
		delete(rm.runners, u)
	}
	rm.mu.Unlock()
	rm.clearSnapshot(u)
	if cancel != nil {
		cancel()
	}
}

func (rm *RunnerManager) runLoop(ctx context.Context, user string) {
	defer func() {
		if r := recover(); r != nil {
			rm.log.Error("runner panic", "user", user, "recover", r)
		}
	}()
	rm.log.Info("runner started", "user", user)
	rm.runTick(ctx, user)
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			rm.log.Info("runner stopped", "user", user)
			return
		case <-tick.C:
			rm.runTick(ctx, user)
		}
	}
}
