import { l as createVNode, h as Fragment, _ as __astro_tag_component__ } from './astro/server_DO_nUfqZ.mjs';
import { c as $$Aside } from './Code_BF6vLxYs.mjs';
import 'clsx';

const frontmatter = {
  "title": "Markets, Events & Tokens",
  "description": "How Polymarket models prediction markets — events, markets, conditions, outcomes, and ERC-1155 token IDs."
};
function getHeadings() {
  return [{
    "depth": 2,
    "slug": "the-four-entities",
    "text": "The four entities"
  }, {
    "depth": 2,
    "slug": "a-concrete-example",
    "text": "A concrete example"
  }, {
    "depth": 2,
    "slug": "yesno-vs-multi-outcome",
    "text": "Yes/No vs multi-outcome"
  }, {
    "depth": 2,
    "slug": "gamma-view-vs-clob-view",
    "text": "Gamma view vs CLOB view"
  }, {
    "depth": 2,
    "slug": "token-ids-are-the-trading-primitive",
    "text": "Token IDs are the trading primitive"
  }, {
    "depth": 2,
    "slug": "how-polygolem-maps-it",
    "text": "How polygolem maps it"
  }, {
    "depth": 2,
    "slug": "reference",
    "text": "Reference"
  }];
}
function _createMdxContent(props) {
  const {Fragment: Fragment$1} = props.components || ({});
  if (!Fragment$1) _missingMdxReference("Fragment");
  return createVNode(Fragment, {
    children: [createVNode(Fragment$1, {
      "set:html": "<p>Polymarket’s data model has four core entities that polygolem surfaces across\nthe Gamma and CLOB clients. Understanding the shape — and the relationship\nbetween Gamma’s view and CLOB’s view of the same market — is the single\nmost important thing for anyone using this SDK.</p>\n<h2 id=\"the-four-entities\">The four entities</h2>\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n<table><thead><tr><th>Entity</th><th>Lives in</th><th>What it represents</th></tr></thead><tbody><tr><td><strong>Event</strong></td><td>Gamma</td><td>A theme or topic. Groups one or more related markets.</td></tr><tr><td><strong>Market</strong></td><td>Gamma + CLOB</td><td>A single tradable proposition, with one or more outcomes.</td></tr><tr><td><strong>Condition</strong></td><td>Conditional Tokens Framework (CTF) on-chain</td><td>The settlement primitive. Identified by a <code dir=\"auto\">conditionID</code>.</td></tr><tr><td><strong>Outcome / Token</strong></td><td>CTF on-chain</td><td>Each possible result is a separate ERC-1155 token under the condition.</td></tr></tbody></table>\n<h2 id=\"a-concrete-example\">A concrete example</h2>\n<p>Take the market <strong>“Will Bitcoin be above $70,000 at 5 PM UTC?”</strong>.</p>\n<ul>\n<li><strong>Event</strong> (Gamma): <code dir=\"auto\">\"Bitcoin price intraday\"</code> — groups multiple BTC markets\nacross different strikes and times.</li>\n<li><strong>Market</strong> (Gamma): <code dir=\"auto\">\"Will Bitcoin be above $70,000 at 5 PM UTC?\"</code> —\na single Yes/No proposition.</li>\n<li><strong>Condition</strong> (CTF): one <code dir=\"auto\">conditionID</code>, e.g. <code dir=\"auto\">0xbd31dc8a...</code>. This is the\non-chain settlement key.</li>\n<li><strong>Outcomes / Tokens</strong>: two ERC-1155 token IDs under the condition:\n<ul>\n<li><code dir=\"auto\">tokenID_yes</code> — pays $1 if BTC ≥ $70k at settlement, else $0</li>\n<li><code dir=\"auto\">tokenID_no</code>  — pays $1 if BTC &#x3C; $70k at settlement, else $0</li>\n</ul>\n</li>\n</ul>\n<p>The two outcome tokens are <strong>complementary</strong>: their prices on a working market\nsum to ~$1 minus the bid-ask spread. Buying one Yes token and one No token at\nissuance equals buying $1 of redemption certainty, which is exactly how new\nshares are minted via the CTF.</p>\n<h2 id=\"yesno-vs-multi-outcome\">Yes/No vs multi-outcome</h2>\n<p>Most Polymarket markets are binary (Yes/No, two tokens). Some — like a\npresidential primary — are multi-outcome:</p>\n<ul>\n<li><strong>Outcomes</strong>: <code dir=\"auto\">[\"Trump\", \"Biden\", \"Other\"]</code></li>\n<li><strong>Tokens</strong>: three ERC-1155 token IDs, one per outcome</li>\n<li><strong>Sum constraint</strong>: prices still sum to ~$1, but split across three lines</li>\n</ul>\n<p>For multi-outcome neg-risk markets, the constraint is enforced by the neg-risk\nadapter contract rather than the bare CTF.</p>\n<h2 id=\"gamma-view-vs-clob-view\">Gamma view vs CLOB view</h2>\n<p>The same market shows up in two places, with different fields:</p>\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n<table><thead><tr><th>Aspect</th><th>Gamma</th><th>CLOB</th></tr></thead><tbody><tr><td>Identifier</td><td><code dir=\"auto\">marketID</code> (Gamma row PK), <code dir=\"auto\">slug</code></td><td><code dir=\"auto\">conditionID</code>, <code dir=\"auto\">tokenID</code></td></tr><tr><td>Best for</td><td>Discovery, search, metadata</td><td>Order book, fees, tick size, trades</td></tr><tr><td>Contains</td><td>Question, description, outcomes, end date, image</td><td>Bid/ask depth, last trade, fee rate</td></tr><tr><td>Auth</td><td>None (read-only)</td><td>None for L0; key+wallet for L1/L2</td></tr></tbody></table>\n<p><code dir=\"auto\">polygolem discover enrich --id &#x3C;conditionID></code> joins both views into a single\nJSON document — see <a href=\"/guides/market-discovery\">Market Discovery</a>.</p>\n<h2 id=\"token-ids-are-the-trading-primitive\">Token IDs are the trading primitive</h2>\n<p>Every CLOB endpoint that touches market data takes a <strong>token ID</strong> (ERC-1155\nasset ID), not a condition ID:</p>\n<div class=\"expressive-code\"><link rel=\"stylesheet\" href=\"/_astro/ec.tm3va.css\"><script type=\"module\" src=\"/_astro/ec.8zarh.js\"></script><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">orderbook</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">get</span><span style=\"--0:#D6DEEB;--1:#403F53\">      </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--token-id</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">71321045679...</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">orderbook</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">tick-size</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--token-id</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">71321045679...</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">clob</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">create-order</span><span style=\"--0:#D6DEEB;--1:#403F53\">  </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--token-id</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">71321045679...</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--side</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">BUY</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">...</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem orderbook get      --token-id 71321045679...polygolem orderbook tick-size --token-id 71321045679...polygolem clob create-order  --token-id 71321045679... --side BUY ...\"><div></div></button></div></figure></div>\n<p>The token ID is the integer encoding of the position in the CTF.</p>\n"
    }), createVNode($$Aside, {
      type: "note",
      "set:html": "<p>A common bug: passing a <code dir=\"auto\">conditionID</code> (32-byte hex starting with <code dir=\"auto\">0x</code>) where a\n<code dir=\"auto\">tokenID</code> (decimal integer, often 70+ digits) is required. Polygolem rejects\nthis at validation time, but you’ll save a round-trip if you grab the right\nidentifier from <code dir=\"auto\">discover enrich</code> first.</p>"
    }), "\n", createVNode(Fragment$1, {
      "set:html": "<h2 id=\"how-polygolem-maps-it\">How polygolem maps it</h2>\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n<table><thead><tr><th>polygolem command / package</th><th>Operates on</th><th>Notes</th></tr></thead><tbody><tr><td><code dir=\"auto\">discover search</code> / <code dir=\"auto\">discover market</code></td><td>Gamma — markets, events</td><td>Returns Gamma fields.</td></tr><tr><td><code dir=\"auto\">discover enrich</code></td><td>Gamma + CLOB</td><td>Joins by <code dir=\"auto\">conditionID</code>.</td></tr><tr><td><code dir=\"auto\">orderbook *</code> / <code dir=\"auto\">pkg/bookreader</code></td><td>CLOB — by <code dir=\"auto\">tokenID</code></td><td>Read-only L0.</td></tr><tr><td><code dir=\"auto\">clob create-order</code></td><td>CLOB — by <code dir=\"auto\">tokenID</code></td><td>Authenticated L2.</td></tr><tr><td><code dir=\"auto\">pkg/marketresolver.ResolveTokenIDs</code></td><td>Gamma → both <code dir=\"auto\">tokenID</code>s</td><td>High-level resolver.</td></tr></tbody></table>\n<h2 id=\"reference\">Reference</h2>\n<ul>\n<li><a href=\"/reference/gamma-api\">Gamma API</a> — full Gamma surface</li>\n<li><a href=\"/reference/clob-api\">CLOB API</a> — full CLOB surface</li>\n<li><a href=\"/concepts/polymarket-api\">Polymarket API Overview</a> — every API at a glance</li>\n</ul>"
    })]
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

const url = "src/content/docs/concepts/markets-events-tokens.mdx";
const file = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/concepts/markets-events-tokens.mdx";
const Content = (props = {}) => MDXContent({
  ...props,
  components: { Fragment: Fragment, ...props.components, },
});
Content[Symbol.for('mdx-component')] = true;
Content[Symbol.for('astro.needsHeadRendering')] = !Boolean(frontmatter.layout);
Content.moduleId = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/concepts/markets-events-tokens.mdx";
__astro_tag_component__(Content, 'astro:jsx');

export { Content, Content as default, file, frontmatter, getHeadings, url };
