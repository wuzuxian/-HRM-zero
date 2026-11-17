import { APIRequestContext, Page } from '@playwright/test';

/**
 * 绩效考核任务交互层的数据准备 helper。
 * 全部走后端 API（:8001 直连，绕过前端代理）：
 *   指标 → 模板 → 方案 → 员工(+账号+角色) → 批次 → 发布 → 生成考核任务
 * 依据 server/desc/api/v1/hrm/{perf_indicator,perf_template,perf_scheme,perf_period}.api。
 */

const API = process.env.E2E_API_URL ?? 'http://localhost:8001';

export const R_SUPER_ROLE_CODE = 'R_SUPER'; // editUser.userRoles 按角色 code 匹配（与前端 roleOptions 一致）

export interface AdminCtx {
  token: string;
}

export interface EmployeeFixture {
  employeeUuid: string;
  employeeNo: string;
  name: string;
  /** 系统账号（用户名=工号，初始密码 123456） */
  username: string;
  password: string;
}

export interface AssessmentFixture {
  periodUuid: string;
  periodName: string;
  employees: EmployeeFixture[];
}

async function apiGet(request: APIRequestContext, path: string, token: string, params?: Record<string, string>) {
  const resp = await request.get(`${API}/api/v1${path}`, {
    headers: { Authorization: `Bearer ${token}` },
    params
  });
  const json = await resp.json();
  if (json.code !== 200 && json.code !== 0) {
    throw new Error(`GET ${path} 失败: code=${json.code} msg=${json.msg}`);
  }
  return json.data;
}

async function apiPost(request: APIRequestContext, path: string, token: string, data: unknown) {
  const resp = await request.post(`${API}/api/v1${path}`, {
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    data
  });
  const json = await resp.json();
  if (json.code !== 200 && json.code !== 0) {
    throw new Error(`POST ${path} 失败: code=${json.code} msg=${json.msg}`);
  }
  return json.data;
}

/** admin 登录拿 token */
export async function adminToken(request: APIRequestContext): Promise<string> {
  const resp = await request.post(`${API}/api/v1/auth/pwd-login`, {
    data: { username: 'admin', password: '123456' }
  });
  const json = await resp.json();
  return json.data.token;
}

/**
 * 创建一条完整考核任务链数据（幂等可重跑：用带时间戳的命名避免撞唯一键）。
 * @param employeeCount 参与考核的员工数（每人一个系统账号，密码 123456）
 */
export async function setupAssessmentData(request: APIRequestContext, employeeCount = 2): Promise<AssessmentFixture> {
  const token = await adminToken(request);
  const ts = Date.now().toString().slice(-8);

  // 1. 组织：优先取已有节点；全新库（迁移不带业务种子数据）没有组织时自建一家公司，
  //    保证 spec 在空库上自足可跑。每 run 独立公司 → 批次 scope 圈定互不串扰。
  const orgTree = await apiGet(request, '/hrm/getOrgTree', token);
  let companyUuid = orgTree?.records?.[0]?.uuid ?? orgTree?.[0]?.uuid;
  if (!companyUuid) {
    await apiPost(request, '/hrm/addOrg', token, {
      orgName: `e2e绩效公司${ts}`,
      orgType: '1',
      status: '1'
    });
    const recreated = await apiGet(request, '/hrm/getOrgTree', token);
    companyUuid = recreated?.records?.[0]?.uuid ?? recreated?.[0]?.uuid;
    if (!companyUuid) throw new Error('无组织节点且自建公司失败，无法圈定批次参与人');
  }

  // 2. 指标（AddResponse 为空，创建后从列表反查）
  const indicatorCode = `E2E-KPI-${ts}`;
  await apiPost(request, '/hrm/addPerfIndicator', token, {
    code: indicatorCode,
    name: `e2e考核指标${ts}`,
    indicatorType: 'kpi',
    scoreRule: 'manual',
    dataSource: 'manual',
    defaultWeight: 100
  });
  const indicatorList = await apiGet(request, '/hrm/getPerfIndicatorList', token, { code: indicatorCode, current: '1', size: '10' });
  const indicatorUuid = indicatorList?.records?.[0]?.uuid;
  if (!indicatorUuid) throw new Error('找不到刚创建的指标');

  // 3. 模板（单指标权重 100）
  await apiPost(request, '/hrm/addPerfTemplate', token, {
    name: `e2e考核模板${ts}`,
    items: [{ indicatorUuid, weight: 100 }]
  });
  const templateList = await apiGet(request, '/hrm/getPerfTemplateList', token, { current: '1', size: '50' });
  const templateUuid = templateList?.records?.find((t: { name: string }) => t.name === `e2e考核模板${ts}`)?.uuid;
  if (!templateUuid) throw new Error('找不到刚创建的模板');

  // 4. 方案
  await apiPost(request, '/hrm/addPerfScheme', token, {
    name: `e2e考核方案${ts}`,
    templateUuid,
    periodType: 'month',
    assessmentMode: 'kpi',
    scoreMode: 'percent',
    payLinkMode: 'coefficient'
  });
  const schemeList = await apiGet(request, '/hrm/getPerfSchemeList', token, { current: '1', size: '50' });
  const schemeUuid = schemeList?.records?.find((s: { name: string }) => s.name === `e2e考核方案${ts}`)?.uuid;
  if (!schemeUuid) throw new Error('找不到刚创建的方案');

  // 5. 先把历史 e2e 员工置为离职（person_status='5' 不被 scope 圈定），
  //    否则每轮 run 的员工都会进入新批次，任务列表越滚越大翻不到目标行
  const oldEmpList = await apiGet(request, '/hrm/getEmployeeList', token, { current: '1', size: '100' });
  for (const old of oldEmpList?.records ?? []) {
    if (old.name?.startsWith('e2e考核员工') && old.personStatus !== '5') {
      await apiPost(request, '/hrm/editEmployee', token, {
        uuid: old.uuid,
        name: old.name,
        personStatus: '5'
      });
    }
  }

  // 5b. 员工 + 账号 + 角色
  // HRM-50 建档即开号收口: addEmployee 在同一事务内自动开号（AddResponse 带 userCreated），
  // 不再走 generateUser 两步造数; 绩效页面确需管理端权限，授角色这一步保留。
  const employees: EmployeeFixture[] = [];
  for (let i = 0; i < employeeCount; i++) {
    const empName = `e2e考核员工${ts}_${i}`;
    const addResp = (await apiPost(request, '/hrm/addEmployee', token, {
      name: empName,
      companyUuid,
      orgUuid: companyUuid,
      personStatus: '1'
    })) as { userCreated?: boolean; username?: string; initialPassword?: string };
    if (!addResp?.userCreated) {
      throw new Error('addEmployee 未自动开号（userCreated=false）: ' + JSON.stringify(addResp));
    }

    const empList = await apiGet(request, '/hrm/getEmployeeList', token, { name: empName, current: '1', size: '10' });
    const empRow = empList?.records?.[0];
    const employeeUuid = empRow?.uuid;
    const employeeNo = empRow?.employeeNo;
    if (!employeeUuid || !employeeNo) throw new Error('getEmployeeList 未返回 uuid/employeeNo: ' + JSON.stringify(empList));
    if (empRow?.userStatus !== '1' || empRow?.loginName !== employeeNo) {
      throw new Error(`员工未关联账号或登录名≠工号: userStatus=${empRow?.userStatus} loginName=${empRow?.loginName}`);
    }

    // 授予超管角色（让账号能看见考核任务菜单与按钮码）
    const userList = await apiGet(request, '/manage/getUserList', token, { username: employeeNo, current: '1', size: '10' });
    const userUuid = userList?.records?.[0]?.uuid;
    if (!userUuid) throw new Error(`找不到员工账号 ${employeeNo}`);
    await apiPost(request, '/manage/editUser', token, {
      uuid: userUuid,
      username: employeeNo,
      userGender: '1',
      nickName: empName,
      userPhone: '',
      userEmail: '',
      status: '1',
      userRoles: [R_SUPER_ROLE_CODE]
    });

    employees.push({ employeeUuid, employeeNo, name: empName, username: employeeNo, password: '123456' });
  }

  // 6. 批次（scope 按公司圈定，状态只含正式员工；AddResponse 为空，创建后从列表反查）
  const periodName = `e2e考核批次${ts}`;
  await apiPost(request, '/hrm/addPerfPeriod', token, {
    name: periodName,
    periodType: 'month',
    schemeUuid,
    scope: JSON.stringify({ companyUuids: [companyUuid], orgUuids: [], positions: [], posts: [], personStatuses: ['1'] }),
    periodStart: '2026-11-01',
    periodEnd: '2026-11-30'
  });
  const periodList = await apiGet(request, '/hrm/getPerfPeriodList', token, { current: '1', size: '50' });
  const periodUuid = periodList?.records?.find((p: { name: string }) => p.name === periodName)?.uuid;
  if (!periodUuid) throw new Error('找不到刚创建的批次');

  // 7. 发布 → 生成考核任务
  await apiPost(request, '/hrm/publishPerfPeriod', token, { uuid: periodUuid });
  const genResult = await apiPost(request, '/hrm/generatePerfAssessment', token, { periodUuid });
  if ((genResult?.generatedCount ?? 0) < employeeCount) {
    throw new Error('generatePerfAssessment 生成数不足: ' + JSON.stringify(genResult));
  }

  // 8. 员工1（employeeCount >= 2 时）API 快速推链到节点 5（进行中），
  //    供申诉/裁决用例使用；其 1→2→3→4 的 UI 交互已由员工0 的用例覆盖
  if (employeeCount >= 2) {
    const list = await apiGet(request, '/hrm/getPerfAssessmentList', token, { periodUuid, current: '1', size: '50' });
    const emp1Task = list?.records?.find((r: { employeeName: string }) => r.employeeName === employees[1].name);
    if (!emp1Task?.uuid) throw new Error('找不到员工1 的考核任务: ' + JSON.stringify(list));
    const indicatorUuid = indicatorList.records[0].uuid;

    // 本人操作：目标确认（1→2）→ 自评（2→3）
    const emp1Token = await userToken(request, employees[1].username, employees[1].password);
    await apiPost(request, '/hrm/confirmPerfAssessmentPlan', emp1Token, { uuid: emp1Task.uuid });
    await apiPost(request, '/hrm/submitPerfAssessmentSelfReview', emp1Token, {
      uuid: emp1Task.uuid,
      items: [{ indicatorUuid, score: 70 }],
      comment: 'e2e 员工1 自评'
    });
    // admin 操作：评审（3→4）→ 校准（4→5）
    await apiPost(request, '/hrm/submitPerfAssessmentLeaderReview', token, {
      uuid: emp1Task.uuid,
      items: [{ indicatorUuid, score: 65 }],
      comment: 'e2e 员工1 评审'
    });
    await apiPost(request, '/hrm/submitPerfAssessmentCalibration', token, {
      uuid: emp1Task.uuid,
      finalScore: 60,
      comment: 'e2e 员工1 校准（压低分引发申诉）'
    });
  }

  return { periodUuid, periodName, employees };
}

/** 任意账号 API 登录拿 token */
async function userToken(request: APIRequestContext, username: string, password: string): Promise<string> {
  const resp = await request.post(`${API}/api/v1/auth/pwd-login`, {
    data: { username, password }
  });
  const json = await resp.json();
  if (json.code !== 200 && json.code !== 0) throw new Error(`账号 ${username} 登录失败: ${json.msg}`);
  return json.data.token;
}

/** UI 登录（通用：任意账号），等待进入系统 */
export async function loginAs(page: Page, username: string, password: string): Promise<void> {
  await page.goto('/');
  await page.waitForURL(/\/(login|home)/, { timeout: 15_000 }).catch(() => {});
  if (!page.url().includes('/login')) {
    // 已是其他登录态：先登出（直接清 storage 再刷新最稳）
    await page.evaluate(() => localStorage.clear());
    await page.goto('/');
    await page.waitForURL(/\/login/, { timeout: 15_000 });
  }
  await page.locator('input[placeholder*="请输入用户名"]').first().fill(username);
  await page.locator('input[type="password"]').first().fill(password);
  await page.locator('button:has-text("确认")').first().click();
  await page.waitForFunction(
    (u) => !!localStorage.getItem('JZERO_ADMIN_token') && localStorage.getItem('JZERO_ADMIN_username') === u,
    username,
    { timeout: 15_000 }
  ).catch(async () => {
    // 有些版本不存 username，退化为只等 token
    await page.waitForFunction(() => !!localStorage.getItem('JZERO_ADMIN_token'), { timeout: 10_000 });
  });
  await page.waitForTimeout(1000);
}
