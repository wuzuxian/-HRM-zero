-- HRM-31 我的培训（我的班次 / 我的培训证书）
-- 规范表：hrm_training_session / hrm_training_session_participant / hrm_training_certificate
-- 后续 HRM-28 培训计划+班次、HRM-30 证书模板与发放 复用这三张表，不另建。
CREATE TABLE `hrm_training_session` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `session_no` varchar(50) NOT NULL COMMENT '班次编号 PX-年-流水',
  `session_name` varchar(100) NOT NULL COMMENT '班次名称',
  `plan_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '培训计划 uuid（HRM-28 预留）',
  `lecturer` varchar(50) NOT NULL DEFAULT '' COMMENT '讲师',
  `location` varchar(200) NOT NULL DEFAULT '' COMMENT '地点',
  `hours` decimal(5,1) NOT NULL DEFAULT 0 COMMENT '学时',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime NOT NULL COMMENT '结束时间',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '状态 1=未开始 2=进行中 3=已结束',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_training_session_participant` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `session_uuid` varchar(36) NOT NULL COMMENT '班次 uuid',
  `user_uuid` varchar(36) NOT NULL COMMENT '学员（manage_user.uuid）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hrm_training_participant` (`session_uuid`, `user_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_training_certificate` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `cert_no` varchar(50) NOT NULL COMMENT '证书号 前缀-年-流水（如 JN-2026-000001）',
  `cert_name` varchar(100) NOT NULL COMMENT '证书名称',
  `cert_type` varchar(20) NOT NULL DEFAULT '' COMMENT '证书类型（HRM-30 预留：合格证/继续教育/结业证…）',
  `template_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '证书模板 uuid（HRM-30 预留）',
  `source` varchar(1) NOT NULL DEFAULT '1' COMMENT '来源 1=班次结业 2=外部导入',
  `session_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '来源班次 uuid',
  `user_uuid` varchar(36) NOT NULL COMMENT '学员（manage_user.uuid）',
  `issue_time` datetime NULL COMMENT '发证时间',
  `status` varchar(1) NOT NULL DEFAULT '0' COMMENT '状态 0=待发 1=已发 2=已撤销',
  `file_name` varchar(255) NOT NULL DEFAULT '' COMMENT '扫描件文件名',
  `file_content` mediumtext NULL COMMENT '扫描件内容 base64',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_training_cert_user` (`user_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 演示数据：2 个班次 + 班次学员（超级管理员）+ 2 张证书（1 已发带扫描件 / 1 待发）
INSERT INTO `hrm_training_session` (`uuid`, `create_time`, `update_time`, `session_no`, `session_name`, `plan_uuid`, `lecturer`, `location`, `hours`, `start_time`, `end_time`, `status`, `remark`)
VALUES
    ('aaa30000-3333-4000-8000-000000000301','2026-11-15 00:00:00','2026-11-15 00:00:00','PX-2026-000001','新员工入职培训','','王老师','总部一楼培训室',14,'2026-11-20 09:00:00','2026-11-21 17:00:00','1',''),
    ('aaa30000-3333-4000-8000-000000000302','2026-11-15 00:00:00','2026-11-15 00:00:00','PX-2026-000002','安全生产培训','','李老师','二楼会议室',6,'2026-10-10 09:00:00','2026-10-10 17:00:00','3','');

INSERT INTO `hrm_training_session_participant` (`uuid`, `create_time`, `update_time`, `session_uuid`, `user_uuid`)
VALUES
    ('aaa30000-3333-4000-8000-000000000311','2026-11-15 00:00:00','2026-11-15 00:00:00','aaa30000-3333-4000-8000-000000000301','1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f'),
    ('aaa30000-3333-4000-8000-000000000312','2026-11-15 00:00:00','2026-11-15 00:00:00','aaa30000-3333-4000-8000-000000000302','1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f');

INSERT INTO `hrm_training_certificate` (`uuid`, `create_time`, `update_time`, `cert_no`, `cert_name`, `cert_type`, `template_uuid`, `source`, `session_uuid`, `user_uuid`, `issue_time`, `status`, `file_name`, `file_content`)
VALUES
    ('aaa30000-3333-4000-8000-000000000321','2026-11-15 00:00:00','2026-11-15 00:00:00','JN-2026-000001','结业证书','结业证','','1','aaa30000-3333-4000-8000-000000000302','1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f','2026-10-12 10:00:00','1','结业证书.pdf','JVBERi0xLjQKMSAwIG9iajw8L1R5cGUvQ2F0YWxvZy9QYWdlcyAyIDAgUj4+ZW5kb2JqCjIgMCBvYmo8PC9UeXBlL1BhZ2VzL0tpZHNbMyAwIFJdL0NvdW50IDE+PmVuZG9iagozIDAgb2JqPDwvVHlwZS9QYWdlL1BhcmVudCAyIDAgUi9NZWRpYUJveFswIDAgNTk1IDg0Ml0vQ29udGVudHMgNCAwIFIvUmVzb3VyY2VzPDwvRm9udDw8L0YxIDUgMCBSPj4+Pj4+ZW5kb2JqCjQgMCBvYmo8PC9MZW5ndGggNjA+PnN0cmVhbQpCVCAvRjEgMjQgVGYgMTAwIDcwMCBUZCAoSmllWWUgWmhlbmdTaHUgRGVtbykgVGogRVQKZW5kc3RyZWFtCmVuZG9iago1IDAgb2JqPDwvVHlwZS9Gb250L1N1YnR5cGUvVHlwZTEvQmFzZUZvbnQvSGVsdmV0aWNhPj5lbmRvYmoKdHJhaWxlcjw8L1Jvb3QgMSAwIFI+PgolJUVPRg=='),
    ('aaa30000-3333-4000-8000-000000000322','2026-11-15 00:00:00','2026-11-15 00:00:00','AQ-2026-000001','安全培训合格证','安全培训证','','1','aaa30000-3333-4000-8000-000000000302','1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f',NULL,'0','','');

-- HRM 菜单目录: 培训管理
INSERT INTO `manage_menu` (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000301','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','1','培训管理',0,'',5,'hrm_training','/hrm/training','layout.base','mdi:certificate-outline','1','route.hrm_training',0,'',0,0,'[]','[]',0,'');

-- 我的培训 page
INSERT INTO `manage_menu` (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000302','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000301','2','我的培训',0,'',1,'hrm_training_my','/hrm/training/my','view.hrm_training_my','mdi:account-school-outline','1','route.hrm_training_my',0,'',0,0,'[]','[{"code":"v1:hrm:training:getMySessionList","desc":"我的班次列表"},{"code":"v1:hrm:training:getMyCertificateList","desc":"我的培训证书列表"},{"code":"v1:hrm:training:downloadMyCertificateFile","desc":"下载/预览证书扫描件"}]',0,'');

-- 我的培训 buttons
INSERT INTO `manage_menu` (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000311','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000302','3','下载/预览证书扫描件',0,'',0,'','','','','1','button.hrm.training.downloadCert',0,'',0,0,'[]','[{"code":"v1:hrm:training:downloadMyCertificateFile","desc":"下载/预览证书扫描件"}]',0,'v1:hrm:training:downloadMyCertificateFile');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO `manage_role_menu` (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000301','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000301',0),
    ('aaa20000-1111-4000-8000-000000000302','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000302',0),
    ('aaa20000-1111-4000-8000-000000000311','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000311',0);
