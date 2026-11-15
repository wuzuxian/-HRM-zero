CREATE TABLE `hrm_schedule_entry` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `schedule_date` varchar(10) NOT NULL COMMENT '排班日期 yyyy-MM-dd',
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工档案 uuid (hrm_employee.uuid)',
  `shift_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '班次 uuid (hrm_shift.uuid)，空=休/未排班',
  `group_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '维护口径所属考勤组 uuid (hrm_attendance_group.uuid) 快照，空=未按组维护',
  `published` varchar(1) NOT NULL DEFAULT '0' COMMENT '发布标记 1=已发布 0=未发布（发布本月=当月排班行整批置 1）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_schedule_entry` (`schedule_date`, `employee_uuid`),
  KEY `idx_hrm_schedule_entry_employee` (`employee_uuid`),
  KEY `idx_hrm_schedule_entry_date` (`schedule_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
