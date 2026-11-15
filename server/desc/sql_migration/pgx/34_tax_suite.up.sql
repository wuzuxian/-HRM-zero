-- HRM-36 个税管理三件套（迁移 34）
-- 三张表 join hrm_employee.uuid；页面挂 薪酬管理 dir（...1301，迁移 25）下 order 3/4/5，
-- 菜单 uuid 段 18xx（grep 全部迁移：10/12/13/14/15/16/17 段已占用，本票声明占用 18xx）。

CREATE TABLE "hrm_tax_profile" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid varchar(36) NOT NULL,
    employee_name varchar(50) NOT NULL DEFAULT '',
    id_card varchar(18) NOT NULL DEFAULT '',
    resident_type varchar(20) NOT NULL DEFAULT 'resident',
    tax_city varchar(50) NOT NULL DEFAULT '',
    default_special_extra varchar(200) NOT NULL DEFAULT '',
    status varchar(1) NOT NULL DEFAULT '1',
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE TABLE "hrm_tax_special_deduction" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid varchar(36) NOT NULL,
    employee_name varchar(50) NOT NULL DEFAULT '',
    deduct_year varchar(4) NOT NULL,
    deduct_month bigint NOT NULL,
    deduct_type varchar(20) NOT NULL,
    amount numeric(12,2) NOT NULL DEFAULT 0,
    source varchar(20) NOT NULL DEFAULT 'manual',
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE TABLE "hrm_tax_opening" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid varchar(36) NOT NULL,
    employee_name varchar(50) NOT NULL DEFAULT '',
    subject_uuid varchar(36) NOT NULL DEFAULT '',
    subject_name varchar(100) NOT NULL DEFAULT '',
    tax_year varchar(4) NOT NULL,
    until_month bigint NOT NULL,
    income_total numeric(14,2) NOT NULL DEFAULT 0,
    special_deduction_total numeric(14,2) NOT NULL DEFAULT 0,
    special_extra_total numeric(14,2) NOT NULL DEFAULT 0,
    exempt_total numeric(14,2) NOT NULL DEFAULT 0,
    other_deduction_total numeric(14,2) NOT NULL DEFAULT 0,
    taxable_income_total numeric(14,2) NOT NULL DEFAULT 0,
    tax_paid_total numeric(14,2) NOT NULL DEFAULT 0,
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX uk_hrm_tax_profile_employee ON "hrm_tax_profile" (employee_uuid);
CREATE INDEX idx_hrm_tax_profile_name ON "hrm_tax_profile" (employee_name);
CREATE UNIQUE INDEX uk_hrm_tax_special_emp_year_month_type ON "hrm_tax_special_deduction" (employee_uuid, deduct_year, deduct_month, deduct_type);
CREATE INDEX idx_hrm_tax_special_emp_year ON "hrm_tax_special_deduction" (employee_uuid, deduct_year);
CREATE UNIQUE INDEX uk_hrm_tax_opening_emp_subject_year ON "hrm_tax_opening" (employee_uuid, subject_uuid, tax_year);
CREATE INDEX idx_hrm_tax_opening_emp_year ON "hrm_tax_opening" (employee_uuid, tax_year);
CREATE INDEX idx_hrm_tax_opening_subject ON "hrm_tax_opening" (subject_uuid);

CREATE TRIGGER update_hrm_tax_profile_update_time BEFORE UPDATE ON "hrm_tax_profile"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

CREATE TRIGGER update_hrm_tax_special_deduction_update_time BEFORE UPDATE ON "hrm_tax_special_deduction"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

CREATE TRIGGER update_hrm_tax_opening_update_time BEFORE UPDATE ON "hrm_tax_opening"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

-- 员工个税档案 page（薪酬管理 dir ...1301 下 order 3）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001801','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','员工个税档案',0,'',3,'hrm_salary_tax_tax-profile','/hrm/salary/tax/tax-profile','view.hrm_salary_tax_tax-profile','mdi:account-card-details','1','route.hrm_salary_tax_tax-profile',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:list","desc":"员工个税档案列表"}]',0,'');

-- 员工个税档案 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001811','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001801','3','新增个税档案',0,'',0,'','','','','1','button.hrm.tax_profile.add',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:add","desc":"新增个税档案"}]',0,'v1:hrm:tax_profile:add'),
    ('aaa10000-1111-4000-8000-000000001812','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001801','3','编辑个税档案',0,'',1,'','','','','1','button.hrm.tax_profile.edit',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:edit","desc":"编辑个税档案"}]',0,'v1:hrm:tax_profile:edit'),
    ('aaa10000-1111-4000-8000-000000001813','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001801','3','删除个税档案',0,'',2,'','','','','1','button.hrm.tax_profile.delete',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:delete","desc":"删除个税档案"}]',0,'v1:hrm:tax_profile:delete'),
    ('aaa10000-1111-4000-8000-000000001814','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001801','3','导出个税档案',0,'',3,'','','','','1','button.hrm.tax_profile.export',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:export","desc":"导出个税档案"}]',0,'v1:hrm:tax_profile:export');

-- 专项附加扣除 page（order 4）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001802','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','专项附加扣除',0,'',4,'hrm_salary_tax_special-deduction','/hrm/salary/tax/special-deduction','view.hrm_salary_tax_special-deduction','mdi:home-heart','1','route.hrm_salary_tax_special-deduction',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:list","desc":"专项附加扣除列表"}]',0,'');

-- 专项附加扣除 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001815','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001802','3','维护专项附加',0,'',0,'','','','','1','button.hrm.tax_special_deduction.maintain',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:maintain","desc":"维护专项附加"}]',0,'v1:hrm:tax_special_deduction:maintain'),
    ('aaa10000-1111-4000-8000-000000001816','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001802','3','删除专项附加',0,'',1,'','','','','1','button.hrm.tax_special_deduction.delete',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:delete","desc":"删除专项附加"}]',0,'v1:hrm:tax_special_deduction:delete'),
    ('aaa10000-1111-4000-8000-000000001817','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001802','3','年度维护专项附加',0,'',2,'','','','','1','button.hrm.tax_special_deduction.maintain_year',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:maintain_year","desc":"年度维护专项附加"}]',0,'v1:hrm:tax_special_deduction:maintain_year'),
    ('aaa10000-1111-4000-8000-000000001818','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001802','3','导入专项附加',0,'',3,'','','','','1','button.hrm.tax_special_deduction.import',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:import","desc":"导入专项附加"}]',0,'v1:hrm:tax_special_deduction:import'),
    ('aaa10000-1111-4000-8000-000000001819','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001802','3','导出专项附加',0,'',4,'','','','','1','button.hrm.tax_special_deduction.export',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:export","desc":"导出专项附加"}]',0,'v1:hrm:tax_special_deduction:export');

-- 个税累计期初 page（order 5）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001803','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','个税累计期初',0,'',5,'hrm_salary_tax_tax-opening','/hrm/salary/tax/tax-opening','view.hrm_salary_tax_tax-opening','mdi:calculator-variant','1','route.hrm_salary_tax_tax-opening',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:list","desc":"个税累计期初列表"}]',0,'');

-- 个税累计期初 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001821','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001803','3','新增累计期初',0,'',0,'','','','','1','button.hrm.tax_opening.add',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:add","desc":"新增累计期初"}]',0,'v1:hrm:tax_opening:add'),
    ('aaa10000-1111-4000-8000-000000001822','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001803','3','编辑累计期初',0,'',1,'','','','','1','button.hrm.tax_opening.edit',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:edit","desc":"编辑累计期初"}]',0,'v1:hrm:tax_opening:edit'),
    ('aaa10000-1111-4000-8000-000000001823','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001803','3','删除累计期初',0,'',2,'','','','','1','button.hrm.tax_opening.delete',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:delete","desc":"删除累计期初"}]',0,'v1:hrm:tax_opening:delete'),
    ('aaa10000-1111-4000-8000-000000001824','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001803','3','导入累计期初',0,'',3,'','','','','1','button.hrm.tax_opening.import',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:import","desc":"导入累计期初"}]',0,'v1:hrm:tax_opening:import'),
    ('aaa10000-1111-4000-8000-000000001825','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001803','3','导出累计期初',0,'',4,'','','','','1','button.hrm.tax_opening.export',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:export","desc":"导出累计期初"}]',0,'v1:hrm:tax_opening:export');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001801','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001801',0),
    ('aaa20000-1111-4000-8000-000000001802','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001802',0),
    ('aaa20000-1111-4000-8000-000000001803','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001803',0),
    ('aaa20000-1111-4000-8000-000000001811','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001811',0),
    ('aaa20000-1111-4000-8000-000000001812','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001812',0),
    ('aaa20000-1111-4000-8000-000000001813','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001813',0),
    ('aaa20000-1111-4000-8000-000000001814','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001814',0),
    ('aaa20000-1111-4000-8000-000000001815','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001815',0),
    ('aaa20000-1111-4000-8000-000000001816','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001816',0),
    ('aaa20000-1111-4000-8000-000000001817','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001817',0),
    ('aaa20000-1111-4000-8000-000000001818','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001818',0),
    ('aaa20000-1111-4000-8000-000000001819','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001819',0),
    ('aaa20000-1111-4000-8000-000000001821','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001821',0),
    ('aaa20000-1111-4000-8000-000000001822','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001822',0),
    ('aaa20000-1111-4000-8000-000000001823','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001823',0),
    ('aaa20000-1111-4000-8000-000000001824','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001824',0),
    ('aaa20000-1111-4000-8000-000000001825','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001825',0);
