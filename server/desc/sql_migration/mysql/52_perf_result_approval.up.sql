-- HRM-45 批次结果审批与结果发布/签收（演示站 9.3 结果，GLOSSARY「绩效结果」）：
--   批次到达 '4'结果已发布（HRM-43 任务链联动后）→ 提交 HR211 批次结果审批单
--   （单号 HR211-{YYYYMMDD}{5位流水}，复用 hrm_bill_sequence，bill_type_code='211'，
--   走 HRM-06 单据公共组件单步审批）。提交时从 hrm_perf_assessment 聚合快照：
--   人数 / 平均得分 / 绩效金额合计 + 公司/部门名称快照。
--   审批通过事务内（OnApproved）：生成 hrm_perf_result 结果行（每参与人一行，
--   UNIQUE(period_uuid,employee_uuid)）+ 批次 AdvancePeriod '4'→'5'已发布 + 结果行
--   published=1。系数映射（HRM-43 契约口径）：A=1.2 B=1.0 C=0.8 D=0.6 其他=1.0；
--   金额制绩效金额=final_score 直取，系数制绩效金额=系数×绩效工资基数（基数 HRM-49
--   联薪时回填，本表先落 0）。签收：员工本人对结果行 signed=1 + sign_time，幂等。
--   salary_synced（已同步薪酬）列由 HRM-49 联薪写入，本票只建列不写。
-- 错误码 10690-10694（server/internal/errcodes/hrm_perf_result.go，1069x 段 grep 空闲）。
-- 菜单 uuid 段 38xx（grep 全部 migration 确认 00xx-37xx 已用、38xx 空闲）：
--   3801 批次结果审批 page / 3802 绩效结果 page（绩效管理 dir，migration 12），
--   3811-3815 审批按钮（v1:hrm:bill:* 共享审批码）+ 3821 签收按钮；role_menu aaa20000 38xx。

DROP TABLE IF EXISTS `hrm_perf_result`;
DROP TABLE IF EXISTS `hrm_perf_result_approval`;

CREATE TABLE `hrm_perf_result_approval` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '审批单号 HR211-{YYYYMMDD}{5位流水}（复用 hrm_bill_sequence，bill_type_code=211）',
  `period_uuid` varchar(36) NOT NULL COMMENT '绩效批次 uuid（hrm_perf_period；一批次一张审批单）',
  `period_no` varchar(32) NOT NULL DEFAULT '' COMMENT '批次编号快照 PERF-年-流水',
  `period_name` varchar(64) NOT NULL DEFAULT '' COMMENT '批次名称快照',
  `company_snapshot` text NOT NULL COMMENT '公司名称快照 JSON 数组（参与人去重）',
  `org_snapshot` text NOT NULL COMMENT '部门名称快照 JSON 数组（参与人去重）',
  `participant_count` int NOT NULL DEFAULT 0 COMMENT '人数（结果行数）',
  `avg_score` decimal(8,2) NOT NULL DEFAULT 0 COMMENT '平均得分（final_score 均值）',
  `total_amount` decimal(12,2) NOT NULL DEFAULT 0 COMMENT '绩效金额合计（结果行 perf_amount 求和）',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态：unsubmitted/approving/approved/rejected/revoked（单据状态机）',
  `submit_by` varchar(64) NOT NULL DEFAULT '' COMMENT '提交人',
  `submit_time` varchar(20) NOT NULL DEFAULT '' COMMENT '提交时间',
  `approve_by` varchar(64) NOT NULL DEFAULT '' COMMENT '审批通过人',
  `approve_time` varchar(20) NOT NULL DEFAULT '' COMMENT '审批通过时间',
  `opinion` varchar(255) NOT NULL DEFAULT '' COMMENT '审批意见',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hrm_perf_result_approval_period` (`period_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_perf_result` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `approval_uuid` varchar(36) NOT NULL COMMENT '批次结果审批单 uuid（hrm_perf_result_approval）',
  `assessment_uuid` varchar(36) NOT NULL COMMENT '考核任务 uuid（hrm_perf_assessment；结果来源）',
  `period_uuid` varchar(36) NOT NULL COMMENT '绩效批次 uuid',
  `period_no` varchar(32) NOT NULL DEFAULT '' COMMENT '批次编号快照',
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工 uuid（hrm_employee）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名快照',
  `company_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '公司 uuid 快照',
  `company_name` varchar(100) NOT NULL DEFAULT '' COMMENT '公司名称快照',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '部门 uuid 快照',
  `org_name` varchar(100) NOT NULL DEFAULT '' COMMENT '部门名称快照',
  `score` decimal(8,2) NOT NULL DEFAULT 0 COMMENT '得分（assessment.final_score；等级制/金额制直存）',
  `grade` varchar(8) NOT NULL DEFAULT '' COMMENT '等级（assessment.final_grade，A/B/C/D 或等级制原值）',
  `coefficient` decimal(5,2) NOT NULL DEFAULT 1.00 COMMENT '绩效系数（等级映射 A=1.2 B=1.0 C=0.8 D=0.6 其他=1.0）',
  `perf_amount` decimal(12,2) NOT NULL DEFAULT 0 COMMENT '绩效金额（金额制=final_score；系数制=系数×绩效工资基数，基数由 HRM-49 联薪时回填）',
  `bonus_amount` decimal(12,2) NOT NULL DEFAULT 0 COMMENT '奖金金额（演示站列，默认 0）',
  `result_status` varchar(1) NOT NULL DEFAULT '1' COMMENT '结果状态：1=正常 2=已作废',
  `published` tinyint NOT NULL DEFAULT 0 COMMENT '已发布：0=否 1=是（审批通过事务内置 1）',
  `signed` tinyint NOT NULL DEFAULT 0 COMMENT '已签收：0=否 1=是（员工本人签收）',
  `sign_time` varchar(20) NOT NULL DEFAULT '' COMMENT '签收时间 YYYY-MM-DD HH:MM:SS',
  `sign_by` varchar(64) NOT NULL DEFAULT '' COMMENT '签收人',
  `salary_synced` tinyint NOT NULL DEFAULT 0 COMMENT '已同步薪酬：0=否 1=是（HRM-49 联薪写入）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hrm_perf_result_period_employee` (`period_uuid`, `employee_uuid`),
  KEY `idx_hrm_perf_result_approval` (`approval_uuid`),
  KEY `idx_hrm_perf_result_assessment` (`assessment_uuid`),
  KEY `idx_hrm_perf_result_employee` (`employee_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 绩效管理目录（migration 12 uuid f0e1d2c3-...4e5f）下新增两个结果页
-- （demo-inventory 9.3 路径：/hrm/performance/result-confirm-list + /hrm/performance/result-list）
INSERT INTO `manage_menu` (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003801','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','批次结果审批',0,'',5,'hrm_performance_result-confirm-list','/hrm/performance/result-confirm-list','view.hrm_performance_result-confirm-list','mdi:clipboard-text-check','1','route.hrm_performance_result-confirm-list',0,'',0,0,'[]','[{"code":"v1:hrm:perf_result:approvalList","desc":"批次结果审批列表"},{"code":"v1:hrm:perf_result:submit","desc":"提交结果审批"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,''),
    ('aaa10000-1111-4000-8000-000000003802','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','绩效结果',0,'',6,'hrm_performance_result-list','/hrm/performance/result-list','view.hrm_performance_result-list','mdi:chart-bar','1','route.hrm_performance_result-list',0,'',0,0,'[]','[{"code":"v1:hrm:perf_result:list","desc":"绩效结果列表"}]',0,''),
    ('aaa10000-1111-4000-8000-000000003811','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003801','3','提交单据',0,'',1,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000003812','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003801','3','审批通过',0,'',2,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000003813','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003801','3','审批驳回',0,'',3,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000003814','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003801','3','撤销单据',0,'',4,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke'),
    ('aaa10000-1111-4000-8000-000000003821','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003802','3','签收结果',0,'',0,'','','','','1','button.hrm.performance.result.sign',0,'',0,0,'[]','[{"code":"v1:hrm:perf_result:sign","desc":"签收结果"}]',0,'v1:hrm:perf_result:sign');

-- 超级管理员角色可见
INSERT INTO `manage_role_menu` (`uuid`, `create_time`, `update_time`, `role_uuid`, `menu_uuid`, `is_home`)
VALUES
    ('aaa20000-1111-4000-8000-000000003801','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003801',0),
    ('aaa20000-1111-4000-8000-000000003802','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003802',0),
    ('aaa20000-1111-4000-8000-000000003811','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003811',0),
    ('aaa20000-1111-4000-8000-000000003812','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003812',0),
    ('aaa20000-1111-4000-8000-000000003813','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003813',0),
    ('aaa20000-1111-4000-8000-000000003814','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003814',0),
    ('aaa20000-1111-4000-8000-000000003821','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003821',0);
