CREATE TABLE `hrm_schedule_cycle` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `cycle_name` varchar(100) NOT NULL COMMENT '周期名称',
  `shift_uuids` varchar(2000) NOT NULL COMMENT '班次序列（逗号分隔 hrm_shift.uuid，有序；周期天数=序列长度派生）',
  `group_uuids` varchar(2000) NOT NULL DEFAULT '' COMMENT '适用考勤组（逗号分隔 hrm_attendance_group.uuid，空=未绑定）',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '状态 1=启用 0=停用',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_schedule_cycle_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
