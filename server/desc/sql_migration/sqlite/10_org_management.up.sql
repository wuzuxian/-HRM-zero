CREATE TABLE hrm_org (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    parent_uuid TEXT NOT NULL,
    org_name TEXT NOT NULL,
    org_type TEXT NOT NULL,
    leader TEXT NOT NULL,
    order_num INTEGER NOT NULL,
    status TEXT NOT NULL,
    remark TEXT NOT NULL
);

CREATE INDEX idx_hrm_org_parent_uuid ON hrm_org (parent_uuid);

-- HRM 菜单目录: 人力资源
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000001','2026-11-15 00:00:00','2026-11-15 00:00:00','1','','1','人力资源',0,'',3,'hrm','/hrm','layout.base','mdi:account-group-outline','1','route.hrm',0,'',0,0,'[]','[]',0,'');

-- 组织管理 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000002','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','2','组织管理',0,'',1,'hrm_org_orginfo','/hrm/org/orginfo','view.hrm_org_orginfo','mdi:sitemap','1','route.hrm_org_orginfo',0,'',0,0,'[]','[{"code":"v1:hrm:org:list","desc":"组织列表"}]',0,'');

-- 组织管理 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000011','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000002','3','新增组织',0,'',0,'','','','','1','button.hrm.org.add',0,'',0,0,'[]','[{"code":"v1:hrm:org:add","desc":"新增组织"}]',0,'v1:hrm:org:add'),
    ('aaa10000-1111-4000-8000-000000000012','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000002','3','新增下级组织',0,'',1,'','','','','1','button.hrm.org.addchild',0,'',0,0,'[]','[{"code":"v1:hrm:org:addChild","desc":"新增下级组织"}]',0,'v1:hrm:org:addChild'),
    ('aaa10000-1111-4000-8000-000000000013','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000002','3','编辑组织',0,'',2,'','','','','1','button.hrm.org.edit',0,'',0,0,'[]','[{"code":"v1:hrm:org:edit","desc":"编辑组织"}]',0,'v1:hrm:org:edit'),
    ('aaa10000-1111-4000-8000-000000000014','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000002','3','删除组织',0,'',3,'','','','','1','button.hrm.org.delete',0,'',0,0,'[]','[{"code":"v1:hrm:org:delete","desc":"删除组织"}]',0,'v1:hrm:org:delete'),
    ('aaa10000-1111-4000-8000-000000000015','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000002','3','导入组织',0,'',4,'','','','','1','button.hrm.org.import',0,'',0,0,'[]','[{"code":"v1:hrm:org:import","desc":"导入组织"}]',0,'v1:hrm:org:import'),
    ('aaa10000-1111-4000-8000-000000000016','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000002','3','导出组织',0,'',5,'','','','','1','button.hrm.org.export',0,'',0,0,'[]','[{"code":"v1:hrm:org:export","desc":"导出组织"}]',0,'v1:hrm:org:export');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000001','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000001',0),
    ('aaa20000-1111-4000-8000-000000000002','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000002',0),
    ('aaa20000-1111-4000-8000-000000000003','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000011',0),
    ('aaa20000-1111-4000-8000-000000000004','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000012',0),
    ('aaa20000-1111-4000-8000-000000000005','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000013',0),
    ('aaa20000-1111-4000-8000-000000000006','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000014',0),
    ('aaa20000-1111-4000-8000-000000000007','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000015',0),
    ('aaa20000-1111-4000-8000-000000000008','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000016',0);

CREATE TRIGGER hrm_org_update_time
AFTER UPDATE ON hrm_org
FOR EACH ROW
BEGIN
    UPDATE hrm_org SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
