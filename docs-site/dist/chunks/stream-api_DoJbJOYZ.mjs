import { l as createVNode, h as Fragment, _ as __astro_tag_component__ } from './astro/server_DO_nUfqZ.mjs';
import { c as $$Aside } from './Code_BF6vLxYs.mjs';
import 'clsx';

const frontmatter = {
  "title": "Stream API",
  "description": "Polygolem's WebSocket support for real-time CLOB market data — connection management, automatic reconnect, and message dedup."
};
function getHeadings() {
  return [{
    "depth": 2,
    "slug": "whats-implemented",
    "text": "What’s implemented"
  }, {
    "depth": 2,
    "slug": "cli-surface",
    "text": "CLI surface"
  }, {
    "depth": 2,
    "slug": "reconnect-semantics",
    "text": "Reconnect semantics"
  }, {
    "depth": 2,
    "slug": "related",
    "text": "Related"
  }];
}
function _createMdxContent(props) {
  const {Fragment: Fragment$1} = props.components || ({});
  if (!Fragment$1) _missingMdxReference("Fragment");
  return createVNode(Fragment, {
    children: [createVNode(Fragment$1, {
      "set:html": "<p>Polymarket exposes a WebSocket at\n<code dir=\"auto\">wss://ws-subscriptions-clob.polymarket.com/ws/</code>. It has two channels:</p>\n<ul>\n<li><strong>Market channel</strong> — public, no auth. Streams order book updates, price\nchanges, and trade events.</li>\n<li><strong>User channel</strong> — authenticated. Streams personal order events.</li>\n</ul>\n<p>Polygolem implements the <strong>market channel</strong> today through <code dir=\"auto\">internal/stream</code>’s\n<code dir=\"auto\">MarketClient</code>. The client handles connection lifecycle, automatic reconnection\nwith backoff, message normalization, and dedup of duplicate events that the\nupstream sometimes delivers across reconnects.</p>\n<h2 id=\"whats-implemented\">What’s implemented</h2>\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n<table><thead><tr><th>Feature</th><th>Status</th></tr></thead><tbody><tr><td>Market channel subscribe (book / price / trade)</td><td>Implemented</td></tr><tr><td>Automatic reconnect with backoff</td><td>Implemented</td></tr><tr><td>Message dedup across reconnects</td><td>Implemented</td></tr><tr><td>User channel (authenticated)</td><td><strong>Not implemented</strong></td></tr></tbody></table>\n<h2 id=\"cli-surface\">CLI surface</h2>\n<p>WebSocket subscriptions are surfaced through <code dir=\"auto\">polygolem stream market</code>.\nThe CLI emits one JSON object per message on stdout, suitable for piping into\n<code dir=\"auto\">jq</code> or a downstream consumer.</p>\n<div class=\"expressive-code\"><link rel=\"stylesheet\" href=\"/_astro/ec.tm3va.css\"><script type=\"module\" src=\"/_astro/ec.8zarh.js\"></script><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">stream</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">market</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--asset-ids</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">&#x3C;token-id-1>,&#x3C;token-id-2></span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--max-messages</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">10</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem stream market --asset-ids <token-id-1>,<token-id-2> --max-messages 10\"><div></div></button></div></figure></div>\n<p><code dir=\"auto\">--max-messages 0</code> streams until the process is interrupted. <code dir=\"auto\">--url</code> can point\nto a local WebSocket test server.</p>\n"
    }), createVNode($$Aside, {
      type: "note",
      "set:html": "<p>For latency-sensitive use cases, prefer the Go SDK directly over the CLI: the\nCLI adds a marshal/unmarshal hop that costs roughly a millisecond per message.</p>"
    }), "\n", createVNode(Fragment$1, {
      "set:html": "<h2 id=\"reconnect-semantics\">Reconnect semantics</h2>\n<p>When the upstream connection drops, the client:</p>\n<ol>\n<li>Closes any open goroutines tied to the old connection.</li>\n<li>Waits an exponentially increasing backoff (capped, with jitter).</li>\n<li>Re-establishes and re-subscribes to the same set of channels.</li>\n<li>Suppresses duplicate events for a short replay window using a content-hash\nset, so consumers don’t see the same trade twice.</li>\n</ol>\n<h2 id=\"related\">Related</h2>\n<ul>\n<li><a href=\"/guides/orderbook-data\">Orderbook Data</a> — REST equivalents for one-shot reads</li>\n<li><a href=\"https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/ARCHITECTURE.md\"><code dir=\"auto\">docs/ARCHITECTURE.md</code></a> — <code dir=\"auto\">internal/stream</code> package boundary</li>\n<li><a href=\"/concepts/polymarket-api\">Polymarket API Overview</a> — every API in one place</li>\n</ul>\n"
    }), createVNode($$Aside, {
      type: "note",
      "set:html": "<p><strong>Source of truth:</strong> the canonical package-level description of <code dir=\"auto\">internal/stream</code>\nlives in\n<a href=\"https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/ARCHITECTURE.md\"><code dir=\"auto\">docs/ARCHITECTURE.md</code></a>.</p>"
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

const url = "src/content/docs/reference/stream-api.mdx";
const file = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/reference/stream-api.mdx";
const Content = (props = {}) => MDXContent({
  ...props,
  components: { Fragment: Fragment, ...props.components, },
});
Content[Symbol.for('mdx-component')] = true;
Content[Symbol.for('astro.needsHeadRendering')] = !Boolean(frontmatter.layout);
Content.moduleId = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/reference/stream-api.mdx";
__astro_tag_component__(Content, 'astro:jsx');

export { Content, Content as default, file, frontmatter, getHeadings, url };
