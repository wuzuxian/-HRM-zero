CREATE TABLE `hrm_attendance_month_lock` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `year_month` varchar(7) NOT NULL COMMENT '月份 YYYY-MM',
  `action` varchar(20) NOT NULL COMMENT '动作 lock=锁定月报 unlock=解锁 recalc=重新计算',
  `operator_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '操作人用户 uuid',
  `operator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '操作人姓名（快照）',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注（锁定/解锁原因）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_month_lock_ym` (`year_month`),
  KEY `idx_hrm_month_lock_action` (`year_month`, `action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
