-- HRM-40 绩效指标库与考核模板（演示站 9.1 设置：指标库 / 模板管理）
DROP TABLE IF EXISTS hrm_perf_indicator;

CREATE TABLE hrm_perf_indicator (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    indicator_type TEXT NOT NULL,
    unit TEXT NOT NULL DEFAULT '',
    score_rule TEXT NOT NULL,
    data_source TEXT NOT NULL,
    default_weight REAL NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT '1',
    remark TEXT NOT NULL DEFAULT ''
);

DROP TABLE IF EXISTS hrm_perf_template;

CREATE TABLE hrm_perf_template (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT '1',
    remark TEXT NOT NULL DEFAULT ''
);

DROP TABLE IF EXISTS hrm_perf_template_item;

CREATE TABLE hrm_perf_template_item (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    template_uuid TEXT NOT NULL,
    indicator_uuid TEXT NOT NULL,
    weight REAL NOT NULL DEFAULT 0
);

CREATE INDEX idx_hrm_perf_template_item_template ON hrm_perf_template_item (template_uuid);

CREATE TRIGGER hrm_perf_indicator_update_time
AFTER UPDATE ON hrm_perf_indicator
FOR EACH ROW
BEGIN
    UPDATE hrm_perf_indicator SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_perf_template_update_time
AFTER UPDATE ON hrm_perf_template
FOR EACH ROW
BEGIN
    UPDATE hrm_perf_template SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_perf_template_item_update_time
AFTER UPDATE ON hrm_perf_template_item
FOR EACH ROW
BEGIN
    UPDATE hrm_perf_template_item SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- 绩效管理目录 + 指标库/模板管理页面 + 按钮权限（route2code 见 internal/handler/route2code.go）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2026-11-15 12:00:00','2026-11-15 12:00:00','1','aaa10000-1111-4000-8000-000000000001','1','绩效管理',0,'',2,'hrm_performance','/hrm/performance','layout.base','mdi:chart-box-outline','1','route.hrm_performance',0,'',0,0,'[]','[]',0,''),
    ('a1b2c3d4-1122-4333-8444-556677889900','2026-11-15 12:00:00','2026-11-15 12:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','指标库',0,'',1,'hrm_performance_setting_indicator','/hrm/performance/setting/indicator','view.hrm_performance_setting_indicator','mdi:target','1','route.hrm_performance_setting_indicator',0,'',0,0,'[]','[{"code":"v1:hrm:perf_indicator:list","desc":"指标库列表"}]',0,''),
    ('a1b2c3d4-1122-4333-8444-556677889901','2026-11-15 12:00:00','2026-11-15 12:00:00','1','a1b2c3d4-1122-4333-8444-556677889900','3','新增指标',0,'',0,'','','','','1','button.hrm.performance.indicator.add',0,'',0,0,'[]','[{"code":"v1:hrm:perf_indicator:add","desc":"新增指标"}]',0,'v1:hrm:perf_indicator:add'),
    ('a1b2c3d4-1122-4333-8444-556677889902','2026-11-15 12:00:00','2026-11-15 12:00:00','1','a1b2c3d4-1122-4333-8444-556677889900','3','编辑指标',0,'',1,'','','','','1','button.hrm.performance.indicator.edit',0,'',0,0,'[]','[{"code":"v1:hrm:perf_indicator:edit","desc":"编辑指标"},{"code":"v1:hrm:perf_indicator:getAll","desc":"获取所有指标"}]',0,'v1:hrm:perf_indicator:edit'),
    ('a1b2c3d4-1122-4333-8444-556677889903','2026-11-15 12:00:00','2026-11-15 12:00:00','1','a1b2c3d4-1122-4333-8444-556677889900','3','删除指标',0,'',2,'','','','','1','button.hrm.performance.indicator.delete',0,'',0,0,'[]','[{"code":"v1:hrm:perf_indicator:delete","desc":"删除指标"}]',0,'v1:hrm:perf_indicator:delete'),
    ('b2c3d4e5-1122-4333-8444-556677889900','2026-11-15 12:00:00','2026-11-15 12:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','模板管理',0,'',2,'hrm_performance_setting_template','/hrm/performance/setting/template','view.hrm_performance_setting_template','mdi:clipboard-text-outline','1','route.hrm_performance_setting_template',0,'',0,0,'[]','[{"code":"v1:hrm:perf_template:list","desc":"考核模板列表"},{"code":"v1:hrm:perf_indicator:getAll","desc":"获取所有指标"}]',0,''),
    ('b2c3d4e5-1122-4333-8444-556677889901','2026-11-15 12:00:00','2026-11-15 12:00:00','1','b2c3d4e5-1122-4333-8444-556677889900','3','新增模板',0,'',0,'','','','','1','button.hrm.performance.template.add',0,'',0,0,'[]','[{"code":"v1:hrm:perf_template:add","desc":"新增模板"},{"code":"v1:hrm:perf_indicator:getAll","desc":"获取所有指标"}]',0,'v1:hrm:perf_template:add'),
    ('b2c3d4e5-1122-4333-8444-556677889902','2026-11-15 12:00:00','2026-11-15 12:00:00','1','b2c3d4e5-1122-4333-8444-556677889900','3','编辑模板',0,'',1,'','','','','1','button.hrm.performance.template.edit',0,'',0,0,'[]','[{"code":"v1:hrm:perf_template:edit","desc":"编辑模板"},{"code":"v1:hrm:perf_template:getDetail","desc":"获取模板详情"},{"code":"v1:hrm:perf_indicator:getAll","desc":"获取所有指标"}]',0,'v1:hrm:perf_template:edit'),
    ('b2c3d4e5-1122-4333-8444-556677889903','2026-11-15 12:00:00','2026-11-15 12:00:00','1','b2c3d4e5-1122-4333-8444-556677889900','3','删除模板',0,'',2,'','','','','1','button.hrm.performance.template.delete',0,'',0,0,'[]','[{"code":"v1:hrm:perf_template:delete","desc":"删除模板"}]',0,'v1:hrm:perf_template:delete');

-- 超级管理员角色可见（casbin 策略在启动时由 role_menu + permissions JSON 同步）
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('c3d4e5f6-1122-4333-8444-556677889901','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f',0),
    ('c3d4e5f6-1122-4333-8444-556677889902','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','a1b2c3d4-1122-4333-8444-556677889900',0),
    ('c3d4e5f6-1122-4333-8444-556677889903','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','a1b2c3d4-1122-4333-8444-556677889901',0),
    ('c3d4e5f6-1122-4333-8444-556677889904','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','a1b2c3d4-1122-4333-8444-556677889902',0),
    ('c3d4e5f6-1122-4333-8444-556677889905','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','a1b2c3d4-1122-4333-8444-556677889903',0),
    ('c3d4e5f6-1122-4333-8444-556677889906','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','b2c3d4e5-1122-4333-8444-556677889900',0),
    ('c3d4e5f6-1122-4333-8444-556677889907','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','b2c3d4e5-1122-4333-8444-556677889901',0),
    ('c3d4e5f6-1122-4333-8444-556677889908','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','b2c3d4e5-1122-4333-8444-556677889902',0),
    ('c3d4e5f6-1122-4333-8444-556677889909','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','b2c3d4e5-1122-4333-8444-556677889903',0);
