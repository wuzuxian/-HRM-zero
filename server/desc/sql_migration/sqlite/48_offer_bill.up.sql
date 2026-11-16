-- HRM-27 录用审批单（HR208）纵切（迁移 48）。
-- 单据底座（HRM-06 契约）：bill_no/status + OnApproved 审批事务内三联动：
--   1) position.IncreaseHiredCountTx 职位已录用+1（每单恰好一次，offer 表 applied 标记幂等）；
--   2) 候选人状态推进至 hired（HRM-26 契约）；
--   3) 创建草稿 HR201 入职单（候选人姓名/计划入职日期/试用期/部门职位按职位单回填，
--      HR 仍可补全入职详情后走入职单自身审批）。
-- 菜单 uuid 段：34xx（录用审批 page/buttons，挂在 10xx 招聘管理 dir aaa10000-...1001 下，
--   order 4：招聘需求=1 候选人=2 职位=3）。

CREATE TABLE hrm_offer_bill (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_no TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'unsubmitted',
    candidate_uuid TEXT NOT NULL,
    candidate_name TEXT NOT NULL DEFAULT '',
    position_uuid TEXT NOT NULL,
    post_name TEXT NOT NULL DEFAULT '',
    org_uuid TEXT NOT NULL DEFAULT '',
    org_name TEXT NOT NULL DEFAULT '',
    plan_entry_date TEXT NOT NULL DEFAULT '',
    probation_months INTEGER NOT NULL DEFAULT 0,
    rank_level TEXT NOT NULL DEFAULT 'junior',
    reporter_name TEXT NOT NULL DEFAULT '',
    entry_bill_uuid TEXT NOT NULL DEFAULT '',
    applied INTEGER NOT NULL DEFAULT 0,
    create_by TEXT NOT NULL DEFAULT '',
    creator_name TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_offer_bill_status ON hrm_offer_bill (status);
CREATE INDEX idx_hrm_offer_bill_candidate ON hrm_offer_bill (candidate_uuid);
CREATE INDEX idx_hrm_offer_bill_position ON hrm_offer_bill (position_uuid);

-- 录用审批 page（挂在 招聘管理 dir aaa10000-...1001 下，order 4）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003401','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001001','2','录用审批',0,'',4,'hrm_recruitment_offer-list','/hrm/recruitment/offer-list','view.hrm_recruitment_offer-list','mdi:file-document-check-outline','1','route.hrm_recruitment_offer-list',0,'',0,0,'[]','[{"code":"v1:hrm:offer:list","desc":"录用审批列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 录用审批 buttons（页面操作 + 共享单据生命周期四码）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
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
