import { test, expect } from '@playwright/test';
import { login } from '../../helpers/auth';

/**
 * 回归：页面跳转空白（2026-11 用户报告事故）
 *
 * 症状：从考核任务 4 个列表页（deep link 或刷新进入）做第一次 SPA 跳转后，
 *       内容区永久空白（本页及后续所有跳转），仅 F5 整页刷新可恢复。
 * 根因：这 4 页 template 根节点前有 HTML 注释，dev 模式 Vue 保留注释节点，
 *       组件根是 fragment [comment, div]；global-content 的
 *       <Transition mode="out-in"> 对 fragment 离场处理有缺陷，旧页 leave 后新页永不 enter。
 *       Vue 3.5.11。修复：注释移入根 div 内部。
 * 任何页面在 template 根部放注释都会触发本 bug，回归测试必须保留。
 */

const PAGES = [
  '/hrm/performance/assessment-task/plan-list',
  '/hrm/performance/assessment-task/self-review-list',
  '/hrm/performance/assessment-task/leader-review-list',
  '/hrm/performance/assessment-task/result-confirm-task-list',
];

test.beforeEach(async ({ page }) => {
  await login(page);
});

// 用户原始复现路径：刷新 plan-list → 菜单点击 self-review-list
test('刷新 plan-list 后菜单跳转 self-review-list 不空白', async ({ page }) => {
  await page.goto(PAGES[0]);
  await page.waitForTimeout(1500);
  await expect(page.locator('.n-data-table').first()).toBeVisible();

  await page.getByRole('menuitem', { name: '自评填报' }).click();
  await page.waitForTimeout(2000);
  // 内容区有真实渲染（空白时 n-data-table 数为 0）
  await expect(page.locator('.n-data-table, .n-card').first()).toBeVisible();
  const tableCount = await page.locator('.n-data-table, .n-card').count();
  expect(tableCount, '跳转后内容区不应空白').toBeGreaterThan(0);
});

// 每个曾出问题的页面：deep link 后第一跳目标页必须渲染
for (const src of PAGES) {
  test(`deep link ${src} → 指标库 第一跳渲染`, async ({ page }) => {
    await page.goto(src);
    await page.waitForTimeout(1500);
    await expect(page.locator('.n-data-table, .n-card').first()).toBeVisible();

    await page.getByRole('menuitem', { name: '指标库' }).click();
    await page.waitForTimeout(2000);
    const tableCount = await page.locator('.n-data-table').count();
    expect(tableCount, `${src} 跳出后指标库页应渲染表格`).toBeGreaterThan(0);
  });
}

// 空白恢复路径回归：F5 可恢复（reloadPage/tab 切换均不可恢复——当时的行为）
test('连续多跳不产生空白（第三跳仍渲染）', async ({ page }) => {
  await page.goto('/hrm/performance/assessment-task/plan-list');
  await page.waitForTimeout(1500);
  await page.getByRole('menuitem', { name: '指标库' }).click();
  await page.waitForTimeout(2000);
  await page.getByRole('menuitem', { name: '模板管理' }).click();
  await page.waitForTimeout(2000);
  await page.getByRole('menuitem', { name: '绩效方案' }).click();
  await page.waitForTimeout(2000);
  const tableCount = await page.locator('.n-data-table').count();
  expect(tableCount, '连续三跳后页面应渲染').toBeGreaterThan(0);
});
