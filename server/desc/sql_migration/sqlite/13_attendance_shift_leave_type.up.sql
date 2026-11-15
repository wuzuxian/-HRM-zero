-- HRM-11 班次管理与假期类型（假勤管理域 假勤设置）
CREATE TABLE hrm_shift (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    shift_name TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    cross_day INTEGER NOT NULL DEFAULT 0,
    rest_start TEXT NOT NULL DEFAULT '',
    rest_end TEXT NOT NULL DEFAULT '',
    standard_hours NUMERIC NOT NULL DEFAULT 0,
    flexible_minutes INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT '1',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE TABLE hrm_leave_type (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type_code TEXT NOT NULL UNIQUE,
    type_name TEXT NOT NULL,
    type_value INTEGER NOT NULL DEFAULT 0,
    balance_enabled INTEGER NOT NULL DEFAULT 0,
    attachment_required INTEGER NOT NULL DEFAULT 0,
    default_days NUMERIC NOT NULL DEFAULT 0,
    order_num INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT '1',
    remark TEXT NOT NULL DEFAULT ''
);

-- 假期类型预置数据（年假/调休假/事假/病假/婚假/产假/陪产假/丧假/其他）
INSERT INTO hrm_leave_type (uuid, create_time, update_time, type_code, type_name, type_value, balance_enabled, attachment_required, default_days, order_num, status, remark)
VALUES
    ('aaa10000-1111-4000-8000-000000000201','2026-11-15 00:00:00','2026-11-15 00:00:00','annual_leave','年假',1,1,0,5,1,'1',''),
    ('aaa10000-1111-4000-8000-000000000202','2026-11-15 00:00:00','2026-11-15 00:00:00','comp_leave','调休假',2,1,0,0,2,'1',''),
    ('aaa10000-1111-4000-8000-000000000203','2026-11-15 00:00:00','2026-11-15 00:00:00','personal_leave','事假',3,0,0,0,3,'1',''),
    ('aaa10000-1111-4000-8000-000000000204','2026-11-15 00:00:00','2026-11-15 00:00:00','sick_leave','病假',4,0,1,0,4,'1',''),
    ('aaa10000-1111-4000-8000-000000000205','2026-11-15 00:00:00','2026-11-15 00:00:00','marriage_leave','婚假',5,0,0,3,5,'1',''),
    ('aaa10000-1111-4000-8000-000000000206','2026-11-15 00:00:00','2026-11-15 00:00:00','maternity_leave','产假',6,0,0,98,6,'1',''),
    ('aaa10000-1111-4000-8000-000000000207','2026-11-15 00:00:00','2026-11-15 00:00:00','paternity_leave','陪产假',7,0,0,15,7,'1',''),
    ('aaa10000-1111-4000-8000-000000000208','2026-11-15 00:00:00','2026-11-15 00:00:00','bereavement_leave','丧假',8,0,0,3,8,'1',''),
    ('aaa10000-1111-4000-8000-000000000209','2026-11-15 00:00:00','2026-11-15 00:00:00','other_leave','其他',9,0,0,0,9,'1','');

-- HRM 菜单目录: 假勤管理
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000101','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','1','假勤管理',0,'',4,'hrm_attendance','/hrm/attendance','layout.base','mdi:calendar-clock','1','route.hrm_attendance',0,'',0,0,'[]','[]',0,'');

-- 班次管理 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000102','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','班次管理',0,'',1,'hrm_attendance_setting_shift','/hrm/attendance/setting/shift','view.hrm_attendance_setting_shift','mdi:clock-outline','1','route.hrm_attendance_setting_shift',0,'',0,0,'[]','[{"code":"v1:hrm:shift:list","desc":"班次列表"}]',0,'');

-- 班次管理 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000111','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000102','3','新增班次',0,'',0,'','','','','1','button.hrm.shift.add',0,'',0,0,'[]','[{"code":"v1:hrm:shift:add","desc":"新增班次"}]',0,'v1:hrm:shift:add'),
    ('aaa10000-1111-4000-8000-000000000112','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000102','3','编辑班次',0,'',1,'','','','','1','button.hrm.shift.edit',0,'',0,0,'[]','[{"code":"v1:hrm:shift:edit","desc":"编辑班次"}]',0,'v1:hrm:shift:edit'),
    ('aaa10000-1111-4000-8000-000000000113','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000102','3','删除班次',0,'',2,'','','','','1','button.hrm.shift.delete',0,'',0,0,'[]','[{"code":"v1:hrm:shift:delete","desc":"删除班次"}]',0,'v1:hrm:shift:delete');

-- 假期类型 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000103','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','假期类型',0,'',2,'hrm_attendance_setting_leave-type','/hrm/attendance/setting/leave-type','view.hrm_attendance_setting_leave-type','mdi:beach-outline','1','route.hrm_attendance_setting_leave-type',0,'',0,0,'[]','[{"code":"v1:hrm:leavetype:list","desc":"假期类型列表"}]',0,'');

-- 假期类型 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000121','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000103','3','新增假期类型',0,'',0,'','','','','1','button.hrm.leavetype.add',0,'',0,0,'[]','[{"code":"v1:hrm:leavetype:add","desc":"新增假期类型"}]',0,'v1:hrm:leavetype:add'),
    ('aaa10000-1111-4000-8000-000000000122','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000103','3','编辑假期类型',0,'',1,'','','','','1','button.hrm.leavetype.edit',0,'',0,0,'[]','[{"code":"v1:hrm:leavetype:edit","desc":"编辑假期类型"}]',0,'v1:hrm:leavetype:edit'),
    ('aaa10000-1111-4000-8000-000000000123','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000103','3','删除假期类型',0,'',2,'','','','','1','button.hrm.leavetype.delete',0,'',0,0,'[]','[{"code":"v1:hrm:leavetype:delete","desc":"删除假期类型"}]',0,'v1:hrm:leavetype:delete');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000101','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000101',0),
    ('aaa20000-1111-4000-8000-000000000102','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000102',0),
    ('aaa20000-1111-4000-8000-000000000103','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000103',0),
    ('aaa20000-1111-4000-8000-000000000104','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000111',0),
    ('aaa20000-1111-4000-8000-000000000105','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000112',0),
    ('aaa20000-1111-4000-8000-000000000106','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000113',0),
    ('aaa20000-1111-4000-8000-000000000107','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000121',0),
    ('aaa20000-1111-4000-8000-000000000108','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000122',0),
    ('aaa20000-1111-4000-8000-000000000109','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000123',0);

CREATE TRIGGER hrm_shift_update_time
AFTER UPDATE ON hrm_shift
FOR EACH ROW
BEGIN
    UPDATE hrm_shift SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_leave_type_update_time
AFTER UPDATE ON hrm_leave_type
FOR EACH ROW
BEGIN
    UPDATE hrm_leave_type SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
