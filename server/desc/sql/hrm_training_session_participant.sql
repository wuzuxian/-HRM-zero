CREATE TABLE `hrm_training_session_participant` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `session_uuid` varchar(36) NOT NULL COMMENT '班次 uuid',
  `user_uuid` varchar(36) NOT NULL COMMENT '学员（manage_user.uuid）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hrm_training_participant` (`session_uuid`, `user_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
