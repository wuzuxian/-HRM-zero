-- HRM-06 单据审批底座（ADR-0001）：共享审批记录表 + 单据编号流水表
DROP TABLE IF EXISTS `hrm_approval_record`;

CREATE TABLE `hrm_approval_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_type_code` varchar(8) NOT NULL COMMENT '单据类型码（如 201=入职、205=请假）',
  `bill_uuid` varchar(36) NOT NULL COMMENT '单据 uuid（各单据表的主键标识）',
  `bill_no` varchar(32) NOT NULL COMMENT '单据编号 HR{类型码}-{YYYYMMDD}{5位流水}',
  `action` varchar(16) NOT NULL COMMENT '动作：submit=提交 approve=审批通过 reject=驳回 revoke=撤销',
  `from_status` varchar(16) NOT NULL COMMENT '动作前单据状态',
  `to_status` varchar(16) NOT NULL COMMENT '动作后单据状态',
  `operator_user_uuid` varchar(36) NOT NULL COMMENT '操作人用户 uuid',
  `operator_name` varchar(64) NOT NULL DEFAULT '' COMMENT '操作人姓名快照',
  `opinion` varchar(255) NOT NULL DEFAULT '' COMMENT '审批意见',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_approval_record_bill` (`bill_type_code`, `bill_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

DROP TABLE IF EXISTS `hrm_bill_sequence`;

CREATE TABLE `hrm_bill_sequence` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_type_code` varchar(8) NOT NULL COMMENT '单据类型码（如 201）',
  `bill_date` char(8) NOT NULL COMMENT '编号日期 YYYYMMDD',
  `last_seq` bigint NOT NULL DEFAULT 0 COMMENT '当日已用最大 5 位流水',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_hrm_bill_sequence_type_date` (`bill_type_code`, `bill_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
