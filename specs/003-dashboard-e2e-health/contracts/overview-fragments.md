# Contract: `GET /api/v1/overview` — fragments

Protected (session cookie). Full shape in `internal/dashboard/overview.go`. This feature only tightens **`market_mood`** copy.

## `market_mood` object

```json
{
  "label": "",
  "score": null,
  "explanation": "<operator-facing string when mood unavailable>",
  "available": false
}
```

**Rules**:

- When `available` is `false`, `explanation` MUST be suitable for operators (no internal engineering jargon).
- When `available` is `true`, `label` SHOULD carry the primary display text.
