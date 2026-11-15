CREATE TABLE `hrm_recruitment_candidate` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(100) NOT NULL COMMENT '姓名',
  `phone` varchar(32) NOT NULL DEFAULT '' COMMENT '手机号',
  `email` varchar(128) NOT NULL DEFAULT '' COMMENT '邮箱',
  `source_channel` varchar(32) NOT NULL DEFAULT 'other' COMMENT '来源渠道 recommend=内部推荐 website=招聘网站 campus=校园招聘 headhunter=猎头 other=其他',
  `education` varchar(32) NOT NULL DEFAULT '' COMMENT '最高学历 doctor=博士 master=硕士 bachelor=本科 associate=大专 senior=高中 other=其他',
  `current_company` varchar(100) NOT NULL DEFAULT '' COMMENT '当前公司',
  `current_position` varchar(100) NOT NULL DEFAULT '' COMMENT '当前岗位',
  `expect_salary` varchar(64) NOT NULL DEFAULT '' COMMENT '期望薪资（自由文本，如 20k-30k）',
  `status` varchar(16) NOT NULL DEFAULT 'new' COMMENT '候选人状态 new=新线索 delivered=已投递 interviewing=面试中 offer=已发Offer hired=录用通过 rejected=不通过',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_recruitment_candidate_status` (`status`),
  KEY `idx_hrm_recruitment_candidate_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
