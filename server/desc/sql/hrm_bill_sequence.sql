CREATE TABLE `hrm_bill_sequence` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_type_code` varchar(8) NOT NULL COMMENT '单据类型码（如 201）',
  `bill_date` char(8) NOT NULL COMMENT '编号日期 YYYYMMDD',
  `last_seq` bigint NOT NULL DEFAULT 0 COMMENT '当日已用最大 5 位流水',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_hrm_bill_sequence_type_date` (`bill_type_code`, `bill_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
