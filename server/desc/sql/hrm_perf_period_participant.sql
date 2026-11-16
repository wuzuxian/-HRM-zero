CREATE TABLE `hrm_perf_period_participant` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `period_uuid` varchar(36) NOT NULL COMMENT '绩效批次 uuid',
  `employee_uuid` varchar(36) NOT NULL COMMENT '参与员工（hrm_employee.uuid）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hrm_perf_period_participant` (`period_uuid`, `employee_uuid`),
  KEY `idx_hrm_perf_period_participant_employee` (`employee_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
