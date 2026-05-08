# playwright-capture

One-shot HTTP traffic capture for Polymarket signup / API-key creation.
Investigation tool — **not part of the polygolem runtime**.

## Why this exists

`indexer_probe` confirms `/auth/api-key` returns `401 Invalid L1 Request
headers` whether we send raw EOA-signed L1 headers with `POLY_ADDRESS=
depositWallet` or the ERC-7739 wrapped form. The browser succeeds at the
same call. We need its actual request shape (endpoint, headers, body) to
replicate it from Go.

## Run

```bash
cd scripts/playwright-capture
npm install
npm run install:browser   # downloads Chromium for Playwright
npm run capture
```

A Chromium window opens at https://polymarket.com. Drive the signup /
login flow yourself. Stdout prints every `clob.polymarket.com /auth/*`
request with full `POLY_*` headers and body in real time. A full HAR
is written to `./captures/<timestamp>.har`.

`Ctrl+C` to finalize the HAR.

## What to capture

The first L1 request the browser issues after a fresh signup. Specifically:

- **Endpoint** — is it `/auth/api-key`, `/auth/derive-api-key`, or something else?
- **POLY_ADDRESS** — is it the EOA, the proxy/safe address, or the deposit wallet?
- **POLY_SIGNATURE** — 65 bytes (raw ECDSA) or longer (wrapped)?
- **POLY_NONCE / POLY_TIMESTAMP** — what range?
- **Cookies / non-POLY headers** — any browser-only token (CF-Ray, _cfuvid, session JWT)?
- **Body** — empty (GET/POST) or signed payload?

Then decode the typed-data the browser signed by reverse-engineering the
hash from POLY_SIGNATURE + POLY_ADDRESS recovery.

## Outputs go in BLOCKERS.md

Once captured, paste the redacted request into `BLOCKERS.md` § B-5 with
the analysis. The capture file itself is gitignored.
