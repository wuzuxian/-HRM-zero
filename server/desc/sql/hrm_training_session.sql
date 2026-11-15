CREATE TABLE `hrm_training_session` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `session_no` varchar(50) NOT NULL COMMENT '班次编号 PX-年-流水',
  `session_name` varchar(100) NOT NULL COMMENT '班次名称',
  `plan_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '培训计划 uuid（HRM-28 预留）',
  `lecturer` varchar(50) NOT NULL DEFAULT '' COMMENT '讲师',
  `location` varchar(200) NOT NULL DEFAULT '' COMMENT '地点',
  `hours` decimal(5,1) NOT NULL DEFAULT 0 COMMENT '学时',
  `start_time` datetime NOT NULL COMMENT '开始时间',
  `end_time` datetime NOT NULL COMMENT '结束时间',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '状态 1=未开始 2=进行中 3=已结束',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
