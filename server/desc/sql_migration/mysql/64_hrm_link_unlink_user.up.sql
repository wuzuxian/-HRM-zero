-- HRM-51 员工↔账号关联/解绑端点权限种子 (spec #53 / ADR 0007): mysql 方言, 与 pgx/sqlite/64 同号同集合。
--   1. 员工档案页下 4 个新动作按钮 (关联账号/解除关联/候选账号搜索/新建并关联账号),
--      route2code: POST /api/v1/hrm/linkEmployeeUser → v1:hrm:employee:linkUser
--                  POST /api/v1/hrm/unlinkEmployeeUser → v1:hrm:employee:unlinkUser
--                  GET  /api/v1/hrm/getLinkableUserList → v1:hrm:employee:linkableUserList
--                  POST /api/v1/hrm/createAndLinkUser → v1:hrm:employee:createAndLinkUser
--      (menu uuid 0417-0420, role_menu 行 aaa20000-...0417-0420, grep 全部迁移确认空闲)
--   2. 用户管理页 permissions 追加只读反查码 (列表「关联员工」列):
--                  GET /api/v1/manage/getLinkedEmployees → v1:manage:user:linkedEmployees
--      只读列不设按钮行, 随页面码授予 (迁移 62 页面级追加的既有模式)。
--   3. 为持有上述菜单的角色补 casbin_rule 存量行 (应用首启 InitCasbinRule 也会重建, 双保险)。

-- casbin_rule 表由 casbin adapter 在应用首启时创建; 全新环境上迁移先于首启执行,
-- 因此触及 casbin_rule 的迁移必须自带幂等建表 (否则 fresh deploy 在此迁移即失败)。
CREATE TABLE IF NOT EXISTS casbin_rule (
    id     BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 员工档案页新动作按钮 (幂等: 已存在则跳过)
INSERT INTO `manage_menu` (`uuid`, `create_time`, `update_time`, `status`, `parent_uuid`, `menu_type`, `menu_name`, `hide_in_menu`, `active_menu`, `order`, `route_name`, `route_path`, `component`, `icon`, `icon_type`, `i18n_key`, `keep_alive`, `href`, `multi_tab`, `fixed_index_in_tab`, `query`, `permissions`, `constant`, `button_code`)
SELECT 'aaa10000-1111-4000-8000-000000000417','2026-11-17 00:00:00','2026-11-17 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','关联账号',0,'',6,'','','','','1','button.hrm.employee.linkUser',0,'',0,0,'[]','[{"code":"v1:hrm:employee:linkUser","desc":"关联系统账号"}]',0,'v1:hrm:employee:linkUser'
WHERE NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM `manage_menu` WHERE `uuid` = 'aaa10000-1111-4000-8000-000000000417') x);

INSERT INTO `manage_menu` (`uuid`, `create_time`, `update_time`, `status`, `parent_uuid`, `menu_type`, `menu_name`, `hide_in_menu`, `active_menu`, `order`, `route_name`, `route_path`, `component`, `icon`, `icon_type`, `i18n_key`, `keep_alive`, `href`, `multi_tab`, `fixed_index_in_tab`, `query`, `permissions`, `constant`, `button_code`)
SELECT 'aaa10000-1111-4000-8000-000000000418','2026-11-17 00:00:00','2026-11-17 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','解除关联',0,'',7,'','','','','1','button.hrm.employee.unlinkUser',0,'',0,0,'[]','[{"code":"v1:hrm:employee:unlinkUser","desc":"解除账号关联"}]',0,'v1:hrm:employee:unlinkUser'
WHERE NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM `manage_menu` WHERE `uuid` = 'aaa10000-1111-4000-8000-000000000418') x);

INSERT INTO `manage_menu` (`uuid`, `create_time`, `update_time`, `status`, `parent_uuid`, `menu_type`, `menu_name`, `hide_in_menu`, `active_menu`, `order`, `route_name`, `route_path`, `component`, `icon`, `icon_type`, `i18n_key`, `keep_alive`, `href`, `multi_tab`, `fixed_index_in_tab`, `query`, `permissions`, `constant`, `button_code`)
SELECT 'aaa10000-1111-4000-8000-000000000419','2026-11-17 00:00:00','2026-11-17 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','候选账号搜索',0,'',8,'','','','','1','button.hrm.employee.linkableUserList',0,'',0,0,'[]','[{"code":"v1:hrm:employee:linkableUserList","desc":"候选账号搜索"}]',0,'v1:hrm:employee:linkableUserList'
WHERE NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM `manage_menu` WHERE `uuid` = 'aaa10000-1111-4000-8000-000000000419') x);

INSERT INTO `manage_menu` (`uuid`, `create_time`, `update_time`, `status`, `parent_uuid`, `menu_type`, `menu_name`, `hide_in_menu`, `active_menu`, `order`, `route_name`, `route_path`, `component`, `icon`, `icon_type`, `i18n_key`, `keep_alive`, `href`, `multi_tab`, `fixed_index_in_tab`, `query`, `permissions`, `constant`, `button_code`)
SELECT 'aaa10000-1111-4000-8000-000000000420','2026-11-17 00:00:00','2026-11-17 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','新建并关联账号',0,'',9,'','','','','1','button.hrm.employee.createAndLinkUser',0,'',0,0,'[]','[{"code":"v1:hrm:employee:createAndLinkUser","desc":"新建账号并关联"}]',0,'v1:hrm:employee:createAndLinkUser'
WHERE NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM `manage_menu` WHERE `uuid` = 'aaa10000-1111-4000-8000-000000000420') x);

-- 超级管理员补授 4 个新按钮 (幂等; 其余角色经菜单管理自行授予)
INSERT INTO `manage_role_menu` (`uuid`, `create_time`, `update_time`, `role_uuid`, `menu_uuid`, `is_home`)
SELECT 'aaa20000-1111-4000-8000-000000000417','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000417',0
WHERE NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM `manage_role_menu` WHERE `role_uuid` = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d' AND `menu_uuid` = 'aaa10000-1111-4000-8000-000000000417') x);

INSERT INTO `manage_role_menu` (`uuid`, `create_time`, `update_time`, `role_uuid`, `menu_uuid`, `is_home`)
SELECT 'aaa20000-1111-4000-8000-000000000418','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000418',0
WHERE NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM `manage_role_menu` WHERE `role_uuid` = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d' AND `menu_uuid` = 'aaa10000-1111-4000-8000-000000000418') x);

INSERT INTO `manage_role_menu` (`uuid`, `create_time`, `update_time`, `role_uuid`, `menu_uuid`, `is_home`)
SELECT 'aaa20000-1111-4000-8000-000000000419','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000419',0
WHERE NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM `manage_role_menu` WHERE `role_uuid` = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d' AND `menu_uuid` = 'aaa10000-1111-4000-8000-000000000419') x);

INSERT INTO `manage_role_menu` (`uuid`, `create_time`, `update_time`, `role_uuid`, `menu_uuid`, `is_home`)
SELECT 'aaa20000-1111-4000-8000-000000000420','2026-11-17 00:00:00','2026-11-17 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000420',0
WHERE NOT EXISTS (SELECT 1 FROM (SELECT 1 FROM `manage_role_menu` WHERE `role_uuid` = '1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d' AND `menu_uuid` = 'aaa10000-1111-4000-8000-000000000420') x);

-- 用户管理页 permissions 追加只读反查码 (幂等: LIKE 守卫, 已追加过则 REPLACE 无匹配)
UPDATE manage_menu SET permissions = REPLACE(permissions, '"}]', '"},{"code":"v1:manage:user:linkedEmployees","desc":"关联员工反查"}]')
WHERE uuid = 'b3c4d5e6-f7a8-4912-83b4-c5d6e7f8a9b0'
  AND permissions LIKE '%v1:manage:user:list%'
  AND permissions NOT LIKE '%v1:manage:user:linkedEmployees%';

-- casbin 存量行: 持有上述新菜单的角色 × 菜单 permissions JSON 里的全部 API 码
-- (与 role set_menus / InitCasbinRule 的重建语义一致; 幂等)
INSERT INTO casbin_rule (p_type, v0, v1)
SELECT DISTINCT 'p', rm.role_uuid, jt.code
FROM manage_role_menu rm
JOIN manage_menu mm ON mm.uuid = rm.menu_uuid
JOIN JSON_TABLE(mm.permissions, '$[*]' COLUMNS (code VARCHAR(255) PATH '$.code')) jt
WHERE rm.menu_uuid IN (
        'aaa10000-1111-4000-8000-000000000417',
        'aaa10000-1111-4000-8000-000000000418',
        'aaa10000-1111-4000-8000-000000000419',
        'aaa10000-1111-4000-8000-000000000420',
        'b3c4d5e6-f7a8-4912-83b4-c5d6e7f8a9b0'
      )
  AND jt.code <> ''
  AND NOT EXISTS (
    SELECT 1 FROM casbin_rule c
    WHERE c.p_type = 'p' AND c.v0 = rm.role_uuid AND c.v1 = jt.code
  );
