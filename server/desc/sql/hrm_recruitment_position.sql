CREATE TABLE `hrm_recruitment_position` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `post_name` varchar(100) NOT NULL COMMENT '岗位名称',
  `requirement_uuid` varchar(36) NOT NULL COMMENT '来源需求单（hrm_recruitment_requirement.uuid，仅 status=approved 可引用）',
  `post_category` varchar(32) NOT NULL DEFAULT 'tech' COMMENT '岗位类别 tech=技术类 product=产品类 design=设计类 operation=运营类 market=市场类 function=职能类 other=其他',
  `owner_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '招聘负责人 uuid',
  `owner_name` varchar(100) NOT NULL DEFAULT '' COMMENT '招聘负责人姓名快照',
  `headcount` int NOT NULL DEFAULT 1 COMMENT '招聘人数',
  `hired_count` int NOT NULL DEFAULT 0 COMMENT '已录用人数（录用审批 HRM-27 通过 IncreaseHiredCount 事务化递增）',
  `publish_date` varchar(10) NOT NULL DEFAULT '' COMMENT '发布日期 YYYY-MM-DD',
  `status` varchar(16) NOT NULL DEFAULT 'open' COMMENT '职位状态 open=招聘中 full=已满 closed=已关闭',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_recruitment_position_requirement` (`requirement_uuid`),
  KEY `idx_hrm_recruitment_position_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
