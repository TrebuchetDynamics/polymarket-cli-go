import { l as createVNode, h as Fragment, _ as __astro_tag_component__ } from './astro/server_DO_nUfqZ.mjs';
import { c as $$Aside } from './Code_BF6vLxYs.mjs';
import 'clsx';

const frontmatter = {
  "title": "Safety Model",
  "description": "Read-only by default, paper-mode locality, live gates, credential redaction — the four-layer safety guarantee in polygolem."
};
function getHeadings() {
  return [{
    "depth": 2,
    "slug": "the-four-gates-in-order",
    "text": "The four gates, in order"
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
      "set:html": "<p>Polygolem is <strong>read-only by default</strong>. Every market-data and discovery command\nruns without credentials and cannot make a state-changing request. Paper mode\nis <strong>local-only</strong>: simulated orders are stored on disk and never reach an\nauthenticated endpoint. Live commands require <strong>all four gates</strong> to pass —\n<code dir=\"auto\">POLYMARKET_LIVE_PROFILE=on</code>, <code dir=\"auto\">live_trading_enabled: true</code>, <code dir=\"auto\">--confirm-live</code>,\nand a successful <code dir=\"auto\">polygolem preflight</code>. Credentials are loaded only at the\nmoment they are needed and are <strong>redacted</strong> in every log, error, and JSON\noutput.</p>\n<p>If any gate fails, the command aborts with a structured error and a non-zero\nexit code. Polygolem will never silently downgrade a live command to paper mode\nor read-only — that would hide operator intent and break automation.</p>\n"
    }), createVNode($$Aside, {
      type: "note",
      "set:html": "<p><strong>Source of truth:</strong> the canonical safety policy lives in\n<a href=\"https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/SAFETY.md\"><code dir=\"auto\">docs/SAFETY.md</code></a>.\nThis page is a summary; if anything here disagrees with the canonical doc,\nthe canonical doc wins.</p>"
    }), "\n", createVNode(Fragment$1, {
      "set:html": "<h2 id=\"the-four-gates-in-order\">The four gates, in order</h2>\n<ol>\n<li><strong><code dir=\"auto\">POLYMARKET_LIVE_PROFILE=on</code></strong> — environment opt-in.</li>\n<li><strong><code dir=\"auto\">live_trading_enabled: true</code></strong> — config-file opt-in.</li>\n<li><strong><code dir=\"auto\">--confirm-live</code></strong> — per-invocation flag.</li>\n<li><strong><code dir=\"auto\">polygolem preflight</code> passes</strong> — wallet readiness, auth readiness, network\nconsistency, API health.</li>\n</ol>\n<p>All four. No partial credit.</p>\n<h2 id=\"related\">Related</h2>\n<ul>\n<li><a href=\"https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/SAFETY.md\"><code dir=\"auto\">docs/SAFETY.md</code></a> — canonical policy</li>\n<li><a href=\"/guides/paper-trading\">Paper Trading</a> — the local-only execution surface</li>\n<li><a href=\"/concepts/architecture\">Architecture</a> — where the gates live in code</li>\n</ul>"
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

const url = "src/content/docs/concepts/safety.mdx";
const file = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/concepts/safety.mdx";
const Content = (props = {}) => MDXContent({
  ...props,
  components: { Fragment: Fragment, ...props.components, },
});
Content[Symbol.for('mdx-component')] = true;
Content[Symbol.for('astro.needsHeadRendering')] = !Boolean(frontmatter.layout);
Content.moduleId = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/concepts/safety.mdx";
__astro_tag_component__(Content, 'astro:jsx');

export { Content, Content as default, file, frontmatter, getHeadings, url };
