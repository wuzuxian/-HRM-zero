-- HRM-34 薪酬方案（薪酬管理域 薪酬设置）
-- 菜单 uuid 段：27xx（开工时已 grep 确认 27xx 空闲，声明占用）。

CREATE TABLE hrm_salary_scheme (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    scheme_name TEXT NOT NULL,
    subject_uuid TEXT NOT NULL,
    subject_name TEXT NOT NULL DEFAULT '',
    person_status TEXT NOT NULL DEFAULT '',
    settle_cycle TEXT NOT NULL DEFAULT 'monthly',
    pay_day INTEGER NOT NULL DEFAULT 1,
    currency TEXT NOT NULL DEFAULT 'CNY',
    status TEXT NOT NULL DEFAULT '1',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE TABLE hrm_salary_scheme_item (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    scheme_uuid TEXT NOT NULL,
    item_uuid TEXT NOT NULL,
    override_amount REAL NULL,
    override_formula TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_salary_scheme_subject ON hrm_salary_scheme (subject_uuid);
CREATE INDEX idx_hrm_salary_scheme_item_scheme ON hrm_salary_scheme_item (scheme_uuid);
CREATE INDEX idx_hrm_salary_scheme_item_item ON hrm_salary_scheme_item (item_uuid);

-- 薪酬方案 page（挂在 13xx 薪酬管理目录下，order 3：主体=1 薪资项目=2）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002702','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','薪酬方案',0,'',3,'hrm_salary_setting_scheme','/hrm/salary/setting/scheme','view.hrm_salary_setting_scheme','mdi:file-cog-outline','1','route.hrm_salary_setting_scheme',0,'',0,0,'[]','[{"code":"v1:hrm:salary_scheme:list","desc":"薪酬方案列表"},{"code":"v1:hrm:salary_scheme:getSchemeDetail","desc":"薪酬方案详情"},{"code":"v1:hrm:salary_scheme:getSchemeItems","desc":"方案项目清单"}]',0,'');

-- 薪酬方案 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002711','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002702','3','新增薪酬方案',0,'',0,'','','','','1','button.hrm.salary_scheme.add',0,'',0,0,'[]','[{"code":"v1:hrm:salary_scheme:add","desc":"新增薪酬方案"}]',0,'v1:hrm:salary_scheme:add'),
    ('aaa10000-1111-4000-8000-000000002712','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002702','3','编辑薪酬方案',0,'',1,'','','','','1','button.hrm.salary_scheme.edit',0,'',0,0,'[]','[{"code":"v1:hrm:salary_scheme:edit","desc":"编辑薪酬方案"}]',0,'v1:hrm:salary_scheme:edit'),
    ('aaa10000-1111-4000-8000-000000002713','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002702','3','删除薪酬方案',0,'',2,'','','','','1','button.hrm.salary_scheme.delete',0,'',0,0,'[]','[{"code":"v1:hrm:salary_scheme:delete","desc":"删除薪酬方案"}]',0,'v1:hrm:salary_scheme:delete');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002701','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002702',0),
    ('aaa20000-1111-4000-8000-000000002711','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002711',0),
    ('aaa20000-1111-4000-8000-000000002712','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002712',0),
    ('aaa20000-1111-4000-8000-000000002713','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002713',0);

CREATE TRIGGER hrm_salary_scheme_update_time
AFTER UPDATE ON hrm_salary_scheme
FOR EACH ROW
BEGIN
    UPDATE hrm_salary_scheme SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_salary_scheme_item_update_time
AFTER UPDATE ON hrm_salary_scheme_item
FOR EACH ROW
BEGIN
    UPDATE hrm_salary_scheme_item SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
