-- HRM-38 薪资确认审批与发薪登记（演示站 8.4 发薪流程，GLOSSARY「薪资批次」四进度标记）：
--   HR209 薪资确认审批单（单号 HR209-{YYYYMMDD}{5位流水}，复用 hrm_bill_sequence，
--   bill_type_code='209'，走 HRM-06 单据公共组件单步审批）。批次选择器口径=已试算
--   （trial_status='1'）且未确认（confirm_status='0'）的批次；提交时快照 批次编号/名称/
--   主体/年月/人数/实发合计。审批通过事务内（OnApproved）：批次 confirm_status='1' +
--   confirm 时间（幂等 applied：UNIQUE(batch_uuid) 单据行 + confirm_status 幂等写回）。
--   HR210 发薪登记（对 confirm_status='1' 且 pay_status='0' 的批次）：发薪金额
--   （默认=实发合计可改）、发薪方式（银行代发/现金/其他）、回执编号、到账时间 →
--   事务内批次 pay_status='1'。一批次一条发薪登记（UNIQUE(batch_uuid)）。
--   confirm_status 列由本票写回（HRM-37 预建）；payslip_status 归 HRM-39。
-- 错误码 10710-10714（server/internal/errcodes/hrm_salary_confirm.go，1071x 段 grep 空闲：
--   1069x=HRM-45 绩效结果审批、10700=工作台）。
-- 菜单 uuid 段 41xx（grep 全部 migration 确认 00xx-38xx 已用、39xx=40xx 空闲，取 41xx）：
--   4101 薪资确认审批 page / 4102 发薪登记 page（13xx 薪酬管理 dir，migration 25，order 8/9），
--   4111-4114 确认审批按钮（v1:hrm:bill:* 共享审批码），4121 发薪登记按钮；role_menu aaa20000 41xx。

DROP TABLE IF EXISTS `hrm_salary_payment`;
DROP TABLE IF EXISTS `hrm_salary_confirm_bill`;

CREATE TABLE `hrm_salary_confirm_bill` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '审批单号 HR209-{YYYYMMDD}{5位流水}（复用 hrm_bill_sequence，bill_type_code=209）',
  `batch_uuid` varchar(36) NOT NULL COMMENT '薪资批次 uuid（hrm_salary_batch；一批次一张确认审批单）',
  `batch_no` varchar(20) NOT NULL DEFAULT '' COMMENT '批次编号快照 SA...',
  `batch_name` varchar(100) NOT NULL DEFAULT '' COMMENT '批次名称快照',
  `subject_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '薪酬主体 uuid 快照',
  `subject_name` varchar(100) NOT NULL DEFAULT '' COMMENT '薪酬主体名称快照',
  `batch_year` int NOT NULL DEFAULT 0 COMMENT '薪资年份快照',
  `batch_month` int NOT NULL DEFAULT 0 COMMENT '薪资月份快照 1-12',
  `headcount` int NOT NULL DEFAULT 0 COMMENT '人数快照（试算人数）',
  `net_total` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '实发合计快照（提交时批次 net_total）',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态：unsubmitted/approving/approved/rejected/revoked（单据状态机）',
  `submit_by` varchar(64) NOT NULL DEFAULT '' COMMENT '提交人',
  `submit_time` varchar(20) NOT NULL DEFAULT '' COMMENT '提交时间',
  `approve_by` varchar(64) NOT NULL DEFAULT '' COMMENT '审批通过人',
  `approve_time` varchar(20) NOT NULL DEFAULT '' COMMENT '审批通过时间',
  `opinion` varchar(255) NOT NULL DEFAULT '' COMMENT '审批意见（最近一次）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '创建人（提交人）用户 uuid',
  `creator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '创建人姓名（快照）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hrm_salary_confirm_bill_batch` (`batch_uuid`),
  KEY `idx_hrm_salary_confirm_bill_batch` (`batch_uuid`),
  KEY `idx_hrm_salary_confirm_bill_no` (`bill_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_salary_payment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `batch_uuid` varchar(36) NOT NULL COMMENT '薪资批次 uuid（hrm_salary_batch；一批次一条发薪登记）',
  `batch_no` varchar(20) NOT NULL DEFAULT '' COMMENT '批次编号快照 SA...',
  `batch_name` varchar(100) NOT NULL DEFAULT '' COMMENT '批次名称快照',
  `subject_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '薪酬主体 uuid 快照',
  `subject_name` varchar(100) NOT NULL DEFAULT '' COMMENT '薪酬主体名称快照',
  `pay_amount` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '发薪金额（默认=批次实发合计，可改）',
  `pay_mode` varchar(16) NOT NULL DEFAULT 'bank' COMMENT '发薪方式：bank=银行代发 cash=现金 other=其他',
  `receipt_no` varchar(64) NOT NULL DEFAULT '' COMMENT '回执编号',
  `arrive_time` varchar(20) NOT NULL DEFAULT '' COMMENT '到账时间',
  `register_by` varchar(36) NOT NULL DEFAULT '' COMMENT '登记人用户 uuid',
  `register_name` varchar(100) NOT NULL DEFAULT '' COMMENT '登记人姓名（快照）',
  `register_time` varchar(20) NOT NULL DEFAULT '' COMMENT '登记时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_hrm_salary_payment_batch` (`batch_uuid`),
  KEY `idx_hrm_salary_payment_batch` (`batch_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 薪酬管理目录（13xx）下两个新页（demo-inventory 8.4 路径：
-- /hrm/salary/confirm-list order 8、/hrm/salary/payment-list order 9）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000004101','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','薪资确认审批',0,'',8,'hrm_salary_confirm-list','/hrm/salary/confirm-list','view.hrm_salary_confirm-list','mdi:clipboard-check','1','route.hrm_salary_confirm-list',0,'',0,0,'[]','[{"code":"v1:hrm:salary_confirm:list","desc":"薪资确认审批列表"},{"code":"v1:hrm:salary_confirm:submit","desc":"提交薪资确认审批"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,''),
    ('aaa10000-1111-4000-8000-000000004102','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','发薪登记',0,'',9,'hrm_salary_payment-list','/hrm/salary/payment-list','view.hrm_salary_payment-list','mdi:cash-check','1','route.hrm_salary_payment-list',0,'',0,0,'[]','[{"code":"v1:hrm:salary_payment:list","desc":"发薪登记列表"},{"code":"v1:hrm:salary_payment:register","desc":"发薪登记"}]',0,'');

-- 确认审批 buttons（共享审批底座 v1:hrm:bill:* + 本域提交码）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000004111','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000004101','3','提交单据',0,'',0,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000004112','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000004101','3','审批通过',0,'',1,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000004113','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000004101','3','审批驳回',0,'',2,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000004114','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000004101','3','撤销单据',0,'',3,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke'),
    ('aaa10000-1111-4000-8000-000000004121','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000004102','3','发薪登记',0,'',0,'','','','','1','button.hrm.salary_payment.register',0,'',0,0,'[]','[{"code":"v1:hrm:salary_payment:register","desc":"发薪登记"}]',0,'v1:hrm:salary_payment:register');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000004101','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000004101',0),
    ('aaa20000-1111-4000-8000-000000004102','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000004102',0),
    ('aaa20000-1111-4000-8000-000000004111','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000004111',0),
    ('aaa20000-1111-4000-8000-000000004112','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000004112',0),
    ('aaa20000-1111-4000-8000-000000004113','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000004113',0),
    ('aaa20000-1111-4000-8000-000000004114','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000004114',0),
    ('aaa20000-1111-4000-8000-000000004121','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000004121',0);
