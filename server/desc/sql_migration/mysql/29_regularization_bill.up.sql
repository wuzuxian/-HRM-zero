-- HRM-08 员工转正单（HR202）纵切（迁移 29）
-- 单据底座（hrm_approval_record/hrm_bill_sequence）见迁移 11。
-- 审批通过回写（spec 用户故事 20 / demo-inventory §3.2）：OnApproved 事务内更新
-- hrm_employee person_status='1'(正式) + formal_date(转正日期，优先单据实际转正日期)，
-- 幂等标记 applied=1；回写失败整体回滚（ADR-0001）。

CREATE TABLE `hrm_regularization_bill` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '单据编号 HR202-YYYYMMDD+5位流水',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态 unsubmitted/approving/approved/rejected/revoked',
  `employee_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '员工档案 uuid（HRM-03 档案引用）',
  `employee_no` varchar(50) NOT NULL DEFAULT '' COMMENT '员工工号（快照）',
  `employee_name` varchar(100) NOT NULL DEFAULT '' COMMENT '员工姓名（快照）',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '员工所属部门（组织 uuid，快照）',
  `company_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '员工所属公司（组织 uuid，快照）',
  `position` varchar(100) NOT NULL DEFAULT '' COMMENT '职位（快照）',
  `post` varchar(100) NOT NULL DEFAULT '' COMMENT '职务（快照）',
  `person_status` varchar(1) NOT NULL DEFAULT '' COMMENT '人员状态（快照，制单时档案值）',
  `hire_date` varchar(10) NOT NULL DEFAULT '' COMMENT '入职日期 YYYY-MM-DD（快照）',
  `expected_formal_date` varchar(10) NOT NULL DEFAULT '' COMMENT '预计转正日期 YYYY-MM-DD（快照）',
  `formal_date` varchar(10) NOT NULL DEFAULT '' COMMENT '实际转正日期 YYYY-MM-DD（审批通过后优先回写档案转正日期）',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `applied` bigint NOT NULL DEFAULT 0 COMMENT '审批回写是否已完成 1=是 0=否（幂等标记）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '创建人（申请人）用户 uuid',
  `creator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '申请人姓名（快照）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_regularization_bill_status` (`status`, `applied`),
  KEY `idx_hrm_regularization_bill_employee` (`employee_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 员工转正 page（人事管理 dir 下，uuid 段 16xx，迁移 16 人事调动 = 601/602，转正取 631/632）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000632','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000601','2','员工转正',0,'',2,'hrm_employee-relation_regular-list','/hrm/employee-relation/regular-list','view.hrm_employee-relation_regular-list','mdi:account-check','1','route.hrm_employee-relation_regular-list',0,'',0,0,'[]','[{"code":"v1:hrm:regularization:list","desc":"员工转正列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 员工转正 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000641','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000632','3','新增转正单',0,'',0,'','','','','1','button.hrm.regularization.add',0,'',0,0,'[]','[{"code":"v1:hrm:regularization:add","desc":"新增转正单"}]',0,'v1:hrm:regularization:add'),
    ('aaa10000-1111-4000-8000-000000000642','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000632','3','编辑转正单',0,'',1,'','','','','1','button.hrm.regularization.edit',0,'',0,0,'[]','[{"code":"v1:hrm:regularization:edit","desc":"编辑转正单"}]',0,'v1:hrm:regularization:edit'),
    ('aaa10000-1111-4000-8000-000000000643','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000632','3','删除转正单',0,'',2,'','','','','1','button.hrm.regularization.delete',0,'',0,0,'[]','[{"code":"v1:hrm:regularization:delete","desc":"删除转正单"}]',0,'v1:hrm:regularization:delete'),
    ('aaa10000-1111-4000-8000-000000000651','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000632','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000000652','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000632','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000000653','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000632','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000000654','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000632','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000631','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000632',0),
    ('aaa20000-1111-4000-8000-000000000641','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000641',0),
    ('aaa20000-1111-4000-8000-000000000642','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000642',0),
    ('aaa20000-1111-4000-8000-000000000643','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000643',0),
    ('aaa20000-1111-4000-8000-000000000651','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000651',0),
    ('aaa20000-1111-4000-8000-000000000652','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000652',0),
    ('aaa20000-1111-4000-8000-000000000653','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000653',0),
    ('aaa20000-1111-4000-8000-000000000654','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000654',0);
