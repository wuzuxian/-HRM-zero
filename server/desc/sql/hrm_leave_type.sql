CREATE TABLE `hrm_leave_type` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `type_code` varchar(50) NOT NULL UNIQUE COMMENT '类型编码 annual_leave/comp_leave/sick_leave/other_leave…',
  `type_name` varchar(50) NOT NULL,
  `type_value` bigint NOT NULL DEFAULT 0 COMMENT '类型值（请假单引用的数字编码）',
  `balance_enabled` bigint NOT NULL DEFAULT 0 COMMENT '余额管理开关 1=是 0=否',
  `attachment_required` bigint NOT NULL DEFAULT 0 COMMENT '需要附件 1=是 0=否',
  `default_days` decimal(5,1) NOT NULL DEFAULT 0 COMMENT '默认天数',
  `order_num` bigint NOT NULL DEFAULT 0 COMMENT '排序',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '启用状态 1=启用 0=停用',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_leave_type_code` (`type_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
