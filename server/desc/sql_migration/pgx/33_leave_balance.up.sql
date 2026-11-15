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

CREATE TABLE "hrm_leave_balance" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    year bigint NOT NULL,
    employee_uuid varchar(36) NOT NULL DEFAULT '',
    employee_no varchar(20) NOT NULL DEFAULT '',
    employee_name varchar(100) NOT NULL DEFAULT '',
    org_uuid varchar(36) NOT NULL DEFAULT '',
    leave_type_code varchar(50) NOT NULL,
    granted numeric(6,1) NOT NULL DEFAULT 0,
    used numeric(6,1) NOT NULL DEFAULT 0,
    frozen numeric(6,1) NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX uk_hrm_leave_balance ON "hrm_leave_balance" (year, employee_uuid, leave_type_code);
CREATE INDEX idx_hrm_leave_balance_employee ON "hrm_leave_balance" (employee_uuid);
CREATE INDEX idx_hrm_leave_balance_year ON "hrm_leave_balance" (year);

CREATE TABLE "hrm_leave_balance_log" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    balance_uuid varchar(36) NOT NULL,
    year bigint NOT NULL,
    employee_uuid varchar(36) NOT NULL DEFAULT '',
    leave_type_code varchar(50) NOT NULL,
    change_type varchar(20) NOT NULL,
    days numeric(6,1) NOT NULL DEFAULT 0,
    granted_after numeric(6,1) NOT NULL DEFAULT 0,
    used_after numeric(6,1) NOT NULL DEFAULT 0,
    frozen_after numeric(6,1) NOT NULL DEFAULT 0,
    reason varchar(500) NOT NULL DEFAULT '',
    bill_uuid varchar(36) NOT NULL DEFAULT '',
    operator_uuid varchar(36) NOT NULL DEFAULT '',
    operator_name varchar(100) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_leave_balance_log_balance ON "hrm_leave_balance_log" (balance_uuid);
CREATE INDEX idx_hrm_leave_balance_log_employee ON "hrm_leave_balance_log" (employee_uuid, year, leave_type_code);

CREATE TRIGGER update_hrm_leave_balance_update_time BEFORE UPDATE ON "hrm_leave_balance"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

CREATE TRIGGER update_hrm_leave_balance_log_update_time BEFORE UPDATE ON "hrm_leave_balance_log"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

-- 假期余额 page (假勤管理 dir aaa...000101 下, order 6; 假勤设置页 order 1-5 已占用)
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001801','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','假期余额',0,'',6,'hrm_attendance_data_leave-balance','/hrm/attendance/data/leave-balance','view.hrm_attendance_data_leave-balance','mdi:wallet-outline','1','route.hrm_attendance_data_leave-balance',0,'',0,0,'[]','[{"code":"v1:hrm:leavebalance:list","desc":"假期余额列表"},{"code":"v1:hrm:leavebalance:logList","desc":"假期余额流水查询"},{"code":"v1:hrm:leavebalance:initPreview","desc":"初始化预览"}]',0,'');

-- 假期余额 buttons
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001811','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001801','3','初始化本年度额度',0,'',0,'','','','','1','button.hrm.leavebalance.init',0,'',0,0,'[]','[{"code":"v1:hrm:leavebalance:init","desc":"初始化本年度额度"}]',0,'v1:hrm:leavebalance:init'),
    ('aaa10000-1111-4000-8000-000000001812','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001801','3','调整余额',0,'',1,'','','','','1','button.hrm.leavebalance.adjust',0,'',0,0,'[]','[{"code":"v1:hrm:leavebalance:adjust","desc":"调整余额"}]',0,'v1:hrm:leavebalance:adjust');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001801','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001801',0),
    ('aaa20000-1111-4000-8000-000000001802','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001811',0),
    ('aaa20000-1111-4000-8000-000000001803','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001812',0);
