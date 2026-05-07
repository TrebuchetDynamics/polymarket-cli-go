# Polymarket Go SDK

[English](README.md) | [中文](README_zh.md)

A comprehensive Go SDK for the Polymarket CLOB (Central Limit Order Book) API, fully implementing all core features from [py-clob-client](https://github.com/Polymarket/py-clob-client).

Follow at X:  @netu5er



## Features

- ✅ **Complete API Coverage**: All endpoints from py-clob-client
- ✅ **Three Authentication Levels**: L0 (read-only), L1 (private key), L2 (full access)
- ✅ **Order Management**: Create, submit, cancel, and query orders
- ✅ **Market Data**: Order books, prices, spreads, and market information
- ✅ **RFQ Support**: Request for Quote functionality
- ✅ **Type Safety**: Strong typing with Go's type system
- ✅ **EIP-712 Signing**: Full support for Ethereum message signing

## Installation

```bash
go get github.com/foxme666/Polymarket-golang@v0.0.1
```

## Quick Start

### Level 0 (Read-Only Mode)

```go
package main

import (
    "fmt"
    "github.com/foxme666/Polymarket-golang/polymarket"
)

func main() {
    // Create read-only client
    client, err := polymarket.NewClobClient(
        "https://clob.polymarket.com",
        137, // Polygon chain ID
        "",  // No private key
        nil, // No API credentials
        nil, // No signature type
        "",  // No funder address
    )
    if err != nil {
        panic(err)
    }

    // Health check
    ok, err := client.GetOK()
    fmt.Println("Server OK:", ok)

    // Get order book
    orderBook, err := client.GetOrderBook("token-id")
    if err != nil {
        panic(err)
    }
    fmt.Printf("OrderBook: %+v\n", orderBook)
}
```

### Level 1 (Private Key Authentication)

```go
client, err := polymarket.NewClobClient(
    "https://clob.polymarket.com",
    137,
    "your-private-key-hex",
    nil,
    nil,
    "",
)

// Create or derive API credentials
creds, err := client.CreateOrDeriveAPIKey(nil)
if err != nil {
    panic(err)
}
fmt.Printf("API Key: %s\n", creds.APIKey)
```

### Level 2 (Full Functionality)

```go
creds := &polymarket.ApiCreds{
    APIKey:      "your-api-key",
    APISecret:   "your-api-secret",
    APIPassphrase: "your-passphrase",
}

client, err := polymarket.NewClobClient(
    "https://clob.polymarket.com",
    137,
    "your-private-key-hex",
    creds,
    nil,
    "",
)

// Query balance
balance, err := client.GetBalanceAllowance(&polymarket.BalanceAllowanceParams{
    AssetType: polymarket.AssetTypeCollateral,
})
if err != nil {
    panic(err)
}
fmt.Printf("Balance: %+v\n", balance)
```

## Examples

### Check Balance

```bash
cd examples/check_balance
export PRIVATE_KEY="your-private-key"
export CHAIN_ID="137"  # Optional, defaults to 137 (Polygon)
export SIGNATURE_TYPE="0"  # Optional, 0=EOA, 1=Magic/Email, 2=Browser
export FUNDER=""  # Optional, for proxy wallets
go run main.go
```

### Create and Post Order

```go
// Create order
orderArgs := &polymarket.OrderArgs{
    TokenID:    "token-id",
    Price:      0.5,
    Size:       100.0,
    Side:       "BUY",
    FeeRateBps: 0,
    Nonce:      1,
    Expiration: 1234567890,
}

order, err := client.CreateOrder(orderArgs, nil)
if err != nil {
    panic(err)
}

// Post order
result, err := client.PostOrder(order, polymarket.OrderTypeGTC)
if err != nil {
    panic(err)
}
fmt.Printf("Order posted: %+v\n", result)
```

### Raw Order Mode (Skip Server Requests)

By default, `CreateOrder` fetches the market's `tick_size`, `neg_risk`, and `fee_rate` from the server. If you want to skip these API calls and provide these values yourself, use the `RawOrder` option:

```go
orderArgs := &polymarket.OrderArgs{
    TokenID:    "token-id",
    Price:      0.56,
    Size:       100.0,
    Side:       "BUY",
    Expiration: 0,
}

// Use RawOrder mode to skip server requests
// Must provide TickSize and NegRisk
tickSize := polymarket.TickSize("0.01")
negRisk := false
options := &polymarket.PartialCreateOrderOptions{
    RawOrder: true,       // Skip server requests for tick_size/neg_risk/fee_rate
    TickSize: &tickSize,  // Required in RawOrder mode
    NegRisk:  &negRisk,   // Required in RawOrder mode
}

order, err := client.CreateOrder(orderArgs, options)
// or
result, err := client.CreateAndPostOrder(orderArgs, options)
```

**Note**: In `RawOrder` mode, `TickSize` and `NegRisk` are **required**. The library still uses `TickSize` to convert price/size to the correct amounts.

### Order Types

`CreateAndPostOrder` supports different order types via the `OrderType` option:

```go
// FAK order (Fill And Kill) - partial fill, cancel remaining
orderType := polymarket.OrderTypeFAK
options := &polymarket.PartialCreateOrderOptions{
    OrderType: &orderType,
}
result, err := client.CreateAndPostOrder(orderArgs, options)

// FOK order (Fill Or Kill) - full fill or cancel
orderType := polymarket.OrderTypeFOK
options := &polymarket.PartialCreateOrderOptions{
    OrderType: &orderType,
}
result, err := client.CreateAndPostOrder(orderArgs, options)
```

| Order Type | Description |
|------------|-------------|
| `GTC` | Good Till Cancel - remains until cancelled (default) |
| `FOK` | Fill Or Kill - full fill or cancel entirely |
| `FAK` | Fill And Kill - partial fill, cancel remaining |
| `GTD` | Good Till Date - expires at specified time (requires `Expiration`) |

### Post Only Orders

Post Only orders are only accepted if they add liquidity to the order book (become a maker order). If a Post Only order would immediately match and become a taker, it will be rejected.

```go
// Use PostOrderWithOptions for Post Only orders
result, err := client.PostOrderWithOptions(signedOrder, polymarket.OrderTypeGTC, true) // postOnly=true

// Or use PostOrders with PostOnly field
args := []polymarket.PostOrdersArgs{
    {Order: signedOrder1, OrderType: polymarket.OrderTypeGTC, PostOnly: true},
    {Order: signedOrder2, OrderType: polymarket.OrderTypeGTD, PostOnly: true},
}
result, err := client.PostOrders(args)
```

**Note**: Post Only is only valid for `GTC` and `GTD` order types.

### Heartbeat API

The Heartbeat API allows you to automatically cancel all orders if your connection is lost. Once started, you must send a heartbeat every 10 seconds or all orders will be cancelled.

```go
// Start heartbeat with an optional ID
heartbeatID := "my-session-123"
result, err := client.PostHeartbeat(&heartbeatID)

// Or send without ID
result, err := client.PostHeartbeat(nil)
```

**Use Case**: Prevent stale orders when your trading system goes offline unexpectedly.

## Web3 Clients

The SDK includes two Web3 clients for on-chain operations:

### PolymarketWeb3Client (Pay Gas)

```go
import "github.com/foxme666/Polymarket-golang/polymarket/web3"

// Create Web3 client (pays gas for transactions)
client, err := web3.NewPolymarketWeb3Client(
    "your-private-key",
    web3.SignatureTypePolyProxy, // 0=EOA, 1=PolyProxy, 2=Safe
    137,                         // Chain ID
    "",                          // RPC URL (empty = default)
)

// Get balances
polBalance, _ := client.GetPOLBalance()
usdcBalance, _ := client.GetUSDCBalance(common.Address{})
tokenBalance, _ := client.GetTokenBalance("token-id", common.Address{})

// Set all necessary approvals
receipts, _ := client.SetAllApprovals()

// Split USDC into positions
receipt, _ := client.SplitPosition(conditionID, 100.0, true) // negRisk=true

// Merge positions back to USDC
receipt, _ := client.MergePosition(conditionID, 100.0, true)

// Transfer USDC
receipt, _ := client.TransferUSDC(recipient, 50.0)

// Transfer conditional tokens
receipt, _ := client.TransferToken("token-id", recipient, 50.0)
```

### PolymarketGaslessWeb3Client (No Gas)

```go
// Create Gasless Web3 client (transactions via relay, no gas required)
// Only supports signature_type=1 (PolyProxy) or signature_type=2 (Safe)
client, err := web3.NewPolymarketGaslessWeb3Client(
    "your-private-key",
    web3.SignatureTypePolyProxy,
    nil,  // Optional: builder credentials
    137,
    "",
)

// Same operations as PolymarketWeb3Client
receipt, _ := client.SplitPosition(conditionID, 100.0, true)
receipt, _ := client.MergePosition(conditionID, 100.0, true)
```

## Project Structure

```
polymarket/
├── client.go                  # Main client structure
├── client_api.go              # API methods (health check, API keys, market data)
├── client_orders.go           # Order management methods (submit, cancel, query)
├── client_order_creation.go   # Order creation methods (CreateOrder, CreateMarketOrder)
├── client_misc.go             # Other features (readonly API keys, order scoring, market queries)
├── rfq_client.go              # RFQ client convenience methods
├── config.go                  # Contract configuration
├── constants.go               # Constants
├── endpoints.go               # API endpoint constants
├── http_client.go             # HTTP client
├── http_helpers.go            # HTTP helper functions (query parameter building)
├── signer.go                  # Signer
├── signing_internal.go        # Signing implementation (EIP-712, HMAC)
├── types.go                   # Type definitions
├── utilities.go               # Utility functions
├── order_summary_wrapper.go   # OrderSummary wrapper
├── headers/                   # Authentication headers (wrapper functions)
│   └── headers.go
├── order_builder/             # Order builder
│   ├── order_builder.go       # Order builder implementation
│   └── helpers.go             # Order builder helper functions
├── rfq/                       # RFQ client
│   ├── rfq_client.go          # RFQ client implementation
│   └── types.go               # RFQ type definitions
└── web3/                      # Web3 clients for on-chain operations
    ├── base_client.go         # Base Web3 client (shared logic)
    ├── web3_client.go         # PolymarketWeb3Client (pay gas)
    ├── gasless_client.go      # PolymarketGaslessWeb3Client (no gas)
    ├── types.go               # Web3 type definitions
    ├── helpers.go             # Web3 helper functions
    ├── abi_loader.go          # ABI loading utilities
    └── abis/                   # Contract ABI files
```

## Implemented Features

### ✅ Core Features
- [x] Basic type definitions (all py-clob-client types)
- [x] Signer - EIP-712 and HMAC signing
- [x] L0/L1/L2 authentication header generation
- [x] HTTP client wrapper
- [x] Client base structure (supports three modes)

### ✅ API Methods
- [x] **Health Check**: `GetOK()`, `GetServerTime()`
- [x] **API Key Management**: `CreateAPIKey()`, `DeriveAPIKey()`, `CreateOrDeriveAPIKey()`, `GetAPIKeys()`, `DeleteAPIKey()`
- [x] **Market Data**: `GetMidpoint()`, `GetMidpoints()`, `GetPrice()`, `GetPrices()`, `GetSpread()`, `GetSpreads()`
- [x] **Order Book**: `GetOrderBook()`, `GetOrderBooks()`, `GetOrderBookHash()`
- [x] **Market Info**: `GetTickSize()`, `GetNegRisk()`, `GetFeeRateBps()` (with caching)
- [x] **Last Trade Price**: `GetLastTradePrice()`, `GetLastTradesPrices()`

### ✅ Order Management
- [x] **Order Submission**: `PostOrder()`, `PostOrders()`, `PostOrderWithOptions()` (with PostOnly support)
- [x] **Order Cancellation**: `Cancel()`, `CancelOrders()`, `CancelAll()`, `CancelMarketOrders()`
- [x] **Order Query**: `GetOrders()`, `GetOrder()`
- [x] **Trade Query**: `GetTrades()`
- [x] **Balance Query**: `GetBalanceAllowance()`
- [x] **Notification Management**: `GetNotifications()`, `DropNotifications()`
- [x] **Heartbeat**: `PostHeartbeat()` - Keep orders alive (auto-cancel after 10s without heartbeat)

### ✅ Order Building and Creation
- [x] Complete order builder implementation (using go-order-utils)
- [x] Order creation methods: `CreateOrder()`, `CreateMarketOrder()`, `CreateAndPostOrder()`
- [x] Market price calculation: `CalculateMarketPrice()`
- [x] Rounding configuration and amount calculation

### ✅ Readonly API Key Management
- [x] `CreateReadonlyAPIKey()` - Create readonly API key
- [x] `GetReadonlyAPIKeys()` - Get readonly API key list
- [x] `DeleteReadonlyAPIKey()` - Delete readonly API key
- [x] `ValidateReadonlyAPIKey()` - Validate readonly API key

### ✅ RFQ Client Features
- [x] `CreateRfqRequest()` - Create RFQ request
- [x] `CancelRfqRequest()` - Cancel RFQ request
- [x] `GetRfqRequests()` - Get RFQ request list
- [x] `CreateRfqQuote()` - Create RFQ quote
- [x] `CancelRfqQuote()` - Cancel RFQ quote
- [x] `GetRfqRequesterQuotes()` - Get quotes on your requests (requester view)
- [x] `GetRfqQuoterQuotes()` - Get quotes you created (quoter view)
- [x] `GetRfqBestQuote()` - Get best RFQ quote
- [x] `AcceptQuote()` - Accept quote (requester side)
- [x] `ApproveOrder()` - Approve order (quoter side)
- [x] `GetRfqConfig()` - Get RFQ configuration

### ✅ Web3 Client Features
- [x] `PolymarketWeb3Client` - On-chain transactions (pays gas)
  - [x] Supports EOA, PolyProxy, and Safe wallets
  - [x] Balance queries (POL, USDC, conditional tokens)
  - [x] Approval management (`SetAllApprovals()`)
  - [x] Position operations (`SplitPosition()`, `MergePosition()`, `RedeemPosition()`, `RedeemPositions()`, `ConvertPositions()`)
  - [x] Token transfers (`TransferUSDC()`, `TransferToken()`)
- [x] `PolymarketGaslessWeb3Client` - Gasless transactions via relay
  - [x] Supports PolyProxy and Safe wallets
  - [x] Same operations as Web3Client without gas fees
  - [x] **Requires Builder credentials** (obtained from Polymarket)
  - [x] **Dynamic Relay Address**: Automatically fetches the current relay node address from `/relay-payload` endpoint
  - [x] **Batch Redeem**: `RedeemPositions()` - Redeem multiple conditionIds in a single transaction

### ✅ Other Features
- [x] Order scoring: `IsOrderScoring()`, `AreOrdersScoring()`
- [x] Market queries: `GetMarkets()`, `GetSimplifiedMarkets()`, `GetSamplingMarkets()`, `GetSamplingSimplifiedMarkets()`
- [x] Market details: `GetMarket()`, `GetMarketTradesEvents()`
- [x] Balance update: `UpdateBalanceAllowance()`
- [x] Order book hash: `GetOrderBookHash()`
- [x] Builder trades: `GetBuilderTrades()`

## Feature Comparison

| Feature | py-clob-client | Go SDK | Status |
|---------|---------------|--------|--------|
| Basic client | ✅ | ✅ | Complete |
| L0/L1/L2 auth | ✅ | ✅ | Complete |
| API key management | ✅ | ✅ | Complete |
| Market data queries | ✅ | ✅ | Complete |
| Order book queries | ✅ | ✅ | Complete |
| Order submit/cancel | ✅ | ✅ | Complete |
| Order queries | ✅ | ✅ | Complete |
| Trade queries | ✅ | ✅ | Complete |
| Balance queries | ✅ | ✅ | Complete |
| Order builder | ✅ | ✅ | Fully implemented |
| RFQ features | ✅ | ✅ | Fully implemented |
| Readonly API keys | ✅ | ✅ | Fully implemented |

## Environment Variables

The example programs use the following environment variables:

- `PRIVATE_KEY` (required unless `PRIVATE_KEY_ENC_FILE` is set): Your Ethereum private key in hex format
- `CHAIN_ID` (optional): Chain ID (default: 137 for Polygon)
- `SIGNATURE_TYPE` (optional): Signature type (0=EOA, 1=Magic/Email, 2=Browser, default: 0)
- `FUNDER` (optional): Funder address for proxy wallets
- `CLOB_HOST` (optional): CLOB API host (default: https://clob.polymarket.com)
- `CLOB_API_KEY` (optional): API key for L2 authentication
- `CLOB_SECRET` (optional): API secret for L2 authentication
- `CLOB_PASSPHRASE` (optional): API passphrase for L2 authentication
- `TOKEN_ID` (optional): Token ID for conditional token balance queries
- `PRIVATE_KEY_ENC_FILE` (optional): Path to an encrypted private key (Fernet token); use instead of `PRIVATE_KEY`
- `CLOB_API_KEY_ENC_FILE` (optional): Path to encrypted `CLOB_API_KEY` file (Fernet token)
- `CLOB_SECRET_ENC_FILE` (optional): Path to encrypted `CLOB_SECRET` file (Fernet token)
- `CLOB_PASSPHRASE_ENC_FILE` (optional): Path to encrypted `CLOB_PASSPHRASE` file (Fernet token)
- `CLOB_*_ENC_FILE` (optional): Only needed if you want to reuse a fixed L2 credential; otherwise L2 creds can be derived from L1.
- `BUILDER_API_KEY_ENC_FILE` / `BUILDER_API_SECRET_ENC_FILE` / `BUILDER_API_PASSPHRASE_ENC_FILE` (optional): Encrypted Builder creds (gasless)
- `POLY_ENC_KEY` (required for *_ENC_FILE): Fernet key or passphrase
- `POLY_ENC_SALT` (required when POLY_ENC_KEY is a passphrase): PBKDF2 salt
- `PRINT_API_CREDS` (optional): Set to `1` to print newly created API credentials
- `DEBUG_HEADERS` (optional): Set to `1` to log auth headers (redacted)
- `DEBUG_HEADERS_FULL` (optional): Set to `1` to log full auth headers

## Changelog

### v0.2.7 (2026-03-14)

#### Bug Fixes

- **Fixed Polygon Bor v2.6.0 `eth_call` Compatibility** - Use raw `eth_call` RPC with `blockOverrides` (`baseFeePerGas: 0`) to bypass baseFee validation
  - Bor v2.6.0 ([announcement](https://forum.polygon.technology/t/bor-v2-6-0-and-erigon-v3-4-0-for-mainnet-and-amoy/21757)) synced upstream go-ethereum's `eth_call` validation logic; nodes now reject calls where `maxFeePerGas` is lower than `baseFee`
  - The Bor node's `setDefaults` injects conflicting gas fields, making it impossible to fix via `CallMsg` gas parameters alone (setting `GasFeeCap`/`GasTipCap` causes "both gasPrice and maxFeePerGas specified"; setting `GasPrice` is ignored)
  - Fix: replace all `ethclient.CallContract` with raw `rpc.Client.CallContext` `eth_call`, using the 4th parameter `blockOverrides: {"baseFeePerGas": "0x0"}` to set block context's baseFee to 0 before gas validation
  - For `estimateGas`, gas fields are removed; all call sites have fallback gas limits
  - This only affects read-only contract calls; real transaction gas price remains unchanged
  - Affects all Polygon RPC providers, not provider-specific

### v0.2.6 (2026-01-28)

#### Bug Fixes

- **Fixed RPC 429 Rate Limit Error** - `waitForReceipt` now includes proper polling interval and timeout
  - Added 4-second polling interval between RPC calls
  - Added 5-minute timeout to prevent infinite waiting
  - Applies to both `PolymarketWeb3Client` and `PolymarketGaslessWeb3Client`

### v0.2.5 (2026-01-24)

#### New Features

- **Unified Batch Redeem** - Single transaction redemption for multiple conditionIds in both Gas and Gasless clients
  - Updated `RedeemPositions(requests []RedeemRequest)` to support single-transaction batching for `PolyProxy` on-chain (previously only serial)
  - Synced API between `PolymarketWeb3Client` and `PolymarketGaslessWeb3Client`
  - Added `ExecuteBatch` for generic multi-call execution on-chain
  - Added auto-discovery example in `examples/gasless_batch_redeem`
  - Dramatically reduces Gas costs and transaction time for large portfolios

### v0.2.3 (2026-01-24)

#### Bug Fixes

- **Gasless Web3 Client - Fixed Invalid Signature Error**
  - Fixed `SignatureParams.relay` to use the dynamic relay address from `/relay-payload` endpoint
  - Previously, the signature used a dynamic relay address while `SignatureParams.relay` used a static configured address, causing signature validation failures
  - Reverted incorrect `to = ProxyFactoryAddress` change that was breaking `ProxyCall.To` target address
  - Now both signature generation and request parameters use consistent dynamic relay address

### v0.2.1 (2026-01-24)

#### Improvements

- **Gasless Web3 Client - Dynamic Relay Address**
  - Added `getRelayPayload()` method to fetch dynamic relay node address from `/relay-payload` endpoint
  - Polymarket's relay service may dynamically assign different relay nodes; the code now fetches the current relay address in real-time



## References

- [py-clob-client](https://github.com/Polymarket/py-clob-client) - Official Python implementation
- [Polymarket CLOB Documentation](https://docs.polymarket.com/developers/CLOB)
- [go-order-utils](https://github.com/polymarket/go-order-utils) - Order building utility library

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
