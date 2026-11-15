-- HRM-41 绩效方案（演示站 9.1 绩效方案）：方案配置 = 关联模板 + 周期类型 + 评估模式 + 评分模式 + 联薪方式 + 启停。
-- 联薪方式 pay_link_mode（coefficient=系数制 amount=金额制）是 HRM-49 绩效联薪接入薪资试算的关键契约字段，勿改名。

DROP TABLE IF EXISTS "hrm_perf_scheme";

CREATE TABLE "hrm_perf_scheme" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name varchar(64) NOT NULL,
    template_uuid varchar(36) NOT NULL,
    period_type varchar(16) NOT NULL,
    assessment_mode varchar(16) NOT NULL,
    score_mode varchar(16) NOT NULL,
    pay_link_mode varchar(16) NOT NULL,
    status varchar(1) NOT NULL DEFAULT '1',
    remark varchar(255) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_perf_scheme_template ON "hrm_perf_scheme" (template_uuid);

CREATE TRIGGER update_hrm_perf_scheme_update_time BEFORE UPDATE ON "hrm_perf_scheme"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

-- 绩效管理目录（migration 12 已建 uuid f0e1d2c3-...4e5f）下新增「绩效方案」页面 + 按钮权限
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('d4e5f6a7-1122-4333-8444-556677889900','2026-11-15 12:00:00','2026-11-15 12:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','绩效方案',0,'',3,'hrm_performance_scheme','/hrm/performance/scheme','view.hrm_performance_scheme','mdi:file-document-edit-outline','1','route.hrm_performance_scheme',0,'',0,0,'[]','[{"code":"v1:hrm:perf_scheme:list","desc":"绩效方案列表"}]',0,''),
    ('d4e5f6a7-1122-4333-8444-556677889901','2026-11-15 12:00:00','2026-11-15 12:00:00','1','d4e5f6a7-1122-4333-8444-556677889900','3','新增方案',0,'',0,'','','','','1','button.hrm.performance.scheme.add',0,'',0,0,'[]','[{"code":"v1:hrm:perf_scheme:add","desc":"新增绩效方案"},{"code":"v1:hrm:perf_template:list","desc":"考核模板列表"}]',0,'v1:hrm:perf_scheme:add'),
    ('d4e5f6a7-1122-4333-8444-556677889902','2026-11-15 12:00:00','2026-11-15 12:00:00','1','d4e5f6a7-1122-4333-8444-556677889900','3','编辑方案',0,'',1,'','','','','1','button.hrm.performance.scheme.edit',0,'',0,0,'[]','[{"code":"v1:hrm:perf_scheme:edit","desc":"编辑绩效方案"},{"code":"v1:hrm:perf_template:list","desc":"考核模板列表"}]',0,'v1:hrm:perf_scheme:edit'),
    ('d4e5f6a7-1122-4333-8444-556677889903','2026-11-15 12:00:00','2026-11-15 12:00:00','1','d4e5f6a7-1122-4333-8444-556677889900','3','删除方案',0,'',2,'','','','','1','button.hrm.performance.scheme.delete',0,'',0,0,'[]','[{"code":"v1:hrm:perf_scheme:delete","desc":"删除绩效方案"}]',0,'v1:hrm:perf_scheme:delete');

-- 超级管理员角色可见
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('c4d5e6f7-1122-4333-8444-556677889901','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','d4e5f6a7-1122-4333-8444-556677889900',0),
    ('c4d5e6f7-1122-4333-8444-556677889902','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','d4e5f6a7-1122-4333-8444-556677889901',0),
    ('c4d5e6f7-1122-4333-8444-556677889903','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','d4e5f6a7-1122-4333-8444-556677889902',0),
    ('c4d5e6f7-1122-4333-8444-556677889904','2026-11-15 12:00:00','2026-11-15 12:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','d4e5f6a7-1122-4333-8444-556677889903',0);
