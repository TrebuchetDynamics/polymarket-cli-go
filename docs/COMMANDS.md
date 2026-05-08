# Commands

Complete command reference for `polygolem`, aligned with the current
`polygolem --help` command tree. Re-run the documentation refresh when commands
are added or changed.

Source of truth for flag semantics: `polygolem <cmd> --help`.

## Conventions

- All commands accept `--json` to emit structured JSON instead of tables.
- Read-only commands do not require credentials.
- Authenticated commands consume environment variables; see
  [Environment Variables](#environment-variables).
- Live-mutating commands require explicit authenticated credentials and pass
  the local safety gates before submitting.
- **Deposit wallet (type 3 / POLY_1271) is the only supported trading mode.**
  EOA, proxy, and Safe are blocked by CLOB V2 for new accounts.

## Commands

### builder

Manage CLOB L2 credentials and legacy builder-relayer HMAC credentials.

**Usage:**

```
polygolem builder [flags]
polygolem builder [command]
```

**Subcommands:** `auto`, `onboard`. Command group; see subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `builder`. |
| `--json` | bool | `false` | Emit JSON output (global). |

### builder auto

Create or derive CLOB L2 credentials via ClobAuth and persist them to a local
0600 env file.

**Usage:**

```
polygolem builder auto [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--clob-url` | string | `""` | CLOB base URL (default: https://clob.polymarket.com). |
| `--env-file` | string | `""` | Target env file (default: ../go-bot/.env.builder). |
| `--force` | bool | `false` | Overwrite existing builder credentials. |
| `-h, --help` | bool | `false` | Help for `auto`. |
| `--no-validate` | bool | `false` | Skip the relayer HMAC liveness check. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
POLYMARKET_PRIVATE_KEY="0x..." polygolem --json builder auto
```

### builder onboard

Capture legacy builder-relayer HMAC credentials from the settings-page manual
flow and persist them to a local 0600 env file.

**Usage:**

```
polygolem builder onboard [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--env-file` | string | `""` | Target env file (default: ../go-bot/.env.builder). |
| `--force` | bool | `false` | Overwrite existing builder credentials. |
| `-h, --help` | bool | `false` | Help for `onboard`. |
| `--no-validate` | bool | `false` | Skip the relayer HMAC liveness check. |
| `--open-browser` | bool | `false` | Attempt to open polymarket.com/settings?tab=builder. |
| `--json` | bool | `false` | Emit JSON output (global). |

### auth

Inspect authentication readiness.

**Usage:**

```
polygolem auth [flags]
polygolem auth [command]
```

**Subcommands:** `headless-onboard`, `status`. Command group; see subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `auth`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem auth --help
```

### bridge

Polymarket Bridge API.

**Usage:**

```
polygolem bridge [flags]
polygolem bridge [command]
```

**Subcommands:** `assets`, `deposit`. Command group; see subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `bridge`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem bridge --help
```

### clob

CLOB market data and authenticated account commands.

**Usage:**

```
polygolem clob [flags]
polygolem clob [command]
```

**Subcommands:** `balance`, `book`, `cancel`, `cancel-all`,
`cancel-market`, `cancel-orders`, `create-api-key`,
`create-api-key-for-address`, `create-builder-fee-key`, `create-order`,
`list-builder-fee-keys`, `market`, `market-order`, `markets`, `order`,
`orders`, `price-history`, `revoke-builder-fee-key`, `tick-size`, `trades`,
`update-balance`. Command group; see subcommands.

**Authenticated mutation note:** `cancel`, `cancel-orders`, `cancel-market`,
and `cancel-all` require `POLYMARKET_PRIVATE_KEY` to derive L2 credentials, but
they reduce or remove open exposure. `create-order` and `market-order` always
sign as sigtype 3 (POLY_1271, deposit wallet) — the only signature type
Polymarket V2 accepts since the 2026-04-28 cutover. Builder attribution is
configured with `--builder-code` or `POLYMARKET_BUILDER_CODE`; malformed
bytes32 values fail preflight and never reach `/order`.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `clob`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem clob --help
```

### deposit-wallet

Deposit wallet onboarding (WALLET-CREATE, nonce, batch, status).

**Usage:**

```
polygolem deposit-wallet [flags]
polygolem deposit-wallet [command]
```

**Subcommands:** `approve`, `batch`, `deploy`, `derive`, `fund`, `nonce`,
`onboard`, `status`. Command group; see subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `deposit-wallet`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem deposit-wallet --help
```

### data

Polymarket Data API analytics.

**Usage:**

```
polygolem data [flags]
polygolem data [command]
```

**Subcommands:** `activity`, `closed-positions`, `holders`, `leaderboard`,
`live-volume`, `markets-traded`, `open-interest`, `positions`, `trades`,
`value`. Command group; see subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `data`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem data --help
```

### discover

Market discovery via Polymarket Gamma API.

**Usage:**

```
polygolem discover [flags]
polygolem discover [command]
```

**Subcommands:** `comments`, `enrich`, `market`, `markets`, `search`,
`series`, `tags`. Command group; see subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `discover`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem discover --help
```

### events

List Polymarket events.

**Usage:**

```
polygolem events [flags]
polygolem events [command]
```

**Subcommands:** `list`. Command group; see subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `events`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem events --help
```

### health

Check Gamma and CLOB API reachability.

**Usage:**

```
polygolem health [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `health`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json health
```

### live

Inspect live gate status.

**Usage:**

```
polygolem live [flags]
polygolem live [command]
```

**Subcommands:** `status`. Command group; see subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `live`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem live --help
```

### orderbook

Read CLOB order book data.

**Usage:**

```
polygolem orderbook [flags]
polygolem orderbook [command]
```

**Subcommands:** `fee-rate`, `get`, `last-trade`, `midpoint`, `price`,
`spread`, `tick-size`. Command group; see subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `orderbook`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem orderbook --help
```

### paper

Inspect local paper trading state.

**Usage:**

```
polygolem paper [flags]
polygolem paper [command]
```

**Subcommands:** `buy`, `positions`, `reset`, `sell`. Command group; see
subcommands.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `paper`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem paper --help
```

### preflight

Inspect local CLI readiness.

**Usage:**

```
polygolem preflight [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `preflight`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json preflight
```

### stream

Polymarket WebSocket streams.

**Usage:**

```
polygolem stream [flags]
polygolem stream [command]
```

**Subcommands:** `market`. Public market stream is implemented. Authenticated
user stream is intentionally not exposed until L2 WebSocket auth has local tests.

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `stream`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem stream --help
```

### version

Print version.

**Usage:**

```
polygolem version [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `version`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem version
```

### auth status

Report authentication readiness without exposing credential material.

**Usage:**

```
polygolem auth status [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `status`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json auth status
```

### auth headless-onboard

Run SIWE login and mint a V2 relayer API key without a browser.

**Usage:**

```
polygolem auth headless-onboard [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--env-file` | string | `""` | Target env file (default: ../go-bot/.env.relayer-v2). |
| `--force` | bool | `false` | Overwrite existing env file. |
| `--gamma-url` | string | `""` | Gamma API base URL (default: https://gamma-api.polymarket.com). |
| `-h, --help` | bool | `false` | Help for `headless-onboard`. |
| `--relayer-url` | string | `""` | Relayer base URL (default: https://relayer-v2.polymarket.com). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
POLYMARKET_PRIVATE_KEY="0x..." polygolem --json auth headless-onboard
```

### bridge assets

List supported bridge assets.

**Usage:**

```
polygolem bridge assets [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `assets`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json bridge assets
```

### bridge deposit

Create deposit addresses.

**Usage:**

```
polygolem bridge deposit <address> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `deposit`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json bridge deposit <wallet-address>
```

### clob balance

Get CLOB balance and allowances.

**Usage:**

```
polygolem clob balance [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--asset-type` | string | `collateral` | Asset type. |
| `-h, --help` | bool | `false` | Help for `balance`. |
| `--output` | string | `json` | Output format (json). |
| `--token-id` | string | `""` | Conditional token id. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob balance --asset-type collateral
```

### clob book

Get L2 order book.

**Usage:**

```
polygolem clob book <token-id> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `book`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob book <token-id>
```

### clob create-api-key

Create or derive bootstrap CLOB API credentials for the EOA. For live
deposit-wallet trading, also create the owner-scoped key with
`clob create-api-key-for-address --owner <deposit-wallet>`.

**Usage:**

```
polygolem clob create-api-key [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `create-api-key`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob create-api-key
```

### clob create-api-key-for-address

Create CLOB API credentials for a deposit wallet or smart-wallet owner address
while signing L1 auth with `POLYMARKET_PRIVATE_KEY`.

**Usage:**

```
polygolem clob create-api-key-for-address [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `create-api-key-for-address`. |
| `--output` | string | `json` | Output format (json). |
| `--owner` | string | `""` | Deposit wallet owner address. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob create-api-key-for-address --owner 0xDepositWallet
```

### clob create-builder-fee-key

Mint a CLOB builder fee key via `POST /auth/builder-api-key`. The returned
`key` is the bytes32-compatible builder attribution value used by V2 orders.

**Usage:**

```
polygolem clob create-builder-fee-key [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `create-builder-fee-key`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob create-builder-fee-key
```

### clob create-order

Create a signed CLOB limit order.

**Usage:**

```
polygolem clob create-order [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--builder-code` | string | `""` | 0x-prefixed bytes32 builder attribution code. |
| `-h, --help` | bool | `false` | Help for `create-order`. |
| `--expiration` | string | `0` | Unix timestamp for GTD orders (`0` = no expiration). |
| `--order-type` | string | `GTC` | Order type. |
| `--output` | string | `json` | Output format (json). |
| `--post-only` | bool | `false` | Post-only order (maker-only, rejected if it would take). |
| `--price` | string | `""` | Limit price. |
| `--side` | string | `buy` | Order side. |
| `--size` | string | `""` | Order size. |
| `--token` | string | `""` | CLOB token id. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob create-order \
  --token <token-id> --side buy --price 0.51 --size 10
polygolem --json clob create-order \
  --token <token-id> --side buy --price 0.51 --size 10 \
  --order-type GTD --expiration 1778125000
polygolem --json clob create-order \
  --token <token-id> --side buy --price 0.51 --size 10 \
  --builder-code "$POLYMARKET_BUILDER_CODE"
polygolem --json clob create-order \
  --token <token-id> --side buy --price 0.51 --size 10 \
  --post-only
```

### clob market

Get CLOB market by condition ID.

**Usage:**

```
polygolem clob market <condition-id> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `market`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob market <condition-id>
```

### clob markets

List CLOB markets with cursor pagination.

**Usage:**

```
polygolem clob markets [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--cursor` | string | `""` | Pagination cursor. |
| `-h, --help` | bool | `false` | Help for `markets`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob markets --cursor <next-cursor>
```

### clob market-order

Create a signed CLOB market/FOK order.

**Usage:**

```
polygolem clob market-order [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--amount` | string | `""` | USDC amount. |
| `--builder-code` | string | `""` | 0x-prefixed bytes32 builder attribution code. |
| `-h, --help` | bool | `false` | Help for `market-order`. |
| `--order-type` | string | `FOK` | Order type. |
| `--output` | string | `json` | Output format (json). |
| `--price` | string | `""` | Limit price. |
| `--side` | string | `buy` | Order side. |
| `--token` | string | `""` | CLOB token id. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob market-order \
  --token <token-id> --side buy --amount 5
```

### clob list-builder-fee-keys

List CLOB builder fee keys for the authenticated wallet.

**Usage:**

```
polygolem clob list-builder-fee-keys [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `list-builder-fee-keys`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob list-builder-fee-keys
```

### clob revoke-builder-fee-key

Revoke one CLOB builder fee key.

**Usage:**

```
polygolem clob revoke-builder-fee-key [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `revoke-builder-fee-key`. |
| `--key` | string | `""` | Builder fee key to revoke. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob revoke-builder-fee-key --key "$POLYMARKET_BUILDER_CODE"
```

### clob orders

List authenticated CLOB orders.

**Usage:**

```
polygolem clob orders [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `orders`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob orders
```

### clob order

Get a single authenticated CLOB order by ID.

**Usage:**

```
polygolem clob order <order-id> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `order`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob order <order-id>
```

### clob cancel

Cancel a single open CLOB order.

**Usage:**

```
polygolem clob cancel <order-id> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `cancel`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob cancel <order-id>
```

### clob cancel-orders

Cancel multiple open CLOB orders. Pass a comma-separated order ID list.

**Usage:**

```
polygolem clob cancel-orders <order-ids> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `cancel-orders`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob cancel-orders <order-id-1>,<order-id-2>
```

### clob cancel-all

Cancel all open CLOB orders for the authenticated account.

**Usage:**

```
polygolem clob cancel-all [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `cancel-all`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob cancel-all
```

### clob cancel-market

Cancel open CLOB orders matching a market condition ID or asset/token ID.

**Usage:**

```
polygolem clob cancel-market [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--asset` | string | `""` | Asset/token ID. |
| `-h, --help` | bool | `false` | Help for `cancel-market`. |
| `--market` | string | `""` | Market condition ID. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob cancel-market --market <condition-id>
polygolem --json clob cancel-market --asset <token-id>
```

### clob price-history

Get CLOB token price history.

**Usage:**

```
polygolem clob price-history <token-id> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `price-history`. |
| `--interval` | string | `1m` | History interval. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob price-history <token-id> --interval 1m
```

### clob tick-size

Get minimum tick size.

**Usage:**

```
polygolem clob tick-size <token-id> [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `tick-size`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob tick-size <token-id>
```

### clob trades

List authenticated CLOB trades.

**Usage:**

```
polygolem clob trades [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `trades`. |
| `--output` | string | `json` | Output format (json). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob trades
```

### clob update-balance

Refresh CLOB balance and allowances.

**Usage:**

```
polygolem clob update-balance [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--asset-type` | string | `collateral` | Asset type. |
| `-h, --help` | bool | `false` | Help for `update-balance`. |
| `--output` | string | `json` | Output format (json). |
| `--token-id` | string | `""` | Conditional token id. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json clob update-balance --asset-type collateral
```

### deposit-wallet approve

Build the standard 6-call approval batch (pUSD + CTF for all 3 V2 exchange
spenders). Without `--submit`, prints the calldata JSON for review. With
`--submit`, signs and submits the WALLET batch via the relayer.

**Usage:**

```
polygolem deposit-wallet approve [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `approve`. |
| `--submit` | bool | `false` | Sign and submit the approval batch. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json deposit-wallet approve --submit
```

### deposit-wallet batch

Sign an EIP-712 `DepositWallet.Batch` message and submit to the relayer.
`--calls-json` must be a JSON array of `DepositWalletCall` objects:
`[{"target":"0x...","value":"0","data":"0x..."}, ...]`. Use `--auto-approve`
to build and submit the standard 6-call approval batch (pUSD + CTF for all
3 V2 exchange spenders).

**Usage:**

```
polygolem deposit-wallet batch [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--calls-json` | string | `""` | JSON array of `DepositWalletCall` objects. |
| `--deadline` | int | `240` | Deadline seconds from now. |
| `-h, --help` | bool | `false` | Help for `batch`. |
| `--nonce` | string | `""` | WALLET nonce (default: fetched from relayer). |
| `--wallet` | string | `""` | Deposit wallet address (default: derived from EOA). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json deposit-wallet batch --calls-json '[{"target":"0x...","value":"0","data":"0x..."}]'
```

### deposit-wallet deploy

Deploy the deposit wallet via relayer WALLET-CREATE.

**Usage:**

```
polygolem deposit-wallet deploy [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `deploy`. |
| `--timeout` | duration | `2m0s` | Max wait time for `--wait`. |
| `--wait` | bool | `false` | Poll until transaction reaches terminal state. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json deposit-wallet deploy --wait
```

### deposit-wallet derive

Derive the deterministic deposit wallet address.

**Usage:**

```
polygolem deposit-wallet derive [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `derive`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json deposit-wallet derive
```

### deposit-wallet fund

Send pUSD from the EOA to the deposit wallet address via direct ERC-20
transfer. `--amount` is in pUSD (e.g. `0.71` for 0.71 pUSD). Uses 6 decimals
internally. Requires POL for gas on Polygon.

**Usage:**

```
polygolem deposit-wallet fund [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--amount` | string | `""` | pUSD amount to transfer (e.g. `0.71`). |
| `-h, --help` | bool | `false` | Help for `fund`. |
| `--rpc-url` | string | `""` | Polygon RPC URL (default: public node). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json deposit-wallet fund --amount 0.71
```

### deposit-wallet nonce

Get the current WALLET nonce for the owner.

**Usage:**

```
polygolem deposit-wallet nonce [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `nonce`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json deposit-wallet nonce
```

### deposit-wallet onboard

Run the complete deposit wallet setup sequence:

1. Derive the deterministic deposit wallet address.
2. Deploy via WALLET-CREATE (skip with `--skip-deploy` if already deployed).
3. Submit the 6-call approval batch for pUSD and CTF (skip with `--skip-approve`).
4. Transfer pUSD from EOA to deposit wallet (requires `--fund-amount`).

After onboarding, sync CLOB:
`polygolem clob update-balance --asset-type collateral`.

**Usage:**

```
polygolem deposit-wallet onboard [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--fund-amount` | string | `""` | pUSD amount to transfer from EOA to deposit wallet (e.g. `0.71`). |
| `-h, --help` | bool | `false` | Help for `onboard`. |
| `--skip-approve` | bool | `false` | Skip approval batch. |
| `--skip-deploy` | bool | `false` | Skip WALLET-CREATE (wallet already deployed). |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json deposit-wallet onboard --fund-amount 0.71
```

### deposit-wallet status

Check deposit wallet deployment status or transaction state.

**Usage:**

```
polygolem deposit-wallet status [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `status`. |
| `--tx-id` | string | `""` | Transaction ID to poll. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json deposit-wallet status --tx-id <tx-id>
```

### discover enrich

Enrich market with CLOB data.

**Usage:**

```
polygolem discover enrich [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `enrich`. |
| `--id` | string | `""` | Market Gamma ID. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json discover enrich --id <gamma-id>
```

### discover market

Get market details.

**Usage:**

```
polygolem discover market [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `market`. |
| `--id` | string | `""` | Market Gamma ID. |
| `--slug` | string | `""` | Market slug. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json discover market --slug <market-slug>
```

### discover markets

List Gamma markets with offset pagination and filters.

**Usage:**

```
polygolem discover markets [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--active` | bool | `true` | Filter active markets. |
| `--ascending` | bool | `false` | Sort ascending. |
| `--closed` | bool | `false` | Filter closed markets. |
| `-h, --help` | bool | `false` | Help for `markets`. |
| `--limit` | int | `20` | Max markets. |
| `--offset` | int | `0` | Pagination offset. |
| `--order` | string | `""` | Gamma order field. |
| `--tag-id` | int | `0` | Filter by tag ID. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json discover markets --limit 20 --active
```

### discover search

Search markets and events.

**Usage:**

```
polygolem discover search [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `search`. |
| `--limit` | int | `10` | Max results. |
| `--query` | string | `""` | Text query. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json discover search --query "bitcoin" --limit 10
```

### discover tags

List Gamma tags/categories or fetch one by ID or slug.

**Usage:**

```
polygolem discover tags [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `tags`. |
| `--id` | string | `""` | Tag ID. |
| `--limit` | int | `100` | Max tags. |
| `--offset` | int | `0` | Pagination offset. |
| `--slug` | string | `""` | Tag slug. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json discover tags --limit 100
polygolem --json discover tags --slug crypto
```

### discover series

List Gamma series or fetch one by ID.

**Usage:**

```
polygolem discover series [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--closed` | bool | `false` | Filter closed series. |
| `-h, --help` | bool | `false` | Help for `series`. |
| `--id` | string | `""` | Series ID. |
| `--limit` | int | `20` | Max series. |
| `--offset` | int | `0` | Pagination offset. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json discover series --limit 20
polygolem --json discover series --id <series-id>
```

### discover comments

List public Gamma comments, fetch one by ID, or list by user.

**Usage:**

```
polygolem discover comments [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--entity-id` | int | `0` | Comment parent entity ID. |
| `--entity-type` | string | `""` | Comment parent entity type. |
| `-h, --help` | bool | `false` | Help for `comments`. |
| `--id` | string | `""` | Comment ID. |
| `--limit` | int | `20` | Max comments. |
| `--offset` | int | `0` | Pagination offset. |
| `--user` | string | `""` | User wallet address. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json discover comments --entity-id <gamma-entity-id> --entity-type market
polygolem --json discover comments --user 0x...
```

### data positions

List open positions for a user.

**Usage:** `polygolem data positions [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--user` | string | `""` | User wallet address. |
| `--limit` | int | `20` | Max rows. |
| `-h, --help` | bool | `false` | Help for `positions`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data positions --user 0x...
```

### data closed-positions

List closed positions for a user.

**Usage:** `polygolem data closed-positions [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--user` | string | `""` | User wallet address. |
| `--limit` | int | `20` | Max rows. |
| `-h, --help` | bool | `false` | Help for `closed-positions`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data closed-positions --user 0x...
```

### data trades

List public Data API trades for a user.

**Usage:** `polygolem data trades [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--user` | string | `""` | User wallet address. |
| `--limit` | int | `20` | Max rows. |
| `-h, --help` | bool | `false` | Help for `trades`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data trades --user 0x... --limit 20
```

### data activity

List public activity for a user.

**Usage:** `polygolem data activity [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--user` | string | `""` | User wallet address. |
| `--limit` | int | `20` | Max rows. |
| `-h, --help` | bool | `false` | Help for `activity`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data activity --user 0x... --limit 20
```

### data holders

List top holders for a token.

**Usage:** `polygolem data holders [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--token-id` | string | `""` | CLOB token ID. |
| `--limit` | int | `20` | Max rows. |
| `-h, --help` | bool | `false` | Help for `holders`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data holders --token-id <token-id> --limit 20
```

### data value

Get total portfolio value for a user.

**Usage:** `polygolem data value [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--user` | string | `""` | User wallet address. |
| `-h, --help` | bool | `false` | Help for `value`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data value --user 0x...
```

### data markets-traded

Get total markets traded for a user.

**Usage:** `polygolem data markets-traded [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--user` | string | `""` | User wallet address. |
| `-h, --help` | bool | `false` | Help for `markets-traded`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data markets-traded --user 0x...
```

### data open-interest

Get open interest for a token.

**Usage:** `polygolem data open-interest [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--token-id` | string | `""` | CLOB token ID. |
| `-h, --help` | bool | `false` | Help for `open-interest`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data open-interest --token-id <token-id>
```

### data leaderboard

List trader leaderboard rows.

**Usage:** `polygolem data leaderboard [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--limit` | int | `20` | Max rows. |
| `-h, --help` | bool | `false` | Help for `leaderboard`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data leaderboard --limit 20
```

### data live-volume

Get live volume summary.

**Usage:** `polygolem data live-volume [flags]`

| Flag | Type | Default | Description |
|---|---|---|---|
| `--limit` | int | `20` | Max rows. |
| `-h, --help` | bool | `false` | Help for `live-volume`. |
| `--json` | bool | `false` | Emit JSON output (global). |

```bash
polygolem --json data live-volume --limit 20
```

### events list

List events.

**Usage:**

```
polygolem events list [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `list`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json events list
```

### live status

Report live gate state without enabling execution.

**Usage:**

```
polygolem live status [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `status`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json live status
```

### orderbook fee-rate

Get fee rate in bps.

**Usage:**

```
polygolem orderbook fee-rate [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `fee-rate`. |
| `--token-id` | string | `""` | CLOB token ID. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json orderbook fee-rate --token-id <token-id>
```

### orderbook get

Get L2 order book.

**Usage:**

```
polygolem orderbook get [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `get`. |
| `--token-id` | string | `""` | CLOB token ID. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json orderbook get --token-id <token-id>
```

### orderbook last-trade

Get last trade price.

**Usage:**

```
polygolem orderbook last-trade [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `last-trade`. |
| `--token-id` | string | `""` | CLOB token ID. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json orderbook last-trade --token-id <token-id>
```

### orderbook midpoint

Get midpoint price.

**Usage:**

```
polygolem orderbook midpoint [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `midpoint`. |
| `--token-id` | string | `""` | CLOB token ID. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json orderbook midpoint --token-id <token-id>
```

### orderbook price

Get best price (BUY side).

**Usage:**

```
polygolem orderbook price [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `price`. |
| `--token-id` | string | `""` | CLOB token ID. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json orderbook price --token-id <token-id>
```

### orderbook spread

Get bid-ask spread.

**Usage:**

```
polygolem orderbook spread [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `spread`. |
| `--token-id` | string | `""` | CLOB token ID. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json orderbook spread --token-id <token-id>
```

### orderbook tick-size

Get minimum tick size.

**Usage:**

```
polygolem orderbook tick-size [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `tick-size`. |
| `--token-id` | string | `""` | CLOB token ID. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json orderbook tick-size --token-id <token-id>
```

### paper buy

Local simulated paper-trading buy against persisted state.

**Usage:**

```
polygolem paper buy [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `buy`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json paper buy
```

### paper positions

List local paper-trading positions.

**Usage:**

```
polygolem paper positions [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `positions`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json paper positions
```

### paper reset

Reset local paper-trading state.

**Usage:**

```
polygolem paper reset [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `reset`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem paper reset
```

### paper sell

Local simulated paper-trading sell against persisted state.

**Usage:**

```
polygolem paper sell [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `-h, --help` | bool | `false` | Help for `sell`. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json paper sell
```

### stream market

Stream public CLOB market events for one or more asset/token IDs.

**Usage:**

```
polygolem stream market [flags]
```

**Flags:**

| Flag | Type | Default | Description |
|---|---|---|---|
| `--asset-ids` | string | `""` | Comma-separated CLOB token IDs. |
| `-h, --help` | bool | `false` | Help for `market`. |
| `--max-messages` | int | `0` | Stop after this many messages; 0 streams until interrupted. |
| `--url` | string | `wss://ws-subscriptions-clob.polymarket.com/ws/market` | WebSocket URL. |
| `--json` | bool | `false` | Emit JSON output (global). |

**Example:**

```bash
polygolem --json stream market --asset-ids <token-id-1>,<token-id-2> --max-messages 10
```

## Environment Variables

| Variable | Required for |
|---|---|
| `POLYMARKET_PRIVATE_KEY` | All authenticated commands. |
| `POLYMARKET_BUILDER_API_KEY` | Deposit-wallet deploy/batch/onboard. |
| `POLYMARKET_BUILDER_SECRET` | Deposit-wallet deploy/batch/onboard. |
| `POLYMARKET_BUILDER_PASSPHRASE` | Deposit-wallet deploy/batch/onboard. |
| `POLYMARKET_BUILDER_CODE` | Optional CLOB V2 order builder attribution. |
| `POLYMARKET_CLOB_BUILDER_CODE` | Alias for `POLYMARKET_BUILDER_CODE`. |
| `RELAYER_API_KEY` | V2 relayer deploy/batch/onboard auth. |
| `RELAYER_API_KEY_ADDRESS` | Owner address for `RELAYER_API_KEY`. |
| `POLYMARKET_RELAYER_URL` | Override relayer URL (default: relayer-v2.polymarket.com). |

Short-form `BUILDER_API_KEY` / `BUILDER_SECRET` / `BUILDER_PASS_PHRASE` are
also accepted.

## Automation

Polygolem is designed to be driven from shell scripts and AI agents.

**Recommended shell harness:**

```bash
#!/usr/bin/env bash
set -euo pipefail
```

**Pair `--json` output with `jq`:**

```bash
polygolem --json version | jq -r '.data.version'
polygolem --json discover search --query "btc" --limit 3 | jq '.data[].title'
polygolem --json orderbook get --token-id "$TOKEN" | jq '.data.bids[0]'
```

**Exit codes:** A non-zero exit always indicates failure. Inspect the JSON
`error.code` field for the specific failure category (see the JSON contract
docs once they land).

**Idempotency:** Read commands are idempotent. Mutation commands
(`clob create-order`, `clob market-order`, `deposit-wallet onboard`,
`deposit-wallet fund`) are not — never retry blindly without first checking
state via the corresponding read command.
