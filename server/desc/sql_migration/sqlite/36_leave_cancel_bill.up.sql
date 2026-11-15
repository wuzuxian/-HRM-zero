-- HRM-19 请假销假单（HR205）（假勤管理域 假勤申请）
-- 表:
--   - hrm_leave_bill 请假销假单：请假类型(9 类预置)/预计开始结束时间/预计天数(自动计算, 0.5 步进)/事由；
--     审批通过事务内扣减假期余额（仅余额管理类型，leavebalance.ConsumeLeave）；
--     销假回填 实际开始/结束时间 + 实际天数（自动计算），余额按差额 多退少补（RefundLeave/ConsumeLeave）。
-- 生命周期（提交/审批/撤销/审批记录）走共享单据底座 v1/hrm/bill，类型码 205（迁移 11）。
-- 契约（写入 /home/wuzuxian/hrm-notes/hrm-19-notes.md，供 HRM-21 与未来销假流程复用）:
--   leavebalance.RefundLeave(ctx, tx, svcCtx, employeeUuid, leaveTypeCode, days, billUuid)
--   - days>0 退回天数：used -= days，记 change_type=refund 流水（正数天数=退回）
--   - 退回后 used < 0 → 10540 拒绝（幂等保护：重复销假由单据 settled 标记拦截）
-- 错误码 1054x 段（server/internal/errcodes/hrm_leave.go）。
-- 菜单 uuid 段 20xx（grep 全部 migration: 0xxx 基础/03xx 假勤设置/07xx 考勤组/09xx-16xx 各域/18xx 假期余额/19xx 个税；
-- 本票取 21xx（开工时取 20xx，但 HRM-16 迁移 35 并行占用了 20xx，合并后改 21xx）: 2101 page / 2111-2118 buttons（含共享单据底座 hrm:bill:* 按钮）/ role_menu aaa20000 21xx）。

CREATE TABLE hrm_leave_bill (
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
    leave_type_code TEXT NOT NULL,
    leave_type_name TEXT NOT NULL DEFAULT '',
    plan_start_time TEXT NOT NULL DEFAULT '',
    plan_end_time TEXT NOT NULL DEFAULT '',
    plan_days NUMERIC NOT NULL DEFAULT 0,
    reason TEXT NOT NULL DEFAULT '',
    actual_start_time TEXT NOT NULL DEFAULT '',
    actual_end_time TEXT NOT NULL DEFAULT '',
    actual_days NUMERIC NOT NULL DEFAULT 0,
    settled INTEGER NOT NULL DEFAULT 0,
    create_by TEXT NOT NULL DEFAULT '',
    creator_name TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_leave_bill_status ON hrm_leave_bill (status, settled);
CREATE INDEX idx_hrm_leave_bill_employee ON hrm_leave_bill (employee_uuid);
CREATE INDEX idx_hrm_leave_bill_type ON hrm_leave_bill (leave_type_code);

CREATE TRIGGER hrm_leave_bill_update_time
AFTER UPDATE ON hrm_leave_bill
FOR EACH ROW
BEGIN
    UPDATE hrm_leave_bill SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- 请假销假 page (假勤管理 dir aaa...000101 下, order 7; 假勤设置 1-4/考勤组/组织默认规则 3-4/假期余额 6 已占用)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002101','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','请假销假',0,'',7,'hrm_attendance_apply_leave-cancel-list','/hrm/attendance/apply/leave-cancel-list','view.hrm_attendance_apply_leave-cancel-list','mdi:calendar-arrow-right','1','route.hrm_attendance_apply_leave-cancel-list',0,'',0,0,'[]','[{"code":"v1:hrm:leave:list","desc":"请假销假列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 请假销假 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:* + 销假）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002111','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','新增请假单',0,'',0,'','','','','1','button.hrm.leave.add',0,'',0,0,'[]','[{"code":"v1:hrm:leave:add","desc":"新增请假单"}]',0,'v1:hrm:leave:add'),
    ('aaa10000-1111-4000-8000-000000002112','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','编辑请假单',0,'',1,'','','','','1','button.hrm.leave.edit',0,'',0,0,'[]','[{"code":"v1:hrm:leave:edit","desc":"编辑请假单"}]',0,'v1:hrm:leave:edit'),
    ('aaa10000-1111-4000-8000-000000002113','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','删除请假单',0,'',2,'','','','','1','button.hrm.leave.delete',0,'',0,0,'[]','[{"code":"v1:hrm:leave:delete","desc":"删除请假单"}]',0,'v1:hrm:leave:delete'),
    ('aaa10000-1111-4000-8000-000000002114','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','销假',0,'',3,'','','','','1','button.hrm.leave.cancel',0,'',0,0,'[]','[{"code":"v1:hrm:leave:cancel","desc":"销假回填实际时间"}]',0,'v1:hrm:leave:cancel'),
    ('aaa10000-1111-4000-8000-000000002115','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','提交单据',0,'',4,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000002116','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','审批通过',0,'',5,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000002117','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','审批驳回',0,'',6,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000002118','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','撤销单据',0,'',7,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002101','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002101',0),
    ('aaa20000-1111-4000-8000-000000002102','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002111',0),
    ('aaa20000-1111-4000-8000-000000002103','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002112',0),
    ('aaa20000-1111-4000-8000-000000002104','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002113',0),
    ('aaa20000-1111-4000-8000-000000002105','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002114',0),
    ('aaa20000-1111-4000-8000-000000002106','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002115',0),
    ('aaa20000-1111-4000-8000-000000002107','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002116',0),
    ('aaa20000-1111-4000-8000-000000002108','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002117',0),
    ('aaa20000-1111-4000-8000-000000002109','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002118',0);
