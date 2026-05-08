import { l as createVNode, h as Fragment, _ as __astro_tag_component__ } from './astro/server_DO_nUfqZ.mjs';
import { c as $$Aside } from './Code_BF6vLxYs.mjs';
import 'clsx';

const frontmatter = {
  "title": "Architecture",
  "description": "The polygolem package layout — public SDK under pkg/, implementation under internal/, CLI as a thin Cobra shell."
};
function getHeadings() {
  return [{
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
      "set:html": "<p>Polygolem is a Go protocol and automation stack for Polymarket with a thin\nCobra-based CLI on top. The codebase is split into three layers:</p>\n<ul>\n<li><strong><code dir=\"auto\">cmd/polygolem</code></strong> — the binary entry point. Just wires Cobra and exits.</li>\n<li><strong><code dir=\"auto\">internal/</code> (21 packages)</strong> — implementation: HTTP clients, signers,\nconfig, paper executor, modes, output, errors. Not part of the public API.</li>\n<li><strong><code dir=\"auto\">pkg/</code> (5 packages)</strong> — the stable public Go SDK: <code dir=\"auto\">bookreader</code>, <code dir=\"auto\">bridge</code>,\n<code dir=\"auto\">gamma</code>, <code dir=\"auto\">marketresolver</code>, <code dir=\"auto\">pagination</code>. Semver-stable.</li>\n</ul>\n<p>Dependencies flow <strong>inward only</strong>: <code dir=\"auto\">cmd → internal → pkg</code>. Public SDK packages\nmust not import internal packages, and internal packages must not depend on\nthe CLI layer.</p>\n<p>The mode boundary (read-only / paper / live) is enforced in <code dir=\"auto\">internal/modes</code>\nand consumed by every command before any side-effecting code path. Live\ncommands additionally route through <code dir=\"auto\">internal/execution</code> after the four-gate\npreflight check.</p>\n"
    }), createVNode($$Aside, {
      type: "note",
      "set:html": "<p><strong>Source of truth:</strong> the canonical architecture document is\n<a href=\"https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/ARCHITECTURE.md\"><code dir=\"auto\">docs/ARCHITECTURE.md</code></a>.\nIt contains the full per-package table and dependency graph; this page is a\nhigh-level summary.</p>"
    }), "\n", createVNode(Fragment$1, {
      "set:html": "<h2 id=\"related\">Related</h2>\n<ul>\n<li><a href=\"https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/ARCHITECTURE.md\"><code dir=\"auto\">docs/ARCHITECTURE.md</code></a> — canonical package map</li>\n<li><a href=\"/reference/sdk\">Go SDK Reference</a> — every public type and method</li>\n<li><a href=\"/concepts/safety\">Safety Model</a> — how the mode boundary is enforced</li>\n</ul>"
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

const url = "src/content/docs/concepts/architecture.mdx";
const file = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/concepts/architecture.mdx";
const Content = (props = {}) => MDXContent({
  ...props,
  components: { Fragment: Fragment, ...props.components, },
});
Content[Symbol.for('mdx-component')] = true;
Content[Symbol.for('astro.needsHeadRendering')] = !Boolean(frontmatter.layout);
Content.moduleId = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/concepts/architecture.mdx";
__astro_tag_component__(Content, 'astro:jsx');

export { Content, Content as default, file, frontmatter, getHeadings, url };
