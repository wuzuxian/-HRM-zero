-- fix: 三个页面动作端点的 API 权限码缺失/错位（60/61 同类遗留，
--   由 route2code 一致性守卫测试 handler/route2code_menu_consistency_test.go 发现）。
--   route2code.go（生成产物）映射：
--     POST /api/v1/hrm/advancePerfPeriod               → v1:hrm:perf_period:advance                （菜单从未授予）
--     POST /api/v1/hrm/finishTrainingSession           → v1:hrm:training_session:finish            （菜单从未授予）
--     POST /api/v1/hrm/maintainTaxSpecialDeductionYear → v1:hrm:tax_special_deduction:maintainYear （种子误写 snake_case maintain_year）
--   → 考核批次"推进"、培训班次"结课"、专项附加扣除"年度维护"三个 UI 按钮点击后
--     Authx 裸 403。同时修 manage_menu.permissions 与 casbin_rule 存量行。
--   豁免：v1:hrm:salary_payslip:mineList / read 为员工自助端点（前端尚无调用方），
--   见测试文件豁免清单，待产品接线时随菜单种子一并授予。

-- casbin_rule 表由 casbin adapter 在应用首启时创建；全新环境上迁移先于首启执行，
-- 因此触及 casbin_rule 的迁移必须自带幂等建表（否则 fresh deploy 在此迁移即失败）。
CREATE TABLE IF NOT EXISTS casbin_rule (
    id     BIGSERIAL PRIMARY KEY,
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL
);


UPDATE manage_menu SET permissions = REPLACE(permissions, '"v1:hrm:tax_special_deduction:maintain_year"', '"v1:hrm:tax_special_deduction:maintainYear"')
WHERE permissions LIKE '%v1:hrm:tax_special_deduction:maintain_year%';

UPDATE manage_menu SET permissions = REPLACE(permissions, '"}]', '"},{"code":"v1:hrm:perf_period:advance","desc":"批次推进"}]')
WHERE route_path = '/hrm/performance/period-list' AND permissions LIKE '%v1:hrm:perf_period:list%';

UPDATE manage_menu SET permissions = REPLACE(permissions, '"}]', '"},{"code":"v1:hrm:training_session:finish","desc":"班次结课"}]')
WHERE route_path = '/hrm/training/session' AND permissions LIKE '%v1:hrm:training_session:list%';

UPDATE casbin_rule SET v1 = 'v1:hrm:tax_special_deduction:maintainYear'
WHERE v1 = 'v1:hrm:tax_special_deduction:maintain_year';

INSERT INTO casbin_rule (p_type, v0, v1)
SELECT DISTINCT 'p', rm.role_uuid, 'v1:hrm:perf_period:advance'
FROM manage_role_menu rm
WHERE rm.menu_uuid = (SELECT uuid FROM manage_menu WHERE route_path = '/hrm/performance/period-list')
  AND NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = rm.role_uuid AND c.v1 = 'v1:hrm:perf_period:advance'
  );

INSERT INTO casbin_rule (p_type, v0, v1)
SELECT DISTINCT 'p', rm.role_uuid, 'v1:hrm:training_session:finish'
FROM manage_role_menu rm
WHERE rm.menu_uuid = (SELECT uuid FROM manage_menu WHERE route_path = '/hrm/training/session')
  AND NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = rm.role_uuid AND c.v1 = 'v1:hrm:training_session:finish'
  );

INSERT INTO casbin_rule (p_type, v0, v1)
SELECT DISTINCT 'p', rm.role_uuid, 'v1:hrm:tax_special_deduction:maintainYear'
FROM manage_role_menu rm
WHERE rm.menu_uuid = (SELECT uuid FROM manage_menu WHERE route_path = '/hrm/salary/tax/special-deduction')
  AND NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = rm.role_uuid AND c.v1 = 'v1:hrm:tax_special_deduction:maintainYear'
  );
