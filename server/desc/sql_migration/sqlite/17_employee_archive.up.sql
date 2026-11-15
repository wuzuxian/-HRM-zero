-- HRM-03 员工档案纵切: 员工主表 + 教育经历/工作经历/家属三个从表 + 人事档案目录菜单 + 员工档案页面与按钮种子
-- 主表 hrm_employee 绑定组织: org_uuid (所属部门) / company_uuid (所属单位);
-- 人员状态 person_status: 1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休 (离职/退休=不在册口径);
-- user_uuid 预留 #6 员工↔账号双向同步 (manage_user.username = employee_no)。

CREATE TABLE hrm_employee (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_no TEXT NOT NULL,
    name TEXT NOT NULL,
    gender TEXT NOT NULL,
    birthday TEXT NOT NULL,
    nation TEXT NOT NULL,
    political_status TEXT NOT NULL,
    marital_status TEXT NOT NULL,
    native_place TEXT NOT NULL,
    blood_type TEXT NOT NULL,
    id_card TEXT NOT NULL,
    height TEXT NOT NULL,
    weight TEXT NOT NULL,
    household_address TEXT NOT NULL,
    address TEXT NOT NULL,
    emergency_contact TEXT NOT NULL,
    emergency_phone TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT NOT NULL,
    avatar TEXT NOT NULL,
    bank_name TEXT NOT NULL,
    bank_account TEXT NOT NULL,
    company_uuid TEXT NOT NULL,
    org_uuid TEXT NOT NULL,
    position TEXT NOT NULL,
    post TEXT NOT NULL,
    person_status TEXT NOT NULL,
    hire_date TEXT NOT NULL,
    formal_date TEXT NOT NULL,
    user_uuid TEXT NOT NULL,
    remark TEXT NOT NULL
);

CREATE INDEX idx_hrm_employee_org_uuid ON hrm_employee (org_uuid);
CREATE INDEX idx_hrm_employee_name ON hrm_employee (name);

CREATE TABLE hrm_employee_education (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid TEXT NOT NULL,
    school TEXT NOT NULL,
    major TEXT NOT NULL,
    education TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    remark TEXT NOT NULL
);

CREATE INDEX idx_hrm_employee_education_employee_uuid ON hrm_employee_education (employee_uuid);

CREATE TABLE hrm_employee_work (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid TEXT NOT NULL,
    company TEXT NOT NULL,
    position TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    remark TEXT NOT NULL
);

CREATE INDEX idx_hrm_employee_work_employee_uuid ON hrm_employee_work (employee_uuid);

CREATE TABLE hrm_employee_family (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid TEXT NOT NULL,
    name TEXT NOT NULL,
    relation TEXT NOT NULL,
    phone TEXT NOT NULL,
    work_unit TEXT NOT NULL,
    remark TEXT NOT NULL
);

CREATE INDEX idx_hrm_employee_family_employee_uuid ON hrm_employee_family (employee_uuid);

-- 人事档案 directory (under 人力资源 dir aaa...000001, order 2)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000401','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','1','人事档案',0,'',2,'hrm_employee','/hrm/employee','layout.base','mdi:account-box-multiple-outline','1','route.hrm_employee',0,'',0,0,'[]','[]',0,'');

-- 员工档案 page (route2code: GET /api/v1/hrm/getEmployeeList → v1:hrm:employee:list)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000402','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000401','2','员工档案',0,'',1,'hrm_employee_employee-archive-list','/hrm/employee/employee-archive-list','view.hrm_employee_employee-archive-list','mdi:card-account-details-outline','1','route.hrm_employee_employee-archive-list',0,'',0,0,'[]','[{"code":"v1:hrm:employee:list","desc":"员工档案列表"},{"code":"v1:hrm:employee:detail","desc":"员工档案详情"}]',0,'');

-- 员工档案 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000411','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','新增员工',0,'',0,'','','','','1','button.hrm.employee.add',0,'',0,0,'[]','[{"code":"v1:hrm:employee:add","desc":"新增员工"}]',0,'v1:hrm:employee:add'),
    ('aaa10000-1111-4000-8000-000000000412','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','编辑员工',0,'',1,'','','','','1','button.hrm.employee.edit',0,'',0,0,'[]','[{"code":"v1:hrm:employee:edit","desc":"编辑员工"}]',0,'v1:hrm:employee:edit'),
    ('aaa10000-1111-4000-8000-000000000413','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','删除员工',0,'',2,'','','','','1','button.hrm.employee.delete',0,'',0,0,'[]','[{"code":"v1:hrm:employee:delete","desc":"删除员工"}]',0,'v1:hrm:employee:delete'),
    ('aaa10000-1111-4000-8000-000000000414','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','导入员工',0,'',3,'','','','','1','button.hrm.employee.import',0,'',0,0,'[]','[{"code":"v1:hrm:employee:import","desc":"导入员工"}]',0,'v1:hrm:employee:import'),
    ('aaa10000-1111-4000-8000-000000000415','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','导出员工',0,'',4,'','','','','1','button.hrm.employee.export',0,'',0,0,'[]','[{"code":"v1:hrm:employee:export","desc":"导出员工"}]',0,'v1:hrm:employee:export'),
    ('aaa10000-1111-4000-8000-000000000416','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','生成用户',0,'',5,'','','','','1','button.hrm.employee.generateuser',0,'',0,0,'[]','[{"code":"v1:hrm:employee:generateUser","desc":"生成系统账号"}]',0,'v1:hrm:employee:generateUser');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000401','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000401',0),
    ('aaa20000-1111-4000-8000-000000000402','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000402',0),
    ('aaa20000-1111-4000-8000-000000000403','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000411',0),
    ('aaa20000-1111-4000-8000-000000000404','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000412',0),
    ('aaa20000-1111-4000-8000-000000000405','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000413',0),
    ('aaa20000-1111-4000-8000-000000000406','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000414',0),
    ('aaa20000-1111-4000-8000-000000000407','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000415',0);

CREATE TRIGGER hrm_employee_update_time
AFTER UPDATE ON hrm_employee
FOR EACH ROW
BEGIN
    UPDATE hrm_employee SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_employee_education_update_time
AFTER UPDATE ON hrm_employee_education
FOR EACH ROW
BEGIN
    UPDATE hrm_employee_education SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_employee_work_update_time
AFTER UPDATE ON hrm_employee_work
FOR EACH ROW
BEGIN
    UPDATE hrm_employee_work SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_employee_family_update_time
AFTER UPDATE ON hrm_employee_family
FOR EACH ROW
BEGIN
    UPDATE hrm_employee_family SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
