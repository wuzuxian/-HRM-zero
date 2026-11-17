# 统一镜像：Go 进程托管前端静态资源，弃用 nginx 反代层

部署 jzero-admin 需要"一个镜像跑起前后端"的能力（交付、演示、单机部署）。决定：构建统一镜像——前端 dist 打进镜像，由 Go 后端进程在同一个 8001 端口直接托管，删除 nginx 这一进程与反代层。

实现契约：

- 后端以 `WEB_DIST` 环境变量开关静态托管（`server/internal/static`）：目录存在且含 `index.html` → 注册静态服务 + SPA history 路由回退（非 `/api` 未匹配路径回 index.html）；目录缺失或变量为空 → 行为与纯 API 服务完全一致（向后兼容，旧 server 镜像不受影响）。
- `/api` 前缀下的未匹配路径保持 JSON 404，不落入 SPA 回退。
- Dockerfile 在 `deploy/docker/Dockerfile`，build context 为仓库根；`scripts/build-docker-image.sh` 只管构建（latest + git sha 双 tag，`--api-only` 跳过前端），运行统一走 `deploy/docker/docker-compose.yaml`。
- "数据库可选"落在部署层而非镜像层：镜像永远内置 sqlite + miniredis 零依赖可跑；mysql/postgres/redis 通过 compose override 文件按需叠加并切换 `DRIVER_NAME`/`DATASOURCE` 等环境变量。

Considered Options: 镜像内 nginx + Go 双进程（零代码改动，但引入进程守护、镜像更大、端口与反代层多一跳，与"一个镜像"的目标相悖）；Go embed 把前端产物编进二进制（后端编译被迫依赖前端构建顺序，且无法单独替换前端文件）；构建期决定带不带数据库（把 DB 二进制打进镜像，违背镜像/数据分离）。

Consequences: 后端新增一个小的静态托管面（约 60 行 + 单测），`rest.WithNotFoundHandler` 与既有 handler 注册共存（已注册路由优先）。前端生产构建本就同源请求（`VITE_SERVICE_BASE_URL` 为空、API 路径绝对 `/api`），无需改前端。旧的 `server/Dockerfile`、`web/Dockerfile`、`deploy/docker-compose/` 全套保留，服务分离部署场景继续可用。
