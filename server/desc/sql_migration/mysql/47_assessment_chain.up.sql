-- HRM-43 考核任务链（演示站 9.2 批次与任务流，GLOSSARY「考核任务」）：
--   批次发布后为每个参与人生成考核任务（考核单号 PA-年-流水，复用 hrm_bill_sequence，
--   bill_type_code='PA'，日期=批次周期开始年份）。五阶段流转（前置守卫：前一节点完成
--   才可下一节点）：目标确认 → 自评填报 → 上级评审 → 校准填报 → 结果确认（可发起申诉）。
--   评分按模板快照 items 权重汇总（percent 百分制→等级 A/B/C/D；grade/amount 直存）。
--   申诉表 hrm_perf_appeal：发起→校准人裁决（维持/调整得分）→关闭。
-- 批次状态联动（联动点，HRM-42 状态机不改动）：
--   - 全部任务到达校准填报（current_node>='4'）→ 联动 AdvancePeriod 至'3'待校准。
--   - 全部任务结果确认完成（result_time 非空）→ 联动 AdvancePeriod 至'4'结果已发布。
--   HRM-45 审批通过后推'4'→'5'已发布（不改本表）。
-- 错误码 10640-10647（server/internal/errcodes/hrm_perf_assessment.go，1063x 已被占用）。
-- 菜单 uuid 段 33xx（grep 全部 migration 确认 00xx-32xx 已用、33xx 空闲）。

DROP TABLE IF EXISTS `hrm_perf_appeal`;
DROP TABLE IF EXISTS `hrm_perf_assessment`;

CREATE TABLE `hrm_perf_assessment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `assessment_no` varchar(32) NOT NULL COMMENT '考核单号 PA-年-流水（复用 hrm_bill_sequence，bill_type_code=PA）',
  `period_uuid` varchar(36) NOT NULL COMMENT '绩效批次 uuid（hrm_perf_period）',
  `employee_uuid` varchar(36) NOT NULL COMMENT '被考核人（hrm_employee.uuid）',
  `template_uuid` varchar(36) NOT NULL COMMENT '考核模板 uuid（快照来源）',
  `items_snapshot` text NOT NULL COMMENT '模板快照 JSON：[{indicatorUuid,indicatorCode,indicatorName,indicatorType,unit,scoreRule,dataSource,weight}]',
  `current_node` varchar(1) NOT NULL DEFAULT '1' COMMENT '当前节点：1=目标确认 2=自评填报 3=上级评审 4=校准填报 5=结果确认',
  `task_status` varchar(1) NOT NULL DEFAULT '0' COMMENT '任务状态：0=进行中 1=已完成 2=申诉中',
  `plan_time` varchar(20) NOT NULL DEFAULT '' COMMENT '目标确认时间 YYYY-MM-DD HH:MM:SS',
  `plan_by` varchar(64) NOT NULL DEFAULT '' COMMENT '目标确认人',
  `self_score` decimal(8,2) NOT NULL DEFAULT 0 COMMENT '自评得分（percent=权重汇总；amount=金额；grade=0）',
  `self_grade` varchar(8) NOT NULL DEFAULT '' COMMENT '自评等级（percent→A/B/C/D；grade=直存）',
  `self_items` text NOT NULL COMMENT '自评逐项得分 JSON：[{indicatorUuid,score}]',
  `self_comment` varchar(255) NOT NULL DEFAULT '' COMMENT '自评说明',
  `self_time` varchar(20) NOT NULL DEFAULT '' COMMENT '自评时间',
  `self_by` varchar(64) NOT NULL DEFAULT '' COMMENT '自评人',
  `review_score` decimal(8,2) NOT NULL DEFAULT 0 COMMENT '上级评审得分',
  `review_grade` varchar(8) NOT NULL DEFAULT '' COMMENT '上级评审等级',
  `review_items` text NOT NULL COMMENT '评审逐项得分 JSON：[{indicatorUuid,score}]',
  `review_comment` varchar(255) NOT NULL DEFAULT '' COMMENT '评审意见',
  `review_time` varchar(20) NOT NULL DEFAULT '' COMMENT '评审时间',
  `review_by` varchar(64) NOT NULL DEFAULT '' COMMENT '评审人（被考核人部门负责人，简化=有评审码者）',
  `final_score` decimal(8,2) NOT NULL DEFAULT 0 COMMENT '最终得分（校准人调整后）',
  `final_grade` varchar(8) NOT NULL DEFAULT '' COMMENT '最终等级',
  `calib_comment` varchar(255) NOT NULL DEFAULT '' COMMENT '校准说明',
  `calib_time` varchar(20) NOT NULL DEFAULT '' COMMENT '校准时间',
  `calib_by` varchar(64) NOT NULL DEFAULT '' COMMENT '校准人',
  `result_time` varchar(20) NOT NULL DEFAULT '' COMMENT '结果确认时间',
  `result_by` varchar(64) NOT NULL DEFAULT '' COMMENT '结果确认人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hrm_perf_assessment_period_employee` (`period_uuid`, `employee_uuid`),
  KEY `idx_hrm_perf_assessment_period` (`period_uuid`),
  KEY `idx_hrm_perf_assessment_employee` (`employee_uuid`),
  KEY `idx_hrm_perf_assessment_node` (`current_node`, `task_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_perf_appeal` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `assessment_uuid` varchar(36) NOT NULL COMMENT '考核任务 uuid（hrm_perf_assessment）',
  `period_uuid` varchar(36) NOT NULL COMMENT '绩效批次 uuid',
  `employee_uuid` varchar(36) NOT NULL COMMENT '发起申诉的被考核人',
  `reason` varchar(500) NOT NULL COMMENT '申诉原因',
  `status` varchar(1) NOT NULL DEFAULT '0' COMMENT '申诉状态：0=待处理 1=已裁决-维持原结果 2=已裁决-调整得分',
  `decision_comment` varchar(500) NOT NULL DEFAULT '' COMMENT '裁决说明',
  `adjusted_score` decimal(8,2) NOT NULL DEFAULT 0 COMMENT '调整后最终得分（status=2 时生效）',
  `adjusted_grade` varchar(8) NOT NULL DEFAULT '' COMMENT '调整后最终等级（status=2 时生效）',
  `handled_by` varchar(64) NOT NULL DEFAULT '' COMMENT '裁决人（校准人）',
  `handled_time` varchar(20) NOT NULL DEFAULT '' COMMENT '裁决时间',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_perf_appeal_assessment` (`assessment_uuid`),
  KEY `idx_hrm_perf_appeal_period` (`period_uuid`),
  KEY `idx_hrm_perf_appeal_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 绩效管理目录（migration 12 uuid f0e1d2c3-...4e5f）下新增四个填报列表页
-- （demo-inventory 9.2 路径：/hrm/performance/assessment-task/{plan,self-review,leader-review,result-confirm}-list）
INSERT INTO `manage_menu` (`uuid`, `create_time`, `update_time`, `status`, `parent_uuid`, `menu_type`, `menu_name`, `hide_in_menu`, `active_menu`, `order`, `route_name`, `route_path`, `component`, `icon`, `icon_type`, `i18n_key`, `keep_alive`, `href`, `multi_tab`, `fixed_index_in_tab`, `query`, `permissions`, `constant`, `button_code`)
VALUES
    ('aaa10000-1111-4000-8000-000000003301','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','目标确认',0,'',5,'hrm_performance_assessment-task_plan-list','/hrm/performance/assessment-task/plan-list','view.hrm_performance_assessment-task_plan-list','mdi:target','1','route.hrm_performance_assessment-task_plan-list',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:list","desc":"考核任务列表"},{"code":"v1:hrm:perf_assessment:detail","desc":"考核任务详情"}]',0,''),
    ('aaa10000-1111-4000-8000-000000003302','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','自评填报',0,'',6,'hrm_performance_assessment-task_self-review-list','/hrm/performance/assessment-task/self-review-list','view.hrm_performance_assessment-task_self-review-list','mdi:pencil-box','1','route.hrm_performance_assessment-task_self-review-list',0,'',0,0,'[]','[]',0,''),
    ('aaa10000-1111-4000-8000-000000003303','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','评审填报',0,'',7,'hrm_performance_assessment-task_leader-review-list','/hrm/performance/assessment-task/leader-review-list','view.hrm_performance_assessment-task_leader-review-list','mdi:account-check','1','route.hrm_performance_assessment-task_leader-review-list',0,'',0,0,'[]','[]',0,''),
    ('aaa10000-1111-4000-8000-000000003304','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','结果确认',0,'',8,'hrm_performance_assessment-task_result-confirm-task-list','/hrm/performance/assessment-task/result-confirm-task-list','view.hrm_performance_assessment-task_result-confirm-task-list','mdi:clipboard-check','1','route.hrm_performance_assessment-task_result-confirm-task-list',0,'',0,0,'[]','[]',0,''),
    ('aaa10000-1111-4000-8000-000000003311','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003301','3','生成考核任务',0,'',0,'','','','','1','button.hrm.performance.assessment.generate',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:generate","desc":"生成考核任务"}]',0,'v1:hrm:perf_assessment:generate'),
    ('aaa10000-1111-4000-8000-000000003312','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003301','3','确认目标',0,'',1,'','','','','1','button.hrm.performance.assessment.plan',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:confirmPlan","desc":"目标确认"}]',0,'v1:hrm:perf_assessment:confirmPlan'),
    ('aaa10000-1111-4000-8000-000000003313','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003302','3','自评提交',0,'',0,'','','','','1','button.hrm.performance.assessment.selfReview',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:submitSelfReview","desc":"自评填报"}]',0,'v1:hrm:perf_assessment:submitSelfReview'),
    ('aaa10000-1111-4000-8000-000000003314','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003303','3','评审提交',0,'',0,'','','','','1','button.hrm.performance.assessment.leaderReview',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:review","desc":"上级评审"}]',0,'v1:hrm:perf_assessment:review'),
    ('aaa10000-1111-4000-8000-000000003315','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003303','3','校准提交',0,'',1,'','','','','1','button.hrm.performance.assessment.calibrate',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:calibrate","desc":"校准填报"}]',0,'v1:hrm:perf_assessment:calibrate'),
    ('aaa10000-1111-4000-8000-000000003316','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003304','3','确认结果',0,'',0,'','','','','1','button.hrm.performance.assessment.confirmResult',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:confirmResult","desc":"结果确认"}]',0,'v1:hrm:perf_assessment:confirmResult'),
    ('aaa10000-1111-4000-8000-000000003317','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003304','3','发起申诉',0,'',1,'','','','','1','button.hrm.performance.assessment.appeal',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:appeal","desc":"发起申诉"}]',0,'v1:hrm:perf_assessment:appeal'),
    ('aaa10000-1111-4000-8000-000000003318','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003304','3','申诉裁决',0,'',2,'','','','','1','button.hrm.performance.assessment.handleAppeal',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:handleAppeal","desc":"申诉裁决"}]',0,'v1:hrm:perf_assessment:handleAppeal');

-- 超级管理员角色可见
INSERT INTO `manage_role_menu` (`uuid`, `create_time`, `update_time`, `role_uuid`, `menu_uuid`, `is_home`)
VALUES
    ('aaa20000-1111-4000-8000-000000003301','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003301',0),
    ('aaa20000-1111-4000-8000-000000003302','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003302',0),
    ('aaa20000-1111-4000-8000-000000003303','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003303',0),
    ('aaa20000-1111-4000-8000-000000003304','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003304',0),
    ('aaa20000-1111-4000-8000-000000003311','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003311',0),
    ('aaa20000-1111-4000-8000-000000003312','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003312',0),
    ('aaa20000-1111-4000-8000-000000003313','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003313',0),
    ('aaa20000-1111-4000-8000-000000003314','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003314',0),
    ('aaa20000-1111-4000-8000-000000003315','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003315',0),
    ('aaa20000-1111-4000-8000-000000003316','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003316',0),
    ('aaa20000-1111-4000-8000-000000003317','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003317',0),
    ('aaa20000-1111-4000-8000-000000003318','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003318',0);
