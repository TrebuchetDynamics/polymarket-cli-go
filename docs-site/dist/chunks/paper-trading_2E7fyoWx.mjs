import { l as createVNode, h as Fragment, _ as __astro_tag_component__ } from './astro/server_DO_nUfqZ.mjs';
import { c as $$Aside } from './Code_BF6vLxYs.mjs';
import 'clsx';

const frontmatter = {
  "title": "Paper Trading",
  "description": "Simulate Polymarket orders against live data with zero risk — paper state is local, never reaches authenticated endpoints."
};
function getHeadings() {
  return [{
    "depth": 2,
    "slug": "how-paper-state-works",
    "text": "How paper state works"
  }, {
    "depth": 2,
    "slug": "a-typical-workflow",
    "text": "A typical workflow"
  }, {
    "depth": 3,
    "slug": "1-find-a-market",
    "text": "1. Find a market"
  }, {
    "depth": 3,
    "slug": "2-place-a-paper-buy",
    "text": "2. Place a paper buy"
  }, {
    "depth": 3,
    "slug": "3-inspect-positions",
    "text": "3. Inspect positions"
  }, {
    "depth": 3,
    "slug": "4-close-it-out-with-a-paper-sell",
    "text": "4. Close it out with a paper sell"
  }, {
    "depth": 3,
    "slug": "5-reset-between-experiments",
    "text": "5. Reset between experiments"
  }, {
    "depth": 2,
    "slug": "what-paper-mode-does-not-do",
    "text": "What paper mode does not do"
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
      "set:html": "<p>Paper mode lets you practice trading workflows against <strong>real, live order books</strong>\nwithout ever touching your wallet, the CLOB authenticated endpoints, or any\non-chain operation.</p>\n"
    }), createVNode($$Aside, {
      type: "caution",
      "set:html": "<p>Paper mode is local-only. Fills are simulated against the current top-of-book\nquote — there is no matching engine, no slippage model, no funding cost. Use it\nfor workflow practice, not for strategy backtesting.</p>"
    }), "\n", createVNode(Fragment$1, {
      "set:html": "<h2 id=\"how-paper-state-works\">How paper state works</h2>\n<ul>\n<li>All paper orders, fills, and positions are stored in a local file under\nyour config directory.</li>\n<li>No HTTP requests with credentials are ever made.</li>\n<li>A failing safety gate cannot accidentally promote a paper command to live —\nthe paper code path simply has no signer wired in.</li>\n</ul>\n<p>See <a href=\"/concepts/safety\">Safety Model</a> for the full guarantee.</p>\n<h2 id=\"a-typical-workflow\">A typical workflow</h2>\n<h3 id=\"1-find-a-market\">1. Find a market</h3>\n<div class=\"expressive-code\"><link rel=\"stylesheet\" href=\"/_astro/ec.tm3va.css\"><script type=\"module\" src=\"/_astro/ec.8zarh.js\"></script><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">discover</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">search</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--query</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#ECC48D;--1:#984E4D\">btc 5m</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--limit</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">5</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem discover search --query &#x22;btc 5m&#x22; --limit 5\"><div></div></button></div></figure></div>\n<p>Note the <code dir=\"auto\">tokenID</code> of the outcome you want to bet on. For neg-risk Yes/No\nmarkets there are two token IDs (one per outcome); pick whichever side you want.</p>\n<h3 id=\"2-place-a-paper-buy\">2. Place a paper buy</h3>\n<div class=\"expressive-code\"><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">paper</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">buy</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">\\</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--token-id</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">71321045679252212594626385532706912750332728571942134274583709853755746957851</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">\\</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--price</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">0.52</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">\\</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--size</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">10</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem paper buy \\  --token-id 71321045679252212594626385532706912750332728571942134274583709853755746957851 \\  --price 0.52 \\  --size 10\"><div></div></button></div></figure></div>\n<p>Output:</p>\n<div class=\"expressive-code\"><figure class=\"frame not-content\"><figcaption class=\"header\"></figcaption><pre data-language=\"json\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#D6DEEB;--1:#403F53\">{</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#7FDBCA;--1:#096E72\">\"order_id\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">: </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#C789D6;--1:#7C5686\">paper-1746625200-abc</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">,</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#7FDBCA;--1:#096E72\">\"status\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">: </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#C789D6;--1:#7C5686\">filled</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">,</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#7FDBCA;--1:#096E72\">\"filled_price\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">: </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#C789D6;--1:#7C5686\">0.52</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">,</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#7FDBCA;--1:#096E72\">\"filled_size\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">: </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#C789D6;--1:#7C5686\">10</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#D6DEEB;--1:#403F53\">}</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"{  &#x22;order_id&#x22;: &#x22;paper-1746625200-abc&#x22;,  &#x22;status&#x22;: &#x22;filled&#x22;,  &#x22;filled_price&#x22;: &#x22;0.52&#x22;,  &#x22;filled_size&#x22;: &#x22;10&#x22;}\"><div></div></button></div></figure></div>\n<h3 id=\"3-inspect-positions\">3. Inspect positions</h3>\n<div class=\"expressive-code\"><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">paper</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">positions</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem paper positions\"><div></div></button></div></figure></div>\n<div class=\"expressive-code\"><figure class=\"frame not-content\"><figcaption class=\"header\"></figcaption><pre data-language=\"json\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#D6DEEB;--1:#403F53\">[</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\"><span style=\"--0:#D6DEEB;--1:#403F53\">  </span></span><span style=\"--0:#D6DEEB;--1:#403F53\">{</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">    </span><span style=\"--0:#7FDBCA;--1:#096E72\">\"token_id\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">: </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#C789D6;--1:#7C5686\">713210...</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">,</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">    </span><span style=\"--0:#7FDBCA;--1:#096E72\">\"side\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">: </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#C789D6;--1:#7C5686\">BUY</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">,</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">    </span><span style=\"--0:#7FDBCA;--1:#096E72\">\"avg_price\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">: </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#C789D6;--1:#7C5686\">0.52</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">,</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">    </span><span style=\"--0:#7FDBCA;--1:#096E72\">\"size\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">: </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#C789D6;--1:#7C5686\">10</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">,</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">    </span><span style=\"--0:#7FDBCA;--1:#096E72\">\"unrealized_pnl\"</span><span style=\"--0:#D6DEEB;--1:#403F53\">: </span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span><span style=\"--0:#C789D6;--1:#7C5686\">0.00</span><span style=\"--0:#D9F5DD;--1:#111111\">\"</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\"><span style=\"--0:#D6DEEB;--1:#403F53\">  </span></span><span style=\"--0:#D6DEEB;--1:#403F53\">}</span></div></div><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#D6DEEB;--1:#403F53\">]</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"[  {    &#x22;token_id&#x22;: &#x22;713210...&#x22;,    &#x22;side&#x22;: &#x22;BUY&#x22;,    &#x22;avg_price&#x22;: &#x22;0.52&#x22;,    &#x22;size&#x22;: &#x22;10&#x22;,    &#x22;unrealized_pnl&#x22;: &#x22;0.00&#x22;  }]\"><div></div></button></div></figure></div>\n<h3 id=\"4-close-it-out-with-a-paper-sell\">4. Close it out with a paper sell</h3>\n<div class=\"expressive-code\"><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">paper</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">sell</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">\\</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--token-id</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">71321045679252212594626385532706912750332728571942134274583709853755746957851</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">\\</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--price</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">0.55</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">\\</span></div></div><div class=\"ec-line\"><div class=\"code\"><span class=\"indent\">  </span><span style=\"--0:#82AAFF;--1:#3B61B0\">--size</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#F78C6C;--1:#AA0982\">10</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem paper sell \\  --token-id 71321045679252212594626385532706912750332728571942134274583709853755746957851 \\  --price 0.55 \\  --size 10\"><div></div></button></div></figure></div>\n<h3 id=\"5-reset-between-experiments\">5. Reset between experiments</h3>\n<p>Wipes all paper state — positions, fills, history:</p>\n<div class=\"expressive-code\"><figure class=\"frame is-terminal not-content\"><figcaption class=\"header\"><span class=\"title\"></span><span class=\"sr-only\">Terminal window</span></figcaption><pre data-language=\"bash\"><code><div class=\"ec-line\"><div class=\"code\"><span style=\"--0:#82AAFF;--1:#3B61B0\">polygolem</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">paper</span><span style=\"--0:#D6DEEB;--1:#403F53\"> </span><span style=\"--0:#ECC48D;--1:#3B61B0\">reset</span></div></div></code></pre><div class=\"copy\"><button title=\"Copy to clipboard\" data-copied=\"Copied!\" data-code=\"polygolem paper reset\"><div></div></button></div></figure></div>\n<h2 id=\"what-paper-mode-does-not-do\">What paper mode does <em>not</em> do</h2>\n<ul>\n<li>It does not call <code dir=\"auto\">clob create-order</code>, <code dir=\"auto\">clob cancel-order</code>, or any L2 endpoint.</li>\n<li>It does not require, read, or accept <code dir=\"auto\">POLYMARKET_PRIVATE_KEY</code>.</li>\n<li>It does not trigger preflight, builder-API checks, or deposit-wallet flows.</li>\n<li>Paper PnL is not tax-reportable or reconciled against any external system.</li>\n</ul>\n<h2 id=\"reference\">Reference</h2>\n<ul>\n<li><code dir=\"auto\">polygolem paper --help</code> — every flag</li>\n<li><a href=\"https://github.com/TrebuchetDynamics/polygolem/blob/main/docs/COMMANDS.md\"><code dir=\"auto\">docs/COMMANDS.md</code></a> — canonical command list</li>\n<li><a href=\"/concepts/safety\">Safety Model</a> — paper vs read-only vs live</li>\n</ul>"
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

const url = "src/content/docs/guides/paper-trading.mdx";
const file = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/guides/paper-trading.mdx";
const Content = (props = {}) => MDXContent({
  ...props,
  components: { Fragment: Fragment, ...props.components, },
});
Content[Symbol.for('mdx-component')] = true;
Content[Symbol.for('astro.needsHeadRendering')] = !Boolean(frontmatter.layout);
Content.moduleId = "/home/xel/git/sages-openclaw/workspace-yunobo/polymarket-mega-bot/go-bot/polygolem/docs-site/src/content/docs/guides/paper-trading.mdx";
__astro_tag_component__(Content, 'astro:jsx');

export { Content, Content as default, file, frontmatter, getHeadings, url };
