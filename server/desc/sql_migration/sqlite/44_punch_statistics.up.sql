-- HRM-22 考勤统计与外部数据导入（假勤管理域 假勤数据；demo-inventory 4.4）
--   仅菜单种子：考勤统计 page（统计卡 + 按人明细 + 查看明细弹窗）。
--   无新业务表 —— 统计聚合走 hrm_punch_record settle_* 快照列（HRM-16 日结落库契约，
--   见 hrm-notes/hrm-16-notes.md「SettleDay 消费契约」）。
--   外部考勤数据导入/手动日结已在 HRM-16 打卡记录页落地（/hrm/importPunchRecord CSV +
--   hrm:punch:import / hrm:punch:settle 按钮），本票不重复实现。
-- API codes: v1:hrm:punch:{statistics,statisticsDetail}
--   （/api/v1/hrm/{getPunchStatistics,getPunchStatisticsDetail}，group v1/hrm/punch ——
--   端点复用 punch.api group，route2code 自动生成 v1:hrm:punch:statistics{,detail}）。
-- 错误码 1063x 段（server/internal/errcodes/hrm_punch_statistics.go）。
-- 菜单 uuid 段 30xx（grep 全部 migration: 0xxx 基础/03xx 假勤设置/07xx 考勤组/09xx-29xx 各域；
-- 本票取 30xx: 3001 考勤统计 page / 3011-3012 buttons；role_menu aaa20000 30xx）。

-- 假勤数据-考勤统计 page (假勤管理 dir aaa...000101 下, order 11)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003001','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','考勤统计',0,'',11,'hrm_attendance_data_punch-statistics','/hrm/attendance/data/punch-statistics','view.hrm_attendance_data_punch-statistics','mdi:chart-box-outline','1','route.hrm_attendance_data_punch-statistics',0,'',0,0,'[]','[{"code":"v1:hrm:punch:statistics","desc":"考勤统计查询"},{"code":"v1:hrm:punch:statisticsDetail","desc":"考勤统计明细"}]',0,'');

-- 考勤统计 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003011','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003001','3','查询统计',0,'',0,'','','','','1','button.hrm.punch.statistics',0,'',0,0,'[]','[{"code":"v1:hrm:punch:statistics","desc":"考勤统计查询"}]',0,'v1:hrm:punch:statistics'),
    ('aaa10000-1111-4000-8000-000000003012','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003001','3','查看考勤明细',0,'',1,'','','','','1','button.hrm.punch.statisticsDetail',0,'',0,0,'[]','[{"code":"v1:hrm:punch:statisticsDetail","desc":"考勤统计明细"}]',0,'v1:hrm:punch:statisticsDetail');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003001','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003001',0),
    ('aaa20000-1111-4000-8000-000000003011','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003011',0),
    ('aaa20000-1111-4000-8000-000000003012','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003012',0);
