-- HRM-33 薪酬主体与薪资项目（薪酬管理域 薪酬设置）
-- 菜单 uuid 段：13xx（12xx 已被 perf 占用 —— 开工时已 grep 确认 12xx/13xx/14xx 均空闲，声明占用 13xx）。

CREATE TABLE hrm_salary_subject (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    subject_name TEXT NOT NULL,
    org_uuid TEXT NOT NULL,
    org_name TEXT NOT NULL DEFAULT '',
    pay_account TEXT NOT NULL DEFAULT '',
    pay_bank TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '1',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE TABLE hrm_salary_item (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    item_code TEXT NOT NULL UNIQUE,
    item_name TEXT NOT NULL,
    item_type TEXT NOT NULL DEFAULT 'fixed',
    item_class TEXT NOT NULL DEFAULT 'earning',
    value_mode TEXT NOT NULL DEFAULT 'manual',
    calc_order INTEGER NOT NULL DEFAULT 0,
    order_num INTEGER NOT NULL DEFAULT 0,
    formula TEXT NOT NULL DEFAULT '',
    is_preset INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT '1',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_salary_subject_org ON hrm_salary_subject (org_uuid);
CREATE INDEX idx_hrm_salary_item_class ON hrm_salary_item (item_class);
CREATE INDEX idx_hrm_salary_item_calc ON hrm_salary_item (calc_order);

-- 薪资项目预置数据：应发 → 应扣 → 实发工资 → 成本
INSERT INTO hrm_salary_item (uuid, create_time, update_time, item_code, item_name, item_type, item_class, value_mode, calc_order, order_num, formula, is_preset, status, remark)
VALUES
    ('aaa10000-1111-4000-8000-000000001301','2026-11-15 00:00:00','2026-11-15 00:00:00','BASIC_SALARY','基本工资','fixed','earning','manual',10,1,'',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001302','2026-11-15 00:00:00','2026-11-15 00:00:00','POST_SALARY','岗位工资','fixed','earning','manual',20,2,'',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001303','2026-11-15 00:00:00','2026-11-15 00:00:00','MERIT_SALARY','绩效工资','data','earning','system',30,3,'',1,'1','绩效结果同步'),
    ('aaa10000-1111-4000-8000-000000001304','2026-11-15 00:00:00','2026-11-15 00:00:00','FIXED_ALLOWANCE','固定补贴','fixed','earning','manual',40,4,'',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001305','2026-11-15 00:00:00','2026-11-15 00:00:00','HEAT_ALLOWANCE','高温补贴','fixed','earning','manual',50,5,'',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001306','2026-11-15 00:00:00','2026-11-15 00:00:00','REWARD_BONUS','奖惩奖励','data','earning','system',60,6,'',1,'1','奖惩单入账'),
    ('aaa10000-1111-4000-8000-000000001307','2026-11-15 00:00:00','2026-11-15 00:00:00','ABSENCE_DEDUCTION','缺勤扣款','data','deduction','system',70,7,'',1,'1','考勤结果取数'),
    ('aaa10000-1111-4000-8000-000000001308','2026-11-15 00:00:00','2026-11-15 00:00:00','SI_PERSONAL','社保个人','data','deduction','system',80,8,'',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001309','2026-11-15 00:00:00','2026-11-15 00:00:00','HF_PERSONAL','公积金个人','data','deduction','system',90,9,'',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001310','2026-11-15 00:00:00','2026-11-15 00:00:00','TAX','个税','formula','deduction','formula',100,10,'',1,'1','按个税口径计算'),
    ('aaa10000-1111-4000-8000-000000001311','2026-11-15 00:00:00','2026-11-15 00:00:00','REWARD_PENALTY','奖惩扣款','data','deduction','system',110,11,'',1,'1','奖惩单入账'),
    ('aaa10000-1111-4000-8000-000000001312','2026-11-15 00:00:00','2026-11-15 00:00:00','OTHER_DEDUCTION','其他应扣','fixed','deduction','manual',120,12,'',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001313','2026-11-15 00:00:00','2026-11-15 00:00:00','NET_SALARY','实发工资','formula','earning','formula',130,13,'应发合计-应扣合计',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001314','2026-11-15 00:00:00','2026-11-15 00:00:00','SI_COMPANY','公司社保','data','cost','system',140,14,'',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001315','2026-11-15 00:00:00','2026-11-15 00:00:00','HF_COMPANY','公司公积金','data','cost','system',150,15,'',1,'1',''),
    ('aaa10000-1111-4000-8000-000000001316','2026-11-15 00:00:00','2026-11-15 00:00:00','LABOR_COST','人工成本','formula','cost','formula',160,16,'应发合计+公司社保+公司公积金',1,'1','');

-- HRM 菜单目录: 薪酬管理
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001301','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','1','薪酬管理',0,'',9,'hrm_salary','/hrm/salary','layout.base','mdi:cash-multiple','1','route.hrm_salary',0,'',0,0,'[]','[]',0,'');

-- 薪酬主体 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001302','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','薪酬主体',0,'',1,'hrm_salary_setting_subject','/hrm/salary/setting/subject','view.hrm_salary_setting_subject','mdi:domain','1','route.hrm_salary_setting_subject',0,'',0,0,'[]','[{"code":"v1:hrm:salary_subject:list","desc":"薪酬主体列表"}]',0,'');

-- 薪酬主体 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001311','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001302','3','新增薪酬主体',0,'',0,'','','','','1','button.hrm.salary_subject.add',0,'',0,0,'[]','[{"code":"v1:hrm:salary_subject:add","desc":"新增薪酬主体"}]',0,'v1:hrm:salary_subject:add'),
    ('aaa10000-1111-4000-8000-000000001312','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001302','3','编辑薪酬主体',0,'',1,'','','','','1','button.hrm.salary_subject.edit',0,'',0,0,'[]','[{"code":"v1:hrm:salary_subject:edit","desc":"编辑薪酬主体"}]',0,'v1:hrm:salary_subject:edit'),
    ('aaa10000-1111-4000-8000-000000001313','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001302','3','删除薪酬主体',0,'',2,'','','','','1','button.hrm.salary_subject.delete',0,'',0,0,'[]','[{"code":"v1:hrm:salary_subject:delete","desc":"删除薪酬主体"}]',0,'v1:hrm:salary_subject:delete');

-- 薪资项目 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001303','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','薪资项目',0,'',2,'hrm_salary_setting_item','/hrm/salary/setting/item','view.hrm_salary_setting_item','mdi:format-list-numbered','1','route.hrm_salary_setting_item',0,'',0,0,'[]','[{"code":"v1:hrm:salary_item:list","desc":"薪资项目列表"},{"code":"v1:hrm:salary_item:getAll","desc":"全量薪资项目"}]',0,'');

-- 薪资项目 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001321','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001303','3','新增薪资项目',0,'',0,'','','','','1','button.hrm.salary_item.add',0,'',0,0,'[]','[{"code":"v1:hrm:salary_item:add","desc":"新增薪资项目"}]',0,'v1:hrm:salary_item:add'),
    ('aaa10000-1111-4000-8000-000000001322','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001303','3','编辑薪资项目',0,'',1,'','','','','1','button.hrm.salary_item.edit',0,'',0,0,'[]','[{"code":"v1:hrm:salary_item:edit","desc":"编辑薪资项目"}]',0,'v1:hrm:salary_item:edit'),
    ('aaa10000-1111-4000-8000-000000001323','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001303','3','删除薪资项目',0,'',2,'','','','','1','button.hrm.salary_item.delete',0,'',0,0,'[]','[{"code":"v1:hrm:salary_item:delete","desc":"删除薪资项目"}]',0,'v1:hrm:salary_item:delete');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001301','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001301',0),
    ('aaa20000-1111-4000-8000-000000001302','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001302',0),
    ('aaa20000-1111-4000-8000-000000001303','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001303',0),
    ('aaa20000-1111-4000-8000-000000001311','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001311',0),
    ('aaa20000-1111-4000-8000-000000001312','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001312',0),
    ('aaa20000-1111-4000-8000-000000001313','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001313',0),
    ('aaa20000-1111-4000-8000-000000001321','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001321',0),
    ('aaa20000-1111-4000-8000-000000001322','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001322',0),
    ('aaa20000-1111-4000-8000-000000001323','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001323',0);

CREATE TRIGGER hrm_salary_subject_update_time
AFTER UPDATE ON hrm_salary_subject
FOR EACH ROW
BEGIN
    UPDATE hrm_salary_subject SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_salary_item_update_time
AFTER UPDATE ON hrm_salary_item
FOR EACH ROW
BEGIN
    UPDATE hrm_salary_item SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
