-- HRM-27 录用审批单（HR208）纵切（迁移 48）。
-- 单据底座（HRM-06 契约）：bill_no/status + OnApproved 审批事务内三联动：
--   1) position.IncreaseHiredCountTx 职位已录用+1（每单恰好一次，offer 表 applied 标记幂等）；
--   2) 候选人状态推进至 hired（HRM-26 契约）；
--   3) 创建草稿 HR201 入职单（候选人姓名/计划入职日期/试用期/部门职位按职位单回填，
--      HR 仍可补全入职详情后走入职单自身审批）。
-- 菜单 uuid 段：34xx（录用审批 page/buttons，挂在 10xx 招聘管理 dir aaa10000-...1001 下，
--   order 4：招聘需求=1 候选人=2 职位=3）。

CREATE TABLE `hrm_offer_bill` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '单据编号 HR208-YYYYMMDD+5位流水',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态 unsubmitted/approving/approved/rejected/revoked',
  `candidate_uuid` varchar(36) NOT NULL COMMENT '候选人（hrm_recruitment_candidate.uuid，仅 status=offer 可发起录用）',
  `candidate_name` varchar(100) NOT NULL DEFAULT '' COMMENT '候选人姓名（快照）',
  `position_uuid` varchar(36) NOT NULL COMMENT '招聘职位（hrm_recruitment_position.uuid，仅 status=open 可录用）',
  `post_name` varchar(100) NOT NULL DEFAULT '' COMMENT '岗位名称（快照）',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '所属部门（组织 uuid，职位单快照回填入职单）',
  `org_name` varchar(100) NOT NULL DEFAULT '' COMMENT '部门名称（快照）',
  `plan_entry_date` varchar(10) NOT NULL DEFAULT '' COMMENT '计划入职日期 YYYY-MM-DD',
  `probation_months` int NOT NULL DEFAULT 0 COMMENT '试用期(月)',
  `rank_level` varchar(32) NOT NULL DEFAULT 'junior' COMMENT '职级 junior=初级 intermediate=中级 senior=高级 expert=专家',
  `reporter_name` varchar(100) NOT NULL DEFAULT '' COMMENT '汇报人',
  `entry_bill_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '审批通过后创建的入职单 uuid（幂等回写标记）',
  `applied` bigint NOT NULL DEFAULT 0 COMMENT '三联动是否已执行 1=是 0=否（幂等标记）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '创建人（申请人）用户 uuid',
  `creator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '申请人姓名（快照）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_offer_bill_status` (`status`),
  KEY `idx_hrm_offer_bill_candidate` (`candidate_uuid`),
  KEY `idx_hrm_offer_bill_position` (`position_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 录用审批 page（挂在 招聘管理 dir aaa10000-...1001 下，order 4）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003401','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001001','2','录用审批',0,'',4,'hrm_recruitment_offer-list','/hrm/recruitment/offer-list','view.hrm_recruitment_offer-list','mdi:file-document-check-outline','1','route.hrm_recruitment_offer-list',0,'',0,0,'[]','[{"code":"v1:hrm:offer:list","desc":"录用审批列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 录用审批 buttons（页面操作 + 共享单据生命周期四码）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003411','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003401','3','新增录用审批',0,'',0,'','','','','1','button.hrm.offer.add',0,'',0,0,'[]','[{"code":"v1:hrm:offer:add","desc":"新增录用审批"}]',0,'v1:hrm:offer:add'),
    ('aaa10000-1111-4000-8000-000000003412','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003401','3','编辑录用审批',0,'',1,'','','','','1','button.hrm.offer.edit',0,'',0,0,'[]','[{"code":"v1:hrm:offer:edit","desc":"编辑录用审批"}]',0,'v1:hrm:offer:edit'),
    ('aaa10000-1111-4000-8000-000000003413','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003401','3','删除录用审批',0,'',2,'','','','','1','button.hrm.offer.delete',0,'',0,0,'[]','[{"code":"v1:hrm:offer:delete","desc":"删除录用审批"}]',0,'v1:hrm:offer:delete'),
    ('aaa10000-1111-4000-8000-000000003421','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003401','3','提交审批',0,'',0,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交审批"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000003422','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003401','3','审批通过',0,'',1,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000003423','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003401','3','审批不通过',0,'',2,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批不通过"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000003424','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003401','3','撤销',0,'',3,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003401','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003401',0),
    ('aaa20000-1111-4000-8000-000000003402','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003411',0),
    ('aaa20000-1111-4000-8000-000000003403','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003412',0),
    ('aaa20000-1111-4000-8000-000000003404','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003413',0),
    ('aaa20000-1111-4000-8000-000000003405','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003421',0),
    ('aaa20000-1111-4000-8000-000000003406','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003422',0),
    ('aaa20000-1111-4000-8000-000000003407','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003423',0),
    ('aaa20000-1111-4000-8000-000000003408','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003424',0);
