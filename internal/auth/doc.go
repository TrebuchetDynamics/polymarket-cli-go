// Package auth provides Polymarket authentication primitives — L0 / L1 / L2
// auth, EIP-712 signing, deposit-wallet CREATE2 derivation, and builder
// attribution.
//
// Used by every signed request to the CLOB and relayer. The default mode
// for polygolem is read-only and never enters this package; mutating
// commands acquire signers here behind explicit gates. Start with the
// Signer types and DeriveDepositWallet for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package auth
