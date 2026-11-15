-- HRM-28 培训计划单（HR216）+ 培训班次（管理端 CRUD）
-- 培训计划走共享单据审批底座（迁移 11: hrm_approval_record / hrm_bill_sequence），
-- 单据类型码 216；审批通过（OnApproved）后置 published=1（发布状态 草稿/已发布）。
-- 班次复用迁移 15 的 hrm_training_session / hrm_training_session_participant（不另建表），
-- plan_uuid 回填培训计划 uuid；班次编号 PX-年-流水 独立编号体系（GLOSSARY 单据编号）。

CREATE TABLE `hrm_training_plan` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '单据编号 HR216-YYYYMMDD+5位流水',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态 unsubmitted/approving/approved/rejected/revoked',
  `plan_name` varchar(100) NOT NULL COMMENT '计划名称',
  `theme` varchar(200) NOT NULL DEFAULT '' COMMENT '主题',
  `plan_year` varchar(4) NOT NULL DEFAULT '' COMMENT '年度',
  `plan_type` varchar(20) NOT NULL DEFAULT 'special' COMMENT '计划类型 special=专项 new_hire=新员工培养 safety=安全培训 skill=技能提升',
  `start_date` varchar(10) NOT NULL DEFAULT '' COMMENT '开始日期 YYYY-MM-DD',
  `end_date` varchar(10) NOT NULL DEFAULT '' COMMENT '结束日期 YYYY-MM-DD',
  `published` tinyint NOT NULL DEFAULT 0 COMMENT '发布状态 0=草稿 1=已发布（审批通过后置 1）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '创建人 uuid',
  `creator_name` varchar(50) NOT NULL DEFAULT '' COMMENT '创建人姓名快照',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_training_plan_status` (`status`, `published`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 演示数据：1 张已发布计划 + 1 张草稿计划；迁移 15 的两个班次回填 plan_uuid 指向已发布计划
INSERT INTO `hrm_training_plan` (uuid, create_time, update_time, bill_no, status, plan_name, theme, plan_year, plan_type, start_date, end_date, published, create_by, creator_name)
VALUES
    ('aaa30000-3333-4000-8000-000000000351','2026-01-01 09:00:00','2026-01-05 10:00:00','HR216-2026010100001','approved','2026年度新员工培养计划','新员工入职培养','2026','new_hire','2026-01-01','2026-12-31',1,'1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f','admin'),
    ('aaa30000-3333-4000-8000-000000000352','2026-11-01 09:00:00','2026-11-01 09:00:00','HR216-2026110100001','unsubmitted','2027年度安全生产专项培训','安全生产专项','2027','safety','2027-01-01','2027-06-30',0,'1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f','admin');

UPDATE `hrm_training_session` SET plan_uuid = 'aaa30000-3333-4000-8000-000000000351' WHERE uuid = 'aaa30000-3333-4000-8000-000000000301';
UPDATE `hrm_training_session` SET plan_uuid = 'aaa30000-3333-4000-8000-000000000351' WHERE uuid = 'aaa30000-3333-4000-8000-000000000302';

-- 培训计划 page（复用迁移 15 的 培训管理 目录 aaa10000-...-000000000301）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000303','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000301','2','培训计划',0,'',2,'hrm_training_plan','/hrm/training/plan','view.hrm_training_plan','mdi:clipboard-text-calendar','1','route.hrm_training_plan',0,'',0,0,'[]','[{"code":"v1:hrm:training_plan:list","desc":"培训计划列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 培训班次 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000304','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000301','2','培训班次',0,'',3,'hrm_training_session','/hrm/training/session','view.hrm_training_session','mdi:google-classroom','1','route.hrm_training_session',0,'',0,0,'[]','[{"code":"v1:hrm:training_session:list","desc":"培训班次列表"},{"code":"v1:hrm:training_session:getSessionParticipantList","desc":"班次学员列表"},{"code":"v1:hrm:training_session:setSessionParticipants","desc":"维护班次学员"}]',0,'');

-- 培训计划 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000321','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000303','3','新增计划',0,'',0,'','','','','1','button.hrm.training.plan.add',0,'',0,0,'[]','[{"code":"v1:hrm:training_plan:add","desc":"新增计划"}]',0,'v1:hrm:training_plan:add'),
    ('aaa10000-1111-4000-8000-000000000322','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000303','3','编辑计划',0,'',1,'','','','','1','button.hrm.training.plan.edit',0,'',0,0,'[]','[{"code":"v1:hrm:training_plan:edit","desc":"编辑计划"}]',0,'v1:hrm:training_plan:edit'),
    ('aaa10000-1111-4000-8000-000000000323','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000303','3','删除计划',0,'',2,'','','','','1','button.hrm.training.plan.delete',0,'',0,0,'[]','[{"code":"v1:hrm:training_plan:delete","desc":"删除计划"}]',0,'v1:hrm:training_plan:delete'),
    ('aaa10000-1111-4000-8000-000000000324','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000303','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000000325','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000303','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000000326','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000303','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000000327','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000303','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 培训班次 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000341','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000304','3','新增班次',0,'',0,'','','','','1','button.hrm.training.session.add',0,'',0,0,'[]','[{"code":"v1:hrm:training_session:add","desc":"新增班次"}]',0,'v1:hrm:training_session:add'),
    ('aaa10000-1111-4000-8000-000000000342','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000304','3','编辑班次',0,'',1,'','','','','1','button.hrm.training.session.edit',0,'',0,0,'[]','[{"code":"v1:hrm:training_session:edit","desc":"编辑班次"}]',0,'v1:hrm:training_session:edit'),
    ('aaa10000-1111-4000-8000-000000000343','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000304','3','删除班次',0,'',2,'','','','','1','button.hrm.training.session.delete',0,'',0,0,'[]','[{"code":"v1:hrm:training_session:delete","desc":"删除班次"}]',0,'v1:hrm:training_session:delete'),
    ('aaa10000-1111-4000-8000-000000000344','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000304','3','班次学员',0,'',3,'','','','','1','button.hrm.training.session.participants',0,'',0,0,'[]','[{"code":"v1:hrm:training_session:setSessionParticipants","desc":"维护班次学员"}]',0,'v1:hrm:training_session:setSessionParticipants');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000303','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000303',0),
    ('aaa20000-1111-4000-8000-000000000304','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000304',0),
    ('aaa20000-1111-4000-8000-000000000321','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000321',0),
    ('aaa20000-1111-4000-8000-000000000322','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000322',0),
    ('aaa20000-1111-4000-8000-000000000323','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000323',0),
    ('aaa20000-1111-4000-8000-000000000324','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000324',0),
    ('aaa20000-1111-4000-8000-000000000325','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000325',0),
    ('aaa20000-1111-4000-8000-000000000326','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000326',0),
    ('aaa20000-1111-4000-8000-000000000327','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000327',0),
    ('aaa20000-1111-4000-8000-000000000341','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000341',0),
    ('aaa20000-1111-4000-8000-000000000342','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000342',0),
    ('aaa20000-1111-4000-8000-000000000343','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000343',0),
    ('aaa20000-1111-4000-8000-000000000344','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000344',0);
