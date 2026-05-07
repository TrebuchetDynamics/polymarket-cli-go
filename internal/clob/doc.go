// Package clob is the CLOB API client — full read plus authenticated
// surface, EIP-712, POLY_1271, and ERC-7739 signing paths.
//
// Wraps Polymarket's central limit order book API. Read endpoints (books,
// midpoints, trades, markets) are usable without credentials. Mutating
// endpoints (create/cancel orders) require an L1 or L2 auth header from
// internal/auth and must be invoked only from live mode after gates pass.
// Start with Client for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package clob
