# Browser Fallback for Polymarket Login

**Last updated:** 2026-05-10

This page is a fallback runbook. The normal path is:

```bash
polygolem auth login
polygolem builder auto
```

Polymarket login signs with the EOA. The deposit wallet remains the trading
wallet for pUSD, POLY_1271 orders, CTF positions, approvals, and redemption.

## When to Use This

Use a browser only when the CLI path is blocked by:

- a temporary Gamma, CLOB, or relayer API change;
- local network, TLS, or cookie policy problems;
- an account policy that asks for extra interactive verification.

If `polygolem auth login` succeeds, you do not need this page.

## Security Warning

`POLYMARKET_PRIVATE_KEY` controls real funds. Importing it into a browser
wallet exposes it to browser extensions, phishing, clipboard capture, and the
wallet extension's own storage model.

Prefer these options:

| Method | Security | Notes |
|---|---|---|
| Hardware wallet | Highest | Key never leaves the device. |
| WalletConnect mobile wallet | High | Key stays on the phone. |
| Dedicated browser profile | Medium | Fresh profile, complete login, delete profile. |
| Raw key import | Low | Use only for small balances or recovery. |

## Manual Steps

1. Confirm the local identity:

   ```bash
   polygolem deposit-wallet derive --json
   ```

2. Open `https://polymarket.com`.

3. Connect the wallet that controls the EOA from `POLYMARKET_PRIVATE_KEY`.
   This is the address shown in the SIWE prompt.

4. Sign the message. The text should match the Polymarket SIWE shape:

   ```text
   polymarket.com wants you to sign in with your Ethereum account:
   0x...

   Welcome to Polymarket! Sign to connect.

   URI: https://polymarket.com
   Version: 1
   Chain ID: 137
   Nonce: ...
   Issued At: ...
   Expiration Time: ...
   ```

5. Return to the CLI and refresh credentials:

   ```bash
   polygolem auth login
   polygolem builder auto --force
   polygolem auth status --check-deposit-key
   ```

6. Disconnect the wallet or remove the imported account from the browser.

## Bot-Generated Keys

If an agent generated the key and you must import it manually:

```bash
polygolem auth export-key --confirm
```

Do this only in a private terminal. Clear shell history afterwards and remove
the browser wallet account after the fallback is complete.

## What This Does Not Change

- It does not create a new deposit wallet.
- It does not move pUSD.
- It does not approve trading or settlement contracts.
- It does not make the deposit wallet sign the SIWE message.

After fallback login, continue with the standard flow in
[ONBOARDING.md](./ONBOARDING.md).
