import { test, expect, request as pwRequest } from '@playwright/test';
import { setupAssessmentData, loginAs, type AssessmentFixture } from '../../helpers/perf-data';

/**
 * 交互层：考核任务全链路（HRM-43/44）
 *
 * 前置（beforeAll 用 API 造数据）：指标→模板→方案→2 名员工(+账号+超管角色)→批次→发布→生成任务
 * UI 链路（serial，阶段间有依赖）：
 *   1. 员工0: 工作台卡片 → 目标确认 → 自评填报（评分弹窗逐项打分）
 *   2. admin: 上级评审（节点3）→ 校准填报（节点4）
 *   3. 员工0: 结果确认（节点5）
 *   4. 员工1: 结果确认阶段发起申诉
 *   5. admin: 申诉裁决（调整得分）
 */

test.describe.configure({ mode: 'serial' });

let fixture: AssessmentFixture;
const SCORES = { self: '90', review: '85', calibrate: '82', adjust: '88' };

test.beforeAll(async ({ }) => {
  const ctx = await pwRequest.newContext();
  fixture = await setupAssessmentData(ctx, 2);
  await ctx.dispose();
  console.log('### 数据准备完成:', fixture.periodName, fixture.employees.map(e => e.username));
});

const row = (page: import('@playwright/test').Page, text: string) =>
  page.locator('.n-data-table tbody tr', { hasText: text });

/**
 * 打开考核任务列表页并用批次 uuid 过滤。
 * 任务列表会跨批次累积（历次 run 数据共存），不过滤的话目标行可能翻不到。
 */
async function gotoAssessmentPage(page: import('@playwright/test').Page, route: string) {
  await page.goto(route);
  await page.waitForTimeout(1500);
  await page.locator('input[placeholder*="批次"]').first().fill(fixture.periodUuid);
  await page.locator('button:has-text("搜索")').first().click();
  await page.waitForTimeout(1000);
}

async function openScoreModalAndSubmit(
  page: import('@playwright/test').Page,
  buttonText: string,
  score: string,
  comment: string,
  modalTitle: string
) {
  await page.getByRole('button', { name: buttonText }).first().click();
  const modal = page.locator('.n-modal', { hasText: modalTitle });
  await expect(modal).toBeVisible();
  // 逐项评分输入（每行一个 NInputNumber）
  const firstScoreInput = modal.locator('.n-input-number input').first();
  await expect(firstScoreInput).toBeVisible();
  await firstScoreInput.fill(score);
  await firstScoreInput.press('Tab');
  // 评论
  await modal.locator('.n-input textarea').first().fill(comment);
  // 弹窗底部确认
  await modal.locator('button:has-text("确认")').last().click();
  await expect(page.locator('.n-message', { hasText: '更新成功' })).toBeVisible({ timeout: 10_000 });
}

// ---------- 1. 员工0：工作台 → 目标确认 → 自评填报 ----------

test('员工0 工作台显示待办卡片并可跳转', async ({ page }) => {
  const emp0 = fixture.employees[0];
  await loginAs(page, emp0.username, emp0.password);

  await page.goto('/hrm/performance/assessment-task/workbench');
  await page.waitForTimeout(1500);

  // 六张汇总卡渲染，目标确认卡有 1 条待办
  const planCard = page.locator('.n-statistic', { hasText: '目标确认' });
  await expect(planCard).toBeVisible();
  await expect(planCard.locator('.text-20px')).toHaveText('1');

  // 点击卡片跳转到目标确认列表
  await planCard.click();
  await page.waitForTimeout(1500);
  expect(page.url()).toContain('/plan-list');
});

test('员工0 确认目标（节点1→2）', async ({ page }) => {
  const emp0 = fixture.employees[0];
  await loginAs(page, emp0.username, emp0.password);

  await gotoAssessmentPage(page, '/hrm/performance/assessment-task/plan-list');
  await expect(row(page, emp0.name).first()).toBeVisible();

  await row(page, emp0.name).first().getByRole('button', { name: '确认目标' }).click();
  await expect(page.locator('.n-message', { hasText: '目标已确认' })).toBeVisible({ timeout: 10_000 });
  // 页面按节点1过滤，确认后行消失
  await expect(row(page, emp0.name)).toHaveCount(0, { timeout: 10_000 });
});

test('员工0 自评填报（节点2→3）', async ({ page }) => {
  const emp0 = fixture.employees[0];
  await loginAs(page, emp0.username, emp0.password);

  await gotoAssessmentPage(page, '/hrm/performance/assessment-task/self-review-list');
  await expect(row(page, emp0.name).first()).toBeVisible();

  await openScoreModalAndSubmit(page, '自评填报', SCORES.self, 'e2e 自评意见：目标达成良好', '自评填报');
  // 页面按节点2过滤，提交后行消失
  await expect(row(page, emp0.name)).toHaveCount(0, { timeout: 10_000 });
});

// ---------- 2. admin：上级评审 → 校准填报 ----------

test('admin 上级评审（节点3→4）', async ({ page }) => {
  const emp0 = fixture.employees[0];
  await loginAs(page, 'admin', '123456');

  await gotoAssessmentPage(page, '/hrm/performance/assessment-task/leader-review-list');
  // 无节点过滤，按员工名定位行
  await expect(row(page, emp0.name).first()).toBeVisible();

  await openScoreModalAndSubmit(page, '上级评审', SCORES.review, 'e2e 评审意见：整体达标', '上级评审');
  // 提交后行节点变 4，出现"校准填报"按钮
  await expect(
    row(page, emp0.name).first().getByRole('button', { name: '校准填报' })
  ).toBeVisible({ timeout: 10_000 });
});

test('admin 校准填报（节点4→5）', async ({ page }) => {
  const emp0 = fixture.employees[0];
  await loginAs(page, 'admin', '123456');

  await gotoAssessmentPage(page, '/hrm/performance/assessment-task/leader-review-list');
  await row(page, emp0.name).first().getByRole('button', { name: '校准填报' }).click();

  const modal = page.locator('.n-modal', { hasText: '校准填报' });
  await expect(modal).toBeVisible();
  const finalScore = modal.locator('.n-input-number input').first();
  await expect(finalScore).not.toHaveValue('');
  await finalScore.fill(SCORES.calibrate);
  await finalScore.press('Tab');
  await modal.locator('.n-input textarea').first().fill('e2e 校准意见');
  await modal.locator('button:has-text("确认")').last().click();
  await expect(page.locator('.n-message', { hasText: '更新成功' })).toBeVisible({ timeout: 10_000 });

  // 节点 4→5：本页（评审填报）节点 tag 应变为"结果确认"
  //（确认结果/发起申诉按钮只在结果确认页渲染，由后续用例覆盖）
  await expect(
    row(page, emp0.name).first().locator('.n-tag', { hasText: '结果确认' })
  ).toBeVisible({ timeout: 10_000 });
});

// ---------- 3. 员工0：结果确认 ----------

test('员工0 确认结果（节点5 完成）', async ({ page }) => {
  const emp0 = fixture.employees[0];
  await loginAs(page, emp0.username, emp0.password);

  await gotoAssessmentPage(page, '/hrm/performance/assessment-task/result-confirm-task-list');
  await expect(row(page, emp0.name).first()).toBeVisible();

  await row(page, emp0.name).first().getByRole('button', { name: '确认结果' }).click();
  await expect(page.locator('.n-message', { hasText: '考核结果已确认' })).toBeVisible({ timeout: 10_000 });
  // 确认后任务完成，行不再有操作按钮（只剩详情）
  await expect(
    row(page, emp0.name).first().getByRole('button', { name: '确认结果' })
  ).toHaveCount(0, { timeout: 10_000 });
});

// ---------- 4. 员工1：发起申诉 ----------

test('员工1 节点5 发起申诉', async ({ page }) => {
  const emp1 = fixture.employees[1];
  await loginAs(page, emp1.username, emp1.password);

  await gotoAssessmentPage(page, '/hrm/performance/assessment-task/result-confirm-task-list');
  await expect(row(page, emp1.name).first()).toBeVisible();

  await row(page, emp1.name).first().getByRole('button', { name: '发起申诉' }).click();
  const modal = page.locator('.n-modal', { hasText: '发起申诉' });
  await expect(modal).toBeVisible();
  await modal.locator('.n-input textarea').first().fill('e2e 申诉原因：自评与校准分差异过大');
  await modal.locator('button:has-text("确认")').last().click();
  await expect(page.locator('.n-message', { hasText: '申诉已提交' })).toBeVisible({ timeout: 10_000 });
});

// ---------- 5. admin：申诉裁决（调整得分） ----------

test('admin 申诉裁决-调整得分', async ({ page }) => {
  const emp1 = fixture.employees[1];
  await loginAs(page, 'admin', '123456');

  await gotoAssessmentPage(page, '/hrm/performance/assessment-task/result-confirm-task-list');
  await expect(row(page, emp1.name).first()).toBeVisible();

  // 操作列在表格横向滚动区最右端，先滚到最右再点
  await page.evaluate(() => {
    const scrollEl = document.querySelector('.n-data-table .n-scrollbar-rail__scrollbar')?.parentElement
      ?? document.querySelector('.n-data-table .n-data-table-base-table-body');
    if (scrollEl) scrollEl.scrollLeft = 99999;
  });
  await page.waitForTimeout(500);

  await row(page, emp1.name).first().getByRole('button', { name: '申诉裁决' }).click();
  const modal = page.locator('.n-modal', { hasText: '申诉裁决' });
  await expect(modal).toBeVisible();

  // 裁决模式：选择"调整得分"→ 填调整后分数
  await modal.locator('.n-radio', { hasText: '调整得分' }).click();
  const adjusted = modal.locator('.n-input-number input').first();
  await expect(adjusted).toBeVisible();
  await adjusted.fill(SCORES.adjust);
  await adjusted.press('Tab');
  await modal.locator('.n-input textarea').first().fill('e2e 裁决意见：酌情上调');
  await modal.locator('button:has-text("确认")').last().click();
  await expect(page.locator('.n-message', { hasText: '申诉裁决完成' })).toBeVisible({ timeout: 10_000 });

  // 裁决后任务完成（taskStatus 已完成），申诉裁决按钮消失
  await expect(
    row(page, emp1.name).first().getByRole('button', { name: '申诉裁决' })
  ).toHaveCount(0, { timeout: 10_000 });
});

// ---------- 6. 只读详情弹窗 ----------

test('考核任务详情弹窗（快照与评分记录）', async ({ page }) => {
  const emp0 = fixture.employees[0];
  await loginAs(page, emp0.username, emp0.password);

  await gotoAssessmentPage(page, '/hrm/performance/assessment-task/plan-list');
  // plan-list 按节点1过滤，emp0 任务已过节点1 → 行应不存在
  await expect(row(page, emp0.name)).toHaveCount(0);

  // 改从工作台进入：切换到 leader-review 页打开详情（无节点过滤）
  await gotoAssessmentPage(page, '/hrm/performance/assessment-task/leader-review-list');
  await row(page, emp0.name).first().getByRole('button', { name: '详情' }).click();
  const modal = page.locator('.n-modal', { hasText: '考核任务详情' });
  await expect(modal).toBeVisible();
  // 快照表格：指标名 + 得分列（弹窗含 NDescriptions 的 table + items 快照 table，
  // 直接断言弹窗内容含快照指标名与得分）
  await expect(modal).toContainText('考核模板快照');
  await expect(modal).toContainText(`E2E-KPI-${fixture.periodName.slice(-8)}`);
  await expect(modal).toContainText('100%');
});
