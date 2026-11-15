CREATE TABLE `hrm_entry_bill_family` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `entry_uuid` varchar(36) NOT NULL COMMENT '入职单 uuid',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '姓名',
  `relation` varchar(50) NOT NULL DEFAULT '' COMMENT '关系',
  `phone` varchar(50) NOT NULL DEFAULT '' COMMENT '电话',
  `work_unit` varchar(200) NOT NULL DEFAULT '' COMMENT '工作单位',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_entry_bill_family_entry_uuid` (`entry_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
