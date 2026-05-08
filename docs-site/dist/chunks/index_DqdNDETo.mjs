import { l as createVNode, h as Fragment, _ as __astro_tag_component__ } from './astro/server_DO_nUfqZ.mjs';
import 'clsx';

const frontmatter = {
  "title": "Polygolem",
  "description": "Safe Polymarket SDK and CLI for Go — deposit wallet (POLY_1271) only"
};
function getHeadings() {
  return [{
    "depth": 2,
    "slug": "what-can-you-do",
    "text": "What can you do?"
  }, {
    "depth": 2,
    "slug": "no-credentials-needed",
    "text": "No credentials needed"
  }, {
    "depth": 2,
    "slug": "ready-to-go-deeper",
    "text": "Ready to go deeper?"
  }];
}
function _createMdxContent(props) {
  const {Card, CardGrid, Fragment: Fragment$1} = props.components || ({});
  if (!Card) _missingMdxReference("Card");
  if (!CardGrid) _missingMdxReference("CardGrid");
  if (!Fragment$1) _missingMdxReference("Fragment");
  return createVNode(Fragment, {
    children: [createVNode(Fragment$1, {
      "set:html": "<p>Polygolem is the single source of truth for <strong>Polymarket protocol access</strong> in Go.\nRead-only by default. No credentials needed to start.</p>\n<p><strong>Trading requires a deposit wallet (type 3 / POLY_1271).</strong> EOA, proxy, and Safe\nare blocked for new API users by Polymarket’s CLOB V2. Polygolem handles the\nfull deposit wallet lifecycle: derive, deploy via relayer, fund, approve, trade.</p>\n<div class=\"expressive-code\"><link rel=\"stylesheet\" href=\"/_astro/ec.tm3va.css\"><script type=\"module\" src=\"/_astro/ec.8zarh.js\"></script><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">search</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--query</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#ECC48D;--1:#984E4D\">btc 5m</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--limit</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">5</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem discover search --query &#x22;btc 5m&#x22; --limit 5\"><div></div></button></div></figure></div>\n<h2 id=\"what-can-you-do\">What can you do?</h2>\n<ul>\n<li><strong>Search markets</strong> — find active prediction markets by keyword, tag, or sport</li>\n<li><strong>Read order books</strong> — get real-time bid/ask depth, midpoints, spreads</li>\n<li><strong>Check market readiness</strong> — tick sizes, fee rates, neg risk status</li>\n<li><strong>Bridge assets</strong> — check supported chains, get deposit quotes</li>\n<li><strong>Paper trade</strong> — simulate orders against live market data</li>\n<li><strong>Build bots</strong> — use the Go SDK to create automated strategies</li>\n</ul>\n<h2 id=\"no-credentials-needed\">No credentials needed</h2>\n<p>Everything is read-only by default. No API keys, no wallet, no risk. Just market data.</p>\n<div class=\"expressive-code\"><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#919F9F;--1:#5F636F\"># Check API health</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">health</span></div></div><div class=\"ec-line\"><div class=\"code\">\n</div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#919F9F;--1:#5F636F\"># Search markets</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">search</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--query</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#ECC48D;--1:#984E4D\">btc 5m</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span></div></div><div class=\"ec-line\"><div class=\"code\">\n</div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#919F9F;--1:#5F636F\"># Get enriched market data (Gamma + CLOB merged)</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">enrich</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--id</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#ECC48D;--1:#984E4D\">0x...</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem healthpolygolem discover search --query &#x22;btc 5m&#x22;polygolem discover enrich --id &#x22;0x...&#x22;\"><div></div></button></div></figure></div>\n<h2 id=\"ready-to-go-deeper\">Ready to go deeper?</h2>\n"
    }), createVNode(CardGrid, {
      children: [createVNode(Card, {
        title: "Installation",
        href: "/getting-started/installation"
      }), createVNode(Card, {
        title: "Quick Start",
        href: "/getting-started/quickstart"
      }), createVNode(Card, {
        title: "Deposit Wallet Lifecycle",
        href: "/guides/deposit-wallet-lifecycle"
      }), createVNode(Card, {
        title: "Signature Types",
        href: "/concepts/deposit-wallets"
      })]
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

const url = "src/content/docs/index.mdx";
const file = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/index.mdx";
const Content = (props = {}) => MDXContent({
  ...props,
  components: { Fragment: Fragment, ...props.components, },
});
Content[Symbol.for('mdx-component')] = true;
Content[Symbol.for('astro.needsHeadRendering')] = !Boolean(frontmatter.layout);
Content.moduleId = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/index.mdx";
__astro_tag_component__(Content, 'astro:jsx');

export { Content, Content as default, file, frontmatter, getHeadings, url };
