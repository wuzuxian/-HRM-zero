CREATE TABLE `hrm_training_cert_template` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `template_name` varchar(100) NOT NULL COMMENT '模板名称',
  `cert_type` varchar(20) NOT NULL DEFAULT '' COMMENT '证书类型（合格证/继续教育/荣誉/安全培训证/结业证/技能证明/上岗证/企业…）',
  `source` varchar(1) NOT NULL DEFAULT '1' COMMENT '来源 1=系统 2=企业',
  `no_prefix` varchar(8) NOT NULL COMMENT '编号前缀（HF/CE/YX/XF/RZ/JN/SG/AQ/HG/JX…），证书号=前缀-年-流水',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '状态 1=启用 2=停用（停用后不可发证）',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_hrm_training_cert_template_prefix` (`no_prefix`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
