-- fix: 考核任务菜单种子的 API 权限码与生成路由映射不一致。
--   route2code.go（生成产物）把 GET /api/v1/hrm/getPerfAssessmentDetail
--   映射为 v1:hrm:perf_assessment:getDetail（与 perf_template:getDetail /
--   employee:detail 等全仓惯例一致），而迁移 47 种子写的是
--   v1:hrm:perf_assessment:detail（少 get 前缀）→ casbin 无此码，
--   超管调 getPerfAssessmentDetail 也 403。
--   后果：自评/评审/校准/详情弹窗全部拉不到模板快照（弹窗 items 空），
--   考核填报链在 UI 上不可用。
--   同时更新 manage_menu.permissions 与 casbin_rule 存量行。

UPDATE manage_menu SET permissions = REPLACE(permissions, '"v1:hrm:perf_assessment:detail"', '"v1:hrm:perf_assessment:getDetail"')
WHERE permissions LIKE '%v1:hrm:perf_assessment:detail%';

UPDATE casbin_rule SET v1 = 'v1:hrm:perf_assessment:getDetail'
WHERE v1 = 'v1:hrm:perf_assessment:detail';
