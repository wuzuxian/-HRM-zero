import { APIRequestContext } from '@playwright/test';

/**
 * 员工档案 ↔ 系统账号（建档即开号/关联/解绑）交互层的数据准备 helper。
 *
 * 全部走后端 API（:8001 直连，绕过前端代理）：
 *   - 建档即开号: addEmployee 在同一事务内经唯一接缝开号（AddResponse 带 userCreated/username/initialPassword，
 *     但不带 uuid → 从列表端点按唯一名反查）
 *   - 关联端点: linkEmployeeUser / unlinkEmployeeUser / getLinkableUserList / createAndLinkUser（员工侧）
 *   - 用户侧: manage/getUserList / manage/deleteUser / manage/getLinkedEmployees
 *   - 批量: generateUser（幂等，已关联跳过）
 *
 * 命名带 Date.now() 时间戳自建自清理：删员工即级联清账号（HRM-52）；
 * 无档案引用的孤儿账号/独立账号走 manage/deleteUser；历史 run 的员工置离职避免堆积。
 *
 * 产品事实（e2e 断言依赖）：
 *   - 离职（personStatus='5'）建档不开号（AddResponse.userCreated=false）→ 之后 editEmployee 改回
 *     正常状态即得到"在职且从未开号"的员工，这是 generateUser 幂等用例里"未生成"侧的正确造数路径
 *     （先开号再解绑的员工再次 generateUser 会撞孤儿账号用户名报 10461，那是既定设计——
 *      再关联的正确路径是候选搜索重关联）。
 *   - 解绑只清档案 user_uuid 指针，账号与角色绑定保留（账号仍可登录）。
 *   - 候选列表只返回未被任何档案占用的账号。
 */

/** 后端 API 直连基址（绕过前端代理）; perf-data 等其他造数 helper 复用同一处定义 */
export const API = process.env.E2E_API_URL ?? 'http://localhost:8001';

/** 员工测试数据统一前缀（历史置离线 + 清理时按名匹配） */
export const EMPLOYEE_NAME_PREFIX = 'e2e账号员工';
export const STANDALONE_USER_PREFIX = 'e2e候选账号';

export interface EmployeeRow {
  uuid: string;
  employeeNo: string;
  name: string;
  /** 登录用户名（=工号，开号后与工号解耦） */
  username: string;
  password: string;
  /** 关联账号 uuid（空串=未关联） */
  userUuid: string;
  /** 用户状态: 1=已生成 0=未生成 */
  userStatus: string;
}

export interface RawResponse {
  code: number;
  msg: string;
  data: unknown;
}

/** Bearer 认证头（apiGet/apiPost/apiGetRaw/apiPostRaw 的唯一出处） */
function authHeaders(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' };
}

/** POST 且不因错误码抛异常（用于断言 1072x 拒绝路径） */
async function apiPostRaw(request: APIRequestContext, path: string, token: string, data: unknown): Promise<RawResponse> {
  const resp = await request.post(`${API}/api/v1${path}`, { headers: authHeaders(token), data });
  return (await resp.json()) as RawResponse;
}

async function apiGetRaw(request: APIRequestContext, path: string, token: string, params?: Record<string, string>): Promise<RawResponse> {
  const resp = await request.get(`${API}/api/v1${path}`, { headers: authHeaders(token), params });
  return (await resp.json()) as RawResponse;
}

/** POST 且 code 非 200/0 抛异常（错误码断言走 apiPostRaw）; perf-data 等复用 */
export async function apiPost(request: APIRequestContext, path: string, token: string, data: unknown): Promise<unknown> {
  const json = await apiPostRaw(request, path, token, data);
  if (json.code !== 200 && json.code !== 0) {
    throw new Error(`POST ${path} 失败: code=${json.code} msg=${json.msg}`);
  }
  return json.data;
}

/** GET 且 code 非 200/0 抛异常; perf-data 等复用 */
export async function apiGet(request: APIRequestContext, path: string, token: string, params?: Record<string, string>): Promise<unknown> {
  const json = await apiGetRaw(request, path, token, params);
  if (json.code !== 200 && json.code !== 0) {
    throw new Error(`GET ${path} 失败: code=${json.code} msg=${json.msg}`);
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

// ---------------- 员工侧 ----------------

/**
 * API 建档（建档即开号由后端完成）。AddResponse 不带 uuid → 从列表按唯一名反查。
 * personStatus 传 '5' 时后端不开号（离职补录），可用于构造"在职未生成"员工。
 */
export async function addEmployeeViaApi(
  request: APIRequestContext,
  token: string,
  opts: { name: string; personStatus?: string }
): Promise<{ row: EmployeeRow; userCreated: boolean; initialPassword: string }> {
  const personStatus = opts.personStatus ?? '1';
  const addResp = (await apiPost(request, '/hrm/addEmployee', token, {
    name: opts.name,
    personStatus
  })) as { userCreated?: boolean; username?: string; initialPassword?: string };

  const row = await findEmployeeByName(request, token, opts.name);
  if (!row) throw new Error(`addEmployee 后列表反查不到员工: ${opts.name}`);
  return {
    row: { ...row, password: addResp.initialPassword ?? '123456' },
    userCreated: addResp.userCreated ?? false,
    initialPassword: addResp.initialPassword ?? '123456'
  };
}

export async function findEmployeeByName(
  request: APIRequestContext,
  token: string,
  name: string
): Promise<EmployeeRow | null> {
  const list = (await apiGet(request, '/hrm/getEmployeeList', token, { name, current: '1', size: '10' })) as {
    records: Array<{
      uuid: string;
      employeeNo: string;
      name: string;
      userStatus: string;
      userUuid: string;
      loginName: string;
    }>;
  };
  const hit = list.records?.find(r => r.name === name);
  if (!hit) return null;
  return {
    uuid: hit.uuid,
    employeeNo: hit.employeeNo,
    name: hit.name,
    username: hit.loginName || hit.employeeNo,
    password: '123456',
    userUuid: hit.userUuid,
    userStatus: hit.userStatus
  };
}

export async function getEmployeeByUuid(
  request: APIRequestContext,
  token: string,
  uuid: string
): Promise<{ userStatus: string; userUuid: string; loginName: string; personStatus: string }> {
  const detail = (await apiGet(request, '/hrm/getEmployeeDetail', token, { uuid })) as {
    employee: { userStatus: string; userUuid: string; loginName: string; personStatus: string };
  };
  return detail.employee;
}

export async function editEmployee(request: APIRequestContext, token: string, body: Record<string, unknown>): Promise<void> {
  await apiPost(request, '/hrm/editEmployee', token, body);
}

/** 解除关联（只清档案指针，账号保留） */
export async function unlinkEmployeeViaApi(request: APIRequestContext, token: string, employeeUuid: string): Promise<void> {
  await apiPost(request, '/hrm/unlinkEmployeeUser', token, { employeeUuid });
}

/** 关联已有账号 */
export async function linkEmployeeViaApi(
  request: APIRequestContext,
  token: string,
  employeeUuid: string,
  userUuid: string
): Promise<void> {
  await apiPost(request, '/hrm/linkEmployeeUser', token, { employeeUuid, userUuid });
}

/** 关联已有账号（原始响应，用于断言一对一冲突拒绝 10720/10722） */
export async function linkEmployeeUserRaw(
  request: APIRequestContext,
  token: string,
  employeeUuid: string,
  userUuid: string
): Promise<RawResponse> {
  return apiPostRaw(request, '/hrm/linkEmployeeUser', token, { employeeUuid, userUuid });
}

/** 候选账号列表（只含未被占用的） */
export async function getLinkableUsers(
  request: APIRequestContext,
  token: string,
  keyword: string
): Promise<Array<{ uuid: string; username: string; nickname: string }>> {
  const data = (await apiGet(request, '/hrm/getLinkableUserList', token, {
    keyword,
    current: '1',
    size: '50'
  })) as { records: Array<{ uuid: string; username: string; nickname: string }> };
  return data.records ?? [];
}

/** 批量生成用户（幂等：已关联跳过） */
export async function generateUserViaApi(
  request: APIRequestContext,
  token: string,
  uuids: string[]
): Promise<{ generated: number; skipped: number }> {
  return (await apiPost(request, '/hrm/generateUser', token, { uuids })) as { generated: number; skipped: number };
}

// ---------------- 用户侧（manage） ----------------

export async function getUserByUsername(
  request: APIRequestContext,
  token: string,
  username: string
): Promise<{ uuid: string; username: string; nickName: string; status: string } | null> {
  const list = (await apiGet(request, '/manage/getUserList', token, { username, current: '1', size: '10' })) as {
    records: Array<{ uuid: string; username: string; nickName: string; status: string }>;
  };
  return list.records?.find(u => u.username === username) ?? null;
}

/** 建一个与任何档案都无关的独立账号（候选搜索用） */
export async function createStandaloneUser(
  request: APIRequestContext,
  token: string,
  username: string,
  nickname: string
): Promise<{ uuid: string; username: string; password: string }> {
  await apiPost(request, '/manage/addUser', token, {
    username,
    nickName: nickname,
    userGender: '1',
    userPhone: '',
    userEmail: '',
    status: '1',
    password: '123456',
    userRoles: []
  });
  const user = await getUserByUsername(request, token, username);
  if (!user) throw new Error(`addUser 后反查不到账号: ${username}`);
  return { uuid: user.uuid, username, password: '123456' };
}

/** 删除账号（原始响应，用于断言"被引用拒绝"） */
export async function deleteUserRaw(
  request: APIRequestContext,
  token: string,
  uuids: string[]
): Promise<RawResponse> {
  return apiPostRaw(request, '/manage/deleteUser', token, { uuids });
}

/** 给账号挂角色（editUser.userRoles 按角色 code 匹配，与前端 roleOptions 一致） */
export async function editUserRoles(
  request: APIRequestContext,
  token: string,
  userUuid: string,
  username: string,
  nickname: string,
  roleCodes: string[]
): Promise<void> {
  await apiPost(request, '/manage/editUser', token, {
    uuid: userUuid,
    username,
    userGender: '1',
    nickName: nickname,
    userPhone: '',
    userEmail: '',
    status: '1',
    userRoles: roleCodes
  });
}

// ---------------- 角色（按钮码授权用例） ----------------

/**
 * 经 manage/getMenuTree 解析授权所需菜单 uuid（首页 + 目录链 + 指定页面 + 页内指定按钮）。
 * getMenuList 只返回顶级节点，按钮/页面要从全量树取；树节点只带 label/pUuid，按菜单名定位。
 */
export async function resolveMenuUuidsForPage(
  request: APIRequestContext,
  token: string,
  pageMenuLabel: string,
  buttonMenuLabel: string,
  homeMenuLabel = '首页'
): Promise<string[]> {
  interface TreeNode {
    uuid: string;
    label: string;
    pUuid: string;
    children: TreeNode[];
  }
  const tree = (await apiGet(request, '/manage/getMenuTree', token)) as TreeNode[];
  const byUuid = new Map<string, TreeNode>();
  const walk = (nodes: TreeNode[]) => {
    for (const n of nodes) {
      byUuid.set(n.uuid, n);
      if (n.children?.length) walk(n.children);
    }
  };
  walk(tree);

  const page = [...byUuid.values()].find(n => n.label === pageMenuLabel);
  if (!page) throw new Error(`菜单树找不到页面: ${pageMenuLabel}`);
  const button = page.children?.find(n => n.label === buttonMenuLabel);
  if (!button) throw new Error(`页面 ${pageMenuLabel} 下找不到按钮: ${buttonMenuLabel} (现有: ${page.children?.map(c => c.label).join(',')})`);
  const home = [...byUuid.values()].find(n => n.label === homeMenuLabel);

  const uuids = new Set<string>();
  // 页面 + 祖先目录链（动态路由只沿 role_menu 建树，目录不授子页面挂不上）
  let cur: TreeNode | undefined = page;
  while (cur) {
    uuids.add(cur.uuid);
    cur = cur.pUuid ? byUuid.get(cur.pUuid) : undefined;
  }
  uuids.add(button.uuid);
  if (home) uuids.add(home.uuid);
  return [...uuids];
}

/** 建角色 + 授菜单（casbin 由 setRoleMenus 同步） */
export async function createRoleWithMenus(
  request: APIRequestContext,
  token: string,
  roleName: string,
  roleCode: string,
  menuUuids: string[]
): Promise<string> {
  await apiPost(request, '/manage/addRole', token, {
    roleName,
    roleCode,
    roleDesc: 'e2e 临时角色',
    status: '1'
  });
  const list = (await apiGet(request, '/manage/getRoleList', token, { roleCode, current: '1', size: '10' })) as {
    records: Array<{ uuid: string; roleCode: string }>;
  };
  const role = list.records?.find(r => r.roleCode === roleCode);
  if (!role) throw new Error(`addRole 后反查不到角色: ${roleCode}`);
  await apiPost(request, '/manage/setRoleMenus', token, { roleUuid: role.uuid, menuUuids });
  return role.uuid;
}

// ---------------- 生命周期 ----------------

/**
 * 历史堆积处理：把此前 run 留下的测试员工置离职（不被新批次/新 run 圈定，列表搜索互不干扰）。
 */
export async function retireHistoricalEmployees(request: APIRequestContext, token: string): Promise<void> {
  const list = (await apiGet(request, '/hrm/getEmployeeList', token, { current: '1', size: '100' })) as {
    records: Array<{ uuid: string; name: string; personStatus: string }>;
  };
  for (const old of list.records ?? []) {
    if (old.name?.startsWith(EMPLOYEE_NAME_PREFIX) && old.personStatus !== '5') {
      await apiPost(request, '/hrm/editEmployee', token, {
        uuid: old.uuid,
        name: old.name,
        personStatus: '5'
      });
    }
  }
}

/**
 * 清理本 run 造的数据：
 *   1. 删员工（deleteEmployee 级联清关联账号与角色绑定，HRM-52）
 *   2. 删员工侧孤儿账号（用例中途解绑过 → 档案指针已清 → 级联删不到，须按用户名显式删；
 *      传全部 fixture 工号即可——仍被引用的删失败只记 warning，由第 1 步级联兜住）
 *   3. 删独立账号/临时角色
 * 单项失败只记录不中断（清理尽力而为，残留由时间戳命名隔离）。
 */
export async function cleanupFixtures(
  request: APIRequestContext,
  token: string,
  opts: {
    employeeUuids: string[];
    /** 可能残留的孤儿账号用户名（员工删完后按用户名反查 uuid 再删） */
    orphanUsernames: string[];
    standaloneUserUuids: string[];
    roleUuids: string[];
  }
): Promise<string[]> {
  const warnings: string[] = [];
  const step = async (label: string, fn: () => Promise<void>) => {
    try {
      await fn();
    } catch (err) {
      warnings.push(`${label}: ${err instanceof Error ? err.message : String(err)}`);
    }
  };

  // 1. 删员工（级联清账号）
  if (opts.employeeUuids.length > 0) {
    await step('删员工', async () => {
      const raw = await apiPostRaw(request, '/hrm/deleteEmployee', token, { uuids: opts.employeeUuids });
      if (raw.code !== 200 && raw.code !== 0) throw new Error(`code=${raw.code} msg=${raw.msg}`);
    });
  }

  // 2. 孤儿账号清扫（档案已删，账号此时必然无引用，可删）
  for (const username of opts.orphanUsernames) {
    await step(`删孤儿账号 ${username}`, async () => {
      const user = await getUserByUsername(request, token, username);
      if (user) {
        const raw = await deleteUserRaw(request, token, [user.uuid]);
        if (raw.code !== 200 && raw.code !== 0) throw new Error(`code=${raw.code} msg=${raw.msg}`);
      }
    });
  }

  // 3. 删独立账号
  if (opts.standaloneUserUuids.length > 0) {
    await step('删独立账号', async () => {
      const raw = await deleteUserRaw(request, token, opts.standaloneUserUuids);
      if (raw.code !== 200 && raw.code !== 0) throw new Error(`code=${raw.code} msg=${raw.msg}`);
    });
  }

  // 4. 删临时角色
  for (const roleUuid of opts.roleUuids) {
    await step(`删角色 ${roleUuid}`, async () => {
      await apiPost(request, '/manage/deleteRole', token, { uuids: [roleUuid] });
    });
  }

  return warnings;
}
