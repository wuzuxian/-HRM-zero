CREATE TABLE `hrm_org` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `parent_uuid` varchar(36) NOT NULL,
  `org_name` varchar(100) NOT NULL,
  `org_type` varchar(1) NOT NULL,
  `leader` varchar(50) NOT NULL,
  `order_num` bigint NOT NULL,
  `status` varchar(1) NOT NULL,
  `remark` varchar(500) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_hrm_org_parent_uuid` (`parent_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
