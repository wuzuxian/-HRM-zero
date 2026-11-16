CREATE TABLE `hrm_training_session_participant` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `session_uuid` varchar(36) NOT NULL COMMENT '班次 uuid',
  `user_uuid` varchar(36) NOT NULL COMMENT '学员（manage_user.uuid）',
  `sign_in_time` datetime NULL COMMENT '签到时间（NULL=未签到）',
  `sign_in_method` varchar(1) NOT NULL DEFAULT '' COMMENT '签到方式 1=扫码自签 2=管理员补签（空=未签）',
  `sign_in_by` varchar(36) NOT NULL DEFAULT '' COMMENT '签到操作人（自签=本人 user_uuid，补签=管理员 user_uuid）',
  `sign_in_reason` varchar(200) NOT NULL DEFAULT '' COMMENT '补签原因',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hrm_training_participant` (`session_uuid`, `user_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
