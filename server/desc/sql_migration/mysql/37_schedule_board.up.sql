-- HRM-15 排班表（假勤管理域 排班管理；demo-inventory 4.3）
-- 表:
--   - hrm_schedule_entry 排班表行：(schedule_date, employee_uuid) 唯一 —— 一人一天一行，
--     shift_uuid 空=休/未排班；group_uuid 维护口径考勤组快照；published 发布标记。
--   - 排班表按考勤组维护：矩阵行成员 = 考勤组适用范围展开（复用 attendancegroup expandGroupMembers
--     同口径，见 hrm-12-notes）。周期生成复用 HRM-14 ExpandCycle（相位锚定 startDate，
--     跨月保持绝对相位需传同一历史锚点日，见 hrm-14-notes）。
-- 发布口径（HRM-16/17/22/23 消费方必读，详见 hrm-notes/hrm-15-notes.md）:
--   - 日结（punch.SettleDay）取班次时【发布与否都生效】：已排班行即生效；
--   - 排班制（group_type=2）员工无排班日（无 hrm_schedule_entry 行或 shift_uuid 空）时，
--     兜底走考勤组/组织默认规则（三级回退 attendancegroup.ResolveEffectiveRule 口径）；
--   - fixed 固定班制员工不读排班表，直接走规则；
--   - published 仅是管理流程标记（未发布月份排班可继续编辑，不拦截打卡）。
-- 错误码 1053x 段（server/internal/errcodes/hrm_schedule_board.go）。
-- 菜单 uuid 段 21xx（grep 全部 migration: 0xxx 基础/03xx 假勤设置/07xx 考勤组/09xx-20xx 各域；
-- 本票取 21xx: 2101 排班表 page / 2111-2117 buttons；role_menu aaa20000 21xx）。

CREATE TABLE `hrm_schedule_entry` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `schedule_date` varchar(10) NOT NULL COMMENT '排班日期 yyyy-MM-dd',
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工档案 uuid (hrm_employee.uuid)',
  `shift_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '班次 uuid (hrm_shift.uuid)，空=休/未排班',
  `group_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '维护口径所属考勤组 uuid 快照，空=未按组维护',
  `published` varchar(1) NOT NULL DEFAULT '0' COMMENT '发布标记 1=已发布 0=未发布',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_schedule_entry` (`schedule_date`, `employee_uuid`),
  KEY `idx_hrm_schedule_entry_employee` (`employee_uuid`),
  KEY `idx_hrm_schedule_entry_date` (`schedule_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 排班表 page (假勤管理 dir aaa...000101 下, order 9；排班管理组: 排班周期 order 6)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002101','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','排班表',0,'',9,'hrm_attendance_scheduling_schedule-board','/hrm/attendance/scheduling/schedule-board','view.hrm_attendance_scheduling_schedule-board','mdi:calendar-month-outline','1','route.hrm_attendance_scheduling_schedule-board',0,'',0,0,'[]','[{"code":"v1:hrm:scheduleboard:matrix","desc":"排班矩阵查询"}]',0,'');

-- 排班表 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
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
