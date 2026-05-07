# Security Policy

`polygolem` handles private keys, signs Polymarket protocol messages,
deploys deposit wallets, and submits on-chain transfers. Vulnerability
reports are taken seriously.

## Reporting a vulnerability

Do **not** file a public GitHub issue.

Use one of:

1. **GitHub Security Advisories** — preferred. Open a private advisory at
   https://github.com/TrebuchetDynamics/polygolem/security/advisories/new.
2. **Email** — `security@trebuchetdynamics.com`. PGP optional; encrypt if
   the report includes secrets or transcripts.

Include: affected version (commit SHA or tag), reproduction steps, a
proof-of-concept where possible, and impact you observed.

## Expected response time

- Acknowledgement within **3 business days**.
- Initial assessment (in scope / not / needs more info) within **7 business days**.
- Coordinated disclosure timeline agreed before any public write-up.

## In scope

- Private-key handling — anywhere a key is read, held in memory, or
  passed to a signer (`internal/auth`, `internal/wallet`, `internal/rpc`).
- Deposit-wallet flow — derive, deploy, batch, approve, fund, onboard
  (`internal/wallet`, `internal/relayer`, `polygolem deposit-wallet *`).
- Signing paths — EIP-712 order signing, POLY_1271 on-chain signature
  verification, ERC-7739 typed-data flows (`internal/clob`,
  `internal/auth`).
- Builder credential handling — API key / secret / passphrase loading,
  redaction in logs and JSON output (`internal/config`,
  `internal/output`).
- JSON-output redaction — any path where `--json` could leak a secret,
  private key, or unredacted credential.
- Order-execution gates — bypasses of read-only / paper / live mode
  separation.

## Out of scope

- Polymarket protocol issues themselves (CLOB pricing, settlement,
  resolution). Report those to Polymarket directly via their channels.
- Polygon network or Ethereum tooling (e.g., go-ethereum) bugs. Report
  upstream.
- Issues that require an attacker who already has the operator's private
  key or shell access on the machine running polygolem.
- Third-party dependencies — file separately upstream; we will track via
  Dependabot updates.

Thank you for keeping `polygolem` users safe.
