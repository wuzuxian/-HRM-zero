# E2E 验收门禁：顶层 e2e/ 目录 + Playwright 分层测试

HRM 曾发生"带页面的需求验收时未做 UI 自动化操作，页面逻辑报错才被发现"的事故。决定：引入 Playwright（TypeScript）作为 UI 自动化框架，建立顶层 `e2e/` 目录（不放进 `web/`，避免污染前端构建），作为后续所有带页面需求的固定验收关卡。

测试分两层：

- **冒烟层**：全部 62 个 HRM 页面可打开、无 JS 控制台报错、核心列表渲染。目标是以最低成本抓"页面报错"类事故。
- **交互层**：真实走增删改/审批/查询等弹窗与按钮流程，按模块逐步铺开，从绩效-考核任务（assessment-task）开始。

运行环境直接使用本地 dev 栈（`go run` + SQLite 自动迁移，前端 `pnpm dev` :9527 代理 :8001，超管 admin/123456），不建独立测试库；交互测试自建数据、自清理。

Considered Options: Cypress（浏览器支持与并行能力不如 Playwright）、塞进 web/ 目录（会污染 soybean-admin 前端构建与依赖树）。

Consequences: `e2e/` 有独立 package.json，需 `pnpm install`；本地跑全量需要前后端同时启动。新需求验收流程增加一步：跑对应模块的 e2e spec 通过后才算验收。
