-- fix: 绩效结果菜单种子的 API 权限码与 route2code.go（生成产物）不一致（58/59 的同类遗留）。
--   route2code.go 映射：
--     GET  /api/v1/hrm/getPerfResultList       → v1:hrm:perf_result:resultList
--     POST /api/v1/hrm/signPerfResult          → v1:hrm:perf_result:signResult
--     POST /api/v1/hrm/submitPerfResultApproval → v1:hrm:perf_result:submitApproval
--   而迁移 52 种子把前端按钮码命名（list/sign/submit）误用作 API 码：
--     v1:hrm:perf_result:list / sign / submit
--   → casbin 无 resultList/signResult/submitApproval，绩效结果页列表 403、
--     签收与审批提交均被网关拒绝（纯后端单测不经过 Authx 发现不了）。
--   仅改 permissions JSON 内的 API 码；button_code 裸码（前端 v-auth 用）不动。

UPDATE manage_menu SET permissions = REPLACE(permissions, '"v1:hrm:perf_result:list"', '"v1:hrm:perf_result:resultList"')
WHERE permissions LIKE '%v1:hrm:perf_result:list%';

UPDATE manage_menu SET permissions = REPLACE(permissions, '"v1:hrm:perf_result:sign"', '"v1:hrm:perf_result:signResult"')
WHERE permissions LIKE '%v1:hrm:perf_result:sign%';

UPDATE manage_menu SET permissions = REPLACE(permissions, '"v1:hrm:perf_result:submit"', '"v1:hrm:perf_result:submitApproval"')
WHERE permissions LIKE '%v1:hrm:perf_result:submit%';

UPDATE casbin_rule SET v1 = 'v1:hrm:perf_result:resultList'
WHERE v1 = 'v1:hrm:perf_result:list';

UPDATE casbin_rule SET v1 = 'v1:hrm:perf_result:signResult'
WHERE v1 = 'v1:hrm:perf_result:sign';

UPDATE casbin_rule SET v1 = 'v1:hrm:perf_result:submitApproval'
WHERE v1 = 'v1:hrm:perf_result:submit';
