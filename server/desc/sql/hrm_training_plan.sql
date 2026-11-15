CREATE TABLE `hrm_training_plan` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '单据编号 HR216-YYYYMMDD+5位流水',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态 unsubmitted/approving/approved/rejected/revoked',
  `plan_name` varchar(100) NOT NULL COMMENT '计划名称',
  `theme` varchar(200) NOT NULL DEFAULT '' COMMENT '主题',
  `plan_year` varchar(4) NOT NULL DEFAULT '' COMMENT '年度',
  `plan_type` varchar(20) NOT NULL DEFAULT 'special' COMMENT '计划类型 special=专项 new_hire=新员工培养 safety=安全培训 skill=技能提升',
  `start_date` varchar(10) NOT NULL DEFAULT '' COMMENT '开始日期 YYYY-MM-DD',
  `end_date` varchar(10) NOT NULL DEFAULT '' COMMENT '结束日期 YYYY-MM-DD',
  `published` tinyint NOT NULL DEFAULT 0 COMMENT '发布状态 0=草稿 1=已发布（审批通过后置 1）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '创建人 uuid',
  `creator_name` varchar(50) NOT NULL DEFAULT '' COMMENT '创建人姓名快照',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_training_plan_status` (`status`, `published`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
