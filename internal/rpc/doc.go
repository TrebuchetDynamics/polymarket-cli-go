// Package rpc provides direct on-chain helpers for Polygon operations
// — primarily ERC-20 pUSD transfers from an EOA used by deposit-wallet
// funding.
//
// Calls go through a configured Polygon RPC endpoint and bypass the
// Polymarket relayer. Used only by live-mode commands behind the funding
// gate; no read-only or paper-mode code path depends on it.
//
// This package is internal and not part of the polygolem public SDK.
package rpc
