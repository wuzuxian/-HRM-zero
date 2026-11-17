-- HRM-37 薪资批次与试算引擎（薪酬管理域）。
-- 菜单 uuid 段：37xx（开工 grep 三方言迁移：34xx offer/35xx month-report 预留不冲突已占报告、
-- 33xx 考核/32xx 签到/30xx+31xx 已占；37xx 空闲，声明占用）。
-- 错误码段：1068x（1067x=考勤月报、1066x=考核任务链、1065x 已占，grep 确认 1068x 空闲）。
-- 批次编号：SA{批次月份首日 yyyyMMdd}{4位流水}，复用共享 hrm_bill_sequence（bill_type_code='SA',
-- bill_date=批次月份首日，per (主体不区分) 月度流水）。

CREATE TABLE `hrm_salary_batch` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `batch_no` varchar(20) NOT NULL COMMENT '批次编号 SA{批次月份首日yyyyMMdd}{4位流水}（复用 hrm_bill_sequence bill_type_code=SA）',
  `batch_name` varchar(100) NOT NULL COMMENT '批次名称',
  `batch_type` varchar(16) NOT NULL DEFAULT 'monthly' COMMENT '批次类型 monthly=月度批次（demo 口径仅月度）',
  `subject_uuid` varchar(36) NOT NULL COMMENT '薪酬主体（hrm_salary_subject.uuid）',
  `subject_name` varchar(100) NOT NULL DEFAULT '' COMMENT '薪酬主体名称快照',
  `batch_year` int NOT NULL COMMENT '薪资年份（如 2026）',
  `batch_month` int NOT NULL COMMENT '薪资月份 1-12',
  `headcount` int NOT NULL DEFAULT 0 COMMENT '试算人数（有有效薪资档案且命中批次的在职员工数）',
  `earning_total` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '应发合计（全员 Σ 应发类明细）',
  `deduction_total` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '应扣合计（全员 Σ 应扣类明细）',
  `net_total` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '实发合计（全员 Σ NET_SALARY）',
  `trial_status` varchar(1) NOT NULL DEFAULT '0' COMMENT '试算进度 0=未试算 1=已试算',
  `confirm_status` varchar(1) NOT NULL DEFAULT '0' COMMENT '确认进度 0=未确认 1=已确认（HRM-38 薪资确认审批写回，本票恒 0）',
  `pay_status` varchar(1) NOT NULL DEFAULT '0' COMMENT '发薪进度 0=未发薪 1=已发薪（HRM-39 发薪登记写回，本票恒 0）',
  `payslip_status` varchar(1) NOT NULL DEFAULT '0' COMMENT '工资条进度 0=未生成 1=已生成（HRM-39 写回，本票恒 0）',
  `trial_time` datetime NULL COMMENT '最近一次试算时间（未试算为 NULL）',
  `version` int NOT NULL DEFAULT 0 COMMENT '试算版本号（每次重算 +1，幂等覆盖口径）',
  `warnings` text NULL COMMENT '最近一次试算告警文本（换行分隔；月报未锁定/绩效缺失/社保未配置等）',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_salary_batch_scope` (`subject_uuid`, `batch_year`, `batch_month`, `batch_type`),
  KEY `idx_hrm_salary_batch_subject` (`subject_uuid`),
  KEY `idx_hrm_salary_batch_no` (`batch_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_salary_batch_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `batch_uuid` varchar(36) NOT NULL COMMENT '薪资批次（hrm_salary_batch.uuid）',
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工（hrm_employee.uuid）',
  `employee_no` varchar(20) NOT NULL DEFAULT '' COMMENT '员工ID（工号快照）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名快照',
  `org_name` varchar(100) NOT NULL DEFAULT '' COMMENT '部门名称快照',
  `item_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '薪资项目（hrm_salary_item.uuid；汇总行次要快照）',
  `item_code` varchar(50) NOT NULL COMMENT '项目编码快照（消费侧 join 键，取数口径见 hrm-33 契约）',
  `item_name` varchar(100) NOT NULL DEFAULT '' COMMENT '项目名称快照',
  `item_class` varchar(20) NOT NULL DEFAULT '' COMMENT '项目分类快照 earning=应发 deduction=应扣 cost=成本',
  `item_type` varchar(20) NOT NULL DEFAULT '' COMMENT '项目类型快照 fixed/data/formula',
  `amount` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '试算金额（应扣类为正数）',
  `source` varchar(30) NOT NULL DEFAULT '' COMMENT '取数来源标记 archive/scheme_override/reward/attendance/perf/tax/formula/manual/zero…',
  `source_detail` varchar(255) NOT NULL DEFAULT '' COMMENT '取数说明（如 奖惩单 2 笔合计 / 系数1.2×基数1000 / 缺勤1天×日薪476.19）',
  `calc_order` int NOT NULL DEFAULT 0 COMMENT '计算顺序快照（项目 calc_order）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_salary_batch_item_batch` (`batch_uuid`),
  KEY `idx_hrm_salary_batch_item_emp` (`batch_uuid`, `employee_uuid`),
  KEY `idx_hrm_salary_batch_item_code` (`batch_uuid`, `item_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 薪资批次 page（挂在 13xx 薪酬管理目录下，order 7：主体=1 项目=2 方案=3 个税=3/4/5 档案=6）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003701','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','薪资批次',0,'',7,'hrm_salary_batch','/hrm/salary/batch','view.hrm_salary_batch','mdi:cash-multiple','1','route.hrm_salary_batch',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:list","desc":"薪资批次列表"},{"code":"v1:hrm:salary_batch:getBatchResult","desc":"查看试算结果"}]',0,'');

-- 薪资批次 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003711','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003701','3','新增薪资批次',0,'',0,'','','','','1','button.hrm.salary_batch.add',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:add","desc":"新增薪资批次"}]',0,'v1:hrm:salary_batch:add'),
    ('aaa10000-1111-4000-8000-000000003712','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003701','3','编辑薪资批次',0,'',1,'','','','','1','button.hrm.salary_batch.edit',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:edit","desc":"编辑薪资批次"}]',0,'v1:hrm:salary_batch:edit'),
    ('aaa10000-1111-4000-8000-000000003713','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003701','3','删除薪资批次',0,'',2,'','','','','1','button.hrm.salary_batch.delete',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:delete","desc":"删除薪资批次"}]',0,'v1:hrm:salary_batch:delete'),
    ('aaa10000-1111-4000-8000-000000003714','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003701','3','薪资试算',0,'',3,'','','','','1','button.hrm.salary_batch.trial',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:trial","desc":"薪资试算"}]',0,'v1:hrm:salary_batch:trial');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003701','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003701',0),
    ('aaa20000-1111-4000-8000-000000003711','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003711',0),
    ('aaa20000-1111-4000-8000-000000003712','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003712',0),
    ('aaa20000-1111-4000-8000-000000003713','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003713',0),
    ('aaa20000-1111-4000-8000-000000003714','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003714',0);
