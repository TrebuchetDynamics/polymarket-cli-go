# Polymarket Coverage Matrix

This matrix tracks the Polymarket surfaces polygolem exposes through the Go
SDK, CLI, docs, and tests. It is intentionally conservative: authenticated user
WebSocket streams and live order execution gates remain documented gaps until
they have local tests.

| Surface | Capabilities | SDK | CLI | Docs | Tests |
|---|---|---|---|---|---|
| Gamma markets | Search, list markets, fetch market by ID/slug, enrich with CLOB data | `internal/gamma`, `pkg/gamma`, `pkg/types`, `pkg/universal` | `discover search`, `discover markets`, `discover market`, `discover enrich` | `README.md`, `docs/COMMANDS.md`, Starlight CLI/Gamma pages | Gamma `httptest` tests, external SDK boundary test, CLI command registration |
| Gamma taxonomy | Tags/categories, series, comments | `internal/gamma`, `pkg/gamma`, `pkg/types`, `pkg/universal` | `discover tags`, `discover series`, `discover comments` | `docs/COMMANDS.md`, Starlight CLI/Gamma pages | External SDK boundary test, CLI command registration |
| CLOB public data | Order book, price, midpoint, spread, tick size, fee rate, last trade, price history, market list, market lookup | `internal/clob`, `pkg/clob`, `pkg/orderbook`, `pkg/types`, `pkg/universal` | `orderbook *`, `clob book`, `clob tick-size`, `clob market`, `clob markets`, `clob price-history` | `README.md`, `docs/COMMANDS.md`, Starlight CLOB/CLI pages | CLOB `httptest` tests, public SDK boundary test, CLI command registration |
| CLOB account reads | L2 API key creation/derivation, owner-scoped API key creation, balance, update balance, orders, order by ID, trades | `internal/clob`, `pkg/clob`, `pkg/universal` | `clob create-api-key`, `clob create-api-key-for-address`, `clob balance`, `clob update-balance`, `clob orders`, `clob order`, `clob trades` | `docs/COMMANDS.md`, Starlight CLOB/CLI pages | CLOB and universal `httptest` tests, public SDK boundary test, CLI command registration |
| CLOB cancellation | Cancel one order, cancel a batch, cancel all, cancel by market/asset | `internal/clob`, `pkg/clob`, `pkg/universal` | `clob cancel`, `clob cancel-orders`, `clob cancel-all`, `clob cancel-market` | `README.md`, `docs/COMMANDS.md`, Starlight CLOB/CLI pages | CLOB and universal `httptest` tests, public SDK boundary test, CLI command registration |
| CLOB placement | Deposit-wallet limit and market/FOK order signing with CLOB V2 payload shape and optional builder attribution | `internal/clob`, `pkg/clob`, `pkg/universal` | `clob create-order`, `clob market-order`, `--builder-code` | `README.md`, `docs/COMMANDS.md`, Starlight CLOB page, `docs/SAFETY.md` | CLOB order-signing and placement tests, public SDK boundary test |
| CLOB builder fee keys | Create, list, and revoke CLOB builder fee keys for V2 order attribution | `internal/clob`, `pkg/clob`, `pkg/universal` | `clob create-builder-fee-key`, `clob list-builder-fee-keys`, `clob revoke-builder-fee-key` | `docs/COMMANDS.md`, Starlight CLOB/CLI pages | CLOB `httptest` tests, public SDK boundary test, CLI command registration |
| Data API | Positions, closed positions, trades, activity, holders, value, markets traded, open interest, leaderboard, live volume | `internal/dataapi`, `pkg/data`, `pkg/types`, `pkg/universal` | `data positions`, `data closed-positions`, `data trades`, `data activity`, `data holders`, `data value`, `data markets-traded`, `data open-interest`, `data leaderboard`, `data live-volume` | `README.md`, `docs/COMMANDS.md`, Starlight Data/CLI pages | Public Data API client tests, external SDK boundary test, universal route tests, CLI command registration |
| Bridge | Supported assets, deposit address creation, deposit status, quotes | `pkg/bridge` | `bridge assets`, `bridge deposit` | `README.md`, `docs/COMMANDS.md`, Starlight bridge guide | Package examples, CLI command registration |
| WebSocket market stream | Public market channel subscription, book/price/last-trade dispatch, reconnect, dedup helpers | `internal/stream`, `pkg/stream`, `pkg/universal` stream constructor | `stream market` | `README.md`, `docs/COMMANDS.md`, Starlight Stream/CLI pages | Local WebSocket SDK test, public SDK boundary test, CLI command registration |
| WebSocket user stream | Authenticated user order/trade stream | Gap | Gap | Documented as planned | Gap: requires L2 WebSocket auth tests |
| Polygon deposit wallet | Derive, deploy, status, nonce, batch, approve, fund, onboard | `internal/auth`, `internal/relayer`, `internal/rpc` | `deposit-wallet *` | `README.md`, `docs/SAFETY.md`, deposit-wallet docs | Existing auth/deposit-wallet tests |
| Polygon wallet actions outside deposit wallet | EOA/proxy/Safe trading modes | Not supported for live trading | Blocked for new production accounts | Documented as unsupported | N/A |

## Current Gaps

- Authenticated user WebSocket stream is still not implemented. Add it only
  after L2 auth header signing and local WebSocket tests exist.
- Data API open-interest currently requires a token ID in the CLI. A future
  all-market variant should be added only after the response shape is captured.
- `docs/COMMANDS.md` and the Starlight CLI reference are generated from the
  Cobra command tree. Run `go run ./cmd/polygolem_docs` after changing CLI
  commands.
- The shared CLI v1 JSON envelope is implemented for success payloads, group
  command usage errors, not-implemented stubs, and common auth failures.
  Remaining JSON-contract work is more precise protocol/transport/chain error
  classification and structured `error.details` for upstream failures.
