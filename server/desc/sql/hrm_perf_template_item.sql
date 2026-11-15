CREATE TABLE `hrm_perf_template_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `template_uuid` varchar(36) NOT NULL COMMENT '考核模板 uuid',
  `indicator_uuid` varchar(36) NOT NULL COMMENT '指标库指标 uuid',
  `weight` decimal(5,2) NOT NULL DEFAULT 0 COMMENT '权重（百分比，模板内权重之和须为 100%）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_perf_template_item_template` (`template_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
