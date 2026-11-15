-- HRM-02 组织架构图：页面菜单 + API 权限码（复用已有 hrm_org 表，仅菜单种子，无新业务表）
-- route2code: GET /api/v1/hrm/getOrgChartTree → v1:hrm:org:orgChart

-- 组织架构图 page (under 人力资源 dir, order 3 to avoid clashing with 组织管理=1 / 绩效管理=2)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000003','2026-11-15 12:00:00','2026-11-15 12:00:00','1','aaa10000-1111-4000-8000-000000000001','2','组织架构图',0,'',3,'hrm_org_org-chart','/hrm/org/org-chart','view.hrm_org_org-chart','mdi:chart-timeline-variant','1','route.hrm_org_org-chart',0,'',0,0,'[]',
     '[{"code":"v1:hrm:org:orgChart","desc":"组织架构图"}]',0,'');

-- 超级管理员 gets the new menu (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000009','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000003',0);
