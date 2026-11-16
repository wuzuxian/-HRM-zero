-- HRM-46 绩效结果列表与驾驶舱（演示站 9.3，GLOSSARY「绩效结果」）：
--   结果列表页菜单已在 migration 52（uuid 3802 绩效结果 /hrm/performance/result-list）。
--   本迁移仅新增 绩效驾驶舱 page（绩效管理 dir f0e1d2c3-...4e5f 下，order 7）：
--   路由 hrm_performance_dashboard / /hrm/performance/dashboard；
--   组件 view.hrm_performance_dashboard（views/hrm/performance/dashboard/）。
--   权限码 v1:hrm:perf_result:dashboard（GET /api/v1/hrm/getPerfDashboard 聚合端点）。
--   驾驶舱口径（pre-decided，见 hrm-46-notes.md）：进行中批次数=批次 status='1'；
--   考核完成率=task_status='1'/总任务；逾期待办=批次 period_end<今天 且任务未完成
--   （task_status!='1'）；风险人数=已发布结果 score<低分阈值 去重员工数；
--   平均得分=已发布结果行 score 均值；申诉中=task_status='2'。
-- 菜单 uuid 段 40xx（grep 全部 migration 确认 00xx-38xx 已用、39xx 在途保留、40xx 空闲）：
--   4051 驾驶舱 page；role_menu aaa20000 40xx。
-- 结果列表新增筛选（公司/部门快照/已同步薪酬）走既有 migration 52 页面权限码
-- v1:hrm:perf_result:list，无新增菜单行。

INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000004051','2026-11-16 12:00:00','2026-11-16 12:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','绩效驾驶舱',0,'',7,'hrm_performance_dashboard','/hrm/performance/dashboard','view.hrm_performance_dashboard','mdi:view-dashboard','1','route.hrm_performance_dashboard',0,'',0,0,'[]','[{"code":"v1:hrm:perf_result:dashboard","desc":"绩效驾驶舱"}]',0,'');

-- 超级管理员角色可见
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000004051','2026-11-16 12:00:00','2026-11-16 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000004051',0);
