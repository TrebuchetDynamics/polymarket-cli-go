import { l as createVNode, h as Fragment, _ as __astro_tag_component__ } from './astro/server_DO_nUfqZ.mjs';
import { c as $$Aside } from './Code_BF6vLxYs.mjs';
import 'clsx';

const frontmatter = {
  "title": "Gamma API",
  "description": "Polymarket's read-only metadata API — markets, events, search, tags, profiles. Wrapped by polygolem's discover commands and pkg/gamma."
};
function getHeadings() {
  return [{
    "depth": 2,
    "slug": "endpoint-categories",
    "text": "Endpoint categories"
  }, {
    "depth": 2,
    "slug": "upstream",
    "text": "Upstream"
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
      "set:html": "<p>The Gamma API (<code dir=\"auto\">https://gamma-api.polymarket.com</code>) is Polymarket’s <strong>read-only\nmetadata layer</strong>. It serves markets, events, search, tags, series, sports\nmetadata, comments, and public profiles. No authentication required.</p>\n<p>Polygolem wraps Gamma in two places:</p>\n<ul>\n<li><strong><code dir=\"auto\">internal/gamma</code></strong> — full typed HTTP client, used by every CLI command.</li>\n<li><strong><code dir=\"auto\">pkg/gamma</code></strong> — the public read-only surface that downstream Go consumers\ncan import. Strict subset of <code dir=\"auto\">internal/gamma</code> with stable types.</li>\n</ul>\n<p>The <code dir=\"auto\">polygolem discover</code> family is the CLI interface to Gamma:</p>\n<div class=\"expressive-code\"><link rel=\"stylesheet\" href=\"/_astro/ec.tm3va.css\"><script type=\"module\" src=\"/_astro/ec.8zarh.js\"></script><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">search</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--query</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#ECC48D;--1:#984E4D\">btc 5m</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--limit</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">10</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">markets</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--limit</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">20</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--active</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">market</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--id</span><span style=\"--0:#D6DEEB;--1:#403F53\">    </span><span style=\"--0:#ECC48D;--1:#3B61B0\">0xbd31dc8a...</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">market</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--slug</span><span style=\"--0:#D6DEEB;--1:#403F53\">  </span><span style=\"--0:#ECC48D;--1:#3B61B0\">will-btc-be-above</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">enrich</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--id</span><span style=\"--0:#D6DEEB;--1:#403F53\">    </span><span style=\"--0:#ECC48D;--1:#3B61B0\">0xbd31dc8a...</span><span style=\"--0:#D6DEEB;--1:#403F53\">   </span><span style=\"--0:#919F9F;--1:#5F636F\"># joins Gamma + CLOB</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">tags</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--limit</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">100</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">series</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--limit</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">20</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">comments</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--entity-id</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">123</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--entity-type</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">market</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem discover search --query &#x22;btc 5m&#x22; --limit 10polygolem discover markets --limit 20 --activepolygolem discover market --id    0xbd31dc8a...polygolem discover market --slug  will-btc-be-abovepolygolem discover enrich --id    0xbd31dc8a...   # joins Gamma + CLOBpolygolem discover tags --limit 100polygolem discover series --limit 20polygolem discover comments --entity-id 123 --entity-type market\"><div></div></button></div></figure></div>\n<h2 id=\"endpoint-categories\">Endpoint categories</h2>\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n<table><thead><tr><th>Category</th><th>Endpoints</th></tr></thead><tbody><tr><td>Markets</td><td>List, by ID, by slug, by token ID</td></tr><tr><td>Events</td><td>List, by ID, by slug, keyset pagination</td></tr><tr><td>Search</td><td>Cross-entity (markets, events, profiles, tags)</td></tr><tr><td>Tags</td><td>List, by ID, by slug, related</td></tr><tr><td>Series</td><td>List, by ID</td></tr><tr><td>Sports</td><td>Teams, market types</td></tr><tr><td>Comments</td><td>List, by ID, by user</td></tr><tr><td>Profiles</td><td>Public profile by wallet</td></tr></tbody></table>\n<h2 id=\"upstream\">Upstream</h2>\n<ul>\n<li>Polymarket public docs: <a href=\"https://docs.polymarket.com/\">docs.polymarket.com</a></li>\n<li>Base URL: <code dir=\"auto\">https://gamma-api.polymarket.com</code></li>\n</ul>\n"
    }), createVNode($$Aside, {
      type: "note",
      "set:html": "<p><strong>Source of truth</strong> for <code dir=\"auto\">polygolem discover</code> flag semantics is\n<a href=\"https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/COMMANDS.md\"><code dir=\"auto\">docs/COMMANDS.md</code></a>\n(see the <code dir=\"auto\">discover</code> section). Run <code dir=\"auto\">polygolem discover --help</code> for live help.</p>"
    }), "\n", createVNode(Fragment$1, {
      "set:html": "<h2 id=\"related\">Related</h2>\n<ul>\n<li><a href=\"/guides/market-discovery\">Market Discovery</a> — task-oriented walkthrough</li>\n<li><a href=\"/concepts/markets-events-tokens\">Markets, Events &#x26; Tokens</a> — data model</li>\n<li><a href=\"/concepts/polymarket-api\">Polymarket API Overview</a> — every API in one place</li>\n</ul>"
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

const url = "src/content/docs/reference/gamma-api.mdx";
const file = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/reference/gamma-api.mdx";
const Content = (props = {}) => MDXContent({
  ...props,
  components: { Fragment: Fragment, ...props.components, },
});
Content[Symbol.for('mdx-component')] = true;
Content[Symbol.for('astro.needsHeadRendering')] = !Boolean(frontmatter.layout);
Content.moduleId = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/reference/gamma-api.mdx";
__astro_tag_component__(Content, 'astro:jsx');

export { Content, Content as default, file, frontmatter, getHeadings, url };
