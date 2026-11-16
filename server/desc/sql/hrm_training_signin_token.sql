CREATE TABLE `hrm_training_signin_token` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `session_uuid` varchar(36) NOT NULL COMMENT '班次 uuid',
  `token` varchar(64) NOT NULL COMMENT '课堂动态码 token（轮换）',
  `expire_time` datetime NOT NULL COMMENT '失效时间（短时效轮换）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '发起刷新的管理员 user_uuid',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_training_signin_token_session` (`session_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
