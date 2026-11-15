-- HRM-09 人事调动单（HR203）+ 服务内嵌调度骨架（迁移 16）
-- 单据底座（hrm_approval_record/hrm_bill_sequence）见迁移 11。

CREATE TABLE hrm_transfer_bill (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_no TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'unsubmitted',
    employee_uuid TEXT NOT NULL DEFAULT '',
    employee_no TEXT NOT NULL DEFAULT '',
    employee_name TEXT NOT NULL DEFAULT '',
    change_type TEXT NOT NULL DEFAULT 'other',
    original_company TEXT NOT NULL DEFAULT '',
    original_org TEXT NOT NULL DEFAULT '',
    original_position TEXT NOT NULL DEFAULT '',
    original_post TEXT NOT NULL DEFAULT '',
    new_company TEXT NOT NULL DEFAULT '',
    new_org TEXT NOT NULL DEFAULT '',
    new_position TEXT NOT NULL DEFAULT '',
    new_post TEXT NOT NULL DEFAULT '',
    new_salary NUMERIC NOT NULL DEFAULT 0,
    transfer_reason TEXT NOT NULL DEFAULT '',
    effective_immediately INTEGER NOT NULL DEFAULT 1,
    effective_date TEXT NOT NULL DEFAULT '',
    applied INTEGER NOT NULL DEFAULT 0,
    create_by TEXT NOT NULL DEFAULT '',
    creator_name TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_transfer_bill_status ON hrm_transfer_bill (status, applied);
CREATE INDEX idx_hrm_transfer_bill_employee ON hrm_transfer_bill (employee_uuid);

-- HRM 菜单目录: 人事管理
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000601','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','1','人事管理',0,'',6,'hrm_employee-relation','/hrm/employee-relation','layout.base','mdi:account-switch','1','route.hrm_employee-relation',0,'',0,0,'[]','[]',0,'');

-- 人事调动 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000602','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000601','2','人事调动',0,'',1,'hrm_employee-relation_transfer-list','/hrm/employee-relation/transfer-list','view.hrm_employee-relation_transfer-list','mdi:swap-horizontal','1','route.hrm_employee-relation_transfer-list',0,'',0,0,'[]','[{"code":"v1:hrm:transfer:list","desc":"人事调动列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 人事调动 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000611','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000602','3','新增调动',0,'',0,'','','','','1','button.hrm.transfer.add',0,'',0,0,'[]','[{"code":"v1:hrm:transfer:add","desc":"新增调动"}]',0,'v1:hrm:transfer:add'),
    ('aaa10000-1111-4000-8000-000000000612','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000602','3','编辑调动',0,'',1,'','','','','1','button.hrm.transfer.edit',0,'',0,0,'[]','[{"code":"v1:hrm:transfer:edit","desc":"编辑调动"}]',0,'v1:hrm:transfer:edit'),
    ('aaa10000-1111-4000-8000-000000000613','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000602','3','删除调动',0,'',2,'','','','','1','button.hrm.transfer.delete',0,'',0,0,'[]','[{"code":"v1:hrm:transfer:delete","desc":"删除调动"}]',0,'v1:hrm:transfer:delete'),
    ('aaa10000-1111-4000-8000-000000000621','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000602','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000000622','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000602','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000000623','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000602','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000000624','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000602','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000601','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000601',0),
    ('aaa20000-1111-4000-8000-000000000602','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000602',0),
    ('aaa20000-1111-4000-8000-000000000603','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000611',0),
    ('aaa20000-1111-4000-8000-000000000604','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000612',0),
    ('aaa20000-1111-4000-8000-000000000605','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000613',0),
    ('aaa20000-1111-4000-8000-000000000606','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000621',0),
    ('aaa20000-1111-4000-8000-000000000607','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000622',0),
    ('aaa20000-1111-4000-8000-000000000608','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000623',0),
    ('aaa20000-1111-4000-8000-000000000609','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000624',0);

CREATE TRIGGER hrm_transfer_bill_update_time
AFTER UPDATE ON hrm_transfer_bill
FOR EACH ROW
BEGIN
    UPDATE hrm_transfer_bill SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
