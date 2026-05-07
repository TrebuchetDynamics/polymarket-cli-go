import { l as createVNode, h as Fragment, _ as __astro_tag_component__ } from './astro/server_DO_nUfqZ.mjs';
import 'clsx';

const frontmatter = {
  "title": "Polymarket API Overview"
};
function getHeadings() {
  return [{
    "depth": 2,
    "slug": "gamma-api-gamma-apipolymarketcom",
    "text": "Gamma API (gamma-api.polymarket.com)"
  }, {
    "depth": 2,
    "slug": "clob-api-clobpolymarketcom",
    "text": "CLOB API (clob.polymarket.com)"
  }, {
    "depth": 2,
    "slug": "data-api-data-apipolymarketcom",
    "text": "Data API (data-api.polymarket.com)"
  }, {
    "depth": 2,
    "slug": "bridge-api-bridgepolymarketcom",
    "text": "Bridge API (bridge.polymarket.com)"
  }, {
    "depth": 2,
    "slug": "websocket-ws-subscriptions-clobpolymarketcom",
    "text": "WebSocket (ws-subscriptions-clob.polymarket.com)"
  }];
}
function _createMdxContent(props) {
  const {Fragment} = props.components || ({});
  if (!Fragment) _missingMdxReference("Fragment");
  return createVNode(Fragment, {
    "set:html": "<p>Understanding the Polymarket ecosystem APIs that polygolem wraps.</p>\n<h2 id=\"gamma-api-gamma-apipolymarketcom\">Gamma API (<code dir=\"auto\">gamma-api.polymarket.com</code>)</h2>\n<p><strong>Read-only. No authentication.</strong></p>\n<p>Market metadata, events, search, tags, series, sports, comments, profiles.</p>\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n<table><thead><tr><th>Category</th><th>Endpoints</th></tr></thead><tbody><tr><td>Markets</td><td>List, by ID, by slug, by token</td></tr><tr><td>Events</td><td>List, by ID, by slug, keyset pagination</td></tr><tr><td>Search</td><td>Cross-entity (markets, events, profiles, tags)</td></tr><tr><td>Tags</td><td>List, by ID, by slug, related</td></tr><tr><td>Series</td><td>List, by ID</td></tr><tr><td>Sports</td><td>Teams, sports metadata, market types</td></tr><tr><td>Comments</td><td>List, by ID, by user</td></tr><tr><td>Profiles</td><td>Public profile by wallet</td></tr></tbody></table>\n<h2 id=\"clob-api-clobpolymarketcom\">CLOB API (<code dir=\"auto\">clob.polymarket.com</code>)</h2>\n<p><strong>Public endpoints (L0): no auth. Authenticated endpoints (L1/L2): require wallet + API key.</strong></p>\n<p>Order book, pricing, market data, orders, trades, rewards.</p>\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n<table><thead><tr><th>Category</th><th>Auth</th><th>Endpoints</th></tr></thead><tbody><tr><td>Market Data</td><td>L0</td><td>Order book, price, midpoint, spread, tick size, fee rate, neg risk, last trade</td></tr><tr><td>Markets</td><td>L0</td><td>List, by condition ID, simplified, sampling</td></tr><tr><td>Orders</td><td>L2</td><td>Place, cancel, query, heartbeat</td></tr><tr><td>Trades</td><td>L2</td><td>Get trades, builder trades</td></tr><tr><td>Rewards</td><td>L2</td><td>Config, earnings, percentages, rebates</td></tr><tr><td>Scoring</td><td>L2</td><td>Order scoring status</td></tr></tbody></table>\n<h2 id=\"data-api-data-apipolymarketcom\">Data API (<code dir=\"auto\">data-api.polymarket.com</code>)</h2>\n<p><strong>Read-only analytics. No auth for most endpoints.</strong></p>\n<p>Positions, volume, leaderboards, open interest.</p>\n<h2 id=\"bridge-api-bridgepolymarketcom\">Bridge API (<code dir=\"auto\">bridge.polymarket.com</code>)</h2>\n<p><strong>Read-only. No auth.</strong></p>\n<p>Supported assets, deposit addresses, quotes, transaction status.</p>\n<h2 id=\"websocket-ws-subscriptions-clobpolymarketcom\">WebSocket (<code dir=\"auto\">ws-subscriptions-clob.polymarket.com</code>)</h2>\n<p><strong>Market channel: no auth. User channel: L2 auth.</strong></p>\n<p>Real-time stream for order books, prices, trades, order events.</p>"
  });
}
function MDXContent(props = {}) {
  const {wrapper: MDXLayout} = props.components || ({});
  return MDXLayout ? createVNode(MDXLayout, {
    ...props,
    children: createVNode(_createMdxContent, {
      ...props
    })
  }) : _createMdxContent(props);
}
function _missingMdxReference(id, component) {
  throw new Error("Expected " + ("component" ) + " `" + id + "` to be defined: you likely forgot to import, pass, or provide it.");
}
const url = "src/content/docs/concepts/polymarket-api.mdx";
const file = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/concepts/polymarket-api.mdx";
const Content = (props = {}) => MDXContent({
  ...props,
  components: { Fragment: Fragment, ...props.components, },
});
Content[Symbol.for('mdx-component')] = true;
Content[Symbol.for('astro.needsHeadRendering')] = !Boolean(frontmatter.layout);
Content.moduleId = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/concepts/polymarket-api.mdx";
__astro_tag_component__(Content, 'astro:jsx');

export { Content, Content as default, file, frontmatter, getHeadings, url };
