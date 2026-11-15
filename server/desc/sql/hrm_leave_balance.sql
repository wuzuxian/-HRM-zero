CREATE TABLE `hrm_leave_balance` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `year` bigint NOT NULL COMMENT '年度',
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工档案 uuid (HRM-03 hrm_employee.uuid)',
  `employee_no` varchar(20) NOT NULL DEFAULT '' COMMENT '工号（快照）',
  `employee_name` varchar(100) NOT NULL DEFAULT '' COMMENT '姓名（快照）',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '部门（组织 uuid 快照）',
  `leave_type_code` varchar(50) NOT NULL COMMENT '假期类型编码 (hrm_leave_type.type_code)',
  `granted` decimal(6,1) NOT NULL DEFAULT 0 COMMENT '应发 = 初始 + 调整',
  `used` decimal(6,1) NOT NULL DEFAULT 0 COMMENT '已扣减（请假消费）',
  `frozen` decimal(6,1) NOT NULL DEFAULT 0 COMMENT '冻结（审批中的请假单占用，预留）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_leave_balance` (`year`, `employee_uuid`, `leave_type_code`),
  KEY `idx_hrm_leave_balance_employee` (`employee_uuid`),
  KEY `idx_hrm_leave_balance_year` (`year`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
