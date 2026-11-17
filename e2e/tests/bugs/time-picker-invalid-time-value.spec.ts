import { test, expect } from '@playwright/test';
import { login } from '../../helpers/auth';

/**
 * 回归：考勤设置页“新增”弹窗打开即抛 RangeError: Invalid time value（2026-11 用户报告）
 *
 * 症状：班次管理 / 考勤组 / 组织默认规则页，点击“新增”后弹窗打开瞬间抛
 *       RangeError: Invalid time value（Promise.then → index.vue add handler）。
 * 根因：NTimePicker 被传入字符串 value（:value="model.startTime as unknown as number"）。
 *       naive-ui TimePicker 的 value 契约是时间戳(number|null)；拿到字符串后经
 *       date-fns format(value, 'HH:mm') 渲染显示文本，date-fns 对非 Date/number 输入
 *       直接 throw RangeError: Invalid time value。shift 的 restStart/restEnd 默认 ''
 *       → 新增弹窗必炸；group/config 编辑回填 null 同理。
 * 修复：三处 modal 统一改为 v-model:formatted-value + value-format="HH:mm"（null 安全，
 *       与库内其他页面 NDatePicker 的 v-model:formatted-value 惯用法一致）。
 */

const PAGES = [
  { route: '/hrm/attendance/setting/shift', name: '班次管理' },
  { route: '/hrm/attendance/setting/group', name: '考勤组' },
  { route: '/hrm/attendance/setting/attendance-config', name: '组织默认规则' },
];

test.beforeEach(async ({ page }) => {
  await login(page);
});

for (const { route, name } of PAGES) {
  test(`${name}(${route}) 点击新增弹窗打开不抛 RangeError`, async ({ page }) => {
    const pageErrors: string[] = [];
    page.on('pageerror', (err) => {
      pageErrors.push(`${err.name}: ${err.message}`);
    });

    await page.goto(route);
    await page.waitForLoadState('networkidle');

    // 新增按钮在 TableHeaderOperation 中，文案为 i18n common.add（“新增”）
    const addBtn = page.locator('button:has-text("新增")').first();
    await expect(addBtn).toBeVisible();
    await addBtn.click();

    // 红色信号：bug 下弹窗渲染直接抛 RangeError（弹窗挂载失败）
    await page.waitForTimeout(2000);
    const rangeErrors = pageErrors.filter(e => e.includes('RangeError: Invalid time value'));
    expect(rangeErrors, `${name} 打开新增弹窗不应抛 RangeError: Invalid time value`).toEqual([]);
  });
}

// 班次管理：弹窗打开后时间选择器显示默认值（bug 下弹窗渲染抛错、TimePicker 输入框为空）
test('班次管理 新增弹窗时间选择器渲染默认值 09:00', async ({ page }) => {
  const pageErrors: string[] = [];
  page.on('pageerror', (err) => {
    pageErrors.push(`${err.name}: ${err.message}`);
  });

  await page.goto('/hrm/attendance/setting/shift');
  await page.waitForLoadState('networkidle');
  await page.locator('button:has-text("新增")').first().click();
  await page.waitForTimeout(2000);

  // NTimePicker 输入框应显示格式化默认值（bug 时弹窗挂载失败、输入框不存在）
  const timeInputs = page.locator('.n-modal .n-time-picker input');
  await expect(timeInputs).toHaveCount(4);
  await expect(timeInputs.nth(0)).toHaveValue('09:00');
  await expect(timeInputs.nth(1)).toHaveValue('18:00');
  expect(
    pageErrors.filter(e => e.includes('RangeError')),
    '新增弹窗不应抛 RangeError'
  ).toEqual([]);
});
