import { test, expect } from '@playwright/test';
import { login, attachErrorCollector, gotoPage } from '../../helpers/auth';
import { HRM_PAGES } from '../../helpers/hrm-pages';

/**
 * 冒烟层：全部 62 个 HRM 页面
 * 断言：页面可达、无 console error / pageerror / 5xx / 请求失败。
 * 忽略规则见 helpers/auth.ts（可在此追加已知噪音）。
 */

const IGNORED_ERROR_PATTERNS: RegExp[] = [
  // 例：/ResizeObserver loop/i,
];

test.beforeEach(async ({ page }) => {
  await login(page);
});

for (const pageInfo of HRM_PAGES) {
  test(`冒烟: ${pageInfo.module} / ${pageInfo.name} [${pageInfo.route}]`, async ({ page }) => {
    const report = attachErrorCollector(page);
    await gotoPage(page, pageInfo.route);

    // 页面主容器渲染出来（soybean 布局的内容区）
    await expect(page.locator('#app')).toBeVisible();
    // 至少等一个表格/卡片/表单类容器渲染（空数据也应有空态）
    await expect(
      page.locator('.n-data-table, .n-card, .n-form, .n-result, .n-tabs, .n-grid').first()
    ).toBeVisible({ timeout: 20_000 });

    const meaningful = report.errors.filter((e) => !IGNORED_ERROR_PATTERNS.some((p) => p.test(e)));
    expect(
      meaningful,
      `${pageInfo.route} 存在页面错误:\n${meaningful.join('\n')}`
    ).toEqual([]);
  });
}
