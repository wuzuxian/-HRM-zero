-- HRM-15 排班表（假勤管理域 排班管理；demo-inventory 4.3）
-- 表: hrm_schedule_entry 排班表行：(schedule_date, employee_uuid) 唯一 —— 一人一天一行，
--   shift_uuid 空=休/未排班；group_uuid 维护口径考勤组快照；published 发布标记。
--   排班表按考勤组维护（成员=考勤组适用范围展开）；周期生成复用 HRM-14 ExpandCycle。
-- 发布口径: 日结取班次时发布与否都生效；排班制无排班日兜底走考勤组/组织默认规则。
-- 错误码 1055x 段（server/internal/errcodes/hrm_schedule_board.go；1053x 实际被 HRM-36 个税占用至 10533）。
-- 菜单 uuid 段 21xx: 2101 排班表 page / 2111-2117 buttons；role_menu aaa20000 21xx。

CREATE TABLE hrm_schedule_entry (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    schedule_date TEXT NOT NULL,
    employee_uuid TEXT NOT NULL,
    shift_uuid TEXT NOT NULL DEFAULT '',
    group_uuid TEXT NOT NULL DEFAULT '',
    published TEXT NOT NULL DEFAULT '0'
);

CREATE UNIQUE INDEX uk_hrm_schedule_entry ON hrm_schedule_entry (schedule_date, employee_uuid);
CREATE INDEX idx_hrm_schedule_entry_employee ON hrm_schedule_entry (employee_uuid);
CREATE INDEX idx_hrm_schedule_entry_date ON hrm_schedule_entry (schedule_date);

CREATE TRIGGER hrm_schedule_entry_update_time
AFTER UPDATE ON hrm_schedule_entry
FOR EACH ROW
BEGIN
    UPDATE hrm_schedule_entry SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- 排班表 page (假勤管理 dir aaa...000101 下, order 9)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002101','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','排班表',0,'',9,'hrm_attendance_scheduling_schedule-board','/hrm/attendance/scheduling/schedule-board','view.hrm_attendance_scheduling_schedule-board','mdi:calendar-month-outline','1','route.hrm_attendance_scheduling_schedule-board',0,'',0,0,'[]','[{"code":"v1:hrm:scheduleboard:matrix","desc":"排班矩阵查询"}]',0,'');

-- 排班表 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002111','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','保存排班单元格',0,'',0,'','','','','1','button.hrm.scheduleboard.savecell',0,'',0,0,'[]','[{"code":"v1:hrm:scheduleboard:saveCell","desc":"保存排班单元格"}]',0,'v1:hrm:scheduleboard:saveCell'),
    ('aaa10000-1111-4000-8000-000000002112','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','批量保存排班',0,'',1,'','','','','1','button.hrm.scheduleboard.batchsave',0,'',0,0,'[]','[{"code":"v1:hrm:scheduleboard:batchSave","desc":"批量保存排班"}]',0,'v1:hrm:scheduleboard:batchSave'),
    ('aaa10000-1111-4000-8000-000000002113','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','周期生成排班',0,'',2,'','','','','1','button.hrm.scheduleboard.generate',0,'',0,0,'[]','[{"code":"v1:hrm:scheduleboard:generate","desc":"周期生成排班"}]',0,'v1:hrm:scheduleboard:generate'),
    ('aaa10000-1111-4000-8000-000000002114','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','复制周排班',0,'',3,'','','','','1','button.hrm.scheduleboard.copyweek',0,'',0,0,'[]','[{"code":"v1:hrm:scheduleboard:copyWeek","desc":"复制周排班"}]',0,'v1:hrm:scheduleboard:copyWeek'),
    ('aaa10000-1111-4000-8000-000000002115','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','导入排班表',0,'',4,'','','','','1','button.hrm.scheduleboard.import',0,'',0,0,'[]','[{"code":"v1:hrm:scheduleboard:import","desc":"导入排班表"}]',0,'v1:hrm:scheduleboard:import'),
    ('aaa10000-1111-4000-8000-000000002116','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','导出排班表',0,'',5,'','','','','1','button.hrm.scheduleboard.export',0,'',0,0,'[]','[{"code":"v1:hrm:scheduleboard:export","desc":"导出排班表"}]',0,'v1:hrm:scheduleboard:export'),
    ('aaa10000-1111-4000-8000-000000002117','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002101','3','发布本月排班',0,'',6,'','','','','1','button.hrm.scheduleboard.publish',0,'',0,0,'[]','[{"code":"v1:hrm:scheduleboard:publish","desc":"发布本月排班"}]',0,'v1:hrm:scheduleboard:publish');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002101','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002101',0),
    ('aaa20000-1111-4000-8000-000000002111','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002111',0),
    ('aaa20000-1111-4000-8000-000000002112','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002112',0),
    ('aaa20000-1111-4000-8000-000000002113','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002113',0),
    ('aaa20000-1111-4000-8000-000000002114','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002114',0),
    ('aaa20000-1111-4000-8000-000000002115','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002115',0),
    ('aaa20000-1111-4000-8000-000000002116','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002116',0),
    ('aaa20000-1111-4000-8000-000000002117','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002117',0);
