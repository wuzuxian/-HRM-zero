CREATE TABLE `hrm_employee_certificate` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工档案 uuid（HRM-03 档案引用）',
  `employee_no` varchar(20) NOT NULL DEFAULT '' COMMENT '员工工号（快照）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名（快照）',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '员工所属部门（组织 uuid，快照）',
  `cert_type` varchar(50) NOT NULL DEFAULT '' COMMENT '证书类型（身份证/学历证书/职业资格证等）',
  `cert_no` varchar(100) NOT NULL DEFAULT '' COMMENT '证号（明文存储，接口出参统一脱敏）',
  `issue_date` varchar(10) NOT NULL DEFAULT '' COMMENT '发证日期 YYYY-MM-DD',
  `expiry_date` varchar(10) NOT NULL DEFAULT '' COMMENT '到期日 YYYY-MM-DD',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_employee_certificate_employee` (`employee_uuid`),
  KEY `idx_hrm_employee_certificate_expiry` (`expiry_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
