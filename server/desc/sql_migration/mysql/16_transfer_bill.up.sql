-- HRM-09 人事调动单（HR203）+ 服务内嵌调度骨架（迁移 16）
-- 单据底座（hrm_approval_record/hrm_bill_sequence）见迁移 11。

CREATE TABLE `hrm_transfer_bill` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '单据编号 HR203-YYYYMMDD+5位流水',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态 unsubmitted/approving/approved/rejected/revoked',
  `employee_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '员工档案 uuid（HRM-03 未合入前允许为空，仅快照工号/姓名）',
  `employee_no` varchar(50) NOT NULL DEFAULT '' COMMENT '员工工号（快照）',
  `employee_name` varchar(100) NOT NULL DEFAULT '' COMMENT '员工姓名（快照）',
  `change_type` varchar(16) NOT NULL DEFAULT 'other' COMMENT '异动类型 post=调岗 org=调部门 salary=调薪 company=调公司 other=其他',
  `original_company` varchar(36) NOT NULL DEFAULT '' COMMENT '原公司（组织 uuid，空表示无）',
  `original_org` varchar(36) NOT NULL DEFAULT '' COMMENT '原部门（组织 uuid，空表示无）',
  `original_position` varchar(100) NOT NULL DEFAULT '' COMMENT '原职位',
  `original_post` varchar(100) NOT NULL DEFAULT '' COMMENT '原岗位',
  `new_company` varchar(36) NOT NULL DEFAULT '' COMMENT '新公司（组织 uuid）',
  `new_org` varchar(36) NOT NULL DEFAULT '' COMMENT '新部门（组织 uuid）',
  `new_position` varchar(100) NOT NULL DEFAULT '' COMMENT '新职位',
  `new_post` varchar(100) NOT NULL DEFAULT '' COMMENT '新岗位',
  `new_salary` decimal(12,2) NOT NULL DEFAULT 0 COMMENT '调整后薪资（调薪时填写，0=不涉及）',
  `transfer_reason` varchar(500) NOT NULL DEFAULT '' COMMENT '调动原因',
  `effective_immediately` bigint NOT NULL DEFAULT 1 COMMENT '是否立即生效 1=是 0=否（0 时由定时任务在生效日处理）',
  `effective_date` varchar(10) NOT NULL DEFAULT '' COMMENT '生效日期 YYYY-MM-DD（立即生效时可空）',
  `applied` bigint NOT NULL DEFAULT 0 COMMENT '生效回写是否已完成 1=是 0=否（幂等标记，定时任务据此扫描）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '创建人（申请人）用户 uuid',
  `creator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '申请人姓名（快照）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_transfer_bill_status` (`status`, `applied`),
  KEY `idx_hrm_transfer_bill_employee` (`employee_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- HRM 菜单目录: 人事管理
INSERT INTO `manage_menu` (`uuid`, `create_time`, `update_time`, `status`, `parent_uuid`, `menu_type`, `menu_name`, `hide_in_menu`, `active_menu`, `order`, `route_name`, `route_path`, `component`, `icon`, `icon_type`, `i18n_key`, `keep_alive`, `href`, `multi_tab`, `fixed_index_in_tab`, `query`, `permissions`, `constant`, `button_code`)
VALUES
    ('aaa10000-1111-4000-8000-000000000401','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','1','人事管理',0,'',6,'hrm_employee-relation','/hrm/employee-relation','layout.base','mdi:account-switch','1','route.hrm_employee-relation',0,'',0,0,'[]','[]',0,'');

-- 人事调动 page
INSERT INTO `manage_menu` (`uuid`, `create_time`, `update_time`, `status`, `parent_uuid`, `menu_type`, `menu_name`, `hide_in_menu`, `active_menu`, `order`, `route_name`, `route_path`, `component`, `icon`, `icon_type`, `i18n_key`, `keep_alive`, `href`, `multi_tab`, `fixed_index_in_tab`, `query`, `permissions`, `constant`, `button_code`)
VALUES
    ('aaa10000-1111-4000-8000-000000000402','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000401','2','人事调动',0,'',1,'hrm_employee-relation_transfer-list','/hrm/employee-relation/transfer-list','view.hrm_employee-relation_transfer-list','mdi:swap-horizontal','1','route.hrm_employee-relation_transfer-list',0,'',0,0,'[]','[{"code":"v1:hrm:transfer:list","desc":"人事调动列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 人事调动 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO `manage_menu` (`uuid`, `create_time`, `update_time`, `status`, `parent_uuid`, `menu_type`, `menu_name`, `hide_in_menu`, `active_menu`, `order`, `route_name`, `route_path`, `component`, `icon`, `icon_type`, `i18n_key`, `keep_alive`, `href`, `multi_tab`, `fixed_index_in_tab`, `query`, `permissions`, `constant`, `button_code`)
VALUES
    ('aaa10000-1111-4000-8000-000000000411','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','新增调动',0,'',0,'','','','','1','button.hrm.transfer.add',0,'',0,0,'[]','[{"code":"v1:hrm:transfer:add","desc":"新增调动"}]',0,'v1:hrm:transfer:add'),
    ('aaa10000-1111-4000-8000-000000000412','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','编辑调动',0,'',1,'','','','','1','button.hrm.transfer.edit',0,'',0,0,'[]','[{"code":"v1:hrm:transfer:edit","desc":"编辑调动"}]',0,'v1:hrm:transfer:edit'),
    ('aaa10000-1111-4000-8000-000000000413','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','删除调动',0,'',2,'','','','','1','button.hrm.transfer.delete',0,'',0,0,'[]','[{"code":"v1:hrm:transfer:delete","desc":"删除调动"}]',0,'v1:hrm:transfer:delete'),
    ('aaa10000-1111-4000-8000-000000000421','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000000422','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000000423','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000000424','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000402','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO `manage_role_menu` (`uuid`, `create_time`, `update_time`, `role_uuid`, `menu_uuid`, `is_home`)
VALUES
    ('aaa20000-1111-4000-8000-000000000401','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000401',0),
    ('aaa20000-1111-4000-8000-000000000402','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000402',0),
    ('aaa20000-1111-4000-8000-000000000403','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000411',0),
    ('aaa20000-1111-4000-8000-000000000404','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000412',0),
    ('aaa20000-1111-4000-8000-000000000405','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000413',0),
    ('aaa20000-1111-4000-8000-000000000406','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000421',0),
    ('aaa20000-1111-4000-8000-000000000407','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000422',0),
    ('aaa20000-1111-4000-8000-000000000408','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000423',0),
    ('aaa20000-1111-4000-8000-000000000409','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000424',0);
