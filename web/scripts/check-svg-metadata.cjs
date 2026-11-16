// Regression check: every SVG in src/assets/svg-icon is compiled exactly the
// way unplugin-icons does in this project — FileSystemIconLoader with the
// project transform (unplugin.ts: only the <svg open tag is rewritten), then
// @vue/compiler-sfc compileTemplate (compiler: 'vue3').
// Inkscape/editor metadata elements (inkscape:grid, sodipodi:namedview,
// rdf:rdf, ...) have colon-containing tag names that Vue compiles to
// resolveComponent(...) calls — producing "[Vue warn]: Failed to resolve
// component" at runtime (see logo.svg fix, 2026-11). This script fails if any
// icon SVG regresses. Duplicate width/height attributes produce non-fatal
// compiler errors (x2) and are tolerated, matching unplugin-icons behavior.
// Usage: node scripts/check-svg-metadata.cjs
const process = require('node:process');
const fs = require('node:fs');
const path = require('node:path');
const { compileTemplate } = require(path.join(__dirname, '..', 'node_modules', '@vue', 'compiler-sfc'));

const iconDir = path.join(__dirname, '..', 'src', 'assets', 'svg-icon');

let failed = false;
for (const file of fs.readdirSync(iconDir).filter(f => f.endsWith('.svg'))) {
  const raw = fs.readFileSync(path.join(iconDir, file), 'utf8');
  // same transform as build/plugins/unplugin.ts FileSystemIconLoader
  const svg = raw.replace(/^<\?xml[^>]*\?>\s*/m, '').replace(/^<svg\s/, '<svg width="1em" height="1em" ');
  const { code } = compileTemplate({ source: svg, id: `local:${file}`, filename: `local-${file}.vue` });
  const offenders = (code.match(/resolveComponent\("[^"]+"\)/g) || []).map(m => m.slice(18, -2));
  if (offenders.length > 0) {
    failed = true;
    console.error(`RED ${file} — compiled metadata as components: ${offenders.join(', ')}`);
  }
}
if (failed) process.exit(1);
console.log(
  `GREEN — no metadata elements compiled as components (${fs.readdirSync(iconDir).filter(f => f.endsWith('.svg')).length} icons checked)`
);
