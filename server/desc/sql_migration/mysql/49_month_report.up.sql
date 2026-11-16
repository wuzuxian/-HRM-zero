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

CREATE TABLE `hrm_attendance_month_report` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `year_month` varchar(7) NOT NULL COMMENT '月份 YYYY-MM',
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工档案 uuid (HRM-03 hrm_employee.uuid)',
  `employee_no` varchar(20) NOT NULL DEFAULT '' COMMENT '工号（快照）',
  `employee_name` varchar(100) NOT NULL DEFAULT '' COMMENT '姓名（快照）',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '部门（组织 uuid 快照）',
  `org_name` varchar(100) NOT NULL DEFAULT '' COMMENT '部门名称（快照）',
  `att_group_name` varchar(100) NOT NULL DEFAULT '' COMMENT '考勤组名称（生效规则快照，三级回退；组织规则=组织默认）',
  `expected_days` bigint NOT NULL DEFAULT 0 COMMENT '应出勤天（区间内工作日数，WorkDays 口径）',
  `attend_days` bigint NOT NULL DEFAULT 0 COMMENT '实出勤天（settle_status ∈ normal/late/early）',
  `late_count` bigint NOT NULL DEFAULT 0 COMMENT '迟到次数',
  `early_count` bigint NOT NULL DEFAULT 0 COMMENT '早退次数',
  `missing_count` bigint NOT NULL DEFAULT 0 COMMENT '缺卡次数',
  `absent_days` bigint NOT NULL DEFAULT 0 COMMENT '旷工天（missing 且上/下班卡均无打卡，同 HRM-22 口径）',
  `workday_overtime_hours` decimal(7,2) NOT NULL DEFAULT 0 COMMENT '工作日加班(h)：status=approved 且 day_type=1 单据基准时长合计',
  `restday_overtime_hours` decimal(7,2) NOT NULL DEFAULT 0 COMMENT '休息日加班(h)：status=approved 且 day_type=2',
  `holiday_overtime_hours` decimal(7,2) NOT NULL DEFAULT 0 COMMENT '节假日加班(h)：status=approved 且 day_type=3',
  `comp_accrue_hours` decimal(7,2) NOT NULL DEFAULT 0 COMMENT '调休入账(h)：status=approved 单据 accrue_hours 合计',
  `row_hash` varchar(64) NOT NULL DEFAULT '' COMMENT '行内容 hash（锁定快照，重算 diff 用）',
  `changed` bigint NOT NULL DEFAULT 0 COMMENT '锁定后变更标记 1=重算后与本行快照不一致（只看锁定后变更过滤用）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_month_report` (`year_month`, `employee_uuid`),
  KEY `idx_hrm_month_report_org` (`org_uuid`),
  KEY `idx_hrm_month_report_ym` (`year_month`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_attendance_month_lock` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `year_month` varchar(7) NOT NULL COMMENT '月份 YYYY-MM',
  `action` varchar(20) NOT NULL COMMENT '动作 lock=锁定月报 unlock=解锁 recalc=重新计算',
  `operator_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '操作人用户 uuid',
  `operator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '操作人姓名（快照）',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注（锁定/解锁原因）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_month_lock_ym` (`year_month`),
  KEY `idx_hrm_month_lock_action` (`year_month`, `action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 假勤数据-考勤月报 page (假勤管理 dir aaa...000101 下, order 12)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003501','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','考勤月报',0,'',12,'hrm_attendance_data_month-report','/hrm/attendance/data/month-report','view.hrm_attendance_data_month-report','mdi:calendar-month-outline','1','route.hrm_attendance_data_month-report',0,'',0,0,'[]','[{"code":"v1:hrm:monthreport:list","desc":"考勤月报查询"},{"code":"v1:hrm:monthreport:detail","desc":"考勤月报日明细"},{"code":"v1:hrm:monthreport:lock","desc":"锁定月报"},{"code":"v1:hrm:monthreport:unlock","desc":"解锁月报"},{"code":"v1:hrm:monthreport:recalc","desc":"重新计算月报"},{"code":"v1:hrm:monthreport:lockRecords","desc":"锁定记录查询"},{"code":"v1:hrm:monthreport:export","desc":"导出考勤月报"}]',0,'');

-- 考勤月报 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
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
