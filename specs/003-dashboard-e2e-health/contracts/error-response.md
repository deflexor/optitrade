# Contract: API error JSON

**Applies**: Any handler using `writeAPIError` in `src/internal/dashboard/errors.go`.

## Response shape

Content-Type: `application/json; charset=utf-8`

```json
{
  "error": "<code>",
  "message": "<human-readable, no secrets>"
}
```

## Examples relevant to positions (unauthenticated examples omitted)

**503** — exchange not configured (nil client), open/closed positions:

```json
{
  "error": "exchange_unavailable",
  "message": "exchange not configured"
}
```

**502** — exchange RPC failure:

```json
{
  "error": "exchange_error",
  "message": "could not load positions"
}
```

**UI expectation**: Client may map `error` + `message` to panel copy; network tab may still show status code — UX should not rely on console absence alone.
