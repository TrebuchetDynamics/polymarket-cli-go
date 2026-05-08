// Headless Playwright driver that signs up to polymarket.com via an injected
// `window.ethereum` provider backed by an EOA we control. No Magic Link, no
// email — we want to observe the EOA-only signup path end-to-end so we can
// replicate it from polygolem.
//
// The provider is a minimal EIP-1193 implementation that satisfies what
// Polymarket's frontend asks (eth_requestAccounts, personal_sign,
// eth_signTypedData_v4, chain switches). All signing happens in Node via
// ethers; the browser provider just relays.
//
// Run:
//   npm run eoa-signup
//
// First run generates a fresh EOA and persists it to ./.eoa-key.json so
// subsequent runs reuse the same address. Override with EOA_KEY=0x… env
// var. Screenshots land in ./captures/screens/, HAR in ./captures/.

import { chromium } from "@playwright/test";
import { Wallet, getBytes } from "ethers";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";

const here = dirname(fileURLToPath(import.meta.url));
const capturesDir = resolve(here, "captures");
const screensDir = resolve(capturesDir, "screens");
const keyFile = resolve(here, ".eoa-key.json");
await mkdir(screensDir, { recursive: true });

let privateKey = process.env.EOA_KEY;
if (!privateKey) {
  if (existsSync(keyFile)) {
    const raw = JSON.parse(await readFile(keyFile, "utf8"));
    privateKey = raw.privateKey;
  } else {
    privateKey = Wallet.createRandom().privateKey;
    await writeFile(keyFile, JSON.stringify({ privateKey }, null, 2), {
      mode: 0o600,
    });
  }
}
const wallet = new Wallet(privateKey);
const eoa = wallet.address;
console.log(`EOA: ${eoa}`);

const stamp = new Date().toISOString().replace(/[:.]/g, "-");
const harPath = resolve(capturesDir, `eoa-signup-${stamp}.har`);

const browser = await chromium.launch({
  headless: process.env.HEADLESS !== "0",
});
const context = await browser.newContext({
  recordHar: { path: harPath, mode: "full", content: "embed" },
  viewport: { width: 1280, height: 900 },
  userAgent:
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36",
});

await context.exposeFunction(
  "_polyEthSign",
  async (method, params) => {
    if (method === "personal_sign") {
      const [message] = params;
      const bytes = getBytes(message);
      return await wallet.signMessage(bytes);
    }
    if (method === "eth_signTypedData_v4" || method === "eth_signTypedData_v3") {
      const [, payloadIn] = params;
      const payload =
        typeof payloadIn === "string" ? JSON.parse(payloadIn) : payloadIn;
      const types = { ...payload.types };
      delete types.EIP712Domain;
      console.log("[SIGN] eth_signTypedData_v4 primaryType:", payload.primaryType);
      console.log("[SIGN] domain:", JSON.stringify(payload.domain));
      console.log("[SIGN] message:", JSON.stringify(payload.message));
      return await wallet.signTypedData(payload.domain, types, payload.message);
    }
    throw new Error(`unsupported sign method: ${method}`);
  },
);

await context.addInitScript((address) => {
  // Persist connected state across page navigations so Polymarket doesn't
  // re-pop the auth modal every time we navigate.
  let connected = (() => {
    try {
      return localStorage.getItem("__polygolem_connected__") === "1";
    } catch {
      return false;
    }
  })();
  const listeners = new Map();
  const provider = {
    isMetaMask: true,
    isConnected: () => true,
    chainId: "0x89",
    networkVersion: "137",
    selectedAddress: null,
    request: async ({ method, params }) => {
      console.log(`[provider] ${method}`, params || "");
      switch (method) {
        case "eth_chainId":
          return "0x89";
        case "net_version":
          return "137";
        case "eth_accounts":
          return connected ? [address] : [];
        case "eth_requestAccounts":
          connected = true;
          try { localStorage.setItem("__polygolem_connected__", "1"); } catch {}
          provider.selectedAddress = address;
          // Fire 'accountsChanged' for any subscribers
          for (const cb of listeners.get("accountsChanged") || [])
            cb([address]);
          for (const cb of listeners.get("connect") || []) cb({ chainId: "0x89" });
          return [address];
        case "wallet_switchEthereumChain":
        case "wallet_addEthereumChain":
          return null;
        case "personal_sign":
        case "eth_signTypedData_v4":
        case "eth_signTypedData_v3":
          return await window._polyEthSign(method, params);
        case "eth_sendTransaction":
          throw { code: 4001, message: "user rejected (stub)" };
        default:
          throw {
            code: 4200,
            message: `Unsupported method by stub provider: ${method}`,
          };
      }
    },
    on: (event, cb) => {
      if (!listeners.has(event)) listeners.set(event, []);
      listeners.get(event).push(cb);
    },
    removeListener: (event, cb) => {
      const arr = listeners.get(event) || [];
      const i = arr.indexOf(cb);
      if (i >= 0) arr.splice(i, 1);
    },
    addListener: function (event, cb) {
      this.on(event, cb);
    },
    enable: async () => provider.request({ method: "eth_requestAccounts" }),
  };
  Object.defineProperty(window, "ethereum", {
    value: provider,
    writable: false,
    configurable: false,
  });
  // Also announce via EIP-6963 (Polymarket's modal scans for it)
  const info = {
    uuid: "00000000-0000-4000-a000-000000000001",
    name: "Polygolem Stub",
    icon: "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciLz4=",
    rdns: "io.polygolem.stub",
  };
  window.dispatchEvent(
    new CustomEvent("eip6963:announceProvider", {
      detail: Object.freeze({ info, provider }),
    }),
  );
  window.addEventListener("eip6963:requestProvider", () => {
    window.dispatchEvent(
      new CustomEvent("eip6963:announceProvider", {
        detail: Object.freeze({ info, provider }),
      }),
    );
  });
}, eoa);

const INTERESTING_HOSTS = [
  "clob.polymarket.com",
  "gamma-api.polymarket.com",
  "relayer-v2.polymarket.com",
  "data-api.polymarket.com",
  "geo.polymarket.com",
  "polymarket.com/api",
  "magic.link",
  "privy.io",
  "walletconnect.com",
  "walletconnect.org",
];
function isInteresting(url) {
  return INTERESTING_HOSTS.some((h) => url.includes(h));
}

const page = await context.newPage();
page.on("console", (msg) => {
  const text = msg.text();
  if (text.startsWith("[provider]") || text.startsWith("[SIGN]"))
    console.log(text);
});
page.on("request", (req) => {
  const url = req.url();
  if (!isInteresting(url)) return;
  console.log(`>>> ${req.method()} ${url}`);
  const headers = req.headers();
  for (const k of Object.keys(headers).sort()) {
    const lk = k.toLowerCase();
    if (lk.startsWith("poly_") || lk === "authorization" || lk === "cookie")
      console.log(`    ${k}: ${headers[k]}`);
  }
  const body = req.postData();
  if (body) console.log(`    body: ${body}`);
});
page.on("response", async (res) => {
  const url = res.url();
  if (!isInteresting(url)) return;
  const s = res.status();
  if (
    url.includes("/auth") ||
    url.includes("/builder") ||
    url.includes("/api-key") ||
    url.includes("/profile") ||
    url.includes("/sign-up") ||
    url.includes("/onboard") ||
    url.includes("/wallet-create") ||
    url.includes("/relayer") ||
    s >= 400
  ) {
    console.log(`<<< ${s} ${url}`);
    try {
      const t = await res.text();
      if (t && t.length < 4000) console.log(`    response: ${t}`);
    } catch {
      /* body consumed */
    }
  }
});

async function shot(name) {
  const file = resolve(screensDir, `${stamp}-${name}.png`);
  await page.screenshot({ path: file, fullPage: false });
  console.log(`[screenshot] ${file}`);
}

console.log(`HAR: ${harPath}`);
console.log(`Loading polymarket.com…`);
await page.goto("https://polymarket.com", { waitUntil: "domcontentloaded" });
await page.waitForTimeout(4000); // let async hydration settle
await shot("01-homepage");

console.log("[click] Sign Up");
await page.getByRole("button", { name: /sign up/i }).first().click();
await page.waitForTimeout(2500);
await shot("02-signup-modal");

console.log("[click] More methods");
await page.getByRole("button", { name: /more methods/i }).click();
await page.waitForTimeout(2000);
await shot("03-more-methods");

console.log("[click] EIP-6963 stub provider (Polygolem Stub)");
await page
  .locator('button:has(img[alt="Polygolem Stub"])')
  .first()
  .click();
await page.waitForTimeout(6000); // give SIWE + profile creation time
await shot("04-after-wallet-connect");

// "Choose a username" modal handler. Try clicking Continue with empty input;
// if it's disabled, fill a username and retry.
console.log("[click] post-signup Continue (username modal)");
try {
  const continueBtn = page.getByRole("button", { name: /^continue$/i }).last();
  const enabled = await continueBtn.isEnabled().catch(() => false);
  if (!enabled) {
    console.log("[fill] username field");
    await page
      .locator('input[type="text"], input[placeholder*="username" i]')
      .first()
      .fill(`polygolem-${Date.now()}`);
    await page.waitForTimeout(500);
  }
  await continueBtn.click({ timeout: 5000 });
  await page.waitForTimeout(2000);
} catch (e) {
  console.log(`[click] username modal continue failed: ${e.message}`);
}
await shot("05-after-username");

console.log("[goto] /settings");
await page.goto("https://polymarket.com/settings", { waitUntil: "domcontentloaded" });
await page.waitForTimeout(3000);
await shot("06-settings");

// Sidebar uses "Builder Codes". Navigate via SPA link rather than goto so
// our React state stays connected.
console.log("[click] Builder Codes (sidebar)");
try {
  await page
    .getByRole("link", { name: /builder codes/i })
    .first()
    .click({ timeout: 5000 });
} catch (e) {
  console.log(`[click] sidebar link failed (${e.message}) — falling back to goto`);
  await page.goto("https://polymarket.com/settings?tab=builder", {
    waitUntil: "domcontentloaded",
  });
}
await page.waitForTimeout(5000);
await shot("07-builder-tab");
console.log("[done] builder-tab screenshot taken");

// Builder Settings: fill name, click Create Builder Profile.
console.log("[fill] Builder Name");
try {
  await page
    .getByPlaceholder(/builder name/i)
    .fill(`polygolem-test-${Date.now()}`);
  await page.waitForTimeout(800);
} catch (e) {
  console.log(`[fill] builder name failed: ${e.message}`);
}
await shot("08-builder-name-filled");

console.log("[click] Create Builder Profile");
try {
  await page
    .getByRole("button", { name: /create builder profile/i })
    .click({ timeout: 5000 });
} catch (e) {
  console.log(`[click] Create Builder Profile failed: ${e.message}`);
}
await page.waitForTimeout(8000);
await shot("09-after-create-builder");

// Now visit Relayer API Keys tab.
console.log("[click] Relayer API Keys");
try {
  await page
    .getByRole("link", { name: /relayer api keys/i })
    .first()
    .click({ timeout: 5000 });
} catch (e) {
  console.log(
    `[click] Relayer API Keys link failed (${e.message}) — falling back to goto`,
  );
  await page.goto("https://polymarket.com/settings?tab=relayer-api-keys", {
    waitUntil: "domcontentloaded",
  });
}
await page.waitForTimeout(5000);
await shot("10-relayer-api-keys");

// Try to click any "create" / "generate" button on this tab.
console.log("[click] Generate API Key");
try {
  const apiKeyCandidates = [
    page.getByRole("button", { name: /create.*api.*key/i }),
    page.getByRole("button", { name: /generate.*api.*key/i }),
    page.getByRole("button", { name: /new.*api.*key/i }),
    page.getByRole("button", { name: /create/i }),
    page.getByRole("button", { name: /generate/i }),
  ];
  for (const c of apiKeyCandidates) {
    if ((await c.count()) > 0) {
      await c.first().click({ timeout: 3000 });
      console.log("[click] api-key candidate matched");
      break;
    }
  }
} catch (e) {
  console.log(`[click] api-key generate failed: ${e.message}`);
}
await page.waitForTimeout(10000); // give signing + minting time
await shot("11-after-api-key");

// Dismiss the "API key created" modal and screenshot the docs panel
console.log("[click] Done (dismiss modal)");
try {
  await page
    .getByRole("button", { name: /^done$/i })
    .first()
    .click({ timeout: 3000 });
  await page.waitForTimeout(1500);
} catch (e) {
  console.log(`[click] Done failed: ${e.message}`);
}
await shot("12-relayer-keys-list");

// Visit Trading tab
console.log("[click] Trading tab");
try {
  await page
    .getByRole("link", { name: /^trading$/i })
    .first()
    .click({ timeout: 5000 });
  await page.waitForTimeout(3000);
} catch (e) {
  console.log(`[click] Trading link failed: ${e.message}`);
}
await shot("13-trading-tab");

// Dump full HTML of Trading tab body to find any CLOB / API key related controls
try {
  const trading = await page.evaluate(() => {
    const main = document.querySelector("main") || document.body;
    return main.innerText.slice(0, 4000);
  });
  console.log(`[trading-text] ${trading.replace(/\n+/g, " | ")}`);
} catch {}

console.log("[done] full flow captured");

await context.close();
await browser.close();
console.log(`HAR saved: ${harPath}`);
