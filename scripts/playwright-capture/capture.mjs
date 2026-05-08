// Drives Chromium headed against polymarket.com so a human can complete signup
// or login interactively, while recording every relevant request to a HAR file
// and printing CLOB auth traffic to stdout in real time.
//
// Goal: decode the exact L1 auth request the browser sends when minting an
// API key for a deposit-wallet-owned account. polygolem currently gets
// `Invalid L1 Request headers` on /auth/api-key with our wrapped ERC-7739
// payload — the browser must send something different.
//
// Usage:
//   cd scripts/playwright-capture
//   npm install
//   npm run install:browser
//   npm run capture           # headed Chromium; needs an X/Wayland display
//   npm run capture:xvfb      # runs under Xvfb so it works on a tty
//   HEADLESS=1 npm run capture  # fully headless (no UI; for scripted flows)
//
// Then in the launched browser: sign up / log in via whichever path you want
// to capture (Email-Magic, Wallet-Connect, MetaMask, etc.). Stdout will print
// every clob.polymarket.com /auth/* request and the full POLY_* headers. A HAR
// file with the full session is written to ./captures/<timestamp>.har.
//
// To interact with a browser running inside Xvfb, attach a VNC server to the
// Xvfb display in a second terminal:
//   x11vnc -display :$XVFB_DISPLAY -localhost -forever -nopw -rfbport 5900
// Then connect a VNC client to localhost:5900 from your normal desktop.
// (Set XVFB_DISPLAY=99 below to pin the display number.)
//
// Press Ctrl+C in the terminal to stop and finalize the HAR.

import { chromium } from "@playwright/test";
import { mkdir } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";

const here = dirname(fileURLToPath(import.meta.url));
const capturesDir = resolve(here, "captures");
await mkdir(capturesDir, { recursive: true });

const stamp = new Date().toISOString().replace(/[:.]/g, "-");
const harPath = resolve(capturesDir, `${stamp}.har`);

const INTERESTING_HOSTS = [
  "clob.polymarket.com",
  "gamma-api.polymarket.com",
  "relayer-v2.polymarket.com",
  "data-api.polymarket.com",
  "geo.polymarket.com",
  "magic.link",
  "privy.io",
  "walletconnect.com",
  "walletconnect.org",
];

function isInteresting(url) {
  return INTERESTING_HOSTS.some((h) => url.includes(h));
}

const headless = process.env.HEADLESS === "1";
const browser = await chromium.launch({ headless, devtools: false });
const context = await browser.newContext({
  recordHar: { path: harPath, mode: "full", content: "embed" },
  viewport: { width: 1280, height: 900 },
});
const page = await context.newPage();

page.on("request", (req) => {
  const url = req.url();
  if (!isInteresting(url)) return;
  console.log(`\n>>> ${req.method()} ${url}`);
  const headers = req.headers();
  for (const k of Object.keys(headers).sort()) {
    const lower = k.toLowerCase();
    if (
      lower.startsWith("poly_") ||
      lower === "authorization" ||
      lower === "cookie" ||
      lower === "x-amz-cf-id" ||
      lower === "user-agent"
    ) {
      console.log(`    ${k}: ${headers[k]}`);
    }
  }
  const body = req.postData();
  if (body) console.log(`    body: ${body}`);
});

page.on("response", async (res) => {
  const url = res.url();
  if (!isInteresting(url)) return;
  const status = res.status();
  const isAuthHost =
    url.includes("magic.link") ||
    url.includes("privy.io") ||
    url.includes("walletconnect");
  const isInterestingPath =
    url.includes("/auth/") ||
    url.includes("/builder") ||
    url.includes("/api-key") ||
    url.includes("/profile") ||
    url.includes("/sign-up") ||
    url.includes("/onboard") ||
    url.includes("/wallet-create") ||
    url.includes("/relayer");
  if (isInterestingPath || isAuthHost || status >= 400) {
    console.log(`<<< ${status} ${url}`);
    try {
      const text = await res.text();
      if (text && text.length < 4000) console.log(`    response: ${text}`);
    } catch {
      /* ignored — body already consumed */
    }
  }
});

await page.goto("https://polymarket.com");

console.log(`\nHAR will be written to: ${harPath}`);
console.log(`Mode: ${headless ? "headless (no UI)" : "headed"}`);
if (process.env.DISPLAY) console.log(`DISPLAY: ${process.env.DISPLAY}`);
console.log(
  "Browser open. Sign up / log in manually, then trigger an API-key creation",
);
console.log(
  "(usually first order placement or 'Trade'). Ctrl+C when you've captured enough.\n",
);

const stop = new Promise((resolvePromise) => {
  process.once("SIGINT", resolvePromise);
  process.once("SIGTERM", resolvePromise);
});
await stop;

console.log("\nFinalizing HAR...");
await context.close();
await browser.close();
console.log(`HAR saved: ${harPath}`);
