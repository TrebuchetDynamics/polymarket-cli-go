// Package wallet provides deposit-wallet primitives — CREATE2 derivation,
// status checks, deploy and batch-signing helpers.
//
// Address derivation is non-mutating and used by read-only deposit-wallet
// commands. Deploy and batch operations sit behind builder credentials
// and the live gate. See docs/DEPOSIT-WALLET-MIGRATION.md for the May 2026
// signature-type migration this package implements.
//
// This package is internal and not part of the polygolem public SDK.
package wallet
