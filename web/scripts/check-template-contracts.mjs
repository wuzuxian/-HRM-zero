#!/usr/bin/env node
/**
 * 守卫：Vue 模板绑定禁止 `as unknown as <type>` 强转组件 value 契约。
 *
 * 实际事故（2026-11）：NTimePicker :value 被传字符串（`:value="model.startTime as unknown as number"`） 后经 date-fns format 抛
 * RangeError: Invalid time value，弹窗整页挂掉 （e2e/tests/bugs/time-picker-invalid-time-value.spec.ts 回归）。 强转只是骗过
 * vue-tsc；契约不匹配应修数据形态（如 NTimePicker 用 formatted-value）。
 *
 * 为什么不是 eslint no-restricted-syntax：本仓库 soybean 配置下模板表达式走 espree 静默回退，TSAsExpression 节点不存在，AST 规则打不中（已验证）。
 * 固定句法模式用正则扫描即可确定命中。
 */
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import process from 'node:process';

const ROOT = new URL('../src', import.meta.url).pathname;

// 只盯模板绑定（:value / :xxx=），script 里的 as 强转是正常 TS 用法
const PATTERN = /(?::[a-zA-Z-]+(?:\.[a-zA-Z-]+)*="[^"]*\bas\s+unknown\s+as\s+)/;

/** @returns {string[]} */
function walk(dir) {
  const out = [];
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    const st = statSync(p);
    if (st.isDirectory()) out.push(...walk(p));
    else if (name.endsWith('.vue')) out.push(p);
  }
  return out;
}

const offenders = [];
for (const file of walk(ROOT)) {
  const lines = readFileSync(file, 'utf8').split('\n');
  lines.forEach((line, i) => {
    if (PATTERN.test(line)) {
      offenders.push(`${file}:${i + 1}: ${line.trim().slice(0, 120)}`);
    }
  });
}

if (offenders.length > 0) {
  console.error(`模板绑定禁止 as-unknown-as 强转组件 value 契约（修数据形态，不要 cast 骗过类型检查）：`);
  for (const o of offenders) console.error(`  ${o}`);
  process.exit(1);
}
console.log('template contract guard: OK');
