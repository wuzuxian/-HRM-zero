---
name: e2e-module-acceptance
description: "HRM/后台模块的 UI 交互层验收测试工作流：用 Playwright 走真实增删改/审批链路（API 造数 + UI 操作），抓'单测过、页面挂'的集成断裂 bug 并修复。当用户要求给某模块写 UI 自动化/E2E/交互层测试、或带页面的新需求要做验收时使用。前置：e2e/ 基建已存在（ADR 0004）。"
---

# 模块 UI 交互层验收（E2E）

依据 [ADR 0004](../../../docs/adr/0004-e2e-acceptance-gate.md)：带页面的需求，交互层 spec 通过才算验收完成。本 skill 是给**一个模块**（如考勤、薪酬、培训）铺交互层测试的完整工作流。首个落地样例：`e2e/tests/performance/assessment-task.spec.ts`（考核任务全链路 9 步）。

## 前置检查

1. `e2e/` 目录存在且 `pnpm install` 完成、chromium 已装；不存在则先按 [e2e/README.md](../../../e2e/README.md) 搭建。
2. 本地栈在跑：后端 `cd server && go run main.go server`（:8001，SQLite 自动迁移），前端 `cd web && pnpm dev`（:9527）。用 `curl -s -o /dev/null -w "%{http_code}" localhost:8001/ 与 localhost:9527/` 探活。
3. 冒烟层对该模块页面全绿（`pnpm test:smoke`）。冒烟红先修冒烟，不做交互。

## Step 1 — 调查模块 API 链路（派 Explore 子代理）

交互层需要真实数据前置。派一个 Explore 子代理回答：

1. 模块主数据（相当于指标/模板/方案）各创建端点：POST 路径 + 必填字段 + **AddResponse 是否返回 uuid**（本仓惯例 AddResponse 为空，须创建后从列表端点反查）。
2. 业务流端点：状态机每步的端点、守卫（本人/权限码）、错误码段。
3. 参与人/归属如何圈定（scope/组织/员工状态），admin 无员工档案时哪些操作会被"本人校验"拒绝（→ 需要造员工账号：`generateUser`（用户名=工号、密码 123456）+ `manage/editUser` 授 `R_SUPER`，**userRoles 传角色 code 而非 uuid**）。
4. 列表口径：返回全部还是当前登录人相关。

## Step 2 — 数据准备 helper（`e2e/helpers/<module>-data.ts`）

铁律（全部来自实战教训）：

- **API 造数，UI 只做被测操作**。链路前置推进（如把任务推到某节点）也是造数，用 API 做；UI 只测真正的交互面。
- 命名带 `Date.now()` 时间戳避免撞唯一键；**自建数据自清理**（员工类先置离职 `personStatus='5'` 再建新数据，否则 scope 会把历次 run 的员工都圈进来，列表越滚越大翻不到目标行）。
- helper 返回 fixture 对象（uuid/username/单号等），spec 的 `test.beforeAll` 里调用一次。
- API 判定：`json.code !== 200 && json.code !== 0` 即抛错（带 path/msg），不要只看 HTTP 状态——本仓业务失败也是 HTTP 200。

## Step 3 — 交互 spec（`e2e/tests/<module>/<flow>.spec.ts`）

模式（照 `assessment-task.spec.ts`）：

- `test.describe.configure({ mode: 'serial' })`，链路步骤 = 多个 test()，fixture 挂模块级变量。
- 每个测试先 `loginAs(page, username, password)`（contexts 间 localStorage 不共享）。
- 列表页导航用带过滤的 helper：**先填过滤条件（批次/单号等）再搜索**，不要裸遍历翻页。
- 行定位 `page.locator('.n-data-table tbody tr', { hasText: 唯一文本 })`；按钮/弹窗 scoped 到行/弹窗，避免 strict mode 命中多个元素。
- 断言成功用 toast：`page.locator('.n-message', { hasText: '更新成功' })`；节点推进类断言用节点 tag 文本或行消失，**不要在错误页面断言别的页面的按钮**。
- 断言文案先查 `web/src/locales/langs/zh-cn.ts`（按钮/弹窗/toast 全是 i18n key，中文文案以 locale 为准；登录按钮文本是 common.confirm="确认"，不是"登录"）。

## Step 4 — 运行与调试循环

`pnpm exec playwright test tests/<module>/ -g <用例名>` 单跑。每次失败按此分类，**不要上来就改产品代码**：

1. **测试 bug**：选择器 strict mode、弹窗未等、断言放错页面、横向滚动没做。改测试。
2. **环境噪音**：vite 首访按需编译超时（重试即过）、iconify 外网图标请求失败。`--retries=1` 消化。
3. **真 bug**：截图 + DB 直查（`sqlite3 server/jzero-admin.db`）+ 后端日志三方取证，确认后修对层级：
   - 前端逻辑 → 改 vue 组件，等 vite 热重载后重跑；
   - 后端逻辑 → 改 go，`go build -o /tmp/hrm-server-e2e .` 重启重跑；
   - 种子数据 → 新增迁移 `server/desc/sql_migration/{sqlite,pgx,mysql}/NN_desc.up.sql`（三方言同步，看 57-59 的注释风格），**不改已应用的旧迁移**。

## Step 5 — 回归与交付

1. 本模块 spec 全绿。
2. `tests/bugs/page-transition-blank.spec.ts` 全绿（前端改动可能碰布局/路由）。
3. `pnpm test:smoke` 全绿（改动跨多页时跑全量）。
4. 交付报告：测试结果表 + 真 bug 清单（严重度/现象/根因/修复文件）+ "模式总结"（哪些是单测抓不到的集成断裂）。

## 本仓已知坑清单（写测试与修 bug 前先过一遍）

- **Vue Transition 空白**：页面组件 template 根部不能放 HTML 注释（dev 模式注释使组件根成 fragment，`out-in` 离场后新页永不进场，仅 F5 可恢复）。新页面一律把注释写在根 div 内部。
- **权限码双消费者**：菜单 `permissions` JSON 同时喂 casbin（API 码 = route2code 生成映射，如 `submitLeaderReview`）和后端业务守卫常量（`errors.go` 的 `PermAssessment*`）。两处必须同名，新增端点时核对 `server/internal/handler/route2code.go`。
- **configcenter 不应用 struct default**：配置默认值不能只靠 `json:",default="` 标签，代码里要空值兜底（见 generate_user.go）。
- **AddResponse 全空**：创建类端点不回 uuid，一律创建后从列表端点按唯一名反查。
- **NMessage 约 3s 自动关闭**：toast 断言要在点击后立即轮询（expect 默认轮询没问题，别在中间插长 sleep）。
- **表格横向滚动**：操作列在 scroll-x 最右端时，click 前先 `evaluate` 把滚动容器 `scrollLeft` 拉满。
- **登录时序**：`goto('/')` 后 SPA 异步重定向到 /login，`loginAs` 里先 `waitForURL(/\/(login|home)/)` 再判断；token 键名 `JZERO_ADMIN_token`。
- **多 tab 长会话**：连续开很多 tab 后的异常优先怀疑 Transition/KeepAlive，刷新只恢复当前页——写长链路测试时用每测试独立登录规避状态累积。
