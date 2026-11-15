CREATE TABLE `hrm_salary_subject` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `subject_name` varchar(100) NOT NULL COMMENT '主体名称',
  `org_uuid` varchar(36) NOT NULL COMMENT '关联公司（hrm_org 中 org_type=1 的公司节点）',
  `org_name` varchar(100) NOT NULL DEFAULT '' COMMENT '关联公司名称快照',
  `pay_account` varchar(64) NOT NULL DEFAULT '' COMMENT '发薪账户',
  `pay_bank` varchar(100) NOT NULL DEFAULT '' COMMENT '发薪银行',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '启停 1=启用 0=停用',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_salary_subject_org` (`org_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

