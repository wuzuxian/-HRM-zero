-- HRM-31 我的培训（我的班次 / 我的培训证书）
-- 规范表：hrm_training_session / hrm_training_session_participant / hrm_training_certificate
-- 后续 HRM-28 培训计划+班次、HRM-30 证书模板与发放 复用这三张表，不另建。
CREATE TABLE "hrm_training_session" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    session_no varchar(50) NOT NULL UNIQUE,
    session_name varchar(100) NOT NULL,
    plan_uuid varchar(36) NOT NULL DEFAULT '',
    lecturer varchar(50) NOT NULL DEFAULT '',
    location varchar(200) NOT NULL DEFAULT '',
    hours numeric(5,1) NOT NULL DEFAULT 0,
    start_time timestamp NOT NULL,
    end_time timestamp NOT NULL,
    status varchar(1) NOT NULL DEFAULT '1',
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE TABLE "hrm_training_session_participant" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    session_uuid varchar(36) NOT NULL,
    user_uuid varchar(36) NOT NULL,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX idx_hrm_training_session_participant ON "hrm_training_session_participant" (session_uuid, user_uuid);

CREATE TABLE "hrm_training_certificate" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cert_no varchar(50) NOT NULL UNIQUE,
    cert_name varchar(100) NOT NULL,
    cert_type varchar(20) NOT NULL DEFAULT '',
    template_uuid varchar(36) NOT NULL DEFAULT '',
    source varchar(1) NOT NULL DEFAULT '1',
    session_uuid varchar(36) NOT NULL DEFAULT '',
    user_uuid varchar(36) NOT NULL,
    issue_time timestamp NULL,
    status varchar(1) NOT NULL DEFAULT '0',
    file_name varchar(255) NOT NULL DEFAULT '',
    file_content text NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_training_cert_user ON "hrm_training_certificate" (user_uuid);

CREATE TRIGGER update_hrm_training_session_update_time BEFORE UPDATE ON "hrm_training_session"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

CREATE TRIGGER update_hrm_training_session_participant_update_time BEFORE UPDATE ON "hrm_training_session_participant"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

CREATE TRIGGER update_hrm_training_certificate_update_time BEFORE UPDATE ON "hrm_training_certificate"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

-- 演示数据：2 个班次 + 班次学员（超级管理员）+ 2 张证书（1 已发带扫描件 / 1 待发）
INSERT INTO "hrm_training_session" (uuid, create_time, update_time, session_no, session_name, plan_uuid, lecturer, location, hours, start_time, end_time, status, remark)
VALUES
    ('aaa30000-3333-4000-8000-000000000301','2026-11-15 00:00:00','2026-11-15 00:00:00','PX-2026-000001','新员工入职培训','','王老师','总部一楼培训室',14,'2026-11-20 09:00:00','2026-11-21 17:00:00','1',''),
    ('aaa30000-3333-4000-8000-000000000302','2026-11-15 00:00:00','2026-11-15 00:00:00','PX-2026-000002','安全生产培训','','李老师','二楼会议室',6,'2026-10-10 09:00:00','2026-10-10 17:00:00','3','');

INSERT INTO "hrm_training_session_participant" (uuid, create_time, update_time, session_uuid, user_uuid)
VALUES
    ('aaa30000-3333-4000-8000-000000000311','2026-11-15 00:00:00','2026-11-15 00:00:00','aaa30000-3333-4000-8000-000000000301','1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f'),
    ('aaa30000-3333-4000-8000-000000000312','2026-11-15 00:00:00','2026-11-15 00:00:00','aaa30000-3333-4000-8000-000000000302','1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f');

INSERT INTO "hrm_training_certificate" (uuid, create_time, update_time, cert_no, cert_name, cert_type, template_uuid, source, session_uuid, user_uuid, issue_time, status, file_name, file_content)
VALUES
    ('aaa30000-3333-4000-8000-000000000321','2026-11-15 00:00:00','2026-11-15 00:00:00','JN-2026-000001','结业证书','结业证','','1','aaa30000-3333-4000-8000-000000000302','1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f','2026-10-12 10:00:00','1','结业证书.pdf','JVBERi0xLjQKMSAwIG9iajw8L1R5cGUvQ2F0YWxvZy9QYWdlcyAyIDAgUj4+ZW5kb2JqCjIgMCBvYmo8PC9UeXBlL1BhZ2VzL0tpZHNbMyAwIFJdL0NvdW50IDE+PmVuZG9iagozIDAgb2JqPDwvVHlwZS9QYWdlL1BhcmVudCAyIDAgUi9NZWRpYUJveFswIDAgNTk1IDg0Ml0vQ29udGVudHMgNCAwIFIvUmVzb3VyY2VzPDwvRm9udDw8L0YxIDUgMCBSPj4+Pj4+ZW5kb2JqCjQgMCBvYmo8PC9MZW5ndGggNjA+PnN0cmVhbQpCVCAvRjEgMjQgVGYgMTAwIDcwMCBUZCAoSmllWWUgWmhlbmdTaHUgRGVtbykgVGogRVQKZW5kc3RyZWFtCmVuZG9iago1IDAgb2JqPDwvVHlwZS9Gb250L1N1YnR5cGUvVHlwZTEvQmFzZUZvbnQvSGVsdmV0aWNhPj5lbmRvYmoKdHJhaWxlcjw8L1Jvb3QgMSAwIFI+PgolJUVPRg=='),
    ('aaa30000-3333-4000-8000-000000000322','2026-11-15 00:00:00','2026-11-15 00:00:00','AQ-2026-000001','安全培训合格证','安全培训证','','1','aaa30000-3333-4000-8000-000000000302','1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f',NULL,'0','','');

-- HRM 菜单目录: 培训管理
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000301','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','1','培训管理',0,'',5,'hrm_training','/hrm/training','layout.base','mdi:certificate-outline','1','route.hrm_training',0,'',0,0,'[]','[]',0,'');

-- 我的培训 page
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000302','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000301','2','我的培训',0,'',1,'hrm_training_my','/hrm/training/my','view.hrm_training_my','mdi:account-school-outline','1','route.hrm_training_my',0,'',0,0,'[]','[{"code":"v1:hrm:training:getMySessionList","desc":"我的班次列表"},{"code":"v1:hrm:training:getMyCertificateList","desc":"我的培训证书列表"},{"code":"v1:hrm:training:downloadMyCertificateFile","desc":"下载/预览证书扫描件"}]',0,'');

-- 我的培训 buttons
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000311','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000302','3','下载/预览证书扫描件',0,'',0,'','','','','1','button.hrm.training.downloadCert',0,'',0,0,'[]','[{"code":"v1:hrm:training:downloadMyCertificateFile","desc":"下载/预览证书扫描件"}]',0,'v1:hrm:training:downloadMyCertificateFile');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000301','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000301',0),
    ('aaa20000-1111-4000-8000-000000000302','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000302',0),
    ('aaa20000-1111-4000-8000-000000000311','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000311',0);
