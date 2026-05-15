# polygolem TODO

## Packaging and trust fixes from skeptical review

These notes came from a blunt review of what looks wrong or fragile about the project. The goal is to make a skeptical operator trust the smallest possible safe path instead of adding more surface area.

### Public positioning

- [ ] Narrow the homepage/README promise to one clear claim: **"Go CLI/SDK for safe Polymarket V2 bot infrastructure."**
- [ ] Replace broad marketing language such as "only production-ready option" with evidence-backed claims.
- [ ] Add a visible "Known limitations" section near the top of the README.
- [ ] Clearly distinguish what is proven from what is experimental, research, or internal archaeology.

### Trust and funds safety

- [ ] Add a threat model covering private keys, signing, orders, funds, deposit wallets, API keys, and relayer credentials.
- [ ] Add a funds-safety checklist before any live trading path.
- [ ] Document redaction guarantees and exactly what the CLI will never log or transmit.
- [ ] Add hard confirmations before live actions.
- [ ] Make safety disclaimers explicit enough that users understand stale markets, wrong token IDs, approvals, settlement assumptions, and command misuse can lose money.

### Onboarding and demo path

- [ ] Add one end-to-end demo path: read-only check → paper trade → live-readiness check → tiny live order with warnings.
- [ ] Make the deposit-wallet/new-user onboarding limitation explicit: one-time browser login is still required for some flows.
- [ ] Clarify what is fully headless versus what requires browser/manual setup.
- [ ] Provide a compatibility matrix for Go version, CLOB versioning, wallet type, signature type, relayer/deposit-wallet support, and supported flows.

### Evidence and validation

- [ ] Show proof points rather than claims: CI, test coverage, live smoke tests, compatibility matrix, known limitations, and failure cases.
- [ ] Add or document aggressive live smoke tests for Polymarket API/relayer behavior changes.
- [ ] Add a safe live-order walkthrough with exact preconditions and maximum spend caps.
- [ ] Link at least one bot or paper-trading example that uses polygolem in practice.

### Scope control and docs cleanup

- [ ] Separate the public reader path from internal archaeology: move obsolete investigations, correction notes, PRDs, probes, and historical blockers out of the main path.
- [ ] Keep `BLOCKERS.md`/history useful but avoid making new users parse old wrong conclusions before they can run the happy path.
- [ ] Re-evaluate whether Polydart, docs site, live probes, open-source project analysis, paper trading, SDK, CLI, and deposit-wallet flows should all be presented at once.
- [ ] Define the core thing polygolem does better than anything else.

### Dependency and version story

- [ ] Revisit `go.mod` declaring `go 1.25.0`; explain or adjust if it weakens installability.
- [ ] Explain the heavy crypto/zk/Ethereum indirect dependency tree so the "simple static binary" story remains credible.

## Blunt diagnosis

The project has value, especially the Polymarket V2/deposit-wallet knowledge encoded in Go, but the packaging is ahead of adoption proof. The weakest point is not code; it is trust. A stranger currently has little reason to run a young repo with funds. The next job is not more features; it is evidence, a smaller happy path, and operator confidence.
