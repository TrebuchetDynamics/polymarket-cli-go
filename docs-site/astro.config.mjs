import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
  site: "https://trebuchetdynamics.github.io/polygolem",
  integrations: [
    starlight({
      title: "Polygolem",
      description: "Safe Polymarket SDK and CLI for Go",
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
            { label: "Introduction", slug: "" },
            { label: "Installation", slug: "getting-started/installation" },
            { label: "Quick Start", slug: "getting-started/quickstart" },
          ],
        },
        {
          label: "Guides",
          items: [
            { label: "Market Discovery", slug: "guides/market-discovery" },
            { label: "Orderbook Data", slug: "guides/orderbook-data" },
            { label: "Paper Trading", slug: "guides/paper-trading" },
            { label: "Bridge & Funding", slug: "guides/bridge-funding" },
            { label: "Go-Bot Integration", slug: "guides/go-bot-integration" },
          ],
        },
        {
          label: "Concepts",
          items: [
            { label: "Polymarket API Overview", slug: "concepts/polymarket-api" },
            { label: "Markets, Events & Tokens", slug: "concepts/markets-events-tokens" },
            { label: "Safety Model", slug: "concepts/safety" },
            { label: "Architecture", slug: "concepts/architecture" },
          ],
        },
        {
          label: "Reference",
          items: [
            { label: "CLI Commands", slug: "reference/cli" },
            { label: "Go SDK", slug: "reference/sdk" },
            { label: "Gamma API", slug: "reference/gamma-api" },
            { label: "CLOB API", slug: "reference/clob-api" },
            { label: "Stream API", slug: "reference/stream-api" },
          ],
        },
      ],
      customCss: ["./src/styles/custom.css"],
    }),
  ],
});
