-- HRM-50 code review (spec #50) — hrm_employee.user_uuid 普通索引: pgx 方言, 与 mysql/sqlite/65 同号同集合。
-- spec 原文要求 "增加普通索引, 用于按账号反查档案"。三个查询点受益:
--   1. hrm/employee/link_user.go   关联时占用校验 (FindOneByCondition user_uuid = ?)
--   2. manage/user/delete.go       删号前批量 IN 反查 (引用校验)
--   3. hrm/employee/base.go        fillLoginNames 批量 IN 回填登录账号
-- 普通索引而非唯一: 一对一由应用层校验保证 (ADR 0007 单列指针 + 应用层校验,
-- 不加库级唯一索引), 且存量大量空串指针 ('' 语义 = 未关联, 见迁移 17) 必须允许重复。
CREATE INDEX IF NOT EXISTS idx_hrm_employee_user_uuid ON "hrm_employee" (user_uuid);
