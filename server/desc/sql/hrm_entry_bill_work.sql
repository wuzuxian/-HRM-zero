CREATE TABLE `hrm_entry_bill_work` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `entry_uuid` varchar(36) NOT NULL COMMENT '入职单 uuid',
  `company` varchar(200) NOT NULL DEFAULT '' COMMENT '工作单位',
  `position` varchar(100) NOT NULL DEFAULT '' COMMENT '职位',
  `start_date` varchar(10) NOT NULL DEFAULT '' COMMENT '开始日期 YYYY-MM-DD',
  `end_date` varchar(10) NOT NULL DEFAULT '' COMMENT '结束日期 YYYY-MM-DD',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_entry_bill_work_entry_uuid` (`entry_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
