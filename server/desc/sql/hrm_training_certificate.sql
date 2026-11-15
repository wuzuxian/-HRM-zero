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
