CREATE TABLE `hrm_salary_scheme_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `scheme_uuid` varchar(36) NOT NULL COMMENT '薪酬方案 uuid',
  `item_uuid` varchar(36) NOT NULL COMMENT '薪资项目 uuid（hrm_salary_item.uuid；不复制快照，试算/展示实时取项目名/编码）',
  `override_amount` decimal(18,2) NULL COMMENT '覆盖金额（固定项可覆盖项目默认值；NULL=继承项目默认）',
  `override_formula` varchar(500) NOT NULL DEFAULT '' COMMENT '覆盖公式（公式项可覆盖项目默认公式；空=继承项目默认）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_salary_scheme_item_scheme` (`scheme_uuid`),
  KEY `idx_hrm_salary_scheme_item_item` (`item_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
