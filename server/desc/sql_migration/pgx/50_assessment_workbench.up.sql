-- HRM-44 考核任务工作台（迁移 50，演示站 9.2：/hrm/performance/assessment-task/workbench）。
-- 只读聚合工单：无新业务表，仅菜单种子。
-- 页面挂在 考核任务 dir（migration 47 uuid aaa10000-...3300）下，order 4：
--   3600 考核任务工作台 page（权限码 v1:hrm:perf_assessment:getMyWorkbench）。
-- 菜单 uuid 段 36xx（grep 全部 migration 确认 00xx-34xx 已用、35xx 在途 month-report、
--   36xx 空闲，本票声明占用 3600）；role_menu aaa20000 3600。

-- 考核任务工作台 page（挂在 考核任务 dir aaa10000-...3300 下，order 4）
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003600','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003300','2','考核任务工作台',0,'',4,'hrm_performance_assessment-task_workbench','/hrm/performance/assessment-task/workbench','view.hrm_performance_assessment-task_workbench','mdi:view-dashboard-outline','1','route.hrm_performance_assessment-task_workbench',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:getMyWorkbench","desc":"考核任务工作台"}]',0,'');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003600','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003600',0);
