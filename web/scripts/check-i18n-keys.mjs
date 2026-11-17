/**
 * [i18n-guard] 反馈环脚本：校验全仓代码引用的 i18n key 在 zh-CN locale 下可解析。
 *
 * 红条件：任一引用 key 在 zh-cn.ts 中解析为 undefined（缺 key / 结构错位）。 用法: node scripts/check-i18n-keys.mjs
 */
import { execFileSync, execSync } from 'node:child_process';
import { promises as fs } from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const srcDir = path.join(root, 'src');

// 1. 收集全仓 $t('...') / $t("...") 静态 key
const files = execSync(`grep -rl --include='*.vue' --include='*.ts' "\\$t(" ${srcDir}`, {
  encoding: 'utf8'
})
  .trim()
  .split('\n')
  .filter(Boolean);

const usedKeys = new Map(); // key -> [file:line]
const contents = await Promise.all(files.map(f => fs.readFile(f, 'utf8')));
for (let i = 0; i < files.length; i += 1) {
  const f = files[i];
  const content = contents[i];
  const re = /\$t\(\s*(['"])([^'"]+)\1/g;
  let m;
  while ((m = re.exec(content))) {
    const line = content.slice(0, m.index).split('\n').length;
    const rel = path.relative(root, f);
    if (!usedKeys.has(m[2])) usedKeys.set(m[2], []);
    usedKeys.get(m[2]).push(`${rel}:${line}`);
  }
}

// 2. 用 tsx 把 zh-cn.ts 作为真实模块加载（与运行时同一数据源）
const modPath = path.join(srcDir, 'locales/langs/zh-cn.ts');
const out = execFileSync(
  'pnpm',
  ['--dir', root, 'exec', 'tsx', '-e', `import mod from '${modPath}'; console.log(JSON.stringify(mod));`],
  { encoding: 'utf8', maxBuffer: 64 * 1024 * 1024 }
);
const zhCN = JSON.parse(out.slice(out.indexOf('{')));

// 3. 解析函数：数值段兼容 weekday.{1..7} 对象键
function resolve(obj, key) {
  return key.split('.').reduce((acc, seg) => {
    if (acc === null || acc === undefined) return undefined;
    if (Object.hasOwn(acc, seg)) return acc[seg];
    return undefined;
  }, obj);
}

// 4. 断言
const missing = [];
const keyList = [...usedKeys.entries()];
for (const [key, sites] of keyList) {
  const isDynamic = key.includes('${');
  const val = isDynamic ? undefined : resolve(zhCN, key);
  // 含变量拼接的 key 无法静态校验，跳过
  if (!isDynamic && typeof val !== 'string') {
    missing.push({ key, sites });
  }
}

if (missing.length > 0) {
  console.log(`\n[i18n-guard] RED — ${missing.length} 个 key 在 zh-CN 下解析失败:\n`);
  for (const { key, sites } of missing) {
    console.log(`  ✗ ${key}`);
    for (const s of sites.slice(0, 3)) console.log(`      at ${s}`);
  }
  process.exit(1);
}

console.log(`[i18n-guard] GREEN — ${usedKeys.size} 个静态引用 key 全部在 zh-CN 下可解析`);
