CREATE TABLE `hrm_employee_contract` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工档案 uuid（HRM-03 档案引用）',
  `employee_no` varchar(20) NOT NULL DEFAULT '' COMMENT '员工工号（快照）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名（快照）',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '员工所属部门（组织 uuid，快照）',
  `contract_no` varchar(50) NOT NULL DEFAULT '' COMMENT '合同编号',
  `start_date` varchar(10) NOT NULL DEFAULT '' COMMENT '合同生效日 YYYY-MM-DD',
  `end_date` varchar(10) NOT NULL DEFAULT '' COMMENT '合同到期日 YYYY-MM-DD',
  `renew_count` bigint NOT NULL DEFAULT 0 COMMENT '续签次数',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '合同状态 1=有效 2=已到期 3=已解除',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_employee_contract_employee` (`employee_uuid`),
  KEY `idx_hrm_employee_contract_end_date` (`end_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
