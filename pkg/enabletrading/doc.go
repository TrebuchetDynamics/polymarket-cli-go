// Package enabletrading exposes the SDK primitives for Polymarket's
// "Enable Trading" UI flow: EOA-signed ClobAuth credential generation and
// deposit-wallet approval-batch signing.
//
// The package intentionally keeps the identity model explicit. Polymarket
// login and CLOB HTTP authentication sign with the EOA; the deposit wallet
// remains the trading wallet for pUSD balances, POLY_1271 order maker/signer
// fields, CTF positions, approvals, and redemption.
package enabletrading
