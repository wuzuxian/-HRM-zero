CREATE TABLE `hrm_perf_period` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `period_no` varchar(32) NOT NULL UNIQUE COMMENT '批次编号 PERF-年-流水',
  `name` varchar(64) NOT NULL COMMENT '批次名称',
  `period_type` varchar(16) NOT NULL COMMENT '批次类型：month=月度 quarter=季度',
  `scheme_uuid` varchar(36) NOT NULL COMMENT '绩效方案（hrm_perf_scheme.uuid）',
  `scope` text NOT NULL COMMENT '适用范围多维筛选条件 JSON：companyUuids/orgUuids(含下级)/positions/posts/personStatuses',
  `period_start` varchar(10) NOT NULL COMMENT '周期开始 YYYY-MM-DD',
  `period_end` varchar(10) NOT NULL COMMENT '周期结束 YYYY-MM-DD',
  `result_publish_date` varchar(10) NOT NULL DEFAULT '' COMMENT '结果发布日期 YYYY-MM-DD',
  `status` varchar(1) NOT NULL DEFAULT '0' COMMENT '生命周期：0=草稿 1=进行中 2=待确认 3=待校准 4=结果已发布 5=已发布',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_perf_period_scheme` (`scheme_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
