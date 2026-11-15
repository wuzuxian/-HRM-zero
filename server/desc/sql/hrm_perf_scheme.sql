CREATE TABLE `hrm_perf_scheme` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `name` varchar(64) NOT NULL COMMENT '方案名称',
  `template_uuid` varchar(36) NOT NULL COMMENT '关联考核模板（hrm_perf_template.uuid）',
  `period_type` varchar(16) NOT NULL COMMENT '周期类型：month=月度 quarter=季度 half_year=半年 year=年度',
  `assessment_mode` varchar(16) NOT NULL COMMENT '评估模式：kpi=KPI kpi_behavior=KPI+行为混合',
  `score_mode` varchar(16) NOT NULL COMMENT '评分模式：percent=百分制 grade=等级制 amount=金额制',
  `pay_link_mode` varchar(16) NOT NULL COMMENT '联薪方式：coefficient=系数制 amount=金额制',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '状态：1=启用 2=停用',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_perf_scheme_template` (`template_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
