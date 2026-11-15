CREATE TABLE `hrm_shift` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `shift_name` varchar(100) NOT NULL,
  `start_time` varchar(5) NOT NULL COMMENT '上班时间 HH:mm',
  `end_time` varchar(5) NOT NULL COMMENT '下班时间 HH:mm，早于上班时间表示跨天（如 22:00-次日06:00）',
  `cross_day` bigint NOT NULL DEFAULT 0 COMMENT '是否跨天 1=是 0=否（保存时按上下班时间推导）',
  `rest_start` varchar(5) NOT NULL DEFAULT '' COMMENT '休息时段开始 HH:mm，空表示无休息',
  `rest_end` varchar(5) NOT NULL DEFAULT '' COMMENT '休息时段结束 HH:mm，空表示无休息',
  `standard_hours` decimal(5,2) NOT NULL DEFAULT 0 COMMENT '标准工时（小时）＝上下班时长－休息时长',
  `flexible_minutes` bigint NOT NULL DEFAULT 0 COMMENT '弹性分钟数（0=不弹性）',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '状态 1=启用 0=停用',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
