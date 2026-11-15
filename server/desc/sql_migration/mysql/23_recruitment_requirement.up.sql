-- HRM-24 招聘需求单（HR207）纵切（迁移 23）
-- 单据底座（hrm_approval_record/hrm_bill_sequence）见迁移 11。
-- 发布衔接契约（spec 用户故事 43→44，HRM-25 招聘职位预留接缝）：
--   审批通过（status='approved'）的需求单可被招聘职位单引用；
--   HRM-25 的职位表用 requirement_uuid 关联本表，并对非 approved 需求拒绝引用。
-- 菜单 uuid 段：10xx（招聘管理 dir + 招聘需求 page/buttons，aaa10000/aaa20000 第 6 段）。

CREATE TABLE `hrm_recruitment_requirement` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '单据编号 HR207-YYYYMMDD+5位流水',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态 unsubmitted/approving/approved/rejected/revoked',
  `post_name` varchar(100) NOT NULL COMMENT '岗位名称',
  `post_category` varchar(32) NOT NULL DEFAULT 'tech' COMMENT '岗位类别 tech=技术类 product=产品类 design=设计类 operation=运营类 market=市场类 function=职能类 other=其他',
  `headcount` bigint NOT NULL DEFAULT 1 COMMENT '需求人数（>=1）',
  `urgency` varchar(16) NOT NULL DEFAULT 'normal' COMMENT '紧急度 normal=普通 urgent=紧急',
  `expect_date` varchar(10) NOT NULL DEFAULT '' COMMENT '期望到岗日期 YYYY-MM-DD',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '所属部门（组织 uuid，HRM-01 组织引用）',
  `org_name` varchar(100) NOT NULL DEFAULT '' COMMENT '所属部门名称（快照）',
  `requirement_desc` varchar(500) NOT NULL DEFAULT '' COMMENT '需求说明/任职要求',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '创建人（申请人）用户 uuid',
  `creator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '申请人姓名（快照）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_recruitment_requirement_status` (`status`),
  KEY `idx_hrm_recruitment_requirement_org` (`org_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- HRM 菜单目录: 招聘管理
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001001','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','1','招聘管理',0,'',8,'hrm_recruitment','/hrm/recruitment','layout.base','mdi:account-search','1','route.hrm_recruitment',0,'',0,0,'[]','[]',0,'');

-- 招聘需求 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001002','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001001','2','招聘需求',0,'',1,'hrm_recruitment_requirement-list','/hrm/recruitment/requirement-list','view.hrm_recruitment_requirement-list','mdi:clipboard-text-search-outline','1','route.hrm_recruitment_requirement-list',0,'',0,0,'[]','[{"code":"v1:hrm:requirement:list","desc":"招聘需求列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 招聘需求 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001011','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001002','3','新增招聘需求',0,'',0,'','','','','1','button.hrm.requirement.add',0,'',0,0,'[]','[{"code":"v1:hrm:requirement:add","desc":"新增招聘需求"}]',0,'v1:hrm:requirement:add'),
    ('aaa10000-1111-4000-8000-000000001012','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001002','3','编辑招聘需求',0,'',1,'','','','','1','button.hrm.requirement.edit',0,'',0,0,'[]','[{"code":"v1:hrm:requirement:edit","desc":"编辑招聘需求"}]',0,'v1:hrm:requirement:edit'),
    ('aaa10000-1111-4000-8000-000000001013','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001002','3','删除招聘需求',0,'',2,'','','','','1','button.hrm.requirement.delete',0,'',0,0,'[]','[{"code":"v1:hrm:requirement:delete","desc":"删除招聘需求"}]',0,'v1:hrm:requirement:delete'),
    ('aaa10000-1111-4000-8000-000000001021','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001002','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000001022','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001002','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000001023','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001002','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000001024','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001002','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001001','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001001',0),
    ('aaa20000-1111-4000-8000-000000001002','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001002',0),
    ('aaa20000-1111-4000-8000-000000001003','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001011',0),
    ('aaa20000-1111-4000-8000-000000001004','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001012',0),
    ('aaa20000-1111-4000-8000-000000001005','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001013',0),
    ('aaa20000-1111-4000-8000-000000001006','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001021',0),
    ('aaa20000-1111-4000-8000-000000001007','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001022',0),
    ('aaa20000-1111-4000-8000-000000001008','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001023',0),
    ('aaa20000-1111-4000-8000-000000001009','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001024',0);
