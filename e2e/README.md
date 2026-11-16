# e2e — HRM UI 自动化测试（验收门禁）

依据 [ADR 0004](../docs/adr/0004-e2e-acceptance-gate.md)。

## 前置

1. 后端：`cd server && go run main.go server`（:8001，SQLite 自动迁移）
2. 前端：`cd web && pnpm dev`（:9527，代理 :8001）
3. 浏览器：`cd e2e && pnpm install && pnpm exec playwright install chromium`

## 运行

```bash
pnpm test            # 全部
pnpm test:smoke      # 冒烟层（62 个 HRM 页面）
pnpm test:assessment # 考核任务交互层
pnpm report          # 打开 HTML 报告
```

环境变量：`E2E_BASE_URL`（默认 http://localhost:9527）、`E2E_USERNAME`/`E2E_PASSWORD`（默认 admin/123456）。

## 结构

```
e2e/
├── playwright.config.ts
├── helpers/
│   ├── auth.ts          # 登录、导航、console/5xx 错误收集器
│   └── hrm-pages.ts     # 62 页路由清单（与 views/hrm 同步维护）
├── tests/
│   ├── smoke/           # 冒烟层：每页可达、无报错
│   └── performance/     # 交互层：考核任务等真实操作流程
└── results/             # 报告与失败工件（gitignore）
```

## 验收约定

带页面的新需求：先补/跑对应模块的交互层 spec，通过才算验收完成。
给某个模块铺交互层测试的完整工作流见 skill：`.zcode/skills/e2e-module-acceptance/SKILL.md`（API 造数模式、调试分类法、本仓已知坑清单）。
