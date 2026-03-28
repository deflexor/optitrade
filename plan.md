Autonomous Deribit options robot

Product goal
Build an autonomous options trading bot for Deribit that trades only liquid BTC/ETH options, uses defined-risk strategies only, and opens/closes positions when the expected edge exceeds fees + slippage + adverse selection.

Core objectives
The bot should classify market regime, choose between low-vol and high-vol playbooks, size positions conservatively, and never exceed preset loss or exposure limits.

Scope
The system will:

ingest Deribit market data and volatility data,
build a regime model,
generate option candidates,
score trades by expected value after costs,
place and manage orders,
monitor fills, exposure, and PnL,
auto-exit or freeze trading on anomaly conditions.

Non-goals
No naked short options, no martingale sizing, no cross-exchange arbitrage, no illiquid strikes, no leverage optimization beyond strict risk caps, and no “trade every signal” behavior.

Data inputs
Use:

public/get_instruments to discover tradable expiries/strikes,
public/ticker and public/get_order_book for pricing and spread checks,
public/get_volatility_index_data for volatility-regime features,
private/get_open_orders, private/get_positions, and private/get_account_summaries for state and risk,
private/cancel_by_label / private/cancel_all_by_instrument for clean exits.

Decision engine

Regime classifier: low vol, normal vol, high vol.
Candidate generator: only liquid expiries/strikes.
Strategy selector: credit spread / iron condor in low vol; debit spread in high vol.
Cost model: subtract estimated fees, spread, and expected slippage.
Risk gate: veto if trade worsens portfolio delta/vega/gamma past limits.
Execution: post-only first, then controlled cross only if edge remains.

Risk controls
Hard limits should include:

max loss per trade,
max daily loss,
max open premium at risk,
max portfolio delta,
max portfolio vega,
max open orders per instrument,
max time-in-trade,
mandatory flatten on feed loss, auth failure, or large book gaps.

Execution rules
Prefer:

post-only limit orders,
combo orders for spreads,
IV orders for options quoting when appropriate,
reduce-only for exits,
small child orders rather than full-size market orders. Deribit’s IV orders update roughly every second and can be stale in fast moves, so the bot should not rely on them alone during sharp spikes.

Architecture suggestion
Use a two-layer design:

Research / strategy service: Python, pandas/numpy, backtesting.
Execution service: Golang

Recommended infra

Go technology stack (recommended)
Core Language: Go 1.26+
WebSocket: gorilla/websocket or nhooyr/websocket
HTTP: net/http or resty
JSON-RPC: custom lightweight client (recommended)
Concurrency: goroutines + channels
worker pools for:
pricing
order placement
risk checks
Storage
SQLite (positions, trades)
Observability
zap or zerolog (logging)

