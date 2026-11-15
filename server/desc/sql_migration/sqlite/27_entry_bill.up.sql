-- HRM-07 员工入职单（HR201）纵切（迁移 27）
-- 单据底座（hrm_approval_record/hrm_bill_sequence）见迁移 11。
-- 入职单快照三张从表（工作经历/教育经历/家庭成员，按 entry_uuid 快照）；
-- 审批通过（OnApproved 事务内）自动创建 hrm_employee 档案：
--   工号自动生成（复用 HRM-03 NextEmployeeNo 最大工号+1）、预计转正=入职日期+试用期月、
--   三从表落员工档案、回写 employee_uuid（幂等：重复审批不重复建档）。

CREATE TABLE hrm_entry_bill (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_no TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'unsubmitted',
    name TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    gender TEXT NOT NULL DEFAULT '',
    birthday TEXT NOT NULL DEFAULT '',
    nation TEXT NOT NULL DEFAULT '',
    political_status TEXT NOT NULL DEFAULT '',
    marital_status TEXT NOT NULL DEFAULT '',
    culture_level TEXT NOT NULL DEFAULT '',
    native_place TEXT NOT NULL DEFAULT '',
    address TEXT NOT NULL DEFAULT '',
    household_address TEXT NOT NULL DEFAULT '',
    emergency_contact TEXT NOT NULL DEFAULT '',
    emergency_phone TEXT NOT NULL DEFAULT '',
    bank_name TEXT NOT NULL DEFAULT '',
    bank_account TEXT NOT NULL DEFAULT '',
    avatar TEXT NOT NULL DEFAULT '',
    position TEXT NOT NULL DEFAULT '',
    post TEXT NOT NULL DEFAULT '',
    probation_months INTEGER NOT NULL DEFAULT 0,
    org_uuid TEXT NOT NULL DEFAULT '',
    company_uuid TEXT NOT NULL DEFAULT '',
    hire_date TEXT NOT NULL DEFAULT '',
    expect_formal_date TEXT NOT NULL DEFAULT '',
    attachment_desc TEXT NOT NULL DEFAULT '',
    remark TEXT NOT NULL DEFAULT '',
    employee_uuid TEXT NOT NULL DEFAULT '',
    create_by TEXT NOT NULL DEFAULT '',
    creator_name TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_entry_bill_status ON hrm_entry_bill (status);
CREATE INDEX idx_hrm_entry_bill_employee ON hrm_entry_bill (employee_uuid);

CREATE TABLE hrm_entry_bill_work (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    entry_uuid TEXT NOT NULL,
    company TEXT NOT NULL DEFAULT '',
    position TEXT NOT NULL DEFAULT '',
    start_date TEXT NOT NULL DEFAULT '',
    end_date TEXT NOT NULL DEFAULT '',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_entry_bill_work_entry_uuid ON hrm_entry_bill_work (entry_uuid);

CREATE TABLE hrm_entry_bill_education (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    entry_uuid TEXT NOT NULL,
    school TEXT NOT NULL DEFAULT '',
    major TEXT NOT NULL DEFAULT '',
    education TEXT NOT NULL DEFAULT '',
    start_date TEXT NOT NULL DEFAULT '',
    end_date TEXT NOT NULL DEFAULT '',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_entry_bill_education_entry_uuid ON hrm_entry_bill_education (entry_uuid);

CREATE TABLE hrm_entry_bill_family (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    entry_uuid TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    relation TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    work_unit TEXT NOT NULL DEFAULT '',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_entry_bill_family_entry_uuid ON hrm_entry_bill_family (entry_uuid);

CREATE TRIGGER hrm_entry_bill_update_time
AFTER UPDATE ON hrm_entry_bill
FOR EACH ROW
BEGIN
    UPDATE hrm_entry_bill SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- 员工入职 page（挂在已有 人事管理 dir aaa...000601 下；transfer 占 order 1,
-- 3.2/3.4 转正/离职页预留 2/3, 本页取 4）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001402','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000601','2','员工入职',0,'',4,'hrm_employee-relation_entry-list','/hrm/employee-relation/entry-list','view.hrm_employee-relation_entry-list','mdi:account-plus-outline','1','route.hrm_employee-relation_entry-list',0,'',0,0,'[]','[{"code":"v1:hrm:entry:list","desc":"员工入职列表"},{"code":"v1:hrm:entry:detail","desc":"入职单详情"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 员工入职 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001411','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','新增入职单',0,'',0,'','','','','1','button.hrm.entry.add',0,'',0,0,'[]','[{"code":"v1:hrm:entry:add","desc":"新增入职单"}]',0,'v1:hrm:entry:add'),
    ('aaa10000-1111-4000-8000-000000001412','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','编辑入职单',0,'',1,'','','','','1','button.hrm.entry.edit',0,'',0,0,'[]','[{"code":"v1:hrm:entry:edit","desc":"编辑入职单"}]',0,'v1:hrm:entry:edit'),
    ('aaa10000-1111-4000-8000-000000001413','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','删除入职单',0,'',2,'','','','','1','button.hrm.entry.delete',0,'',0,0,'[]','[{"code":"v1:hrm:entry:delete","desc":"删除入职单"}]',0,'v1:hrm:entry:delete'),
    ('aaa10000-1111-4000-8000-000000001421','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000001422','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000001423','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000001424','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001401','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001402',0),
    ('aaa20000-1111-4000-8000-000000001402','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001411',0),
    ('aaa20000-1111-4000-8000-000000001403','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001412',0),
    ('aaa20000-1111-4000-8000-000000001404','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001413',0),
    ('aaa20000-1111-4000-8000-000000001405','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001421',0),
    ('aaa20000-1111-4000-8000-000000001406','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001422',0),
    ('aaa20000-1111-4000-8000-000000001407','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001423',0),
    ('aaa20000-1111-4000-8000-000000001408','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001424',0);
