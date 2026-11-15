CREATE TABLE `hrm_tax_special_deduction` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工（hrm_employee.uuid）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名快照',
  `deduct_year` varchar(4) NOT NULL COMMENT '扣除年度（如 2026）',
  `deduct_month` bigint NOT NULL COMMENT '扣除月份 1-12',
  `deduct_type` varchar(20) NOT NULL COMMENT '扣除项 CHILDREN_EDU/CONTINUING_EDU/CRITICAL_ILLNESS/HOUSE_LOAN/HOUSE_RENT/SUPPORT_ELDERLY/INFANT_CARE',
  `amount` decimal(12,2) NOT NULL DEFAULT 0 COMMENT '扣除金额（元/月）',
  `source` varchar(20) NOT NULL DEFAULT 'manual' COMMENT '来源 manual=手工维护 import=导入',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_tax_special_emp_year_month_type` (`employee_uuid`, `deduct_year`, `deduct_month`, `deduct_type`),
  KEY `idx_hrm_tax_special_emp_year` (`employee_uuid`, `deduct_year`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
