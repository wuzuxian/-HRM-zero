CREATE TABLE `hrm_attendance_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `org_uuid` varchar(36) NOT NULL UNIQUE COMMENT '组织 uuid（每组织一条默认规则）',
  `start_time` varchar(5) NOT NULL DEFAULT '' COMMENT '上班时间 HH:mm',
  `end_time` varchar(5) NOT NULL DEFAULT '' COMMENT '下班时间 HH:mm',
  `work_days` varchar(200) NOT NULL DEFAULT '' COMMENT '考勤日，如 1,2,3,4,5',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_attendance_config_org_uuid` (`org_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
