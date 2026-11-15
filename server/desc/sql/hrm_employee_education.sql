CREATE TABLE `hrm_employee_education` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL,
  `school` varchar(100) NOT NULL,
  `major` varchar(100) NOT NULL,
  `education` varchar(20) NOT NULL,
  `start_date` varchar(10) NOT NULL,
  `end_date` varchar(10) NOT NULL,
  `remark` varchar(500) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_hrm_employee_education_employee_uuid` (`employee_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
