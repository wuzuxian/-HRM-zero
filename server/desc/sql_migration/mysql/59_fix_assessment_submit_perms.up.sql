-- fix: 考核任务按钮菜单 API 权限码与生成路由映射不一致（迁移 58 的同类遗留）。
--   route2code.go（生成产物）映射：
--     POST /api/v1/hrm/submitPerfAssessmentLeaderReview → v1:hrm:perf_assessment:submitLeaderReview
--     POST /api/v1/hrm/submitPerfAssessmentCalibration  → v1:hrm:perf_assessment:submitCalibration
--   而迁移 47 按钮菜单"评审提交/校准提交"种子的 API 码写的是
--   v1:hrm:perf_assessment:review / v1:hrm:perf_assessment:calibrate
--   （这两个是前端按钮码 hrm:perf_assessment:review|calibrate 的命名，误用作 API 码）
--   → casbin 无 submitLeaderReview/submitCalibration，上级评审与校准提交 403，
--   考核链在节点 3/4 卡死（自评因码名恰好一致而幸免）。
--   仅改 permissions JSON 内的 API 码；button_code 裸码（前端 v-auth 用）不动。

UPDATE manage_menu SET permissions = REPLACE(permissions, '"v1:hrm:perf_assessment:review"', '"v1:hrm:perf_assessment:submitLeaderReview"')
WHERE permissions LIKE '%v1:hrm:perf_assessment:review%';

UPDATE manage_menu SET permissions = REPLACE(permissions, '"v1:hrm:perf_assessment:calibrate"', '"v1:hrm:perf_assessment:submitCalibration"')
WHERE permissions LIKE '%v1:hrm:perf_assessment:calibrate%';

UPDATE casbin_rule SET v1 = 'v1:hrm:perf_assessment:submitLeaderReview'
WHERE v1 = 'v1:hrm:perf_assessment:review';

UPDATE casbin_rule SET v1 = 'v1:hrm:perf_assessment:submitCalibration'
WHERE v1 = 'v1:hrm:perf_assessment:calibrate';
