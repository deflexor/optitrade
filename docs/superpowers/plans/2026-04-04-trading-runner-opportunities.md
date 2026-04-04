# Trading runner, Opportunities UI, and OKX selector — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship Milestone 1 end-to-end: per-operator goroutine runner inside the dashboard, OKX-based put-credit-spread discovery for BTC, ranked opportunities with real gates (`cost.ScoreCandidate`, `risk.Engine.Check`, max-loss-% of equity), BFF APIs + optional SSE, `/opportunities` UI with paused banner, header bot mode, settings risk field, and execution hooks (open/cancel/close) for the opportunity state machine; remove legacy `/positions` routes and refresh Playwright.

**Architecture:** Extend `dashboard_operator_settings` for `account_status`, `bot_mode`, and `max_loss_equity_pct`. Add `RunnerManager` (map username → cancel + worker) started from `runDashboardCmdFull` after `NewServer`, with reconcile on settings PUT. New package `internal/opportunities` orchestrates universe fetch (OKX public REST), parameterized builders in `internal/strategy` (OKX `instId` naming), ranking, and snapshot merge with persisted rows in `dashboard_opportunity`. Map OKX books to `deribit.OrderBook` so existing `strategy.TouchFromBook` and `cost.ScoreCandidate` stay unchanged for M1. Load trading policy via `OPTITRADE_POLICY_PATH` (`config.LoadFile`). Default regime label `normal` until a live regime feed is wired (`OPTITRADE_DASHBOARD_REGIME_LABEL` override).

**Tech stack:** Go 1.26 (`src/go.mod`), SQLite + `state.ApplyMigrations`, existing `internal/okx` private client + new unsigned public helper, `internal/config`, `internal/regime`, `internal/cost`, `internal/risk`, `internal/market`, React 19 + Vite 8 (`web/`), Playwright e2e.

**Milestone 1 status:** Tasks **1–20** below are marked **[x]** complete on `master` (2026-04). **Milestone 2–3** checkboxes at the end of this file remain open.

---

## File structure (before tasks)

| Path | Responsibility |
|------|----------------|
| `src/internal/state/migrations/0006_dashboard_trading_prefs.sql` | `ALTER TABLE dashboard_operator_settings` add `account_status`, `bot_mode`, `max_loss_equity_pct`. |
| `src/internal/state/migrations/0007_dashboard_opportunities.sql` | `dashboard_opportunity` table for server-assigned ids and lifecycle JSON. |
| `src/internal/strategy/okx_inst.go` | OKX option `instId` formatting + `VerticalPutCreditOKX(uly, expiryYYYYMMDD, shortStrike, width int64)`. |
| `src/internal/strategy/okx_inst_test.go` | Table tests: known inputs → expected `instId` strings. |
| `src/internal/strategy/structure_templates.go` | Rename live misuse: keep `BuildLegsForStructure` only for certification; doc + `//go:build cert` or move demo legs to `_test.go` / `Example` — **do not** call from selector. |
| `src/internal/okx/public.go` | `PublicClient` (no signing): `GetInstruments`, `GetOrderBook` parsing to `deribit.OrderBook`. |
| `src/internal/okx/public_test.go` | `httptest` golden JSON → parsed books. |
| `src/internal/opportunities/types.go` | `Snapshot`, `Row`, `LegQuote`, enums `RowStatus`, JSON tags for API. |
| `src/internal/opportunities/selector.go` | `Selector` struct, `RunTick(ctx, in SelectorInput) (Snapshot, error)` — M1: only `credit_spread` + BTC + nearest expiries. |
| `src/internal/opportunities/selector_test.go` | Stub books + policy fixture → ranked rows, gate vetoes. |
| `src/internal/audit/nop.go` | `NopDecisionLogger` implementing `DecisionLogger` for runner when no audit repo. |
| `src/internal/dashboard/operator_settings_store.go` | Scan/merge new columns; `OperatorSettingsPatch` fields. |
| `src/internal/dashboard/settings_handlers.go` + `settings_crypto_test.go` | Expose `max_loss_equity_pct` in catalog + PUT validation **1–50** integer. |
| `src/internal/dashboard/trading_handlers.go` | `GET /trading/status`, `PUT /trading/mode`. |
| `src/internal/dashboard/opportunities_handlers.go` | `GET /opportunities`, `GET /opportunities/stream`, POST open/cancel/close. |
| `src/internal/dashboard/runner_manager.go` | Start/stop/reconcile, eligibility, `recover` on loop. |
| `src/internal/dashboard/runner_tick.go` | Per-user tick: equity, selector, merge DB rows, auto-open branch. |
| `src/internal/dashboard/opportunity_store.go` | CRUD for `dashboard_opportunity`. |
| `src/internal/dashboard/server.go` | Register routes; hold `*RunnerManager` + policy pointer. |
| `src/cmd/optitrade/main.go` | Load policy, construct `RunnerManager`, `srv.SetRunners(...)`, `Start` after Listen, `Shutdown` cancel. |
| `src/internal/state/sqlite/store_test.go` | Expect **7** migration rows after 0006–0007 land. |
| `web/src/pages/OpportunitiesPage.tsx` | Table + paused banner + disabled messaging. |
| `web/src/App.tsx` | Routes: `/opportunities`; remove `/positions*`. |
| `web/src/pages/Overview.tsx` | Link text → Opportunities. |
| `web/e2e/dashboard-conformance.spec.ts` | Replace positions navigation with opportunities. |
| `web/e2e/opportunities.spec.ts` | Bot mode control + paused empty state. |

---

### Task 1: Migration 0006 — trading prefs columns

**Files:**
- Create: `src/internal/state/migrations/0006_dashboard_trading_prefs.sql`
- Modify: `src/internal/state/sqlite/store_test.go` (migration count `5` → `7` only after **both** 0006 and 0007 exist; if you land 0006 alone first, use `6` temporarily)
- Test: `src/internal/state/sqlite/store_test.go`

- [x] **Step 1: Write migration SQL**

Create `src/internal/state/migrations/0006_dashboard_trading_prefs.sql`:

```sql
-- Trading prefs: admin account_status, operator bot_mode, max loss % of equity (spec 2026-04-04).
ALTER TABLE dashboard_operator_settings ADD COLUMN account_status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE dashboard_operator_settings ADD COLUMN bot_mode TEXT NOT NULL DEFAULT 'manual';
ALTER TABLE dashboard_operator_settings ADD COLUMN max_loss_equity_pct INTEGER NOT NULL DEFAULT 10;
```

- [x] **Step 2: Run migration test**

Run: `cd /home/dfr/optitrade/src && go test ./internal/state/sqlite/... -run TestOpenAppliesMigrationsEmptyDB -v`

Expected: **FAIL** if expected count still `5` (got 6). After updating expected count to `6`, **PASS**.

- [x] **Step 3: Update expected migration count (after 0006 only)**

In `store_test.go`, change:

```go
if n != 5 {
	t.Fatalf("migrations recorded: got %d want 5", n)
}
```

to `if n != 6` until 0007 is added; after Task 2, use `7`.

- [x] **Step 4: Run full sqlite package tests**

Run: `cd /home/dfr/optitrade/src && go test ./internal/state/sqlite/... -count=1`

Expected: PASS

- [x] **Step 5: Commit**

```bash
git add src/internal/state/migrations/0006_dashboard_trading_prefs.sql src/internal/state/sqlite/store_test.go
git commit -m "feat(dashboard): add trading prefs columns migration"
```

---

### Task 2: Migration 0007 — opportunity rows

**Files:**
- Create: `src/internal/state/migrations/0007_dashboard_opportunities.sql`
- Modify: `src/internal/state/sqlite/store_test.go` (`n != 7`)

- [x] **Step 1: Add migration file**

`src/internal/state/migrations/0007_dashboard_opportunities.sql`:

```sql
CREATE TABLE IF NOT EXISTS dashboard_opportunity (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  status TEXT NOT NULL,
  strategy_name TEXT NOT NULL,
  legs_json TEXT NOT NULL,
  meta_json TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_dashboard_opportunity_user ON dashboard_opportunity(username);
CREATE INDEX IF NOT EXISTS idx_dashboard_opportunity_updated ON dashboard_opportunity(updated_at);
```

- [x] **Step 2: Set migration count to 7**

In `store_test.go`: `if n != 7 { ... }`

- [x] **Step 3: Test**

Run: `cd /home/dfr/optitrade/src && go test ./internal/state/sqlite/... -count=1`

Expected: PASS

- [x] **Step 4: Commit**

```bash
git add src/internal/state/migrations/0007_dashboard_opportunities.sql src/internal/state/sqlite/store_test.go
git commit -m "feat(dashboard): add dashboard_opportunity table"
```

---

### Task 3: OKX option `instId` helpers

**Files:**
- Create: `src/internal/strategy/okx_inst.go`
- Create: `src/internal/strategy/okx_inst_test.go`

- [x] **Step 1: Failing test**

`src/internal/strategy/okx_inst_test.go`:

```go
package strategy

import "testing"

func TestOKXOptionInstID_put(t *testing.T) {
	got := OKXOptionInstID("BTC", "20260327", 95000, false)
	want := "BTC-USD-260327-95000-P"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestVerticalPutCreditOKX(t *testing.T) {
	legs, err := VerticalPutCreditOKX("BTC", "20260327", 95000, 500)
	if err != nil {
		t.Fatal(err)
	}
	if len(legs) != 2 {
		t.Fatalf("legs: %d", len(legs))
	}
	if legs[0].Side != LegSell || legs[1].Side != LegBuy {
		t.Fatalf("sides: %+v", legs)
	}
	if legs[0].Instrument != "BTC-USD-260327-95000-P" {
		t.Fatalf("short %q", legs[0].Instrument)
	}
	if legs[1].Instrument != "BTC-USD-260327-90000-P" {
		t.Fatalf("long %q", legs[1].Instrument)
	}
}
```

Run: `cd /home/dfr/optitrade/src && go test ./internal/strategy/... -run TestOKXOptionInstID -v`

Expected: FAIL undefined `OKXOptionInstID`

- [x] **Step 2: Implementation**

`src/internal/strategy/okx_inst.go`:

```go
package strategy

import (
	"fmt"
	"strings"
)

// OKXOptionInstID builds OKX v5 OPTION instId: {BASE}-USD-{YYMMDD}-{strike}-{C|P}.
// expiryYYYYMMDD must be 8 digits (UTC listing date).
func OKXOptionInstID(base, expiryYYYYMMDD string, strike int64, call bool) string {
	cp := "P"
	if call {
		cp = "C"
	}
	b := strings.ToUpper(strings.TrimSpace(base))
	e := strings.TrimSpace(expiryYYYYMMDD)
	return fmt.Sprintf("%s-USD-%s-%d-%s", b, e, strike, cp)
}

// VerticalPutCreditOKX is the OKX-named variant of VerticalPutCredit.
func VerticalPutCreditOKX(base, expiryYYYYMMDD string, shortStrike int64, width int) ([]LegSpec, error) {
	if width <= 0 {
		width = DefaultStrikeWidth
	}
	if shortStrike <= int64(width) {
		return nil, fmt.Errorf("put credit: short strike too low for width %d", width)
	}
	longStrike := shortStrike - int64(width)
	legs := []LegSpec{
		{Instrument: OKXOptionInstID(base, expiryYYYYMMDD, shortStrike, false), Side: LegSell},
		{Instrument: OKXOptionInstID(base, expiryYYYYMMDD, longStrike, false), Side: LegBuy},
	}
	mustAssertDefinedRiskInDev(legs)
	return legs, nil
}
```

- [x] **Step 3: Verify tests pass**

Run: `cd /home/dfr/optitrade/src && go test ./internal/strategy/... -run 'TestOKXOptionInstID|TestVerticalPutCreditOKX' -v`

Expected: PASS

- [x] **Step 4: Commit**

```bash
git add src/internal/strategy/okx_inst.go src/internal/strategy/okx_inst_test.go
git commit -m "feat(strategy): OKX option instId and put credit spread builder"
```

---

### Task 4: Quarantine demo `BuildLegsForStructure` from production path

**Files:**
- Modify: `src/internal/strategy/structure_templates.go`
- Modify: `src/internal/strategy/sc001_playbook_certification_test.go` (if build tag needed)

- [x] **Step 1: Add file header comment** in `structure_templates.go` stating: **Certification / examples only — selector must use venue-specific builders (e.g. VerticalPutCreditOKX).**

- [x] **Step 2: Ensure no production import** — `grep -R "BuildLegsForStructure" src/internal` — only `strategy` tests and `opportunities` (none until you add it). **Do not** reference from `opportunities` package.

- [x] **Step 3: Commit**

```bash
git add src/internal/strategy/structure_templates.go
git commit -m "docs(strategy): mark BuildLegsForStructure as cert-only path"
```

---

### Task 5: OKX public REST client + order book mapping

**Files:**
- Create: `src/internal/okx/public.go`
- Create: `src/internal/okx/public_test.go`

- [x] **Step 1: Failing test** — `public_test.go` spins `httptest.Server` returning minimal JSON for `/api/v5/market/books` with one bid/ask; assert `deribit.OrderBook` fields.

Example test body:

```go
package okx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dfr/optitrade/src/internal/deribit"
)

func TestPublicClient_OrderBookToDeribitShape(t *testing.T) {
	const raw = `{"code":"0","data":[{"asks":[["0.05","1","0","1"]],"bids":[["0.04","2","0","2"]]}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(raw))
	}))
	defer srv.Close()

	pc := &PublicClient{BaseURL: srv.URL, HTTP: srv.Client()}
	book, err := pc.GetOrderBookDeribit(context.Background(), "BTC-USD-260327-95000-P", 5)
	if err != nil {
		t.Fatal(err)
	}
	if book.InstrumentName != "BTC-USD-260327-95000-P" {
		t.Fatalf("inst %q", book.InstrumentName)
	}
	if len(book.Bids) != 1 || len(book.Asks) != 1 {
		t.Fatalf("depth %+v", book)
	}
	if book.Bids[0].Price != 0.04 || book.Asks[0].Price != 0.05 {
		t.Fatalf("prices %+v %+v", book.Bids[0], book.Asks[0])
	}
}
```

Run test before implementation: **FAIL** undefined `PublicClient`.

- [x] **Step 2: Implement** `PublicClient` with `GetOrderBookDeribit(ctx, instId string, sz int) (deribit.OrderBook, error)` parsing OKX `data[0].bids` / `asks` as `[][4]string` price first element.

- [x] **Step 3: Add** `GetInstruments(ctx, instType, uly string) ([]InstrumentSummary, error)` returning `InstId`, `ExpTime` / `ListTime` fields you need for expiry filter (struct fields per OKX JSON).

- [x] **Step 4: Pass tests**

Run: `cd /home/dfr/optitrade/src && go test ./internal/okx/... -count=1`

- [x] **Step 5: Commit**

```bash
git add src/internal/okx/public.go src/internal/okx/public_test.go
git commit -m "feat(okx): public market data and deribit-shaped order books"
```

---

### Task 6: Nop audit logger for risk engine in dashboard

**Files:**
- Create: `src/internal/audit/nop.go`
- Create: `src/internal/audit/nop_test.go`

- [x] **Step 1: Test**

```go
package audit

import (
	"context"
	"testing"
)

func TestNopDecisionLogger(t *testing.T) {
	var l DecisionLogger = NopDecisionLogger{}
	if err := l.LogDecision(context.Background(), DecisionRecord{CorrelationID: "x"}); err != nil {
		t.Fatal(err)
	}
}
```

Run: `go test ./internal/audit/... -run Nop -v` → FAIL undefined.

- [x] **Step 2: Implement**

```go
package audit

import "context"

// NopDecisionLogger satisfies DecisionLogger without persisting (dashboard runner until audit DB wired).
type NopDecisionLogger struct{}

func (NopDecisionLogger) LogDecision(_ context.Context, _ DecisionRecord) error { return nil }
```

- [x] **Step 3: Commit**

```bash
git add src/internal/audit/nop.go src/internal/audit/nop_test.go
git commit -m "feat(audit): nop decision logger for BFF runner"
```

---

### Task 7: Opportunities selector (M1) with stubbed books

**Files:**
- Create: `src/internal/opportunities/types.go`
- Create: `src/internal/opportunities/selector.go`
- Create: `src/internal/opportunities/selector_test.go`
- Use: `config/examples/policy.example.json` or copy minimal JSON into testdata

- [x] **Step 1: types.go** — define:

```go
type RowStatus string

const (
	StatusCandidate RowStatus = "candidate"
	StatusOpening   RowStatus = "opening"
	StatusActive    RowStatus = "active"
	StatusPartial   RowStatus = "partial"
)

type LegQuote struct {
	Instrument string  `json:"instrument"`
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
}

type Row struct {
	ID             string     `json:"id"`
	StrategyName   string     `json:"strategy_name"`
	Status         RowStatus  `json:"status"`
	Legs           []LegQuote `json:"legs"`
	GreeksNote     string     `json:"greeks_note,omitempty"`
	MaxProfit      string     `json:"max_profit"`
	MaxLoss        string     `json:"max_loss"`
	Recommend      string     `json:"recommendation"`
	Rationale      string     `json:"rationale"`
	ExpectedEdge   string     `json:"expected_edge"`
	SortKey        float64    `json:"-"`
}

type Snapshot struct {
	UpdatedAtMs int64 `json:"updated_at_ms"`
	Rows        []Row `json:"rows"`
}
```

- [x] **Step 2: Failing test** — build `Selector` with injected `BookFetcher` interface:

```go
type BookFetcher interface {
	FetchBook(ctx context.Context, inst string) (deribit.OrderBook, error)
}
```

Test: two credit spreads, wide spread veto vs tight spread pass `cost.ScoreCandidate`; use real `config.LoadBytes` minimal policy JSON from existing tests in `internal/cost/score_test.go` as template.

- [x] **Step 3: Implement** `selector.go`:

  - Input: `underlying BTC`, list of `(expiry, shortStrike)` candidates from caller (selector does **not** call OKX in this task — caller passes expansions).
  - For each expansion `VerticalPutCreditOKX`, fetch books, compute **v1 edge** as net credit at mid: `shortMid - longMid` (positive for credit spread); `MaxLoss` = width − credit (in underlying coin units — document assumption), `ExpectedEdge` = `fmt.Sprintf("%f", edge)`.
  - Call `cost.ScoreCandidate(policy, regimeLabel, candInput, books, nil, cost.IVSanityOptions{})`.
  - Call `risk.Engine.Check` with `CandidateRisk{Instruments: [...], MaxLossQuote: ...}` and empty positions — use `risk.NewEngine(policy, audit.NopDecisionLogger{})`; for veto tests use a policy with `limits.max_portfolio_delta` set to `"0"` so `Check` vetoes.
  - Rank by `bd.EdgeAfterCosts` descending using the existing third return value from `cost.ScoreCandidate` (`var bd cost.CostBreakdown` — use `ok, veto, bd := cost.ScoreCandidate(...)`).

- [x] **Step 4: Tests pass**

Run: `cd /home/dfr/optitrade/src && go test ./internal/opportunities/... -count=1`

- [x] **Step 5: Commit**

```bash
git add src/internal/opportunities/
git commit -m "feat(opportunities): M1 selector with cost and risk gates"
```

---

### Task 8: Operator settings store — read/write new fields

**Files:**
- Modify: `src/internal/dashboard/operator_settings_store.go`
- Modify: `src/internal/dashboard/settings_crypto_test.go` or new `operator_settings_store_test.go`

- [x] **Step 1: Extend structs**

```go
type OperatorSettingsRow struct {
	// ... existing ...
	AccountStatus     string // active|disabled
	BotMode           string // manual|auto|paused
	MaxLossEquityPct  int    // 1-50
}

type OperatorSettingsPatch struct {
	// ... existing ...
	AccountStatus    *string
	BotMode          *string
	MaxLossEquityPct *int
}
```

- [x] **Step 2: SQL** — `SELECT` / `INSERT` / `UPDATE` in `GetDecrypting` and `Put` to include three columns; validate `account_status` in `{"active","disabled"}`, `bot_mode` in `{"manual","auto","paused"}`, `max_loss_equity_pct` 1–50.

- [x] **Step 3: Test** round-trip via in-memory sqlite opened with `sqlite.Open` in temp dir.

- [x] **Step 4: Commit**

```bash
git add src/internal/dashboard/operator_settings_store.go src/internal/dashboard/*_test.go
git commit -m "feat(dashboard): persist trading prefs on operator settings"
```

---

### Task 9: Settings API — max loss % field

**Files:**
- Modify: `src/internal/dashboard/settings_handlers.go` (`settingsFieldCatalog`, `writeSettingsJSON`, PUT decoder)
- Modify: `web/src/pages/SettingsPage.tsx`

- [x] **Step 1: Go** — add field `max_loss_equity_pct` kind `number`, help text referencing equity cap for opportunities.

- [x] **Step 2: Handler test** — extend existing settings test: PUT `{"max_loss_equity_pct": 25}` → GET shows 25.

- [x] **Step 3: React** — controlled number input 1–50, load/save via existing settings API client pattern.

- [x] **Step 4: Run** `cd /home/dfr/optitrade/web && npm test && npm run lint` and `go test ./internal/dashboard/...`

- [x] **Step 5: Commit**

```bash
git add src/internal/dashboard/settings_handlers.go web/src/pages/SettingsPage.tsx
git commit -m "feat(settings): max loss percent of equity"
```

---

### Task 10: Trading status + mode API

**Files:**
- Create: `src/internal/dashboard/trading_handlers.go`
- Create: `src/internal/dashboard/trading_handlers_test.go`
- Modify: `src/internal/dashboard/server.go`

- [x] **Step 1: Implement** `handleTradingStatus`:

```json
{"account_status":"active","bot_mode":"manual","runner_running":true}
```

`runner_running` requires `RunnerManager` reference — for the test, pass `nil` manager returning `false` until Task 11.

- [x] **Step 2: Implement** `PUT /trading/mode` body `{"bot_mode":"paused"}` → updates settings, invalidates exchange cache, returns 200.

- [x] **Step 3: Register** in `server.go`:

```go
protected.Handle("GET /trading/status", http.HandlerFunc(s.handleTradingStatus))
protected.Handle("PUT /trading/mode", http.HandlerFunc(s.handleTradingModePut))
```

- [x] **Step 4: Handler tests** with `httptest` + mock settings store or real sqlite.

- [x] **Step 5: Commit**

```bash
git add src/internal/dashboard/trading_handlers.go src/internal/dashboard/trading_handlers_test.go src/internal/dashboard/server.go
git commit -m "feat(dashboard): trading status and bot mode endpoints"
```

---

### Task 11: Opportunity store + HTTP handlers (snapshot)

**Files:**
- Create: `src/internal/dashboard/opportunity_store.go`
- Create: `src/internal/dashboard/opportunities_handlers.go`
- Create: `src/internal/dashboard/opportunities_handlers_test.go`
- Modify: `src/internal/dashboard/server.go`

- [x] **Step 1: OpportunityStore** methods: `ListByUser`, `Upsert`, `Get`, `UpdateStatus`.

- [x] **Step 2: GET /opportunities** returns merge of runner in-memory snapshot (from `RunnerManager.Snapshot(user)`) + DB rows for non-candidate statuses. If `bot_mode == paused` or `account_status != active`, return `{"paused":true,"rows":[]}` or `{"disabled":true,...}` per spec.

- [x] **Step 3: Tests** with fake runner snapshot + sqlite.

- [x] **Step 4: Commit**

```bash
git add src/internal/dashboard/opportunity_store.go src/internal/dashboard/opportunities_handlers.go src/internal/dashboard/opportunities_handlers_test.go src/internal/dashboard/server.go
git commit -m "feat(dashboard): opportunities snapshot API"
```

---

### Task 12: RunnerManager + reconcile

**Files:**
- Create: `src/internal/dashboard/runner_manager.go`
- Create: `src/internal/dashboard/runner_manager_test.go`
- Modify: `src/internal/dashboard/server.go` (optional setter `AttachRunnerManager`)
- Modify: `src/internal/dashboard/settings_handlers.go` (after successful PUT, call `rm.Reconcile(ctx)`)

- [x] **Step 1: Types**

```go
type RunnerManager struct {
	log     *slog.Logger
	mu      sync.Mutex
	runners map[string]runnerHandle
	// deps: settings, crypto, policy, db, public OKX factory
}

type runnerHandle struct {
	cancel context.CancelFunc
}
```

- [x] **Step 2: Eligibility** — start goroutine when: `account_status==active`, `bot_mode` in `{manual,auto}`, provider OKX, `validateOperatorSettings` ok, keys present.

- [x] **Step 3: Reconcile** — on settings change, stop ineligible users; start newly eligible; restart if keys changed (`UpdatedAtMs`).

- [x] **Step 4: Test** — table: rows A/B modes → expected start/stop counts using stub deps.

- [x] **Step 5: Commit**

```bash
git add src/internal/dashboard/runner_manager.go src/internal/dashboard/runner_manager_test.go src/internal/dashboard/settings_handlers.go src/internal/dashboard/server.go
git commit -m "feat(dashboard): RunnerManager eligibility and reconcile"
```

---

### Task 13: Runner tick — OKX universe + selector

**Files:**
- Create: `src/internal/dashboard/runner_tick.go`
- Modify: `src/internal/dashboard/runner_manager.go`

- [x] **Step 1: In runner loop** (separate func `runUserRunner`):

  - `defer recover()` → log stack, `time.Sleep(2s)`, continue.
  - `ctx` from cancel.
  - Load equity via existing `okxExchange.GetAccountSummaries` for BTC.
  - Public client: list BTC-USD options, filter expiries ≤ `now+30d`, strikes within ±`W` of spot index (use OKX index ticker public endpoint), build short strikes grid from policy width (`strategy.DefaultStrikeWidth` or policy if exposed).
  - Call `opportunities.Selector.RunTick`.
  - Cap **top N = 20** rows.
  - Store snapshot under username mutex.

- [x] **Step 2: Max loss vs equity** — skip `recommend=open` when `max_loss > equity * (pct/100)` (use float64 carefully, compare in quote currency consistently — document USD vs coin; OKX `EqUsd` path already used in overview).

- [x] **Step 3: Integration test** with `httptest` OKX public mocks (or skip network with injected `InstrumentSource` interface in manager constructor).

- [x] **Step 4: Commit**

```bash
git add src/internal/dashboard/runner_tick.go src/internal/dashboard/runner_manager.go
git commit -m "feat(dashboard): runner tick with OKX universe and selector"
```

---

### Task 14: SSE `/opportunities/stream`

**Files:**
- Modify: `src/internal/dashboard/opportunities_handlers.go`
- Modify: `src/internal/dashboard/opportunities_handlers_test.go`

- [x] **Step 1: Handler** sets `Content-Type: text/event-stream`, `Cache-Control: no-cache`, flusher, sends `data: {json}\n\n` every 1s while ctx alive (same payload as GET).

- [x] **Step 2: Test** — start handler in goroutine, read first chunk contains `"rows"`.

- [x] **Step 3: Commit**

```bash
git add src/internal/dashboard/opportunities_handlers.go src/internal/dashboard/opportunities_handlers_test.go src/internal/dashboard/server.go
git commit -m "feat(dashboard): opportunities SSE stream"
```

---

### Task 15: Execution — open / cancel / close

**Files:**
- Modify: `src/internal/dashboard/opportunities_handlers.go`
- Modify: `src/internal/okx/private.go` (or new `trade.go`) — `BatchPlaceOrders`
- Modify: `src/internal/dashboard/okx_exchange.go` if writer surface needed

- [x] **Step 1: State machine** — `POST /opportunities/{id}/open`: only from `candidate`; persist `opening`. `cancel`: `opening` → `candidate` or deleted. `close`: `active|partial` → enqueue close flow (reuse reduce-only placement pattern from `close.go` but keyed by opportunity legs JSON).

- [x] **Step 2: OKX batch** — implement `POST /api/v5/trade/batch-orders` on signed client; map legs to orders (`tdMode`, `posSide` per OKX options docs).

- [x] **Step 3: Handler tests** with fake exchange implementing new interface.

- [x] **Step 4: Commit**

```bash
git add src/internal/dashboard/opportunities_handlers.go src/internal/okx/
git commit -m "feat(dashboard): opportunity open cancel close with OKX batch orders"
```

---

### Task 16: Auto mode auto-open

**Files:**
- Modify: `src/internal/dashboard/runner_tick.go`

- [x] **Step 1: When** `bot_mode==auto`, for top row where `recommendation==open` and no existing DB row for same `candidate_key`, call shared `OpenOpportunity(ctx, user, rowID)` (same as HTTP handler).

- [x] **Step 2: Test** — inject fake opener counter; assert single increment per tick.

- [x] **Step 3: Commit**

```bash
git add src/internal/dashboard/runner_tick.go
git commit -m "feat(dashboard): auto mode opens top gated opportunity"
```

---

### Task 17: Wire main — policy load + runner lifecycle

**Files:**
- Modify: `src/cmd/optitrade/main.go`
- Modify: `src/internal/dashboard/server.go`

- [x] **Step 1: Load policy**

```go
var pol *config.Policy
if p := config.PolicyPathFromEnv(); p != "" {
	var err error
	pol, err = config.LoadFile(p)
	if err != nil {
		return fmt.Errorf("dashboard policy: %w", err)
	}
}
```

If unset, log **Warn** and run runner with **nil** policy causing selector to skip scoring (or fail closed — pick one and test).

- [x] **Step 2: Construct** `RunnerManager` with db, settings, crypto, policy, logger; `server.AttachRunnerManager(rm)`; after `Listen`, `go rm.Start(context.Background())`; on shutdown `rm.Stop()`.

- [x] **Step 3: Manual run** — `optitrade dashboard -listen :8080` with `OPTITRADE_POLICY_PATH` set; verify no panic.

- [x] **Step 4: Commit**

```bash
git add src/cmd/optitrade/main.go src/internal/dashboard/server.go
git commit -m "feat(dashboard): load policy and start RunnerManager"
```

---

### Task 18: Frontend — Opportunities page + header mode

**Files:**
- Modify: `web/src/App.tsx`
- Create: `web/src/pages/OpportunitiesPage.tsx`
- Modify: `web/src/api/client.ts` (methods `getTradingStatus`, `putTradingMode`, `getOpportunities`, subscribe SSE or poll)
- Modify: `web/src/pages/Overview.tsx`

- [x] **Step 1: Header** — segmented control `manual | auto | paused` calling `PUT /api/v1/trading/mode`; muted indicator when `runner_running` false.

- [x] **Step 2: Page** — table columns per spec; paused banner using `@/components/ui/alert`; `EventSource` to `/api/v1/opportunities/stream` with cookie auth (if cookies not sent, fall back to poll `GET /opportunities` every 2s).

- [x] **Step 3: Routes** — `/opportunities` primary; remove `PositionsPage` routes.

- [x] **Step 4: `npm test && npm run lint`**

- [x] **Step 5: Commit**

```bash
git add web/src/
git commit -m "feat(web): Opportunities page and bot mode header"
```

---

### Task 19: Remove positions + e2e refresh

**Files:**
- Delete or orphan: `web/src/pages/PositionsPage.tsx`, `web/src/pages/PositionDetail.tsx` (delete if no imports)
- Modify: `web/e2e/dashboard-conformance.spec.ts`
- Create: `web/e2e/opportunities.spec.ts`
- Modify: `web/e2e/positions-degraded.spec.ts` — delete or rewrite for opportunities degraded state

- [x] **Step 1: Playwright** — login → `/opportunities` heading **Opportunities**; set paused → banner visible.

- [x] **Step 2: Remove obsolete assertions** referencing `/positions`.

- [x] **Step 3: `cd web && npx playwright test`**

- [x] **Step 4: Commit**

```bash
git add web/
git commit -m "test(e2e): opportunities route and retire positions UI"
```

---

### Task 20: Admin account_status (API-only)

**Files:**
- Modify: `src/internal/dashboard/trading_handlers.go` or admin-only PUT on `/settings` guarded by env `OPTITRADE_DASHBOARD_ADMIN_USER`

- [x] **Step 1: Allow** `account_status` patch only when `requestUser == os.Getenv("OPTITRADE_DASHBOARD_ADMIN_USER")` (non-empty env).

- [x] **Step 2: Test** unauthorized user cannot disable.

- [x] **Step 3: Commit**

```bash
git add src/internal/dashboard/
git commit -m "feat(dashboard): optional admin account_status patch"
```

---

## Milestone 2–3 (follow-on tasks, same plan file)

- [ ] **M2-A:** Extend selector for additional `AllowedStructures` entries (call debit, iron condor) with OKX builders.
- [ ] **M2-B:** Second underlying (ETH) from operator `currencies`.
- [ ] **M2-C:** Rate-limit batching for books (token bucket per user).
- [ ] **M3-A:** Replace v1 edge proxy with fair-value estimate + documented haircut.
- [ ] **M3-B:** Optional OI/volume filters from OKX tickers.
- [ ] **M3-C:** WebSocket push instead of SSE.

---

## Self-review

**Spec coverage**

| Spec section | Tasks |
|--------------|-------|
| Runner in dashboard + RunnerManager | 12, 13, 17 |
| Eligibility + paused/disabled | 8, 10, 12, 11 |
| Strategy maturity + OKX builders | 3, 4, 7 |
| Universe + books + greeks note | 5, 13 (greeks note placeholder string OK for M1; enrich when OKX ticker exposes greeks) |
| cost.ScoreCandidate + risk.Check + max loss % | 7, 8, 9, 13 |
| API surface | 10, 11, 14, 15 |
| Header bot mode + Opportunities page | 18 |
| Remove /positions | 19 |
| Auto mode | 16 |
| account_status | 8, 20 |
| Testing (Go + Playwright) | per-task tests + 19 |

**Gaps addressed in follow-up:** Full greeks from venue (M1 uses `greeks_note` empty or "n/a"); orphan exchange positions (explicitly out of scope); separate worker process (non-goal).

**Placeholder scan:** None intentional; all test and SQL bodies are concrete.

**Type consistency:** `bot_mode` / `account_status` strings match across Go structs, JSON API, and React. `RowStatus` aligns with UI labels.

---

## Execution handoff

**Plan complete and saved to `docs/superpowers/plans/2026-04-04-trading-runner-opportunities.md`. Two execution options:**

**1. Subagent-Driven (recommended)** — Dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints for review.

**Which approach?**
