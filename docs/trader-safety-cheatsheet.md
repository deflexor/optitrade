# Trader safety cheatsheet: Optitrade Deribit options bot

**Canonical location**: `docs/trader-safety-cheatsheet.md` (repo root)  
**Feature**: `001-autonomous-deribit-options-robot`  
**Last updated**: 2026-03-28

**Disclaimer**: This is operational and risk-education material, not investment, tax, or legal advice. Cryptocurrency options and automated trading can result in **total loss** of capital. The bot does **not** guarantee any outcome. Read the exchange rules and margin documentation for your account type before enabling automation.

---

## 1. What the bot can and cannot do

| Can (when implemented to spec) | Cannot |
|--------------------------------|--------|
| Restrict strategies to **defined-risk** templates (no naked short options per product scope) | Remove **market tail risk**, **gap risk**, or **exchange liquidation** risk |
| Enforce **configured** caps (loss, Greeks, premium at risk, order counts) in software | Guarantee caps match **exchange maintenance margin** at all times |
| Prefer **maker/post-only** style execution and **cost vetoes** | Eliminate **adverse selection**, model error, or **partial fill** leg risk |
| Enter **protective mode** on bad data / auth / book quality signals | React faster than exchange and network reality; **detection delay exists** |

**Bottom line**: The bot is an **assistant with guardrails**, not insurance. Treat worst case as **meaningful loss of the capital you allocate** to this subaccount.

---

## 2. Account setup (recommended order)

1. **Paper / testnet first**  
   Run with Deribit test credentials until you trust logging, reconciliation, and kill behavior.

2. **Dedicated subaccount**  
   Fund only what you are willing to lose. Keep long-term savings and unrelated strategies elsewhere.

3. **Margin mode literacy**  
   Know whether you use **portfolio margin** or **standard margin** and how options, perps, and balances **interact** in the same account. Re-read the exchange help and margin pages **whenever** you change product mix.

4. **Capital buffer**  
   Keep **extra equity** above "minimum to trade" so spot/vol shocks do not push you near **maintenance**. The bot's internal limits are not a substitute for a human margin cushion.

5. **API keys**  
   - **No withdrawal** permission if the exchange allows separation of scopes.  
   - **IP allowlist** if available; use static egress IPs for the machine running the bot.  
   - Rotate keys if leaked or if an operator leaves.

6. **Labels and inventory**  
   Decide naming conventions for **order labels** so you can **cancel by label** during conflict or drill. Avoid manual orders on the same instruments unless you understand interaction with automation.

---

## 3. Bot configuration checklist (before mainnet)

Work through policy JSON validated against the embedded schema in-tree: `execution/internal/config/policy.schema.json` (validated via `config.LoadFile`). Start from `config/examples/policy.example.json` (**testnet** `environment` only for examples in this repo).

- [ ] **max_loss_per_trade** and **max_daily_loss** are small relative to subaccount equity.  
- [ ] **max_open_premium_at_risk** and **max_portfolio_delta / vega** reflect your real pain tolerance, not optimistic backtests.  
- [ ] **Liquidity filters** (min top size, max spread bps) exclude the strikes you would never manually trade.  
- [ ] **Cost model** fees and slippage buffers are **pessimistic** vs recent live fills.  
- [ ] **protective_mode** thresholds (book gap, feed staleness) are set for your latency profile.  
- [ ] **Playbooks** per regime only list structures you understand (credit spread, iron condor, debit spread, etc.).  
- [ ] **time_in_trade** limits match how long you are willing to hold through event risk.

---

## 4. Day-of operations

- [ ] Confirm **heartbeat**: logs show fresh market data and successful auth.  
- [ ] Confirm **reconciliation**: open orders and positions match the exchange UI within an agreed tolerance.  
- [ ] Have a **manual kill** procedure: process stop, cancel-all or cancel-by-label, flatten if policy says so.  
- [ ] Monitor **margin ratio** on the exchange alongside bot risk snapshots.  
- [ ] After major releases or config edits, watch the first session actively.

---

## 5. When to **stop** trading immediately

Pull automation (and consider flattening) if any of the following occur:

- Repeated **disconnects**, auth errors, or obvious **stale** books while the bot still tries to trade.  
- **Partial fills** leaving odd-legged exposure you do not understand.  
- **Margin** trending toward maintenance despite calm-looking bot metrics.  
- Any **unexpected** strategy, instrument, or size in audit logs.  
- You cannot **explain** the last three trades to another trader in plain language.

---

## 6. "Chances not to get bust" (plain language)

There is **no probability** the team can honestly quote: outcomes depend on markets, parameters, code maturity, and luck.

**Risk goes down when** capital at risk is small, limits are conservative, structures are defined-risk, costs are modeled with slack, and humans monitor margin and incidents.

**Risk goes up when** those are reversed, or when automation runs unattended on mainnet early in the project lifecycle.

---

## 7. Where to read more in this repo

- **Quickstart (CLI, testnet observe, smoke-order gate):** `docs/quickstart.md`  
- **Incident steps (kill switch, cancel-all, reconciliation pointer):** `docs/runbook-incident.md`  
- Product requirements: `kitty-specs/001-autonomous-deribit-options-robot/spec.md`  
- Architecture and stack: `kitty-specs/001-autonomous-deribit-options-robot/plan.md`  
- Research (including safety section): `kitty-specs/001-autonomous-deribit-options-robot/research.md`  
- Exchange API reference (external): see `kitty-specs/001-autonomous-deribit-options-robot/research/source-register.csv` for curated links

---

## 8. Revision

Update this cheatsheet when Deribit changes margin UI, when new playbooks ship, or after **postmortems** from testnet/mainnet incidents.
