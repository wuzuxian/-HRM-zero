-- HRM-21 加班申请单（HR217）（假勤管理域 假勤申请）
-- 表:
--   - hrm_overtime_bill 加班申请单：加班日期/当天类型(工作日/休息日/法定节假日)/起止时间/
--     申请时长(自动计算, 最小计算单位向下取整)/事由/来源(员工申请)/核定时长/补偿方式；
--     提交时按 HRM-13 ResolveOvertimeRule 判定 需审批/免审批 并快照 倍数；
--     需审批→走审批底座(typeCode 217)；免审批→提交即底座自动通过（同一事务内 SubmitBill+ApproveBill）。
--     审批人 核定（核定时长 ≤ 申请时长）+ 补偿方式(调休/其他)；
--     审批通过事务内按 CompMultiplier 折算调休时长（最小计算单位向下取整、月上限守卫），
--     补偿方式=调休时 leavebalance.AccrueCompLeave 入账 comp_leave 余额（流水 change_type=accrue）。
-- 生命周期（提交/审批/撤销/审批记录）走共享单据底座 v1/hrm/bill，类型码 217（迁移 11）。
-- 契约消费:
--   overtimerule.ResolveOvertimeRule(ctx, svcCtx, groupUuid, orgUuid, dayType) （HRM-13）
--   attendancegroup.ResolveEffectiveRule(ctx, svcCtx, employee)      （HRM-12 定位考勤组）
--   leavebalance.AccrueCompLeave(ctx, tx, svcCtx, employeeUuid, hours, billUuid) （HRM-18）
-- 错误码 1057x 段（server/internal/errcodes/hrm_overtime.go）。
-- 菜单 uuid 段 26xx（grep 全部 migration: 0xxx 基础/03xx 假勤设置/07xx 考勤组/09xx-16xx 各域/
-- 18xx 假期余额/19xx 个税/20xx 打卡/21xx 排班表/23xx 请假销假/25xx 职位; 22xx/24xx 空;
-- 本票取 26xx: 2601 page / 2611-2616 buttons（含共享单据底座 hrm:bill:* 按钮 + 核定）/
-- role_menu aaa20000 26xx）。

CREATE TABLE hrm_overtime_bill (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_no TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'unsubmitted',
    employee_uuid TEXT NOT NULL DEFAULT '',
    employee_no TEXT NOT NULL DEFAULT '',
    employee_name TEXT NOT NULL DEFAULT '',
    org_uuid TEXT NOT NULL DEFAULT '',
    org_name TEXT NOT NULL DEFAULT '',
    overtime_date TEXT NOT NULL,
    day_type TEXT NOT NULL,
    start_time TEXT NOT NULL DEFAULT '',
    end_time TEXT NOT NULL DEFAULT '',
    apply_hours NUMERIC NOT NULL DEFAULT 0,
    reason TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'employee',
    rule_source TEXT NOT NULL DEFAULT '',
    comp_multiplier NUMERIC NOT NULL DEFAULT 1,
    confirm_hours NUMERIC NOT NULL DEFAULT 0,
    comp_type TEXT NOT NULL DEFAULT '',
    accrue_hours NUMERIC NOT NULL DEFAULT 0,
    settled INTEGER NOT NULL DEFAULT 0,
    create_by TEXT NOT NULL DEFAULT '',
    creator_name TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_overtime_bill_status ON hrm_overtime_bill (status, settled);
CREATE INDEX idx_hrm_overtime_bill_employee ON hrm_overtime_bill (employee_uuid);
CREATE INDEX idx_hrm_overtime_bill_date ON hrm_overtime_bill (overtime_date);

CREATE TRIGGER hrm_overtime_bill_update_time
AFTER UPDATE ON hrm_overtime_bill
FOR EACH ROW
BEGIN
    UPDATE hrm_overtime_bill SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- 加班申请 page (假勤管理 dir aaa...000101 下, order 10; 请假销假 7/打卡记录 8/排班表 9 已占用)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002601','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','加班申请',0,'',10,'hrm_attendance_apply_overtime-bill-list','/hrm/attendance/apply/overtime-bill-list','view.hrm_attendance_apply_overtime-bill-list','mdi:clock-plus-outline','1','route.hrm_attendance_apply_overtime-bill-list',0,'',0,0,'[]','[{"code":"v1:hrm:overtimebill:list","desc":"加班申请列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 加班申请 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:* + 核定）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002611','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002601','3','新增加班单',0,'',0,'','','','','1','button.hrm.overtimebill.add',0,'',0,0,'[]','[{"code":"v1:hrm:overtimebill:add","desc":"新增加班单"}]',0,'v1:hrm:overtimebill:add'),
    ('aaa10000-1111-4000-8000-000000002612','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002601','3','编辑加班单',0,'',1,'','','','','1','button.hrm.overtimebill.edit',0,'',0,0,'[]','[{"code":"v1:hrm:overtimebill:edit","desc":"编辑加班单"}]',0,'v1:hrm:overtimebill:edit'),
    ('aaa10000-1111-4000-8000-000000002613','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002601','3','删除加班单',0,'',2,'','','','','1','button.hrm.overtimebill.delete',0,'',0,0,'[]','[{"code":"v1:hrm:overtimebill:delete","desc":"删除加班单"}]',0,'v1:hrm:overtimebill:delete'),
    ('aaa10000-1111-4000-8000-000000002614','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002601','3','核定加班单',0,'',3,'','','','','1','button.hrm.overtimebill.confirm',0,'',0,0,'[]','[{"code":"v1:hrm:overtimebill:confirm","desc":"核定时长与补偿方式"}]',0,'v1:hrm:overtimebill:confirm'),
    ('aaa10000-1111-4000-8000-000000002615','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002601','3','提交单据',0,'',4,'','','','','1','button.hrm.overtimebill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:overtimebill:submit","desc":"提交加班单"}]',0,'v1:hrm:overtimebill:submit'),
    ('aaa10000-1111-4000-8000-000000002616','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002601','3','审批通过',0,'',5,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000002617','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002601','3','审批驳回',0,'',6,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000002618','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002601','3','撤销单据',0,'',7,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002601','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002601',0),
    ('aaa20000-1111-4000-8000-000000002602','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002611',0),
    ('aaa20000-1111-4000-8000-000000002603','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002612',0),
    ('aaa20000-1111-4000-8000-000000002604','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002613',0),
    ('aaa20000-1111-4000-8000-000000002605','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002614',0),
    ('aaa20000-1111-4000-8000-000000002606','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002615',0),
    ('aaa20000-1111-4000-8000-000000002607','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002616',0),
    ('aaa20000-1111-4000-8000-000000002608','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002617',0),
    ('aaa20000-1111-4000-8000-000000002609','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002618',0);
