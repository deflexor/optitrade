# Contracts: Operator Trading Dashboard

Machine-readable API contract for the Go dashboard BFF (`/api/v1/*`).

| File | Purpose |
|------|---------|
| [dashboard-api.openapi.yaml](dashboard-api.openapi.yaml) | OpenAPI 3.0: auth, snapshot, P/L series, positions, close/rebalance previews and actions. |

## Versioning

- Bump `info.version` when breaking JSON shapes or path semantics.
- Frontend Axios client SHOULD use the same version string in `X-Client-Version` header once implemented (optional).

## Security

- Cookie name and attributes documented in implementation; OpenAPI lists `cookie` security scheme for protected routes.
- Never return password fields or session secrets in JSON bodies.
