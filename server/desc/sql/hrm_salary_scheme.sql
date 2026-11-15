CREATE TABLE `hrm_salary_scheme` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `scheme_name` varchar(100) NOT NULL COMMENT '方案名称',
  `subject_uuid` varchar(36) NOT NULL COMMENT '关联薪酬主体（hrm_salary_subject.uuid）',
  `subject_name` varchar(100) NOT NULL DEFAULT '' COMMENT '关联主体名称快照',
  `person_status` varchar(100) NOT NULL DEFAULT '' COMMENT '适用人员状态（逗号分隔的 hrm_employee.person_status 值 1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休）',
  `settle_cycle` varchar(20) NOT NULL DEFAULT 'monthly' COMMENT '结算周期 monthly=月度',
  `pay_day` bigint NOT NULL DEFAULT 1 COMMENT '发薪日（几号，1-31）',
  `currency` varchar(10) NOT NULL DEFAULT 'CNY' COMMENT '币种（ISO 4217 编码）',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '启停 1=启用 0=停用',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_salary_scheme_subject` (`subject_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
