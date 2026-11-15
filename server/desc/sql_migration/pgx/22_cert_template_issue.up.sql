-- HRM-30 证书模板与证书发放
-- 证书模板 CRUD（模板名称/证书类型/来源 系统|企业/启用停用/编号前缀）；
-- 证书发放复用迁移 15 的 hrm_training_certificate（不另建表、不改列）：
--   批量发证（班次结业学员按模板生成证书号=前缀-年-流水，source='1'，template_uuid 回填），
--   导入外部证书（source='2'），预览/下载扫描件，撤销（status 1→2）。
-- 证书号流水复用迁移 11 的 hrm_bill_sequence（bill_type_code=编号前缀, bill_date=4位年, 按前缀+年递增），
-- 与班次编号 PX（HRM-28 GenerateSessionNo）同一存储/分配模式。

CREATE TABLE "hrm_training_cert_template" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    template_name varchar(100) NOT NULL,
    cert_type varchar(20) NOT NULL DEFAULT '',
    source varchar(1) NOT NULL DEFAULT '1',
    no_prefix varchar(8) NOT NULL,
    status varchar(1) NOT NULL DEFAULT '1',
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX uni_hrm_training_cert_template_prefix ON "hrm_training_cert_template" (no_prefix);

CREATE TRIGGER update_hrm_training_cert_template_update_time BEFORE UPDATE ON "hrm_training_cert_template"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

-- 演示数据：2 个系统模板（结业证 JN / 安全培训证 AQ，与迁移 15 演示证书号前缀一致）+ 1 个企业停用模板
INSERT INTO "hrm_training_cert_template" (uuid, create_time, update_time, template_name, cert_type, source, no_prefix, status, remark)
VALUES
    ('aaa30000-3333-4000-8000-000000000361','2026-11-15 00:00:00','2026-11-15 00:00:00','结业证书模板','结业证','1','JN','1','班次结业发证'),
    ('aaa30000-3333-4000-8000-000000000362','2026-11-15 00:00:00','2026-11-15 00:00:00','安全培训合格证模板','安全培训证','1','AQ','1','安全生产培训发证'),
    ('aaa30000-3333-4000-8000-000000000363','2026-11-15 00:00:00','2026-11-15 00:00:00','企业荣誉证书模板','荣誉','2','YX','2','企业自建模板（演示停用）');

-- 证书模板 page（复用迁移 15 的 培训管理 目录 aaa10000-...-000000000301）
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000305','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000301','2','证书模板',0,'',4,'hrm_training_cert-template','/hrm/training/cert-template','view.hrm_training_cert-template','mdi:certificate','1','route.hrm_training_cert-template',0,'',0,0,'[]','[{"code":"v1:hrm:cert_template:list","desc":"证书模板列表"}]',0,'');

-- 证书发放 page
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000306','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000301','2','证书发放',0,'',5,'hrm_training_cert-issue','/hrm/training/cert-issue','view.hrm_training_cert-issue','mdi:card-account-details-star','1','route.hrm_training_cert-issue',0,'',0,0,'[]','[{"code":"v1:hrm:certificate:list","desc":"证书发放列表"},{"code":"v1:hrm:certificate:downloadFile","desc":"下载/预览证书扫描件"}]',0,'');

-- 证书模板 buttons
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000365','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000305','3','新增模板',0,'',0,'','','','','1','button.hrm.training.certTemplate.add',0,'',0,0,'[]','[{"code":"v1:hrm:cert_template:add","desc":"新增模板"}]',0,'v1:hrm:cert_template:add'),
    ('aaa10000-1111-4000-8000-000000000366','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000305','3','编辑模板',0,'',1,'','','','','1','button.hrm.training.certTemplate.edit',0,'',0,0,'[]','[{"code":"v1:hrm:cert_template:edit","desc":"编辑模板"}]',0,'v1:hrm:cert_template:edit'),
    ('aaa10000-1111-4000-8000-000000000367','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000305','3','启用/停用模板',0,'',2,'','','','','1','button.hrm.training.certTemplate.toggleStatus',0,'',0,0,'[]','[{"code":"v1:hrm:cert_template:toggleStatus","desc":"启用/停用模板"}]',0,'v1:hrm:cert_template:toggleStatus'),
    ('aaa10000-1111-4000-8000-000000000368','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000305','3','删除模板',0,'',3,'','','','','1','button.hrm.training.certTemplate.delete',0,'',0,0,'[]','[{"code":"v1:hrm:cert_template:delete","desc":"删除模板"}]',0,'v1:hrm:cert_template:delete');

-- 证书发放 buttons
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000371','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000306','3','批量发证',0,'',0,'','','','','1','button.hrm.training.certificate.batchIssue',0,'',0,0,'[]','[{"code":"v1:hrm:certificate:batchIssue","desc":"批量发证"}]',0,'v1:hrm:certificate:batchIssue'),
    ('aaa10000-1111-4000-8000-000000000372','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000306','3','导入外部证书',0,'',1,'','','','','1','button.hrm.training.certificate.import',0,'',0,0,'[]','[{"code":"v1:hrm:certificate:importExternal","desc":"导入外部证书"}]',0,'v1:hrm:certificate:importExternal'),
    ('aaa10000-1111-4000-8000-000000000373','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000306','3','发放证书',0,'',2,'','','','','1','button.hrm.training.certificate.issue',0,'',0,0,'[]','[{"code":"v1:hrm:certificate:issue","desc":"发放证书（待发→已发）"}]',0,'v1:hrm:certificate:issue'),
    ('aaa10000-1111-4000-8000-000000000374','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000306','3','撤销证书',0,'',3,'','','','','1','button.hrm.training.certificate.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:certificate:revoke","desc":"撤销证书"}]',0,'v1:hrm:certificate:revoke'),
    ('aaa10000-1111-4000-8000-000000000375','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000306','3','上传证书扫描件',0,'',4,'','','','','1','button.hrm.training.certificate.uploadFile',0,'',0,0,'[]','[{"code":"v1:hrm:certificate:uploadFile","desc":"上传证书扫描件"}]',0,'v1:hrm:certificate:uploadFile');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000305','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000305',0),
    ('aaa20000-1111-4000-8000-000000000306','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000306',0),
    ('aaa20000-1111-4000-8000-000000000365','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000365',0),
    ('aaa20000-1111-4000-8000-000000000366','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000366',0),
    ('aaa20000-1111-4000-8000-000000000367','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000367',0),
    ('aaa20000-1111-4000-8000-000000000368','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000368',0),
    ('aaa20000-1111-4000-8000-000000000371','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000371',0),
    ('aaa20000-1111-4000-8000-000000000372','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000372',0),
    ('aaa20000-1111-4000-8000-000000000373','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000373',0),
    ('aaa20000-1111-4000-8000-000000000374','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000374',0),
    ('aaa20000-1111-4000-8000-000000000375','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000375',0);
