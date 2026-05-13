# Arenaton Native Comments V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a native Arenaton market discussion slice where users can read Arenaton comments publicly and post wallet-signed comments from Alpha Terminal market detail.

**Architecture:** Keep `/api/comments*` as the source of truth and do not add Polymarket external comments to the client in this slice. Flutter gets comment models and REST methods in `ArenatonAlphaApi`, then renders an `Arenaton comments` section on the existing market detail page.

**Tech Stack:** Go HTTP handlers and store boundary in `server-arenaton/internal/pulse`; Flutter/Dart provider-based Alpha Terminal UI in `arenaton-flutter/lib/features/alpha_terminal`.

---

### Task 1: Server Comment API Confidence

**Files:**
- Modify: `server-arenaton/internal/pulse/comments_test.go`
- Modify if needed: `server-arenaton/internal/pulse/comments.go`

- [ ] **Step 1: Write failing tests**

Add tests using the existing `newTestHandler` fake store from `server_test.go` to prove:
- `GET /api/comments/nonce` returns a nonce for a valid wallet.
- a valid EOA signature creates a visible comment.
- the same nonce cannot be reused.
- `GET /api/comments?provider=polymarket&sourceMarketId=...` returns only the matching Arenaton comments.
- report and hide routes call the store and hidden comments do not show in later list responses.

- [ ] **Step 2: Run red tests**

Run:

```bash
cd server-arenaton && go test ./internal/pulse -run 'TestComments' -count=1
```

Expected: at least one new test fails before fake-store/comment behavior is completed.

- [ ] **Step 3: Implement minimal server/test support**

Prefer improving the fake store in `server_test.go` or adding local test helpers over changing production behavior. Only change `comments.go` if a tested v1 contract is wrong, especially keeping list responses native-only for normal `/api/comments` calls.

- [ ] **Step 4: Run green tests**

Run:

```bash
cd server-arenaton && go test ./internal/pulse -run 'TestComments' -count=1
```

Expected: all comment tests pass.

### Task 2: Flutter Comment Models And API

**Files:**
- Modify: `arenaton-flutter/lib/features/alpha_terminal/domain/alpha/alpha_terminal_models.dart`
- Modify: `arenaton-flutter/lib/features/alpha_terminal/data/sources/arenaton_alpha/arenaton_alpha_api.dart`
- Modify: `arenaton-flutter/test/services/arenaton_alpha_api_test.dart`

- [ ] **Step 1: Write failing API tests**

Add tests proving:
- `listArenatonComments` calls `/api/comments` with `provider=polymarket`, `sourceMarketId`, and optional `conditionId`/`paperMirrorId`.
- `createArenatonComment` fetches `/api/comments/nonce`, signs the exact message format expected by the server, and posts the signed payload.
- response JSON maps into an `ArenatonComment`.

- [ ] **Step 2: Run red tests**

Run:

```bash
cd arenaton-flutter && flutter test test/services/arenaton_alpha_api_test.dart
```

Expected: new comment API tests fail because methods/models do not exist.

- [ ] **Step 3: Implement minimal models and API methods**

Add `ArenatonComment` and a signing callback based API:

```dart
Future<ArenatonComment> createArenatonComment(
  String sourceMarketId, {
  required String wallet,
  required String body,
  required Future<String> Function(String message) signMessage,
  String conditionId = '',
  String paperMirrorId = '',
});
```

Use `Uri.replace(queryParameters: ...)` for reads. Use the same message format as the server:

```text
ArenatonComments
BodyHash: <sha3-256 body hash>
Nonce: <nonce>
Wallet: <wallet>
```

- [ ] **Step 4: Run green API tests**

Run:

```bash
cd arenaton-flutter && flutter test test/services/arenaton_alpha_api_test.dart
```

Expected: service tests pass.

### Task 3: Market Detail Discussion UI

**Files:**
- Modify: `arenaton-flutter/lib/features/alpha_terminal/presentation/detail/external_market_detail_screen.dart`
- Modify: `arenaton-flutter/test/screens/external_market_detail_screen_test.dart`

- [ ] **Step 1: Write failing widget tests**

Add tests proving:
- market detail shows an `Arenaton comments` section after mapping.
- public comments are visible without a private Alpha session.
- composer can post through `WalletController.signPersonalMessage` when a wallet is available.

- [ ] **Step 2: Run red tests**

Run:

```bash
cd arenaton-flutter && flutter test test/screens/external_market_detail_screen_test.dart
```

Expected: new UI tests fail because the section and callbacks do not exist.

- [ ] **Step 3: Implement UI**

Load comments during `_loadDetail`, merge newly-created comments in local state, and render a compact section near decision notes. Use `widget.market.conditionId` as both `sourceMarketId` and `conditionId` for the first client slice, and include `_paperMirror?.paperMarketId` when present.

- [ ] **Step 4: Run green widget tests**

Run:

```bash
cd arenaton-flutter && flutter test test/screens/external_market_detail_screen_test.dart
```

Expected: screen tests pass.

### Task 4: Final Verification

Run:

```bash
cd server-arenaton && go test ./internal/pulse
cd arenaton-flutter && flutter test test/services/arenaton_alpha_api_test.dart test/screens/external_market_detail_screen_test.dart
cd arenaton-flutter && flutter analyze lib/features/alpha_terminal test/services/arenaton_alpha_api_test.dart test/screens/external_market_detail_screen_test.dart
```

Expected: commands pass or any remaining failures are reported with exact output and cause.
