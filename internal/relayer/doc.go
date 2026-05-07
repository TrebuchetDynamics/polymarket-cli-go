// Package relayer is the builder relayer client — WALLET-CREATE,
// WALLET batch, nonce reads, and operation polling.
//
// Used by the deposit-wallet flow to deploy a CREATE2 deposit wallet,
// submit batched calls under POLY_1271, and poll relayer operations until
// they reach a terminal state. Requires builder credentials supplied by
// internal/config. Start with Client for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package relayer
