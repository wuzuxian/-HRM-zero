-- HRM-17 个人考勤看板（假勤管理域 我的假勤 4.1）
-- 无新业务表：今日卡片/本月统计/考勤日历全部从 hrm_punch_record settle_* 列聚合
-- （HRM-16 落库契约），年假可用从 hrm_leave_balance 聚合（HRM-18），
-- 加班时长从 hrm_overtime_bill 已审批通过单据 confirm_hours 聚合（HRM-21）。
-- 本迁移仅落菜单种子：假勤管理 dir aaa...000101 下 我的假勤-个人看板 page（order 11）
-- + 刷新看板 button；role_menu aaa20000 rows。
-- 菜单 uuid 段 28xx（grep 全部 migration: 20xx 打卡/21xx 排班周期/23xx 请假销假/24xx 补卡/
-- 25xx 职位/26xx 加班，28xx 空闲）：2801 个人看板 page / 2811 刷新看板 button。
-- 错误码 1061x 段（server/internal/errcodes/hrm_dashboard.go）。

-- 我的假勤-个人看板 page (假勤管理 dir aaa...000101 下, order 11; 1-10 已占用)
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002801','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','个人看板',0,'',11,'hrm_attendance_my_dashboard','/hrm/attendance/my/dashboard','view.hrm_attendance_my_dashboard','mdi:view-dashboard-outline','1','route.hrm_attendance_my_dashboard',0,'',0,0,'[]','[{"code":"v1:hrm:dashboard:getMyDashboard","desc":"个人考勤看板"}]',0,'');

-- 个人看板 buttons
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002811','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002801','3','刷新看板',0,'',0,'','','','','1','button.hrm.dashboard.view',0,'',0,0,'[]','[{"code":"v1:hrm:dashboard:getMyDashboard","desc":"刷新个人考勤看板"}]',0,'v1:hrm:dashboard:getMyDashboard');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002801','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002801',0),
    ('aaa20000-1111-4000-8000-000000002811','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002811',0);
