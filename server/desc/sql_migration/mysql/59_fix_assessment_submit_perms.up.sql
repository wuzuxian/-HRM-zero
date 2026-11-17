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

-- casbin_rule 表由 casbin adapter 在应用首启时创建；全新环境上迁移先于首启执行，
-- 因此触及 casbin_rule 的迁移必须自带幂等建表（否则 fresh deploy 在此迁移即失败）。
CREATE TABLE IF NOT EXISTS casbin_rule (
    id     BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL
);


UPDATE manage_menu SET permissions = REPLACE(permissions, '"v1:hrm:perf_assessment:review"', '"v1:hrm:perf_assessment:submitLeaderReview"')
WHERE permissions LIKE '%v1:hrm:perf_assessment:review%';

UPDATE manage_menu SET permissions = REPLACE(permissions, '"v1:hrm:perf_assessment:calibrate"', '"v1:hrm:perf_assessment:submitCalibration"')
WHERE permissions LIKE '%v1:hrm:perf_assessment:calibrate%';

UPDATE casbin_rule SET v1 = 'v1:hrm:perf_assessment:submitLeaderReview'
WHERE v1 = 'v1:hrm:perf_assessment:review';

UPDATE casbin_rule SET v1 = 'v1:hrm:perf_assessment:submitCalibration'
WHERE v1 = 'v1:hrm:perf_assessment:calibrate';
