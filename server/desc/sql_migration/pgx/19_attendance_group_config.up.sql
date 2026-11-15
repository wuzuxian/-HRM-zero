-- HRM-12 考勤组与组织默认规则（三级回退：考勤组 > 本部门规则 > 最近上级组织规则）
-- 表:
--   hrm_attendance_group 考勤组（班制 1=固定班制 2=排班制 3=自由工时; 适用范围=部门含下级+指定人员; 负责人; 启停;
--     default_shift_uuids/unassigned_shift_uuid 预留 HRM-13 排班周期/班次引用）
--   hrm_attendance_config 组织默认考勤规则（每组织一条）
-- 菜单 uuid 段: 7xx（4xx=员工档案, 6xx=人事调动，7xx 本工单独占）

CREATE TABLE "hrm_attendance_group" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    group_name varchar(100) NOT NULL,
    group_type varchar(1) NOT NULL,
    work_days varchar(200) NOT NULL DEFAULT '',
    start_time varchar(5) NOT NULL DEFAULT '',
    end_time varchar(5) NOT NULL DEFAULT '',
    flexible_minutes bigint NOT NULL DEFAULT 0,
    unassigned_shift_uuid varchar(36) NOT NULL DEFAULT '',
    default_shift_uuids varchar(2000) NOT NULL DEFAULT '',
    org_uuids varchar(2000) NOT NULL DEFAULT '',
    member_uuids varchar(2000) NOT NULL DEFAULT '',
    managers varchar(500) NOT NULL DEFAULT '',
    status varchar(1) NOT NULL DEFAULT '1',
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE TABLE "hrm_attendance_config" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    org_uuid varchar(36) NOT NULL UNIQUE,
    start_time varchar(5) NOT NULL DEFAULT '',
    end_time varchar(5) NOT NULL DEFAULT '',
    work_days varchar(200) NOT NULL DEFAULT '',
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_attendance_config_org_uuid ON "hrm_attendance_config" (org_uuid);

CREATE TRIGGER update_hrm_attendance_group_update_time BEFORE UPDATE ON "hrm_attendance_group"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

CREATE TRIGGER update_hrm_attendance_config_update_time BEFORE UPDATE ON "hrm_attendance_config"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

-- 考勤组 page (under 假勤设置 dir aaa...000101)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000701','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','考勤组',0,'',3,'hrm_attendance_setting_group','/hrm/attendance/setting/group','view.hrm_attendance_setting_group','mdi:account-group-outline','1','route.hrm_attendance_setting_group',0,'',0,0,'[]','[{"code":"v1:hrm:attendancegroup:list","desc":"考勤组列表"},{"code":"v1:hrm:attendancegroup:effectiveRule","desc":"生效规则查询"}]',0,'');

-- 考勤组 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000711','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000701','3','新增考勤组',0,'',0,'','','','','1','button.hrm.attendancegroup.add',0,'',0,0,'[]','[{"code":"v1:hrm:attendancegroup:add","desc":"新增考勤组"}]',0,'v1:hrm:attendancegroup:add'),
    ('aaa10000-1111-4000-8000-000000000712','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000701','3','编辑考勤组',0,'',1,'','','','','1','button.hrm.attendancegroup.edit',0,'',0,0,'[]','[{"code":"v1:hrm:attendancegroup:edit","desc":"编辑考勤组"}]',0,'v1:hrm:attendancegroup:edit'),
    ('aaa10000-1111-4000-8000-000000000713','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000701','3','删除考勤组',0,'',2,'','','','','1','button.hrm.attendancegroup.delete',0,'',0,0,'[]','[{"code":"v1:hrm:attendancegroup:delete","desc":"删除考勤组"}]',0,'v1:hrm:attendancegroup:delete'),
    ('aaa10000-1111-4000-8000-000000000714','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000701','3','考勤组启停',0,'',3,'','','','','1','button.hrm.attendancegroup.togglestatus',0,'',0,0,'[]','[{"code":"v1:hrm:attendancegroup:toggleStatus","desc":"考勤组启用/停用"}]',0,'v1:hrm:attendancegroup:toggleStatus'),
    ('aaa10000-1111-4000-8000-000000000715','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000701','3','生效人数查询',0,'',4,'','','','','1','button.hrm.attendancegroup.members',0,'',0,0,'[]','[{"code":"v1:hrm:attendancegroup:members","desc":"考勤组生效人数查询"}]',0,'v1:hrm:attendancegroup:members'),
    ('aaa10000-1111-4000-8000-000000000716','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000701','3','班次引用查询',0,'',5,'','','','','1','button.hrm.attendancegroup.shiftusage',0,'',0,0,'[]','[{"code":"v1:hrm:attendancegroup:shiftUsage","desc":"考勤组班次引用查询"}]',0,'v1:hrm:attendancegroup:shiftUsage');

-- 组织默认规则 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000702','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','组织默认规则',0,'',4,'hrm_attendance_setting_attendance-config','/hrm/attendance/setting/attendance-config','view.hrm_attendance_setting_attendance-config','mdi:office-building-cog-outline','1','route.hrm_attendance_setting_attendance-config',0,'',0,0,'[]','[{"code":"v1:hrm:attendanceconfig:list","desc":"组织默认规则列表"}]',0,'');

-- 组织默认规则 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000721','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000702','3','保存组织规则',0,'',0,'','','','','1','button.hrm.attendanceconfig.save',0,'',0,0,'[]','[{"code":"v1:hrm:attendanceconfig:save","desc":"保存组织默认规则（快速开始/新增/编辑共用）"}]',0,'v1:hrm:attendanceconfig:save'),
    ('aaa10000-1111-4000-8000-000000000722','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000702','3','删除组织规则',0,'',1,'','','','','1','button.hrm.attendanceconfig.delete',0,'',0,0,'[]','[{"code":"v1:hrm:attendanceconfig:delete","desc":"删除组织默认规则"}]',0,'v1:hrm:attendanceconfig:delete');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000701','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000701',0),
    ('aaa20000-1111-4000-8000-000000000702','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000702',0),
    ('aaa20000-1111-4000-8000-000000000703','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000711',0),
    ('aaa20000-1111-4000-8000-000000000704','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000712',0),
    ('aaa20000-1111-4000-8000-000000000705','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000713',0),
    ('aaa20000-1111-4000-8000-000000000706','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000714',0),
    ('aaa20000-1111-4000-8000-000000000707','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000715',0),
    ('aaa20000-1111-4000-8000-000000000708','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000716',0),
    ('aaa20000-1111-4000-8000-000000000709','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000721',0),
    ('aaa20000-1111-4000-8000-000000000710','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000722',0);
