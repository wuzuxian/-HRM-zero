-- HRM-36 个税管理三件套（迁移 34）
-- 三张表 join hrm_employee.uuid；页面挂 薪酬管理 dir（...1301，迁移 25）下 order 3/4/5，
-- 菜单 uuid 段 19xx（18xx 已被 HRM-18 假期余额迁移 33 占用——合并 feat/hrm 时撞车发现，二次声明占用 19xx，grep 已确认 19xx 空闲）（grep 全部迁移：10-17 段已占用；18xx 已被 HRM-18（迁移 33）占用，本票声明占用 19xx）。

CREATE TABLE hrm_tax_profile (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid TEXT NOT NULL,
    employee_name TEXT NOT NULL DEFAULT '',
    id_card TEXT NOT NULL DEFAULT '',
    resident_type TEXT NOT NULL DEFAULT 'resident',
    tax_city TEXT NOT NULL DEFAULT '',
    default_special_extra TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '1',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE TABLE hrm_tax_special_deduction (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid TEXT NOT NULL,
    employee_name TEXT NOT NULL DEFAULT '',
    deduct_year TEXT NOT NULL,
    deduct_month INTEGER NOT NULL,
    deduct_type TEXT NOT NULL,
    amount REAL NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'manual',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE TABLE hrm_tax_opening (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid TEXT NOT NULL,
    employee_name TEXT NOT NULL DEFAULT '',
    subject_uuid TEXT NOT NULL DEFAULT '',
    subject_name TEXT NOT NULL DEFAULT '',
    tax_year TEXT NOT NULL,
    until_month INTEGER NOT NULL,
    income_total REAL NOT NULL DEFAULT 0,
    special_deduction_total REAL NOT NULL DEFAULT 0,
    special_extra_total REAL NOT NULL DEFAULT 0,
    exempt_total REAL NOT NULL DEFAULT 0,
    other_deduction_total REAL NOT NULL DEFAULT 0,
    taxable_income_total REAL NOT NULL DEFAULT 0,
    tax_paid_total REAL NOT NULL DEFAULT 0,
    remark TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX uk_hrm_tax_profile_employee ON hrm_tax_profile (employee_uuid);
CREATE INDEX idx_hrm_tax_profile_name ON hrm_tax_profile (employee_name);
CREATE UNIQUE INDEX uk_hrm_tax_special_emp_year_month_type ON hrm_tax_special_deduction (employee_uuid, deduct_year, deduct_month, deduct_type);
CREATE INDEX idx_hrm_tax_special_emp_year ON hrm_tax_special_deduction (employee_uuid, deduct_year);
CREATE UNIQUE INDEX uk_hrm_tax_opening_emp_subject_year ON hrm_tax_opening (employee_uuid, subject_uuid, tax_year);
CREATE INDEX idx_hrm_tax_opening_emp_year ON hrm_tax_opening (employee_uuid, tax_year);
CREATE INDEX idx_hrm_tax_opening_subject ON hrm_tax_opening (subject_uuid);

-- 员工个税档案 page（薪酬管理 dir ...1301 下 order 3）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001901','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','员工个税档案',0,'',3,'hrm_salary_tax_tax-profile','/hrm/salary/tax/tax-profile','view.hrm_salary_tax_tax-profile','mdi:account-card-details','1','route.hrm_salary_tax_tax-profile',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:list","desc":"员工个税档案列表"}]',0,'');

-- 员工个税档案 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001911','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001901','3','新增个税档案',0,'',0,'','','','','1','button.hrm.tax_profile.add',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:add","desc":"新增个税档案"}]',0,'v1:hrm:tax_profile:add'),
    ('aaa10000-1111-4000-8000-000000001912','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001901','3','编辑个税档案',0,'',1,'','','','','1','button.hrm.tax_profile.edit',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:edit","desc":"编辑个税档案"}]',0,'v1:hrm:tax_profile:edit'),
    ('aaa10000-1111-4000-8000-000000001913','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001901','3','删除个税档案',0,'',2,'','','','','1','button.hrm.tax_profile.delete',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:delete","desc":"删除个税档案"}]',0,'v1:hrm:tax_profile:delete'),
    ('aaa10000-1111-4000-8000-000000001914','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001901','3','导出个税档案',0,'',3,'','','','','1','button.hrm.tax_profile.export',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:export","desc":"导出个税档案"}]',0,'v1:hrm:tax_profile:export');

-- 专项附加扣除 page（order 4）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001902','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','专项附加扣除',0,'',4,'hrm_salary_tax_special-deduction','/hrm/salary/tax/special-deduction','view.hrm_salary_tax_special-deduction','mdi:home-heart','1','route.hrm_salary_tax_special-deduction',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:list","desc":"专项附加扣除列表"}]',0,'');

-- 专项附加扣除 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001915','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','维护专项附加',0,'',0,'','','','','1','button.hrm.tax_special_deduction.maintain',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:maintain","desc":"维护专项附加"}]',0,'v1:hrm:tax_special_deduction:maintain'),
    ('aaa10000-1111-4000-8000-000000001916','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','删除专项附加',0,'',1,'','','','','1','button.hrm.tax_special_deduction.delete',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:delete","desc":"删除专项附加"}]',0,'v1:hrm:tax_special_deduction:delete'),
    ('aaa10000-1111-4000-8000-000000001917','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','年度维护专项附加',0,'',2,'','','','','1','button.hrm.tax_special_deduction.maintain_year',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:maintain_year","desc":"年度维护专项附加"}]',0,'v1:hrm:tax_special_deduction:maintain_year'),
    ('aaa10000-1111-4000-8000-000000001918','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','导入专项附加',0,'',3,'','','','','1','button.hrm.tax_special_deduction.import',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:import","desc":"导入专项附加"}]',0,'v1:hrm:tax_special_deduction:import'),
    ('aaa10000-1111-4000-8000-000000001919','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','导出专项附加',0,'',4,'','','','','1','button.hrm.tax_special_deduction.export',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:export","desc":"导出专项附加"}]',0,'v1:hrm:tax_special_deduction:export');

-- 个税累计期初 page（order 5）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001903','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','个税累计期初',0,'',5,'hrm_salary_tax_tax-opening','/hrm/salary/tax/tax-opening','view.hrm_salary_tax_tax-opening','mdi:calculator-variant','1','route.hrm_salary_tax_tax-opening',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:list","desc":"个税累计期初列表"}]',0,'');

-- 个税累计期初 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001921','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','新增累计期初',0,'',0,'','','','','1','button.hrm.tax_opening.add',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:add","desc":"新增累计期初"}]',0,'v1:hrm:tax_opening:add'),
    ('aaa10000-1111-4000-8000-000000001922','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','编辑累计期初',0,'',1,'','','','','1','button.hrm.tax_opening.edit',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:edit","desc":"编辑累计期初"}]',0,'v1:hrm:tax_opening:edit'),
    ('aaa10000-1111-4000-8000-000000001923','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','删除累计期初',0,'',2,'','','','','1','button.hrm.tax_opening.delete',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:delete","desc":"删除累计期初"}]',0,'v1:hrm:tax_opening:delete'),
    ('aaa10000-1111-4000-8000-000000001924','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','导入累计期初',0,'',3,'','','','','1','button.hrm.tax_opening.import',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:import","desc":"导入累计期初"}]',0,'v1:hrm:tax_opening:import'),
    ('aaa10000-1111-4000-8000-000000001925','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','导出累计期初',0,'',4,'','','','','1','button.hrm.tax_opening.export',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:export","desc":"导出累计期初"}]',0,'v1:hrm:tax_opening:export');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001901','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001901',0),
    ('aaa20000-1111-4000-8000-000000001902','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001902',0),
    ('aaa20000-1111-4000-8000-000000001903','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001903',0),
    ('aaa20000-1111-4000-8000-000000001911','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001911',0),
    ('aaa20000-1111-4000-8000-000000001912','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001912',0),
    ('aaa20000-1111-4000-8000-000000001913','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001913',0),
    ('aaa20000-1111-4000-8000-000000001914','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001914',0),
    ('aaa20000-1111-4000-8000-000000001915','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001915',0),
    ('aaa20000-1111-4000-8000-000000001916','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001916',0),
    ('aaa20000-1111-4000-8000-000000001917','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001917',0),
    ('aaa20000-1111-4000-8000-000000001918','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001918',0),
    ('aaa20000-1111-4000-8000-000000001919','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001919',0),
    ('aaa20000-1111-4000-8000-000000001921','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001921',0),
    ('aaa20000-1111-4000-8000-000000001922','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001922',0),
    ('aaa20000-1111-4000-8000-000000001923','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001923',0),
    ('aaa20000-1111-4000-8000-000000001924','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001924',0),
    ('aaa20000-1111-4000-8000-000000001925','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001925',0);

CREATE TRIGGER hrm_tax_profile_update_time
AFTER UPDATE ON hrm_tax_profile
FOR EACH ROW
BEGIN
    UPDATE hrm_tax_profile SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_tax_special_deduction_update_time
AFTER UPDATE ON hrm_tax_special_deduction
FOR EACH ROW
BEGIN
    UPDATE hrm_tax_special_deduction SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_tax_opening_update_time
AFTER UPDATE ON hrm_tax_opening
FOR EACH ROW
BEGIN
    UPDATE hrm_tax_opening SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
