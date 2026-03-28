# Contracts: Autonomous Deribit Options Robot

This directory holds **machine-readable contracts** at trust boundaries for the execution service.

| File | Purpose |
|------|---------|
| [config-policy.schema.json](config-policy.schema.json) | JSON Schema for operator risk and playbook configuration loaded at startup. |
| [event-envelope.schema.json](event-envelope.schema.json) | Optional: normalized audit/event envelope for logs or future streaming consumers. |

Internal Deribit JSON-RPC method mapping and type definitions live in Go package `execution/internal/deribit` (implementation), not duplicated here.

## Versioning

- Bump `version` inside policy JSON when semantics change.
- Execution MUST reject config that fails schema validation (constitution: validate untrusted input at boundaries).
