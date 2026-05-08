import { l as createVNode, h as Fragment, _ as __astro_tag_component__ } from './astro/server_DO_nUfqZ.mjs';
import 'clsx';

const frontmatter = {
  "title": "Coverage Matrix",
  "description": "Polygolem coverage across Gamma, CLOB, Data API, Bridge, WebSocket, and Polygon deposit-wallet actions."
};
function getHeadings() {
  return [{
    "depth": 2,
    "slug": "matrix",
    "text": "Matrix"
  }, {
    "depth": 2,
    "slug": "gaps",
    "text": "Gaps"
  }];
}
function _createMdxContent(props) {
  const {Fragment} = props.components || ({});
  if (!Fragment) _missingMdxReference("Fragment");
  return createVNode(Fragment, {
    "set:html": "<h2 id=\"matrix\">Matrix</h2>\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n<table><thead><tr><th>Surface</th><th>SDK</th><th>CLI</th><th>Status</th></tr></thead><tbody><tr><td>Gamma markets</td><td><code dir=\"auto\">internal/gamma</code>, <code dir=\"auto\">pkg/universal</code></td><td><code dir=\"auto\">discover search</code>, <code dir=\"auto\">discover markets</code>, <code dir=\"auto\">discover market</code>, <code dir=\"auto\">discover enrich</code></td><td>Covered</td></tr><tr><td>Gamma taxonomy</td><td><code dir=\"auto\">internal/gamma</code>, <code dir=\"auto\">pkg/universal</code></td><td><code dir=\"auto\">discover tags</code>, <code dir=\"auto\">discover series</code>, <code dir=\"auto\">discover comments</code></td><td>Covered</td></tr><tr><td>CLOB public data</td><td><code dir=\"auto\">internal/clob</code>, <code dir=\"auto\">pkg/bookreader</code>, <code dir=\"auto\">pkg/universal</code></td><td><code dir=\"auto\">orderbook *</code>, <code dir=\"auto\">clob book</code>, <code dir=\"auto\">clob market</code>, <code dir=\"auto\">clob markets</code>, <code dir=\"auto\">clob price-history</code></td><td>Covered</td></tr><tr><td>CLOB account reads</td><td><code dir=\"auto\">internal/clob</code>, <code dir=\"auto\">pkg/universal</code></td><td><code dir=\"auto\">clob create-api-key</code>, <code dir=\"auto\">clob balance</code>, <code dir=\"auto\">clob update-balance</code>, <code dir=\"auto\">clob orders</code>, <code dir=\"auto\">clob order</code>, <code dir=\"auto\">clob trades</code></td><td>Covered</td></tr><tr><td>CLOB cancellation</td><td><code dir=\"auto\">internal/clob</code>, <code dir=\"auto\">pkg/universal</code></td><td><code dir=\"auto\">clob cancel</code>, <code dir=\"auto\">clob cancel-orders</code>, <code dir=\"auto\">clob cancel-market</code>, <code dir=\"auto\">clob cancel-all</code></td><td>Covered</td></tr><tr><td>CLOB placement</td><td><code dir=\"auto\">internal/clob</code>, <code dir=\"auto\">pkg/universal</code></td><td><code dir=\"auto\">clob create-order</code>, <code dir=\"auto\">clob market-order</code></td><td>Deposit wallet only</td></tr><tr><td>Data API</td><td><code dir=\"auto\">internal/dataapi</code>, <code dir=\"auto\">pkg/universal</code></td><td><code dir=\"auto\">data *</code></td><td>Covered</td></tr><tr><td>Bridge</td><td><code dir=\"auto\">pkg/bridge</code></td><td><code dir=\"auto\">bridge assets</code>, <code dir=\"auto\">bridge deposit</code></td><td>Covered</td></tr><tr><td>Public WebSocket</td><td><code dir=\"auto\">internal/stream</code>, <code dir=\"auto\">pkg/universal</code></td><td><code dir=\"auto\">stream market</code></td><td>Covered</td></tr><tr><td>User WebSocket</td><td>Planned</td><td>Planned</td><td>Gap</td></tr><tr><td>Deposit wallet</td><td><code dir=\"auto\">internal/auth</code>, <code dir=\"auto\">internal/relayer</code>, <code dir=\"auto\">internal/rpc</code></td><td><code dir=\"auto\">deposit-wallet *</code></td><td>Covered</td></tr></tbody></table>\n<p>The repository copy with test and documentation columns lives in\n<a href=\"https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/POLYMARKET-COVERAGE-MATRIX.md\"><code dir=\"auto\">docs/POLYMARKET-COVERAGE-MATRIX.md</code></a>.</p>\n<h2 id=\"gaps\">Gaps</h2>\n<ul>\n<li>Authenticated user WebSocket streams wait for L2 WebSocket auth tests.</li>\n<li>Data API all-market open interest is not exposed by the CLI yet; the current\ncommand requires <code dir=\"auto\">--token-id</code>.</li>\n<li>The command reference is manually aligned with the CLI tree. A checked-in\ngenerator is still a documentation tooling gap.</li>\n</ul>"
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
const url = "src/content/docs/reference/coverage-matrix.mdx";
const file = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/reference/coverage-matrix.mdx";
const Content = (props = {}) => MDXContent({
  ...props,
  components: { Fragment: Fragment, ...props.components, },
});
Content[Symbol.for('mdx-component')] = true;
Content[Symbol.for('astro.needsHeadRendering')] = !Boolean(frontmatter.layout);
Content.moduleId = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/reference/coverage-matrix.mdx";
__astro_tag_component__(Content, 'astro:jsx');

export { Content, Content as default, file, frontmatter, getHeadings, url };
