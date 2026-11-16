-- HRM-23 考勤月报与锁定（假勤管理域 假勤数据；demo-inventory 4.4 考勤月报）
-- 表:
--   - hrm_attendance_month_report 考勤月报行：(year_month, employee_uuid) 唯一。
--     聚合源（HRM-16/HRM-22 契约，Go 内存聚合照抄 HRM-22）:
--       应出勤/实出勤/迟到/早退/缺卡/旷工 ← hrm_punch_record settle_* 快照列
--         （旷工 = missing 且上/下班卡均无打卡，同 HRM-22 缺勤口径）；
--       工作日/休息日/节假日加班(h) ← hrm_overtime_bill status=approved 按 day_type 分桶的
--         折算基准时长（核定 confirm_hours>0 取核定，否则申请时长）；
--       调休入账(h) ← hrm_overtime_bill status=approved 的 accrue_hours 合计。
--     锁定 = 当前实时聚合结果写入本表作为快照 + row_hash；解锁/重算产生 diff 时
--     changed=1 标记（只看锁定后变更过滤用）。
--   - hrm_attendance_month_lock 锁定记录流水：lock/unlock/recalc 谁何时操作。
-- 消费契约（HRM-37 试算）:
--   monthreport.GetMonthReportRows(ctx, svcCtx, yearMonth string, orgUuids []string)
--     → []MonthReportRow（锁定月取快照行、未锁定月实时聚合并带 unlocked=true，见
--     server/internal/logic/v1/hrm/monthreport/domain.go）。
-- 错误码 1067x 段（server/internal/errcodes/hrm_month_report.go；1066x 考核任务链已占）。
-- 菜单 uuid 段 35xx（grep 全部 migration: 0xxx-33xx 已用、34xx 在途加班 offer 票——下票避让；
-- 本票取 35xx: 3501 考勤月报 page / 3511-3515 buttons / role_menu aaa20000 35xx）。

CREATE TABLE "hrm_attendance_month_report" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    year_month varchar(7) NOT NULL,
    employee_uuid varchar(36) NOT NULL DEFAULT '',
    employee_no varchar(20) NOT NULL DEFAULT '',
    employee_name varchar(100) NOT NULL DEFAULT '',
    org_uuid varchar(36) NOT NULL DEFAULT '',
    org_name varchar(100) NOT NULL DEFAULT '',
    att_group_name varchar(100) NOT NULL DEFAULT '',
    expected_days bigint NOT NULL DEFAULT 0,
    attend_days bigint NOT NULL DEFAULT 0,
    late_count bigint NOT NULL DEFAULT 0,
    early_count bigint NOT NULL DEFAULT 0,
    missing_count bigint NOT NULL DEFAULT 0,
    absent_days bigint NOT NULL DEFAULT 0,
    workday_overtime_hours numeric(7,2) NOT NULL DEFAULT 0,
    restday_overtime_hours numeric(7,2) NOT NULL DEFAULT 0,
    holiday_overtime_hours numeric(7,2) NOT NULL DEFAULT 0,
    comp_accrue_hours numeric(7,2) NOT NULL DEFAULT 0,
    row_hash varchar(64) NOT NULL DEFAULT '',
    changed bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX uk_hrm_month_report ON "hrm_attendance_month_report" (year_month, employee_uuid);
CREATE INDEX idx_hrm_month_report_org ON "hrm_attendance_month_report" (org_uuid);
CREATE INDEX idx_hrm_month_report_ym ON "hrm_attendance_month_report" (year_month);

CREATE TABLE "hrm_attendance_month_lock" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    year_month varchar(7) NOT NULL,
    action varchar(20) NOT NULL,
    operator_uuid varchar(36) NOT NULL DEFAULT '',
    operator_name varchar(100) NOT NULL DEFAULT '',
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_month_lock_ym ON "hrm_attendance_month_lock" (year_month);

-- 假勤数据-考勤月报 page (假勤管理 dir aaa...000101 下, order 12)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003501','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','考勤月报',0,'',12,'hrm_attendance_data_month-report','/hrm/attendance/data/month-report','view.hrm_attendance_data_month-report','mdi:calendar-month-outline','1','route.hrm_attendance_data_month-report',0,'',0,0,'[]','[{"code":"v1:hrm:monthreport:list","desc":"考勤月报查询"},{"code":"v1:hrm:monthreport:detail","desc":"考勤月报日明细"},{"code":"v1:hrm:monthreport:lock","desc":"锁定月报"},{"code":"v1:hrm:monthreport:unlock","desc":"解锁月报"},{"code":"v1:hrm:monthreport:recalc","desc":"重新计算月报"},{"code":"v1:hrm:monthreport:lockRecords","desc":"锁定记录查询"},{"code":"v1:hrm:monthreport:export","desc":"导出考勤月报"}]',0,'');

-- 考勤月报 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003511','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003501','3','锁定记录',0,'',0,'','','','','1','button.hrm.monthreport.lockRecords',0,'',0,0,'[]','[{"code":"v1:hrm:monthreport:lockRecords","desc":"锁定记录查询"}]',0,'v1:hrm:monthreport:lockRecords'),
    ('aaa10000-1111-4000-8000-000000003512','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003501','3','重新计算',0,'',1,'','','','','1','button.hrm.monthreport.recalc',0,'',0,0,'[]','[{"code":"v1:hrm:monthreport:recalc","desc":"重新计算月报"}]',0,'v1:hrm:monthreport:recalc'),
    ('aaa10000-1111-4000-8000-000000003513','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003501','3','锁定月报',0,'',2,'','','','','1','button.hrm.monthreport.lock',0,'',0,0,'[]','[{"code":"v1:hrm:monthreport:lock","desc":"锁定月报"}]',0,'v1:hrm:monthreport:lock'),
    ('aaa10000-1111-4000-8000-000000003514','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003501','3','解锁',0,'',3,'','','','','1','button.hrm.monthreport.unlock',0,'',0,0,'[]','[{"code":"v1:hrm:monthreport:unlock","desc":"解锁月报"}]',0,'v1:hrm:monthreport:unlock'),
    ('aaa10000-1111-4000-8000-000000003515','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003501','3','导出',0,'',4,'','','','','1','button.hrm.monthreport.export',0,'',0,0,'[]','[{"code":"v1:hrm:monthreport:export","desc":"导出考勤月报"}]',0,'v1:hrm:monthreport:export');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003501','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003501',0),
    ('aaa20000-1111-4000-8000-000000003511','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003511',0),
    ('aaa20000-1111-4000-8000-000000003512','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003512',0),
    ('aaa20000-1111-4000-8000-000000003513','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003513',0),
    ('aaa20000-1111-4000-8000-000000003514','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003514',0),
    ('aaa20000-1111-4000-8000-000000003515','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003515',0);
