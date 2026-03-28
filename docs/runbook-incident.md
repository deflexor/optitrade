# Incident runbook: Optitrade Deribit bot

**Canonical**: `docs/runbook-incident.md`  
**Last updated**: 2026-03-28  
**Audience**: on-call operator with exchange UI + shell access to the host running `optitrade`.

---

## 1. Stop automation immediately (kill switch)

1. **Stop the process** running the bot (systemd example):

   ```bash
   sudo systemctl stop optitrade
   ```

   If you use another supervisor, stop that unit instead (Docker: `docker stop …`, Kubernetes: scale deployment to 0, etc.).

2. **Cancel exposure on the exchange** (choose one path; confirm in Deribit UI):

   - **Cancel all on an instrument** (supported by the execution client — use whatever wrapper you deploy, or the exchange UI “cancel all” for the subaccount).
   - **Cancel by label** if your deployment tags orders with a known `label` (e.g. operational drill label from policy).

3. **Verify** in the Deribit UI: open orders for the subaccount should be empty or only intentional manual rests.

4. **Flatten** only if your risk policy says so: close positions via controlled limit orders or manual workflow—do **not** rely on the bot while debugging an incident.

---

## 2. Auth / key hygiene

- If keys may be compromised: **revoke** the Deribit API key, **issue a new pair**, update the host environment (`DERIBIT_CLIENT_ID` /DERIBIT_CLIENT_SECRET`), restart only after validation on **testnet**.
- Prefer keys **without withdrawal** and with **IP allowlist** where available.
- Rotate keys when staff with access offboards; document rotation in your ticket.

---

## 3. Reconciliation (SC-004 reference)

After any window where the bot placed or managed orders:

1. Export or snapshot **open orders** and **positions** from the exchange UI.
2. Compare to local `order_record` / audit outputs your deployment persists (SQLite path is deployment-specific).
3. For each **local open** order, expect a matching exchange `order_id`, or a **terminal** state reached via `get_order_state` (see `internal/execution/reconcile.go`).
4. **Unexplained orphan legs** (position leg not attributable to a known fill or order) must be triaged manually; file a bug if the bounded procedure in code cannot explain the drift.

---

## 4. Who to ping

Escalation is **organization-specific**. Fill in below for your deployment:

| Role           | Contact / channel |
|----------------|-------------------|
| Primary on-call | *TBD*             |
| Trading / risk  | *TBD*             |
| Infra / SRE     | *TBD*             |

---

## 5. Post-incident

- Capture timestamps, exchange screenshots, and log excerpts (redact secrets).
- If automation contributed, open a remediation ticket: config change, code fix, or runbook update.
- Update `docs/trader-safety-cheatsheet.md` if a new failure mode was discovered.
