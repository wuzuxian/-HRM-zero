-- HRM-35 员工薪资档案（薪酬管理域）
-- 菜单 uuid 段：29xx（开工时已 grep 三方言迁移：22xx/28xx 空闲但 28xx 已被并行 dashboard 票登记预留，
-- 本票取 29xx，声明占用）。

CREATE TABLE hrm_salary_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid TEXT NOT NULL,
    employee_no TEXT NOT NULL DEFAULT '',
    employee_name TEXT NOT NULL DEFAULT '',
    company_uuid TEXT NOT NULL DEFAULT '',
    company_name TEXT NOT NULL DEFAULT '',
    org_uuid TEXT NOT NULL DEFAULT '',
    org_name TEXT NOT NULL DEFAULT '',
    subject_uuid TEXT NOT NULL,
    subject_name TEXT NOT NULL DEFAULT '',
    scheme_uuid TEXT NOT NULL DEFAULT '',
    scheme_name TEXT NOT NULL DEFAULT '',
    fixed_salary REAL NOT NULL DEFAULT 0,
    fixed_allowance REAL NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT '1',
    effective_date TEXT NOT NULL,
    expire_date TEXT NOT NULL DEFAULT '',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_salary_archive_employee ON hrm_salary_archive (employee_uuid);
CREATE INDEX idx_hrm_salary_archive_subject ON hrm_salary_archive (subject_uuid);
CREATE INDEX idx_hrm_salary_archive_scheme ON hrm_salary_archive (scheme_uuid);
CREATE INDEX idx_hrm_salary_archive_effective ON hrm_salary_archive (effective_date);
CREATE INDEX idx_hrm_salary_archive_expire ON hrm_salary_archive (expire_date);

-- 员工薪资档案 page（挂在 13xx 薪酬管理目录下，order 6：主体=1 项目=2 方案=3 个税=3/4/5 冲突取 6）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002901','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','员工薪资档案',0,'',6,'hrm_salary_archive-list','/hrm/salary/archive-list','view.hrm_salary_archive-list','mdi:account-cash-outline','1','route.hrm_salary_archive-list',0,'',0,0,'[]','[{"code":"v1:hrm:salary_archive:list","desc":"员工薪资档案列表"},{"code":"v1:hrm:salary_archive:getEffectiveArchive","desc":"取员工生效档案"}]',0,'');

-- 员工薪资档案 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002911','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002901','3','新增薪资档案',0,'',0,'','','','','1','button.hrm.salary_archive.add',0,'',0,0,'[]','[{"code":"v1:hrm:salary_archive:add","desc":"新增薪资档案"}]',0,'v1:hrm:salary_archive:add'),
    ('aaa10000-1111-4000-8000-000000002912','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002901','3','编辑薪资档案',0,'',1,'','','','','1','button.hrm.salary_archive.edit',0,'',0,0,'[]','[{"code":"v1:hrm:salary_archive:edit","desc":"编辑薪资档案"}]',0,'v1:hrm:salary_archive:edit'),
    ('aaa10000-1111-4000-8000-000000002913','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002901','3','删除薪资档案',0,'',2,'','','','','1','button.hrm.salary_archive.delete',0,'',0,0,'[]','[{"code":"v1:hrm:salary_archive:delete","desc":"删除薪资档案"}]',0,'v1:hrm:salary_archive:delete');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002901','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002901',0),
    ('aaa20000-1111-4000-8000-000000002911','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002911',0),
    ('aaa20000-1111-4000-8000-000000002912','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002912',0),
    ('aaa20000-1111-4000-8000-000000002913','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002913',0);

CREATE TRIGGER hrm_salary_archive_update_time
AFTER UPDATE ON hrm_salary_archive
FOR EACH ROW
BEGIN
    UPDATE hrm_salary_archive SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
