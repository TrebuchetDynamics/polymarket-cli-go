import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  site: "https://polygolem.trebuchetdynamics.com",
  integrations: [
    starlight({
      title: "Polygolem",
      description: "Safe Polymarket SDK and CLI for Go — V2 deposit wallet (POLY_1271)",
      logo: {
        src: "./src/assets/logo.svg",
      },
      social: {
        github: "https://github.com/TrebuchetDynamics/polygolem",
      },
      sidebar: [
        {
          label: "Getting Started",
          items: [
            { label: "Introduction", link: "/docs/" },
            { label: "Installation", link: "/docs/getting-started/installation" },
            { label: "Quick Start", link: "/docs/getting-started/quickstart" },
          ],
        },
        {
          label: "Guides",
          items: [
            { label: "Builder & Relayer Keys", link: "/docs/guides/builder-auto" },
            { label: "Market Discovery", link: "/docs/guides/market-discovery" },
            { label: "Universal Client", link: "/docs/guides/universal-client" },
            { label: "Deposit Wallet Lifecycle", link: "/docs/guides/deposit-wallet-lifecycle" },
            { label: "Redeem Winners", link: "/docs/guides/redeem-winners" },
            { label: "Orderbook Data", link: "/docs/guides/orderbook-data" },
            { label: "Paper Trading", link: "/docs/guides/paper-trading" },
            { label: "Bridge & Funding", link: "/docs/guides/bridge-funding" },
            { label: "Go-Bot Integration", link: "/docs/guides/go-bot-integration" },
          ],
        },
        {
          label: "Concepts",
          items: [
            { label: "Polymarket API Overview", link: "/docs/concepts/polymarket-api" },
            { label: "Smart Contracts", link: "/docs/concepts/contracts" },
            { label: "Deposit Wallets (POLY_1271)", link: "/docs/concepts/deposit-wallets" },
            { label: "Secrets Management", link: "/docs/concepts/secrets-management" },
            { label: "POLY_1271 Signing Chain", link: "/docs/concepts/poly-1271-signing" },
            { label: "Markets, Events & Tokens", link: "/docs/concepts/markets-events-tokens" },
            { label: "Safety Model", link: "/docs/concepts/safety" },
            { label: "Architecture", link: "/docs/concepts/architecture" },
          ],
        },
        {
          label: "Reference",
          items: [
            { label: "CLI Commands", link: "/docs/reference/cli" },
            { label: "Go SDK Contracts", link: "/docs/reference/sdk" },
            { label: "Protocol Types", link: "/docs/reference/polytypes" },
            { label: "Internal Packages", link: "/docs/reference/internal-packages" },
            { label: "Gamma API", link: "/docs/reference/gamma-api" },
            { label: "CLOB V2 API", link: "/docs/reference/clob-api" },
            { label: "Data API", link: "/docs/reference/data-api" },
            { label: "Stream API", link: "/docs/reference/stream-api" },
            { label: "Coverage Matrix", link: "/docs/reference/coverage-matrix" },
          ],
        },
      ],
      customCss: ["./src/styles/custom.css"],
    }),
  ],
});
