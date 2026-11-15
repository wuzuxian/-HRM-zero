CREATE TABLE `hrm_perf_indicator` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `code` varchar(32) NOT NULL COMMENT '指标编码（唯一）',
  `name` varchar(64) NOT NULL COMMENT '指标名称',
  `indicator_type` varchar(16) NOT NULL COMMENT '指标类型：kpi=结果指标(KPI) behavior=行为能力',
  `unit` varchar(32) NOT NULL DEFAULT '' COMMENT '计量单位',
  `score_rule` varchar(16) NOT NULL COMMENT '评分规则：formula=公式评分 manual=手工评分',
  `data_source` varchar(16) NOT NULL COMMENT '数据来源：system=系统取数 manual=手工填报',
  `default_weight` decimal(5,2) NOT NULL DEFAULT 0 COMMENT '默认权重（百分比，如 40.00 表示 40%）',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '状态：1=启用 2=停用',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_hrm_perf_indicator_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
