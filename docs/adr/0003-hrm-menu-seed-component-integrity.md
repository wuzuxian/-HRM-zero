# HRM 动态路由菜单种子必须引用前端真实存在的组件

日期：2026-11-16 · 状态：已采纳 · 关联：迁移 47→57、bug 走查（绩效子菜单 No match / 假期余额 v-auth / 排班表 parentNode null）

## 背景

绩效管理「考核任务」及其 5 个子菜单点击全部报 `No match for {"name":"hrm_performance_assessment-task_plan-list"}`，「绩效结果」页面同源 403。根因是**数据侧种子与前端组件映射脱节**，且失败发生在前端动态路由装配的深链路上，症状与根因相距甚远。

## 决策

1. **目录菜单（menu_type='1'）的 component 必须为 `layout.base`**；只有叶子菜单（menu_type='2'）才允许 `view.*`。`view.*` 的值必须严格等于前端 `web/src/views/` 下真实存在的目录路径（`imports.ts` 的 views 映射由 elegant-router 从文件系统生成，文件不存在 = 映射无此 key）。
2. 每张 HRM 菜单种子迁移提交前，跑一次**种子 ↔ views 映射交叉检查**（脚本见本文末尾），确保每个 `view.*` 都能解析到真实组件。
3. 前端插件目录 `web/src/plugins/` 是全局指令（`app.directive`）的唯一注册点，新增指令经 `plugins/index.ts` re-export + `main.ts` 调用 `setup*Directives(app)` 注册，且注册时机必须在 `setupStore` 之后（指令内部依赖 pinia store）。

## 教训（为什么这类 bug 值得记录）

### 1. transform 的 try/catch 会吞掉整个子树

`web/src/router/elegant/transform.ts` 的 `transformElegantRouteToVueRoute` 在 `getViewName` 抛 `View component not found` 时 `return []` —— **丢弃的不只是出错的路由，还有它下面挂的全部子路由**。于是一个目录菜单 component 种错（`view.hrm_performance_assessment-task`，而 `assessment-task/` 下没有 index.vue），导致 5 个完全正常的子菜单路由集体未注册，控制台却只显示子菜单的 `No match for {name:...}`，从报错完全看不出父路由有问题。

排查手法：症状路由名 → `transform.ts` 的 `routeMap` 找到其父链 → 逐级检查父 component 是否能解析（`imports.ts` 是否有该 key）。全仓交叉比对（见下）能一次性暴露所有同类隐患。

### 2. 指令未注册是静默的权限失效

`v-auth` 在 16 个 HRM 页面用了 41 处，但全仓没有任何 `app.directive('auth', ...)` 注册。Vue 对未注册指令只发一条 `[Vue warn]: Failed to resolve directive: auth` **警告而非错误**，页面照常渲染——实际后果是权限按钮**全部恒渲染**，权限控制彻底失效，而走查时容易被当作纯警告忽略。反过来也一样：新增页面写 `v-auth` 前先确认指令已注册，否则按钮显示与否只取决于运气。

修复：`web/src/plugins/directives.ts` 注册全局 `auth` 指令，无权限时 `mounted` 移除元素；判定复用 `useAuth().hasAuth`（authStore.userInfo.buttons）。注意 TSX 渲染函数里（如表格列 render）指令同样生效。

### 3. 延时回调访问已卸载 DOM

`layouts/modules/global-tab/index.vue` 的 `scrollToActiveTab` 在 `setTimeout(…, 50)` 回调里读取 tab 节点；路由快速切换时节点已 detach，抛 `Cannot read properties of null (reading 'parentNode')`。任何「先测量 DOM 再延时滚动/聚焦」的代码都要在回调入口检查 `el.isConnected`。另外该报错在此案例里是被 bug 1 的路由失败连锁放大的——根因修掉后频次大降，防御代码保留兜底。

### 4. v-auth 指令在注册时不能碰 pinia/router

指令工厂若在 `setupAppDirectives(app)`（`setupRouter` 之前）执行 `useAuth()` → 提前创建 auth store → store 内 `useRoute()` 拿不到路由注入，控制台报 `injection "Symbol(route location)" not found` 且 store 带着坏的 route 引用存活。规则：**指令的 store/路由访问一律放进 `mounted` 等生命周期钩子**（此时组件上下文的注入可用），注册阶段只做 `app.directive(name, def)`。

### 5. NDatePicker 的 formatted-value 绝不能是空串

`v-model:formatted-value` 绑定的值只能是 `null`（未选择）或合法格式串；`''` 会让 naive-ui 在 setup 期 `format("")` 抛 `RangeError: Invalid time value`，组件崩溃 → 整页空白 → KeepAlive 状态被污染，**后续打开的页面也全部空白**。本项目踩过两处：排班表 `month = ref('')`（页面常驻直接崩）和 6 个弹窗 `xxxDate: ''` 默认值（打开弹窗即崩，且编辑回填 DB `NOT NULL DEFAULT ''` 列时同样塞进空串）。约定：

- 弹窗 model 的日期字段初值与编辑回填一律 `|| null` 归一；
- TS 类型层把对应请求字段放宽为 `string | null`（视图层 `Omit & { x: string | null }`）；
- 提交时 `?? ''` 适配回 API 类型（后端 Go `string` 字段把 JSON `null` 解码为 `""`，校验语义不变；可选字段用 `?? undefined` 走 omitempty）；
- naive-ui `required: true` 规则对 `null` 照样拦截，客户端必填校验不受影响。

## 验证

- 交叉检查脚本（模拟迁移 57 应用后）输出 `remaining missing view components: NONE`。
- `npm run typecheck` ✓、`npm run build` ✓、server `go build ./...` ✓。
- 手工验证路径：登录 → 绩效管理 → 考核任务 5 个子菜单可点击进入；假期余额页控制台无 `auth` 指令警告且无权限账号看不到调整/初始化按钮；排班表切换页签不再抛 parentNode null。

## 附：种子交叉检查脚本

```bash
python3 - <<'EOF'
import re, glob
imp = open('web/src/router/elegant/imports.ts').read()
views = set(k.strip('"\'') for k in re.findall(r'([\"\']?[\w\-]+[\"\']?)\s*:\s*\(\)', imp))
bad = []
for f in sorted(glob.glob('server/desc/sql_migration/sqlite/*.up.sql')):
    txt = open(f).read()
    if 'manage_menu' not in txt: continue
    for m in re.finditer(r"'view\.([\w-]+)'", txt):
        if m.group(1) not in views:
            bad.append((f.split('/')[-1], m.group(0)))
for f, c in bad: print(f, c, 'MISSING')
print('total:', len(bad))
EOF
```

输出非 0 即红灯（新种子必须先补前端页面文件，或改 component 语义）。
