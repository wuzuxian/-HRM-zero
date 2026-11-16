-- final-review F5 (spec 用户故事 23: 每类人事单据支持导出):
-- 共享导出端点 GET /api/v1/hrm/exportBillList 的按钮码种子。挂在人事管理目录
-- （uuid 601）下，permissions code 与 route2code 生成码 v1:hrm:bill:exportBillList
-- 一致（authx 按 casbin 精确匹配该码）。菜单 uuid 段 69x（grep 全部 migration
-- 确认 6xx 段仅 601-684 已用，69x 空闲）。

INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000691','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000601','3','导出单据',0,'',0,'','','','','1','button.hrm.bill.export',0,'',0,0,'[]','[{"code":"v1:hrm:bill:exportBillList","desc":"导出单据列表"}]',0,'v1:hrm:bill:exportBillList');

-- 超级管理员 gets the button (casbin rules derived at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000691','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000691',0);
