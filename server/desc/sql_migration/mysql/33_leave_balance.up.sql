-- HRM-18 假期余额（初始化/调整/流水）（假勤管理域 假期数据）
-- 表:
--   - hrm_leave_balance 假期余额行：(year, employee_uuid, leave_type_code) 唯一
--     应发 granted = 初始 + 调整; 可用 = granted - used - frozen
--   - hrm_leave_balance_log 假期余额流水：每次 初始化/调整/请假扣减(consume)/调休入账(accrue) 都记一条
--     （变更类型/数量/变更后三值/操作人/时间/关联单据 bill_uuid）
-- 消费契约（下游单据调用，事务内记流水+更新余额）:
--   - HRM-19 请假单扣减: leavebalance.ConsumeLeave(ctx, tx, employeeUuid, leaveTypeCode, days, billUuid)
--   - HRM-21 加班调休入账: leavebalance.AccrueCompLeave(ctx, tx, employeeUuid, hours, billUuid)
-- 错误码 1051x 段（server/internal/errcodes/hrm_leave_balance.go）。
-- 菜单 uuid 段 18xx（grep 全部 migration: 0xxx 基础/03xx 假勤设置/07xx 考勤组/09xx-16xx 各域，
-- 17xx 注释提及但实际未用；本票取 18xx: 1801 page / 1811-1812 buttons / role_menu aaa20000 18xx）。

CREATE TABLE `hrm_leave_balance` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `year` bigint NOT NULL COMMENT '年度',
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工档案 uuid (HRM-03 hrm_employee.uuid)',
  `employee_no` varchar(20) NOT NULL DEFAULT '' COMMENT '工号（快照）',
  `employee_name` varchar(100) NOT NULL DEFAULT '' COMMENT '姓名（快照）',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '部门（组织 uuid 快照）',
  `leave_type_code` varchar(50) NOT NULL COMMENT '假期类型编码 (hrm_leave_type.type_code)',
  `granted` decimal(6,1) NOT NULL DEFAULT 0 COMMENT '应发 = 初始 + 调整',
  `used` decimal(6,1) NOT NULL DEFAULT 0 COMMENT '已扣减（请假消费）',
  `frozen` decimal(6,1) NOT NULL DEFAULT 0 COMMENT '冻结（审批中的请假单占用，预留）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_leave_balance` (`year`, `employee_uuid`, `leave_type_code`),
  KEY `idx_hrm_leave_balance_employee` (`employee_uuid`),
  KEY `idx_hrm_leave_balance_year` (`year`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_leave_balance_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `balance_uuid` varchar(36) NOT NULL COMMENT '余额行 uuid (hrm_leave_balance.uuid)',
  `year` bigint NOT NULL,
  `employee_uuid` varchar(36) NOT NULL,
  `leave_type_code` varchar(50) NOT NULL,
  `change_type` varchar(20) NOT NULL COMMENT '变更类型 init=初始化 adjust=调整 consume=请假扣减 accrue=调休入账',
  `days` decimal(6,1) NOT NULL COMMENT '变更数量（正=增，负=减）',
  `granted_after` decimal(6,1) NOT NULL DEFAULT 0 COMMENT '变更后应发',
  `used_after` decimal(6,1) NOT NULL DEFAULT 0 COMMENT '变更后已用',
  `frozen_after` decimal(6,1) NOT NULL DEFAULT 0 COMMENT '变更后冻结',
  `reason` varchar(500) NOT NULL DEFAULT '' COMMENT '调整原因/说明',
  `bill_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '关联单据 uuid（请假单/加班单，无关联为空）',
  `operator_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '操作人用户 uuid（系统动作为空）',
  `operator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '操作人姓名（快照）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_leave_balance_log_balance` (`balance_uuid`),
  KEY `idx_hrm_leave_balance_log_employee` (`employee_uuid`, `year`, `leave_type_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 假期余额 page (假勤管理 dir aaa...000101 下, order 6; 假勤设置页 order 1-5 已占用)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001801','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','假期余额',0,'',6,'hrm_attendance_data_leave-balance','/hrm/attendance/data/leave-balance','view.hrm_attendance_data_leave-balance','mdi:wallet-outline','1','route.hrm_attendance_data_leave-balance',0,'',0,0,'[]','[{"code":"v1:hrm:leavebalance:list","desc":"假期余额列表"},{"code":"v1:hrm:leavebalance:logList","desc":"假期余额流水查询"},{"code":"v1:hrm:leavebalance:initPreview","desc":"初始化预览"}]',0,'');

-- 假期余额 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001811','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001801','3','初始化本年度额度',0,'',0,'','','','','1','button.hrm.leavebalance.init',0,'',0,0,'[]','[{"code":"v1:hrm:leavebalance:init","desc":"初始化本年度额度"}]',0,'v1:hrm:leavebalance:init'),
    ('aaa10000-1111-4000-8000-000000001812','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001801','3','调整余额',0,'',1,'','','','','1','button.hrm.leavebalance.adjust',0,'',0,0,'[]','[{"code":"v1:hrm:leavebalance:adjust","desc":"调整余额"}]',0,'v1:hrm:leavebalance:adjust');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001801','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001801',0),
    ('aaa20000-1111-4000-8000-000000001802','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001811',0),
    ('aaa20000-1111-4000-8000-000000001803','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001812',0);
