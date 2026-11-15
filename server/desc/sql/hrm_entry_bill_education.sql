CREATE TABLE `hrm_entry_bill_education` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `entry_uuid` varchar(36) NOT NULL COMMENT '入职单 uuid',
  `school` varchar(200) NOT NULL DEFAULT '' COMMENT '学校',
  `major` varchar(200) NOT NULL DEFAULT '' COMMENT '专业',
  `education` varchar(50) NOT NULL DEFAULT '' COMMENT '学历',
  `start_date` varchar(10) NOT NULL DEFAULT '' COMMENT '开始日期 YYYY-MM-DD',
  `end_date` varchar(10) NOT NULL DEFAULT '' COMMENT '结束日期 YYYY-MM-DD',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_entry_bill_education_entry_uuid` (`entry_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
