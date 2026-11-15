-- HRM-07 员工入职单（HR201）纵切（迁移 27）
-- 单据底座（hrm_approval_record/hrm_bill_sequence）见迁移 11。
-- 入职单快照三张从表（工作经历/教育经历/家庭成员，按 entry_uuid 快照）；
-- 审批通过（OnApproved 事务内）自动创建 hrm_employee 档案：
--   工号自动生成（复用 HRM-03 NextEmployeeNo 最大工号+1）、预计转正=入职日期+试用期月、
--   三从表落员工档案、回写 employee_uuid（幂等：重复审批不重复建档）。

CREATE TABLE `hrm_entry_bill` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '单据编号 HR201-YYYYMMDD+5位流水',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态 unsubmitted/approving/approved/rejected/revoked',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '姓名',
  `phone` varchar(50) NOT NULL DEFAULT '' COMMENT '手机号',
  `email` varchar(100) NOT NULL DEFAULT '' COMMENT '邮箱',
  `gender` varchar(4) NOT NULL DEFAULT '' COMMENT '性别 1=男 2=女',
  `birthday` varchar(10) NOT NULL DEFAULT '' COMMENT '出生日期 YYYY-MM-DD',
  `nation` varchar(50) NOT NULL DEFAULT '' COMMENT '民族',
  `political_status` varchar(50) NOT NULL DEFAULT '' COMMENT '政治面貌',
  `marital_status` varchar(4) NOT NULL DEFAULT '' COMMENT '婚姻状况 1=未婚 2=已婚 3=离异 4=丧偶',
  `culture_level` varchar(50) NOT NULL DEFAULT '' COMMENT '文化程度',
  `native_place` varchar(200) NOT NULL DEFAULT '' COMMENT '籍贯',
  `address` varchar(500) NOT NULL DEFAULT '' COMMENT '现居住地',
  `household_address` varchar(500) NOT NULL DEFAULT '' COMMENT '户口所在地',
  `emergency_contact` varchar(100) NOT NULL DEFAULT '' COMMENT '紧急联系人',
  `emergency_phone` varchar(50) NOT NULL DEFAULT '' COMMENT '紧急联系人电话',
  `bank_name` varchar(200) NOT NULL DEFAULT '' COMMENT '工资开户行',
  `bank_account` varchar(100) NOT NULL DEFAULT '' COMMENT '工资卡账户',
  `avatar` varchar(500) NOT NULL DEFAULT '' COMMENT '员工照片',
  `position` varchar(100) NOT NULL DEFAULT '' COMMENT '职位',
  `post` varchar(100) NOT NULL DEFAULT '' COMMENT '职务',
  `probation_months` int NOT NULL DEFAULT 0 COMMENT '试用期(月)，预计转正=入职日期+试用期月',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '所属部门（组织 uuid）',
  `company_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '所属公司（组织 uuid）',
  `hire_date` varchar(10) NOT NULL DEFAULT '' COMMENT '入职日期 YYYY-MM-DD',
  `expect_formal_date` varchar(10) NOT NULL DEFAULT '' COMMENT '预计转正日期 YYYY-MM-DD',
  `attachment_desc` varchar(500) NOT NULL DEFAULT '' COMMENT '附件说明',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `employee_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '审批通过后创建的员工档案 uuid（幂等回写标记）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '创建人（申请人）用户 uuid',
  `creator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '申请人姓名（快照）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_entry_bill_status` (`status`),
  KEY `idx_hrm_entry_bill_employee` (`employee_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_entry_bill_work` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `entry_uuid` varchar(36) NOT NULL COMMENT '入职单 uuid',
  `company` varchar(200) NOT NULL DEFAULT '' COMMENT '工作单位',
  `position` varchar(100) NOT NULL DEFAULT '' COMMENT '职位',
  `start_date` varchar(10) NOT NULL DEFAULT '' COMMENT '开始日期 YYYY-MM-DD',
  `end_date` varchar(10) NOT NULL DEFAULT '' COMMENT '结束日期 YYYY-MM-DD',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_entry_bill_work_entry_uuid` (`entry_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_entry_bill_education` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `entry_uuid` varchar(36) NOT NULL COMMENT '入职单 uuid',
  `school` varchar(200) NOT NULL DEFAULT '' COMMENT '学校',
  `major` varchar(200) NOT NULL DEFAULT '' COMMENT '专业',
  `education` varchar(50) NOT NULL DEFAULT '' COMMENT '学历',
  `start_date` varchar(10) NOT NULL DEFAULT '' COMMENT '开始日期 YYYY-MM-DD',
  `end_date` varchar(10) NOT NULL DEFAULT '' COMMENT '结束日期 YYYY-MM-DD',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_entry_bill_education_entry_uuid` (`entry_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_entry_bill_family` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `entry_uuid` varchar(36) NOT NULL COMMENT '入职单 uuid',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '姓名',
  `relation` varchar(50) NOT NULL DEFAULT '' COMMENT '关系',
  `phone` varchar(50) NOT NULL DEFAULT '' COMMENT '电话',
  `work_unit` varchar(200) NOT NULL DEFAULT '' COMMENT '工作单位',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_entry_bill_family_entry_uuid` (`entry_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 员工入职 page（挂在已有 人事管理 dir aaa...000601 下；transfer 占 order 1,
-- 3.2/3.4 转正/离职页预留 2/3, 本页取 4）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001402','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000601','2','员工入职',0,'',4,'hrm_employee-relation_entry-list','/hrm/employee-relation/entry-list','view.hrm_employee-relation_entry-list','mdi:account-plus-outline','1','route.hrm_employee-relation_entry-list',0,'',0,0,'[]','[{"code":"v1:hrm:entry:list","desc":"员工入职列表"},{"code":"v1:hrm:entry:detail","desc":"入职单详情"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 员工入职 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001411','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','新增入职单',0,'',0,'','','','','1','button.hrm.entry.add',0,'',0,0,'[]','[{"code":"v1:hrm:entry:add","desc":"新增入职单"}]',0,'v1:hrm:entry:add'),
    ('aaa10000-1111-4000-8000-000000001412','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','编辑入职单',0,'',1,'','','','','1','button.hrm.entry.edit',0,'',0,0,'[]','[{"code":"v1:hrm:entry:edit","desc":"编辑入职单"}]',0,'v1:hrm:entry:edit'),
    ('aaa10000-1111-4000-8000-000000001413','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','删除入职单',0,'',2,'','','','','1','button.hrm.entry.delete',0,'',0,0,'[]','[{"code":"v1:hrm:entry:delete","desc":"删除入职单"}]',0,'v1:hrm:entry:delete'),
    ('aaa10000-1111-4000-8000-000000001421','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000001422','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000001423','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000001424','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001402','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001401','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001402',0),
    ('aaa20000-1111-4000-8000-000000001402','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001411',0),
    ('aaa20000-1111-4000-8000-000000001403','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001412',0),
    ('aaa20000-1111-4000-8000-000000001404','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001413',0),
    ('aaa20000-1111-4000-8000-000000001405','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001421',0),
    ('aaa20000-1111-4000-8000-000000001406','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001422',0),
    ('aaa20000-1111-4000-8000-000000001407','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001423',0),
    ('aaa20000-1111-4000-8000-000000001408','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001424',0);
