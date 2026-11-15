-- HRM-14 排班周期（假勤管理域 排班管理）
-- 表: hrm_schedule_cycle 排班周期
--   - 循环使用的班次序列（shift_uuids 有序引用 hrm_shift.uuid，逗号分隔），周期天数=序列长度派生
--   - group_uuids 适用考勤组（逗号分隔 hrm_attendance_group.uuid，空=未绑定）
--   - 消费契约（HRM-15 排班表）: ExpandCycle(cycleShiftIds []int64, startDate, endDate) map[dateKey]shiftUuid
--     按 纯函数 server/internal/logic/v1/hrm/schedulecycle/expand.go 周期从起始日循环展开、落月截断
-- 菜单 uuid 段: 16xx（10xx-15xx 已被假勤/证书/薪酬/加班占用, 16xx 本工单独占）

CREATE TABLE `hrm_schedule_cycle` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `cycle_name` varchar(100) NOT NULL COMMENT '周期名称',
  `shift_uuids` varchar(2000) NOT NULL COMMENT '班次序列（逗号分隔 hrm_shift.uuid，有序；周期天数=序列长度派生）',
  `group_uuids` varchar(2000) NOT NULL DEFAULT '' COMMENT '适用考勤组（逗号分隔 hrm_attendance_group.uuid，空=未绑定）',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '状态 1=启用 0=停用',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_schedule_cycle_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 排班周期 page (under 假勤管理 dir aaa...000101)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001601','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','排班周期',0,'',6,'hrm_attendance_scheduling_schedule-cycle','/hrm/attendance/scheduling/schedule-cycle','view.hrm_attendance_scheduling_schedule-cycle','mdi:calendar-refresh','1','route.hrm_attendance_scheduling_schedule-cycle',0,'',0,0,'[]','[{"code":"v1:hrm:schedulecycle:list","desc":"排班周期列表"}]',0,'');

-- 排班周期 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001611','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001601','3','新增排班周期',0,'',0,'','','','','1','button.hrm.schedulecycle.add',0,'',0,0,'[]','[{"code":"v1:hrm:schedulecycle:add","desc":"新增排班周期"}]',0,'v1:hrm:schedulecycle:add'),
    ('aaa10000-1111-4000-8000-000000001612','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001601','3','编辑排班周期',0,'',1,'','','','','1','button.hrm.schedulecycle.edit',0,'',0,0,'[]','[{"code":"v1:hrm:schedulecycle:edit","desc":"编辑排班周期"}]',0,'v1:hrm:schedulecycle:edit'),
    ('aaa10000-1111-4000-8000-000000001613','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001601','3','删除排班周期',0,'',2,'','','','','1','button.hrm.schedulecycle.delete',0,'',0,0,'[]','[{"code":"v1:hrm:schedulecycle:delete","desc":"删除排班周期"}]',0,'v1:hrm:schedulecycle:delete'),
    ('aaa10000-1111-4000-8000-000000001614','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001601','3','排班周期启停',0,'',3,'','','','','1','button.hrm.schedulecycle.togglestatus',0,'',0,0,'[]','[{"code":"v1:hrm:schedulecycle:toggleStatus","desc":"排班周期启用/停用"}]',0,'v1:hrm:schedulecycle:toggleStatus');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001601','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001601',0),
    ('aaa20000-1111-4000-8000-000000001602','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001611',0),
    ('aaa20000-1111-4000-8000-000000001603','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001612',0),
    ('aaa20000-1111-4000-8000-000000001604','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001613',0),
    ('aaa20000-1111-4000-8000-000000001605','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001614',0);
