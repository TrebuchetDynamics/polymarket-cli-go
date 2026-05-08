import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  site: "https://trebuchetdynamics.github.io/polygolem",
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
            { label: "Introduction", link: "/" },
            { label: "Installation", link: "/getting-started/installation" },
            { label: "Quick Start", link: "/getting-started/quickstart" },
          ],
        },
        {
          label: "Guides",
          items: [
            { label: "Builder Auto (Zero-Browser)", link: "/guides/builder-auto" },
            { label: "Market Discovery", link: "/guides/market-discovery" },
            { label: "Universal Client", link: "/guides/universal-client" },
            { label: "Deposit Wallet Lifecycle", link: "/guides/deposit-wallet-lifecycle" },
            { label: "Orderbook Data", link: "/guides/orderbook-data" },
            { label: "Paper Trading", link: "/guides/paper-trading" },
            { label: "Bridge & Funding", link: "/guides/bridge-funding" },
            { label: "Go-Bot Integration", link: "/guides/go-bot-integration" },
          ],
        },
        {
          label: "Concepts",
          items: [
            { label: "Polymarket API Overview", link: "/concepts/polymarket-api" },
            { label: "Smart Contracts", link: "/concepts/contracts" },
            { label: "Deposit Wallets (POLY_1271)", link: "/concepts/deposit-wallets" },
            { label: "Secrets Management", link: "/concepts/secrets-management" },
            { label: "Markets, Events & Tokens", link: "/concepts/markets-events-tokens" },
            { label: "Safety Model", link: "/concepts/safety" },
            { label: "Architecture", link: "/concepts/architecture" },
          ],
        },
        {
          label: "Reference",
          items: [
            { label: "CLI Commands", link: "/reference/cli" },
            { label: "Go SDK Contracts", link: "/reference/sdk" },
            { label: "Protocol Types", link: "/reference/polytypes" },
            { label: "Internal Packages", link: "/reference/internal-packages" },
            { label: "Gamma API", link: "/reference/gamma-api" },
            { label: "CLOB V2 API", link: "/reference/clob-api" },
            { label: "Data API", link: "/reference/data-api" },
            { label: "Stream API", link: "/reference/stream-api" },
            { label: "Coverage Matrix", link: "/reference/coverage-matrix" },
          ],
        },
      ],
      customCss: ["./src/styles/custom.css"],
    }),
  ],
});
