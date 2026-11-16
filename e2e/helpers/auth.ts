import { Page, expect } from '@playwright/test';

export const BASE_URL = process.env.E2E_BASE_URL ?? 'http://localhost:9527';

export const ADMIN = {
  username: process.env.E2E_USERNAME ?? 'admin',
  password: process.env.E2E_PASSWORD ?? '123456',
};

/** 登录并等待进入系统首页（未登录则自动跳 /login） */
export async function login(page: Page): Promise<void> {
  await page.goto('/');
  // SPA 路由重定向到 /login 是异步的：goto 返回时 URL 可能还是 /，
  // 等 URL 稳定（出现 /login 或出现路由跳转完成）后再判断
  await page.waitForURL(/\/(login|home|home\/)/, { timeout: 15_000 }).catch(() => {});
  if (!page.url().includes('/login')) {
    return;
  }
  await page.locator('input[placeholder*="请输入用户名"]').first().fill(ADMIN.username);
  await page.locator('input[type="password"]').first().fill(ADMIN.password);
  // 登录按钮文本是 i18n 的 common.confirm（"确认"）
  await page.locator('button:has-text("确认")').first().click();
  await expect(page).not.toHaveURL(/\/login/, { timeout: 20_000 });
  // 等 token 真正写入 localStorage（loginByToken 是异步链）
  await page.waitForFunction(() => !!localStorage.getItem('JZERO_ADMIN_token'), { timeout: 15_000 });
}

/** 登录后直接导航到某路由，等待页面网络空闲 */
export async function gotoPage(page: Page, route: string): Promise<void> {
  await page.goto(route);
  await page.waitForLoadState('networkidle');
}

/**
 * 收集页面 console error / pageerror / 严重请求失败。
 * 必须在 goto 之前调用，测试结束时断言 report.errors 为空。
 */
export interface ConsoleReport {
  errors: string[];
}

export function attachErrorCollector(page: Page): ConsoleReport {
  const errors: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      errors.push(`[console.error] ${msg.text()}`);
    }
  });
  page.on('pageerror', (err) => {
    errors.push(`[pageerror] ${err.message}`);
  });
  page.on('requestfailed', (req) => {
    const failure = req.failure()?.errorText ?? 'unknown';
    // 静态资源可重试失败不阻断；仅记录非 abort 类
    if (failure !== 'net::ERR_ABORTED') {
      errors.push(`[requestfailed] ${req.method()} ${req.url()} ${failure}`);
    }
  });
  page.on('response', (res) => {
    if (res.status() >= 500) {
      errors.push(`[http ${res.status()}] ${res.request().method()} ${res.url()}`);
    }
  });
  return { errors };
}
