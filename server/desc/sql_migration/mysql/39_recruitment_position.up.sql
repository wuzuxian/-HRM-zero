-- HRM-25 招聘职位（HRM-25）纵切（迁移 39）
-- 职位为普通实体（非单据底座）：无 bill_no/审批状态，状态 招聘中(open)/已满(full)/已关闭(closed)。
-- 来源需求单接缝（HRM-24 契约）：requirement_uuid 引用 hrm_recruitment_requirement.uuid，
--   仅 status='approved' 的需求单可被引用（add/edit 侧拒绝非 approved，errcode 1058x）。
-- 已录用人数联动（HRM-27 录用审批契约）：IncreaseHiredCount(ctx, tx, positionUuid) 事务化递增，
--   达到 headcount 自动置 status='full'。
-- 菜单 uuid 段：25xx（招聘职位 page/buttons，挂在 10xx 招聘管理 dir aaa10000-...1001 下）。

CREATE TABLE `hrm_recruitment_position` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `post_name` varchar(100) NOT NULL COMMENT '岗位名称',
  `requirement_uuid` varchar(36) NOT NULL COMMENT '来源需求单（hrm_recruitment_requirement.uuid，仅 status=approved 可引用）',
  `post_category` varchar(32) NOT NULL DEFAULT 'tech' COMMENT '岗位类别 tech=技术类 product=产品类 design=设计类 operation=运营类 market=市场类 function=职能类 other=其他',
  `owner_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '招聘负责人 uuid',
  `owner_name` varchar(100) NOT NULL DEFAULT '' COMMENT '招聘负责人姓名快照',
  `headcount` int NOT NULL DEFAULT 1 COMMENT '招聘人数',
  `hired_count` int NOT NULL DEFAULT 0 COMMENT '已录用人数（录用审批 HRM-27 通过 IncreaseHiredCount 事务化递增）',
  `publish_date` varchar(10) NOT NULL DEFAULT '' COMMENT '发布日期 YYYY-MM-DD',
  `status` varchar(16) NOT NULL DEFAULT 'open' COMMENT '职位状态 open=招聘中 full=已满 closed=已关闭',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_recruitment_position_requirement` (`requirement_uuid`),
  KEY `idx_hrm_recruitment_position_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 招聘职位 page（挂在 招聘管理 dir aaa10000-...1001 下，order 3）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002501','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001001','2','招聘职位',0,'',3,'hrm_recruitment_position-list','/hrm/recruitment/position-list','view.hrm_recruitment_position-list','mdi:briefcase-outline','1','route.hrm_recruitment_position-list',0,'',0,0,'[]','[{"code":"v1:hrm:position:list","desc":"招聘职位列表"}]',0,'');

-- 招聘职位 buttons（页面操作）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002511','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002501','3','新增职位',0,'',0,'','','','','1','button.hrm.position.add',0,'',0,0,'[]','[{"code":"v1:hrm:position:add","desc":"新增职位"}]',0,'v1:hrm:position:add'),
    ('aaa10000-1111-4000-8000-000000002512','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002501','3','编辑职位',0,'',1,'','','','','1','button.hrm.position.edit',0,'',0,0,'[]','[{"code":"v1:hrm:position:edit","desc":"编辑职位"}]',0,'v1:hrm:position:edit'),
    ('aaa10000-1111-4000-8000-000000002513','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002501','3','删除职位',0,'',2,'','','','','1','button.hrm.position.delete',0,'',0,0,'[]','[{"code":"v1:hrm:position:delete","desc":"删除职位"}]',0,'v1:hrm:position:delete'),
    ('aaa10000-1111-4000-8000-000000002514','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002501','3','职位状态流转',0,'',3,'','','','','1','button.hrm.position.updatestatus',0,'',0,0,'[]','[{"code":"v1:hrm:position:updateStatus","desc":"职位状态流转"}]',0,'v1:hrm:position:updateStatus');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002501','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002501',0),
    ('aaa20000-1111-4000-8000-000000002502','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002511',0),
    ('aaa20000-1111-4000-8000-000000002503','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002512',0),
    ('aaa20000-1111-4000-8000-000000002504','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002513',0),
    ('aaa20000-1111-4000-8000-000000002505','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002514',0);
