CREATE TABLE `hrm_employee_family` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL,
  `name` varchar(50) NOT NULL,
  `relation` varchar(20) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `work_unit` varchar(100) NOT NULL,
  `remark` varchar(500) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_hrm_employee_family_employee_uuid` (`employee_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
