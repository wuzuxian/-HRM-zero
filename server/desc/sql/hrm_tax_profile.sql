CREATE TABLE `hrm_tax_profile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工（hrm_employee.uuid）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名快照',
  `id_card` varchar(18) NOT NULL DEFAULT '' COMMENT '身份证号快照',
  `resident_type` varchar(20) NOT NULL DEFAULT 'resident' COMMENT '居民类型 resident=居民个人 non_resident=非居民个人',
  `tax_city` varchar(50) NOT NULL DEFAULT '' COMMENT '税城市',
  `default_special_extra` varchar(200) NOT NULL DEFAULT '' COMMENT '默认专项附加（逗号分隔扣除项编码）',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '启停 1=启用 0=停用',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_tax_profile_employee` (`employee_uuid`),
  KEY `idx_hrm_tax_profile_name` (`employee_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
