import { test, expect, request as pwRequest, type Page, type Locator } from '@playwright/test';
import { loginAs } from '../../helpers/perf-data';
import {
  EMPLOYEE_NAME_PREFIX,
  STANDALONE_USER_PREFIX,
  addEmployeeViaApi,
  adminToken,
  cleanupFixtures,
  createRoleWithMenus,
  createStandaloneUser,
  editEmployee,
  editUserRoles,
  findEmployeeByName,
  generateUserViaApi,
  getEmployeeByUuid,
  linkEmployeeUserRaw,
  resolveMenuUuidsForPage,
  retireHistoricalEmployees,
  unlinkEmployeeViaApi,
  getUserByUsername
} from '../../helpers/employee-data';

/**
 * 交互层验收：员工档案 ↔ 系统账号打通（ticket #56 / HRM-54，ADR 0004 验收门禁）
 *
 * 覆盖 12 条用例链路（serial，状态沿测试序推进）：
 *   1. 新增即开号，成功提示含登录用户名与初始密码
 *   2. 用提示里的凭据登录，能看到 考勤打卡/个人看板/我的培训
 *   3. 同一账号菜单不含 评审填报/考核任务工作台（默认角色 R_EMPLOYEE）
 *   4. 列表「登录账号」= 工号，「用户状态」= 已生成
 *   5. 解绑后「用户状态」= 未生成，且该账号仍能登录
 *   6. 候选搜索重新关联后恢复已生成
 *   7. 候选列表不含已被其他员工占用的账号
 *   8. 一对一冲突：明确错误提示（占用者工号+姓名）且不写入
 *      —— UI 候选下拉结构性地排除被占用账号（用例 7），故冲突拒绝走 API 契约断言
 *   9. 用户管理「关联员工」列显示 工号 + 姓名
 *  10. 用户管理删除被引用账号被拒，提示先解绑，账号不消失
 *  11. 勾选 已生成+未生成 点「生成用户」，提示生成 1 跳过 1
 *  12. 非超管但被授予按钮码的角色能看到「生成用户」按钮
 *
 * 前置推进全走 API（:8001），UI 只做被测动作。
 * 每个 test 独立 loginAs（规避多 tab 长会话累积 Transition/KeepAlive 状态）；
 * toast 断言紧跟点击（NMessage ~3s 自动关闭）；操作列点击前拉满横向滚动。
 */

test.describe.configure({ mode: 'serial' });

const ARCHIVE_ROUTE = '/hrm/employee/employee-archive-list';
const USER_ROUTE = '/manage/user';
const EMPLOYEE_PAGE_MENU_LABEL = '员工档案';
const GENERATE_USER_BUTTON_LABEL = '生成用户';

const ts = Date.now().toString().slice(-9);
const NAME_A = `${EMPLOYEE_NAME_PREFIX}${ts}A`; // UI 新增（建档即开号）
const NAME_B = `${EMPLOYEE_NAME_PREFIX}${ts}B`; // API 建，已关联（占用对照）
const NAME_C = `${EMPLOYEE_NAME_PREFIX}${ts}C`; // API 建，已关联（候选下拉宿主）
const NAME_D = `${EMPLOYEE_NAME_PREFIX}${ts}D`; // API 建 → 解绑 → 未生成（冲突目标）
const NAME_E = `${EMPLOYEE_NAME_PREFIX}${ts}E`; // 离职建档(不开号) → 改正式 → 在职未生成（批量生成目标）

interface EmployeeFixture {
  uuid: string;
  employeeNo: string;
  name: string;
  username: string;
  password: string;
  userUuid: string;
}

let apiCtx: pwRequest.APIRequestContext;
let token: string;
let empA: EmployeeFixture;
let empB: EmployeeFixture;
let empC: EmployeeFixture;
let empD: EmployeeFixture;
let empE: EmployeeFixture;
let standaloneUser: { uuid: string; username: string; password: string };
let tempRoleUuid: string;
let grantUser: { uuid: string; username: string; password: string };

const row = (page: Page, text: string) => page.locator('.n-data-table tbody tr', { hasText: text });

/** 表格横向滚动拉满（操作列在最右端，不滚动点不到） */
async function scrollTableRight(page: Page) {
  await page.evaluate(() => {
    document
      .querySelectorAll<HTMLElement>('.n-data-table-base-table-body, .n-data-table-base-table .n-scrollbar-container')
      .forEach(el => {
        el.scrollLeft = el.scrollWidth;
      });
  });
}

/** 表头文本 → 该行对应列的单元格（用 DOM 对齐 th/td，避免硬编码列序） */
async function cellByHeader(page: Page, rowLocator: Locator, headerText: string): Promise<Locator> {
  const headers = await page.locator('.n-data-table thead th').allTextContents();
  const idx = headers.findIndex(h => h.includes(headerText));
  if (idx < 0) throw new Error(`表头未找到列: ${headerText} (实际: ${headers.join('|')})`);
  return rowLocator.locator('td').nth(idx);
}

/** 逐级展开左侧菜单目录（NMenu 多级嵌套，目录项本身是 menuitem，点击切换展开） */
async function openMenuPath(page: Page, ...labels: string[]) {
  for (const label of labels) {
    const item = page.getByRole('menuitem', { name: label }).first();
    if (!(await item.isVisible().catch(() => false))) {
      throw new Error(`菜单目录不可见: ${label}`);
    }
    const expanded = await item.evaluate(el =>
      el.closest('.n-submenu')?.classList.contains('n-submenu--expanded')
    );
    if (!expanded) {
      await item.click();
      await page.waitForTimeout(400);
    }
  }
}

/** admin 打开列表页并按输入框 placeholder 定位搜索（LIKE 前缀；搜索表单默认收在折叠面板里）。
 *  员工档案页（placeholder 含「姓名」）与用户管理页（placeholder 含「用户名」）的搜索
 *  折叠面板结构一致, 收敛为同一实现; 不用 networkidle（dev 模式 Vite 常驻连接会饿死
 *  networkidle），一律等具体元素。 */
async function gotoListAndSearch(page: Page, route: string, placeholderKeyword: string, keyword: string) {
  await page.goto(route);
  const searchInput = page.locator(`input[placeholder*="${placeholderKeyword}"]`).first();
  await page
    .locator(`.n-collapse-item__header, input[placeholder*="${placeholderKeyword}"]`)
    .first()
    .waitFor({ state: 'visible', timeout: 15_000 });
  if (!(await searchInput.isVisible().catch(() => false))) {
    await page.locator('.n-collapse-item__header', { hasText: '搜索' }).first().click();
    await searchInput.waitFor({ state: 'visible', timeout: 5000 }).catch(() => {});
  }
  await searchInput.fill(keyword);
  await page.locator('button:has-text("搜索")').first().click();
  await page.waitForTimeout(800);
}

const gotoArchiveAndSearch = (page: Page, keyword: string) => gotoListAndSearch(page, ARCHIVE_ROUTE, '姓名', keyword);

/** 打开某员工行的编辑弹窗并切到「系统账号」Tab */
async function openEditModalAccountTab(page: Page, employeeName: string) {
  await scrollTableRight(page);
  await row(page, employeeName).first().getByRole('button', { name: '编辑' }).click();
  const modal = page.locator('.n-modal', { hasText: '编辑员工' });
  await expect(modal).toBeVisible();
  await modal.locator('.n-tabs-tab', { hasText: '系统账号' }).click();
  return modal;
}

/** 等待并断言一条 message toast（NMessage ~3s 自动关闭，必须紧跟点击） */
async function expectMessage(page: Page, text: string) {
  await expect(page.locator('.n-message', { hasText: text }).first()).toBeVisible({ timeout: 8_000 });
}

test.beforeAll(async () => {
  apiCtx = await pwRequest.newContext();
  token = await adminToken(apiCtx);
  await retireHistoricalEmployees(apiCtx, token);

  // B/C: API 建档即开号（已关联）
  const b = await addEmployeeViaApi(apiCtx, token, { name: NAME_B });
  empB = b.row;
  const c = await addEmployeeViaApi(apiCtx, token, { name: NAME_C });
  empC = c.row;

  // D: API 建档开号后解绑 → 未生成（孤儿账号保留，账号仍存在）
  const d = await addEmployeeViaApi(apiCtx, token, { name: NAME_D });
  empD = d.row;
  await unlinkEmployeeViaApi(apiCtx, token, empD.uuid);
  empD.userUuid = '';

  // E: 离职建档不开号 → 改回正式 → 在职且从未开号（generateUser 的"未生成"侧；
  //    若走"开号后解绑"，再次 generateUser 会撞孤儿账号用户名 10461，那是既定设计）
  const e = await addEmployeeViaApi(apiCtx, token, { name: NAME_E, personStatus: '5' });
  expect(e.userCreated, '离职建档不开号').toBe(false);
  empE = e.row;
  await editEmployee(apiCtx, token, { uuid: empE.uuid, name: empE.name, personStatus: '1' });

  // 独立账号（候选对照）+ 授码角色与非超管账号（用例 12）
  standaloneUser = await createStandaloneUser(apiCtx, token, `${STANDALONE_USER_PREFIX}${ts}`, `e2e候选${ts}`);
  const menuUuids = await resolveMenuUuidsForPage(apiCtx, token, EMPLOYEE_PAGE_MENU_LABEL, GENERATE_USER_BUTTON_LABEL);
  tempRoleUuid = await createRoleWithMenus(
    apiCtx,
    token,
    `e2e生成用户角色${ts}`,
    `R_E2E_GEN_${ts}`,
    menuUuids
  );
  // 登录表单 REG_USER_NAME 限 4-16 字符: e2egrant(8)+ts后8位=16
  grantUser = await createStandaloneUser(apiCtx, token, `e2egrant${ts.slice(-8)}`, `e2e授码员${ts}`);
  // editUser.userRoles 按角色 code 匹配
  await editUserRoles(apiCtx, token, grantUser.uuid, grantUser.username, `e2e授码员${ts}`, [`R_E2E_GEN_${ts}`]);

  console.log(
    '### 造数完成:',
    JSON.stringify({ A: NAME_A, B: empB.employeeNo, C: empC.employeeNo, D: empD.employeeNo, E: empE.employeeNo, U: standaloneUser.username })
  );
});

test.afterAll(async () => {
  // 兜底：中途失败没来得及注册的 fixture 按本 run 的唯一名反查补齐，避免残留
  const names: Record<string, string> = { A: NAME_A, B: NAME_B, C: NAME_C, D: NAME_D, E: NAME_E };
  const lookedUpUuids: string[] = [];
  const orphanUsernames = new Set<string>();
  for (const n of Object.values(names)) {
    const found = await findEmployeeByName(apiCtx, token, n).catch(() => null);
    if (!found) continue;
    lookedUpUuids.push(found.uuid);
    orphanUsernames.add(found.employeeNo);
  }
  // 已注册 fixture 的工号也全部进孤儿清扫（仍被引用的删失败由员工级联兜住）
  for (const e of [empA, empB, empC, empD, empE]) {
    if (e?.employeeNo) orphanUsernames.add(e.employeeNo);
  }
  const warnings = await cleanupFixtures(apiCtx, token, {
    employeeUuids: [
      ...new Set([...[empA, empB, empC, empD, empE].filter(Boolean).map(e => e.uuid), ...lookedUpUuids])
    ],
    orphanUsernames: [...orphanUsernames],
    standaloneUserUuids: [standaloneUser?.uuid, grantUser?.uuid].filter(Boolean) as string[],
    roleUuids: tempRoleUuid ? [tempRoleUuid] : []
  });
  await apiCtx.dispose();
  if (warnings.length > 0) {
    console.warn('### 清理告警:', warnings.join(' | '));
  }
});

// ---------- 1. 新增即开号：成功提示含凭据 ----------
test('新增员工即开号，成功提示含登录用户名与初始密码', async ({ page }) => {
  await loginAs(page, 'admin', '123456');
  await page.goto(ARCHIVE_ROUTE);
  await page.getByRole('button', { name: '新增' }).first().waitFor({ state: 'visible', timeout: 15_000 });

  await page.getByRole('button', { name: '新增' }).first().click();
  const modal = page.locator('.n-modal', { hasText: '新增员工' });
  await expect(modal).toBeVisible();
  await modal.locator('input[placeholder*="姓名"]').first().fill(NAME_A);
  // toast/对话框断言紧跟保存点击
  await modal.locator('button:has-text("确认")').last().click();

  // 建档即开号: $dialog.success(title=员工与账号已创建, content=登录账号: x 初始密码: y)
  const dialog = page.locator('.n-dialog', { hasText: '员工与账号已创建' });
  await expect(dialog).toBeVisible({ timeout: 10_000 });
  const content = (await dialog.locator('.n-dialog__content').innerText()) ?? '';
  const match = content.match(/登录账号:\s*(\S+)\s+初始密码:\s*(\S+)/);
  expect(match, `凭据提示应含 登录账号/初始密码: ${content}`).toBeTruthy();
  const username = match![1];
  const password = match![2];
  expect(password.length, '初始密码非空').toBeGreaterThan(0);

  await dialog.locator('button:has-text("确认")').click();

  // 反查档案行（AddResponse 不带 uuid；断言 username=工号 且已生成）
  const apiRow = await findEmployeeByName(apiCtx, token, NAME_A);
  expect(apiRow, '新增后列表可反查').toBeTruthy();
  expect(apiRow!.username, '登录用户名=工号').toBe(apiRow!.employeeNo);
  expect(apiRow!.userUuid, '已关联账号指针').not.toBe('');
  expect(apiRow!.userStatus, '用户状态=已生成').toBe('1');
  expect(username, '弹窗用户名=工号').toBe(apiRow!.employeeNo);

  empA = { uuid: apiRow!.uuid, employeeNo: apiRow!.employeeNo, name: NAME_A, username, password, userUuid: apiRow!.userUuid };
});

// ---------- 2. 用凭据登录：自助菜单可见 ----------
test('新开账号凭据登录，能看到 考勤打卡/个人看板/我的培训', async ({ page }) => {
  await loginAs(page, empA.username, empA.password);

  await openMenuPath(page, '人力资源', '假勤管理');
  await expect(page.getByRole('menuitem', { name: '考勤打卡' })).toBeVisible();
  await expect(page.getByRole('menuitem', { name: '个人看板' })).toBeVisible();

  await openMenuPath(page, '培训管理');
  await expect(page.getByRole('menuitem', { name: '我的培训' })).toBeVisible();

  // 深链一页证明菜单不是摆设：考勤打卡页真实渲染
  await page.goto('/hrm/attendance/my/attendance-punch');
  await expect(page.locator('.n-card, .n-grid, .n-data-table').first()).toBeVisible({ timeout: 15_000 });
});

// ---------- 3. 默认角色菜单不含管理者视角 ----------
test('同一账号菜单不含 评审填报/考核任务工作台', async ({ page }) => {
  await loginAs(page, empA.username, empA.password);

  await openMenuPath(page, '人力资源', '绩效管理', '考核任务');
  // 正向对照：自服务三页在菜单里
  await expect(page.getByRole('menuitem', { name: '目标确认' })).toBeVisible();
  // 负向：管理者视角页面不授
  await expect(page.getByRole('menuitem', { name: '评审填报' })).toHaveCount(0);
  await expect(page.getByRole('menuitem', { name: '考核任务工作台' })).toHaveCount(0);
});

// ---------- 4. 列表：登录账号=工号，用户状态=已生成 ----------
test('员工列表「登录账号」=工号、「用户状态」=已生成', async ({ page }) => {
  await loginAs(page, 'admin', '123456');
  await gotoArchiveAndSearch(page, NAME_A);

  const target = row(page, NAME_A).first();
  await expect(target).toBeVisible();

  const loginCell = await cellByHeader(page, target, '登录账号');
  await expect(loginCell).toHaveText(empA.employeeNo);
  const statusCell = await cellByHeader(page, target, '用户状态');
  await expect(statusCell).toHaveText(/已生成/);
});

// ---------- 5. 解绑：未生成 + 账号仍能登录 ----------
test('解绑后「用户状态」=未生成，且该账号仍能登录', async ({ page }) => {
  await loginAs(page, 'admin', '123456');
  await gotoArchiveAndSearch(page, NAME_A);

  const modal = await openEditModalAccountTab(page, NAME_A);
  await expect(modal.locator('.n-tag', { hasText: '已关联' })).toBeVisible();

  await modal.locator('button', { hasText: '解除关联' }).first().click();
  // popconfirm 正向按钮（naive-ui zhCN: 确认）
  await page.locator('.n-popconfirm').locator('button:has-text("确认")').first().click();
  await expectMessage(page, '已解除关联'); // toast 紧跟点击

  await expect(modal.locator('.n-tag', { hasText: '未关联' })).toBeVisible({ timeout: 8_000 });
  await modal.locator('button:has-text("取消")').last().click();

  // API: 用户状态=未生成
  const detail = await getEmployeeByUuid(apiCtx, token, empA.uuid);
  expect(detail.userStatus, '解绑后未生成').toBe('0');
  expect(detail.userUuid, '指针已清').toBe('');

  // 账号仍能登录（解绑只清指针，账号与角色绑定保留）
  await loginAs(page, empA.username, empA.password);
  await expect(page.locator('.n-menu').first()).toBeVisible();
});

// ---------- 6. 候选搜索重新关联：恢复已生成 ----------
/** 「关联」卡片里的远程搜索候选 NSelect（filterable 的 placeholder 不是 input 属性 → 点选后用键盘输入） */
async function candidateSelect(page: Page, modal: Locator) {
  // 「关联」卡片是「新建账号并关联」卡片之前的第一个 NCard
  const linkCard = modal.locator('.n-card', { hasText: '关联' }).first();
  const select = linkCard.locator('.n-select').first();
  await select.click();
  return select;
}

async function searchCandidate(page: Page, modal: Locator, keyword: string) {
  await candidateSelect(page, modal);
  await page.keyboard.type(keyword);
}

test('候选搜索重新关联后恢复已生成', async ({ page }) => {
  await loginAs(page, 'admin', '123456');
  await gotoArchiveAndSearch(page, NAME_A);

  const modal = await openEditModalAccountTab(page, NAME_A);
  await expect(modal.locator('.n-tag', { hasText: '未关联' })).toBeVisible();

  // 远程搜索候选（解绑后的账号已不被占用 → 出现在候选）
  await searchCandidate(page, modal, empA.username);
  const option = page.locator('.n-base-select-option', { hasText: empA.username }).first();
  await expect(option).toBeVisible({ timeout: 8_000 });
  await option.click();

  const linkCard = modal.locator('.n-card', { hasText: '关联' }).first();
  await linkCard.locator('button', { hasText: '关联' }).first().click();
  await expectMessage(page, '关联成功'); // toast 紧跟点击

  await expect(modal.locator('.n-tag', { hasText: `已关联: ${empA.username}` })).toBeVisible({ timeout: 8_000 });
  await modal.locator('button:has-text("取消")').last().click();

  const detail = await getEmployeeByUuid(apiCtx, token, empA.uuid);
  expect(detail.userStatus, '重关联后恢复已生成').toBe('1');
  expect(detail.userUuid, '指针回到原账号').toBe(empA.userUuid);
});

// ---------- 7. 候选列表排除被占用账号 ----------
test('候选列表不含已被其他员工占用的账号', async ({ page }) => {
  await loginAs(page, 'admin', '123456');
  await gotoArchiveAndSearch(page, NAME_C);

  const modal = await openEditModalAccountTab(page, NAME_C);

  // 关键字 e2e：命中独立账号（昵称 e2e候选*）；员工账号昵称 e2e账号员工* 但均已被占用
  await searchCandidate(page, modal, 'e2e');
  await page
    .locator('.n-base-select-option')
    .first()
    .waitFor({ state: 'visible', timeout: 8_000 })
    .catch(() => {});

  const optionTexts = await page.locator('.n-base-select-option').allInnerTexts();
  const joined = optionTexts.join('\n');
  expect(optionTexts.length, '至少有候选（独立账号）').toBeGreaterThan(0);
  expect(joined).toContain(standaloneUser.username);
  // 被占用账号（A/B/C 自身）一律不出现
  expect(joined, 'A 的账号被 A 占用不出现').not.toContain(empA.username);
  expect(joined, 'B 的账号被 B 占用不出现').not.toContain(empB.username);
  expect(joined, 'C 的账号被 C 占用不出现').not.toContain(empC.username);
  await modal.locator('button:has-text("取消")').last().click();
});

// ---------- 8. 一对一冲突：明确错误提示且不写入 ----------
test('一对一冲突拒绝（占用者报出+已关联拒绝）且不写入', async ({ page }) => {
  // UI 候选下拉已结构性排除被占用账号（用例 7），冲突拒绝是 API 契约：
  // 把 A 已持有的账号关联给 D（未关联）→ 10720 占用，错误提示报出占用者工号+姓名
  const occupied = await linkEmployeeUserRaw(apiCtx, token, empD.uuid, empA.userUuid);
  expect(occupied.code, '占用拒绝 10720').toBe(10720);
  expect(occupied.msg).toContain(empA.employeeNo);
  expect(occupied.msg).toContain(empA.name);

  // 员工已关联其他账号时再关联 → 10722 已关联
  const already = await linkEmployeeUserRaw(apiCtx, token, empB.uuid, empA.userUuid);
  expect(already.code, '已关联拒绝 10722').toBe(10722);

  // 不写入：D 仍未关联，A 的关联未被夺走
  const dDetail = await getEmployeeByUuid(apiCtx, token, empD.uuid);
  expect(dDetail.userStatus, 'D 仍为未生成').toBe('0');
  expect(dDetail.userUuid, 'D 指针仍为空').toBe('');
  const aDetail = await getEmployeeByUuid(apiCtx, token, empA.uuid);
  expect(aDetail.userUuid, 'A 的关联未被变更').toBe(empA.userUuid);
});

/** 用户管理页搜索（与员工档案页共用 gotoListAndSearch） */
const gotoUserAndSearch = (page: Page, username: string) => gotoListAndSearch(page, USER_ROUTE, '用户名', username);

// ---------- 9. 用户管理「关联员工」列 ----------
test('用户管理「关联员工」列显示工号 + 姓名', async ({ page }) => {
  await loginAs(page, 'admin', '123456');
  await gotoUserAndSearch(page, empA.username);

  const target = row(page, empA.username).first();
  await expect(target).toBeVisible();
  const linkedCell = await cellByHeader(page, target, '关联员工');
  await expect(linkedCell).toHaveText(`${empA.employeeNo} ${empA.name}`);
});

// ---------- 10. 删除被引用账号被拒 ----------
test('用户管理删除被引用账号被拒，提示先解绑', async ({ page }) => {
  await loginAs(page, 'admin', '123456');
  await gotoUserAndSearch(page, empA.username);

  const target = row(page, empA.username).first();
  await expect(target).toBeVisible();
  await target.getByRole('button', { name: '删除' }).click();
  await page.locator('.n-popconfirm').locator('button:has-text("确认")').first().click();

  // 后端错误码消息经 toast 带出（紧跟点击断言）
  await expectMessage(page, '请先在员工档案中解除关联');

  // 账号未被删除
  const user = await getUserByUsername(apiCtx, token, empA.username);
  expect(user, '被引用账号不可删除').toBeTruthy();
});

// ---------- 11. 批量生成用户：生成 1 跳过 1 ----------
test('勾选 已生成+未生成 点「生成用户」，提示生成 1 跳过 1', async ({ page }) => {
  await loginAs(page, 'admin', '123456');
  await gotoArchiveAndSearch(page, EMPLOYEE_NAME_PREFIX + ts);

  // 同页勾选：A（已生成 → skip）+ E（在职未生成 → generate）
  await row(page, NAME_A).first().locator('.n-checkbox').first().click();
  await row(page, NAME_E).first().locator('.n-checkbox').first().click();

  await page.getByRole('button', { name: '生成用户' }).first().click();
  await expectMessage(page, '已生成 1 个系统账号'); // toast 紧跟点击
  await expectMessage(page, '跳过 1 个');

  // E 获得账号（用户名=工号），A 保持已生成
  const eDetail = await getEmployeeByUuid(apiCtx, token, empE.uuid);
  expect(eDetail.userStatus, 'E 已生成').toBe('1');
  expect(eDetail.loginName, 'E 登录名=工号').toBe(empE.employeeNo);
  const aDetail = await getEmployeeByUuid(apiCtx, token, empA.uuid);
  expect(aDetail.userStatus, 'A 未被重复生成').toBe('1');

  // 幂等回读：generateUser API 对 A/E 再跑一遍 → generated=0 skipped=2
  const idem = await generateUserViaApi(apiCtx, token, [empA.uuid, empE.uuid]);
  expect(idem.generated).toBe(0);
  expect(idem.skipped).toBe(2);
});

// ---------- 12. 非超管被授予按钮码 → 可见「生成用户」 ----------
test('非超管被授予按钮码的角色能看到「生成用户」按钮', async ({ page }) => {
  await loginAs(page, grantUser.username, grantUser.password);
  await page.goto(ARCHIVE_ROUTE);
  await expect(page.getByRole('button', { name: '生成用户' })).toBeVisible({ timeout: 15_000 });
  // 非超管且仅授了该按钮：未授码的其他按钮（新增）不可见，证明是精确授权而非超管旁路
  await expect(page.getByRole('button', { name: '新增' })).toHaveCount(0);
});
