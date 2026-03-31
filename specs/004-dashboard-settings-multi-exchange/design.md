# Design: Per-operator Settings (DB-backed credentials) + OKX

**Feature**: `004-dashboard-settings-multi-exchange`  
**Status**: Approved (session 2026-04-01)  
**Depends on**: `002-dashboard-operator-trading` (sessions, BFF), `003-dashboard-e2e-health` (health/overview contracts)

## 1. Goals

- Move **venue trading credentials** off environment variables for the **operator dashboard**: each **allowlisted operator** (session subject) has their own saved configuration.
- Add a **Settings** page in the SPA that loads/saves this configuration, shows **short help text per field**, and surfaces **warnings** when required values are missing or invalid.
- Support **Deribit** and **OKX**, selectable per operator, including OKX **demo trading** as defined by OKX API v5 (Demo Trading API keys + `x-simulated-trading: 1` on REST; WebSocket hosts differ for demo — `wspap.okx.com` vs `/ws.okx.com`).
- Keep **one server-side secret** outside the DB to **encrypt** operator credentials at rest (operator-approved): env var or mounted key file — **not** venue API keys.

## 2. Non-goals (initial delivery)

- **`observe` / CLI / smoke-order** flows: may continue to use `DERIBIT_*` until a separate change aligns them; dashboard code path **must not** depend on venue keys in env.
- **HSM / Vault / cloud SM**: out of scope; document as future hardening.
- **Full OKX feature parity** with Deribit (especially options-specific shapes): adapter implements what the **current dashboard** needs (positions, account summary, user trades, close/order paths where applicable); gaps are explicit errors or degraded UI with operator-safe messages.

## 3. Security

| Topic | Decision |
|--------|-----------|
| Venue secrets | Stored only in SQLite, AES-GCM (or age-like construct) with secrets **never** logged. |
| Master key | `OPTITRADE_SETTINGS_SECRET` (raw key material, 32-byte recommended) **or** `OPTITRADE_SETTINGS_KEY_FILE` pointing to a file containing the key; fail fast at startup if missing when settings encryption is required. |
| API responses | GET returns **masked** secrets (e.g. last 4 chars / `••••`); full values only accepted on PUT body, never echoed back. |
| Sessions | Existing session cookie model unchanged; settings are keyed by **username** from `SessionStore.LookupUsername`. |

## 4. Data model (SQLite)

New migration under `src/internal/state/migrations/`:

- Table e.g. **`dashboard_operator_settings`**: `username TEXT PRIMARY KEY`, non-secret columns for **provider** (`deribit` \| `okx`), **deribit** testnet/mainnet or explicit base URL policy, **okx_demo** boolean, **`currencies`** override (replaces `OPTITRADE_DASHBOARD_CURRENCIES` for that user when set), **`secrets_blob`** (nonce + ciphertext of JSON payload), **`updated_at`**.

Secrets JSON (plaintext before encryption) holds: Deribit client id/secret; OKX api key, secret, passphrase — only the subset needed for the selected provider.

## 5. BFF API

- **`GET /api/v1/settings`** (authenticated): structured **field catalog** (id, label, help, required, type, warning if applicable) plus **current values** (masked secrets) and **provider**.
- **`PUT /api/v1/settings`** (authenticated): partial or full update; validate provider-specific requirements; re-encrypt secrets blob on change.

Field catalog can be **server-authored** (Go) so UI and backend stay aligned; warnings derive from validation (e.g. OKX selected but passphrase empty).

## 6. Exchange resolution

- Remove **single global** `Exchange` on `Server` from env (`ExchangeFromEnv` unused for dashboard).
- For each protected handler: **`username` → load settings → decrypt secrets → construct client** (with small TTL cache keyed by username + `updated_at` hash to avoid decrypting every request).
- **Health / trading mode**: derive **test vs live** from saved operator settings (Deribit base URL / OKX demo flag), not from `DERIBIT_BASE_URL` env for the dashboard.

## 7. Exchange abstraction

Today handlers use `exchangeReader` with **Deribit** types end-to-end. Target:

- Introduce **dashboard-neutral** types (or a narrow `TradingBackend` interface) for: account summary slice, positions list, user trades, place-order close path.
- **Deribit**: adapter from existing `internal/deribit.REST`.
- **OKX**: new `internal/okx` REST v5 client (signing headers, optional `x-simulated-trading: 1`), adapter mapping into the same dashboard-neutral types.

OKX REST base: `https://www.okx.com` (demo uses header, not a different REST host per OKX v5 docs).

## 8. Frontend

- New **Settings** route, linked from shell/nav.
- Form sections per provider; switching provider shows relevant fields and help.
- Inline warnings next to fields; optional banner if exchange unusable until settings complete.

## 9. Testing

- Unit tests: encryption roundtrip, validation, resolver cache.
- Handler tests: mock trading backend per username.
- Extend Playwright: settings save (mock or stubbed BFF if needed); ensure no `DERIBIT_*` required for dashboard e2e when settings present.

## 10. Rollout notes

- Document: operators **must** set server encryption key; first login → Settings → save credentials.
- Remove dashboard reliance on `DERIBIT_CLIENT_ID` / `DERIBIT_CLIENT_SECRET` / `DERIBIT_BASE_URL` for trading; update README quickstart for dashboard accordingly.

## 11. Open technical decisions (implementation phase)

- Exact **neutral** position/trade schema vs minimal JSON compatibility with current UI (may map OKX fields into existing response shapes where honest).
- **Close / place order** on OKX: product and `instId` mapping from current Deribit-centric flows — may require operator-visible limitations until instrument mapping exists.
