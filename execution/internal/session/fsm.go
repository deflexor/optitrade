// Package session implements the execution service session FSM (data-model.md State machine).
//
// NFR / SC-003 timing (FR-009):
// Spec target: worst-case ~60s from unhealthy feed to blocked submits includes detection
// and control-plane latency. This package applies transitions when Observe* runs; MVP does
// not guarantee full end-to-end 60s without wiring ObserveMarket/ObserveConnectivity into a
// single main loop tick rate and measuring in production. Tune policy thresholds
// (feed_stale_ms, ws_down_grace_ms) plus poll interval to meet SC-003.
//
// Transition table (authoritative; data-model.md session section):
//
//	From state           | Event / condition              | To state
//	---------------------|--------------------------------|------------------
//	running              | market quality (stale/gap/wide)| protective_flatten
//	running              | WS down >= ws_down_grace_ms   | protective_flatten
//	running              | RPC auth failure               | protective_flatten
//	running              | operator pause                 | paused
//	any                  | operator freeze                | frozen
//	protective_flatten   | flatten complete               | frozen
//	protective_flatten   | max_flatten_wait elapsed       | frozen
//	paused               | operator resume                | running
package session

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dfr/optitrade/execution/internal/config"
	"github.com/dfr/optitrade/execution/internal/market"
)

// State is a session FSM state (session_state in data-model.md).
type State string

const (
	StateRunning           State = "running"
	StatePaused            State = "paused"
	StateProtectiveFlatten State = "protective_flatten"
	StateFrozen            State = "frozen"
)

const (
	reasonMarketQuality  = "market_quality"
	reasonWSDown         = "ws_down"
	reasonRPCAuth        = "rpc_auth_failure"
	reasonFlattenDone    = "flatten_complete"
	reasonFlattenWait    = "flatten_deadline"
	reasonOperatorPause  = "operator_pause"
	reasonOperatorFrz    = "operator_freeze"
	reasonOperatorResume = "operator_resume"
)

// ErrSubmitBlocked is returned when the FSM disallows placement.
var ErrSubmitBlocked = errors.New("session: submit blocked")

// FSM is a thread-safe session state machine.
type FSM struct {
	mu sync.RWMutex

	state State
	now   func() time.Time

	// wsDownSince is set while the WS client reports disconnected (nil when connected).
	wsDownSince *time.Time
	// protectiveSince is set when entering protective_flatten.
	protectiveSince *time.Time

	lastReason string
}

// NewFSM returns an FSM in state running.
func NewFSM() *FSM {
	return &FSM{
		state: StateRunning,
		now:   time.Now,
	}
}

// NewFSMForTest returns an FSM with a clock stub (tests only).
func NewFSMForTest(now func() time.Time) *FSM {
	return &FSM{
		state: StateRunning,
		now:   now,
	}
}

// State returns the current session state.
func (f *FSM) State() State {
	if f == nil {
		return StateFrozen
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

// LastReason returns the reason string for the last transition (best-effort, for tests/logs).
func (f *FSM) LastReason() string {
	if f == nil {
		return ""
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.lastReason
}

// AllowSubmit enforces FR-009 placement policy: no new risk-increasing orders in
// protective_flatten; paused and frozen block all submits.
func (f *FSM) AllowSubmit(reduceOnly bool) error {
	if f == nil {
		return fmt.Errorf("%w: nil fsm", ErrSubmitBlocked)
	}
	f.mu.RLock()
	s := f.state
	f.mu.RUnlock()

	switch s {
	case StateRunning:
		return nil
	case StateProtectiveFlatten:
		if reduceOnly {
			return nil
		}
		return fmt.Errorf("%w: protective_flatten disallows non-reduce-only orders", ErrSubmitBlocked)
	case StatePaused, StateFrozen:
		return fmt.Errorf("%w: state=%s", ErrSubmitBlocked, s)
	default:
		return fmt.Errorf("%w: unknown state %q", ErrSubmitBlocked, s)
	}
}

// NotifyWS records connection health for ObserveConnectivity.
// up false means the socket is down (read error before reconnect, or Close).
func (f *FSM) NotifyWS(up bool) {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if !up {
		if f.wsDownSince == nil {
			t := f.now()
			f.wsDownSince = &t
		}
		return
	}
	f.wsDownSince = nil
}

// ObserveMarket triggers protective_flatten from running when quality flags are set
// (feed loss, book gap, wide spread). Flags are produced by market.BuildSnapshotFromBook
// using policy protective_mode thresholds.
func (f *FSM) ObserveMarket(flags market.QualityFlag, _ *config.ProtectiveMode) {
	if f == nil || flags == market.FlagNone {
		return
	}
	const bad = market.StaleBook | market.Gap | market.WideSpread
	if flags&bad == 0 {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.state != StateRunning {
		return
	}
	f.transitionLocked(StateProtectiveFlatten, reasonMarketQuality)
}

// ObserveConnectivity moves running -> protective_flatten when WS has been down for
// ws_down_grace_ms (0 or nil policy field means immediate on first observation while down).
func (f *FSM) ObserveConnectivity(pm *config.ProtectiveMode, now time.Time) {
	if f == nil {
		return
	}
	grace := wsDownGrace(pm)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.state != StateRunning || f.wsDownSince == nil {
		return
	}
	downFor := now.Sub(*f.wsDownSince)
	if downFor < grace {
		return
	}
	f.transitionLocked(StateProtectiveFlatten, reasonWSDown)
}

// NotifyRPCAuthFailure moves running -> protective_flatten (private REST / RPC auth errors).
func (f *FSM) NotifyRPCAuthFailure() {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.state != StateRunning {
		return
	}
	f.transitionLocked(StateProtectiveFlatten, reasonRPCAuth)
}

// ObserveFlattenDeadline escalates protective_flatten -> frozen after max_flatten_wait_ms
// to mitigate deadlock between flatten and reconciliation.
func (f *FSM) ObserveFlattenDeadline(pm *config.ProtectiveMode, now time.Time) {
	if f == nil {
		return
	}
	maxWait, ok := flattenMaxWait(pm)
	if !ok {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.state != StateProtectiveFlatten || f.protectiveSince == nil {
		return
	}
	if now.Sub(*f.protectiveSince) < maxWait {
		return
	}
	f.transitionLocked(StateFrozen, reasonFlattenWait)
}

// NotifyFlattenComplete moves protective_flatten -> frozen when exits are done (operator policy).
func (f *FSM) NotifyFlattenComplete() {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.state != StateProtectiveFlatten {
		return
	}
	f.transitionLocked(StateFrozen, reasonFlattenDone)
}

// OperatorPause moves any state except frozen to paused.
func (f *FSM) OperatorPause() {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.state == StateFrozen {
		return
	}
	f.transitionLocked(StatePaused, reasonOperatorPause)
}

// OperatorResume moves paused -> running.
func (f *FSM) OperatorResume() {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.state != StatePaused {
		return
	}
	f.state = StateRunning
	f.lastReason = reasonOperatorResume
	f.protectiveSince = nil
}

// OperatorFreeze moves to frozen from any state.
func (f *FSM) OperatorFreeze() {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.state == StateFrozen {
		return
	}
	f.transitionLocked(StateFrozen, reasonOperatorFrz)
}

func wsDownGrace(pm *config.ProtectiveMode) time.Duration {
	if pm == nil || pm.WSDownGraceMs == nil {
		return 0
	}
	if *pm.WSDownGraceMs <= 0 {
		return 0
	}
	return time.Duration(*pm.WSDownGraceMs) * time.Millisecond
}

func flattenMaxWait(pm *config.ProtectiveMode) (time.Duration, bool) {
	if pm == nil || pm.MaxFlattenWaitMs == nil || *pm.MaxFlattenWaitMs <= 0 {
		return 0, false
	}
	return time.Duration(*pm.MaxFlattenWaitMs) * time.Millisecond, true
}

func (f *FSM) transitionLocked(to State, reason string) {
	from := f.state
	if from == to {
		return
	}
	if !allowedTransition(from, to, reason) {
		return
	}
	f.state = to
	f.lastReason = reason
	if to == StateProtectiveFlatten {
		t := f.now()
		f.protectiveSince = &t
	}
	if to != StateProtectiveFlatten {
		f.protectiveSince = nil
	}
	if to == StatePaused {
		f.wsDownSince = nil
	}
}

func allowedTransition(from, to State, reason string) bool {
	if reason == reasonOperatorFrz && to == StateFrozen {
		return from != StateFrozen
	}
	if reason == reasonOperatorPause && to == StatePaused {
		return from != StateFrozen
	}
	switch from {
	case StateRunning:
		switch to {
		case StateProtectiveFlatten:
			return reason == reasonMarketQuality || reason == reasonWSDown || reason == reasonRPCAuth
		case StatePaused:
			return reason == reasonOperatorPause
		}
	case StateProtectiveFlatten:
		switch to {
		case StateFrozen:
			return reason == reasonFlattenDone || reason == reasonFlattenWait
		case StatePaused:
			return reason == reasonOperatorPause
		}
	case StatePaused, StateFrozen:
		return false
	}
	return false
}
