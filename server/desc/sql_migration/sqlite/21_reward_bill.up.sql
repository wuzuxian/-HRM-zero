-- HRM-07 奖惩单（HR215）纵切（迁移 21）
-- 单据底座（hrm_approval_record/hrm_bill_sequence）见迁移 11。
-- 薪资数据源契约（spec 用户故事 53）：审批通过后 synced=1，HRM-34 薪资项目/HRM-37 试算
-- 按 (employee_uuid, salary_month, synced=1, status='approved') 取
-- reward_type='reward' → 奖惩奖励（应发），reward_type='punishment' → 奖惩扣款（应扣）。

CREATE TABLE hrm_reward_bill (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_no TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'unsubmitted',
    employee_uuid TEXT NOT NULL DEFAULT '',
    employee_no TEXT NOT NULL DEFAULT '',
    employee_name TEXT NOT NULL DEFAULT '',
    reward_type TEXT NOT NULL DEFAULT 'reward',
    reward_level TEXT NOT NULL DEFAULT 'company',
    doc_no TEXT NOT NULL DEFAULT '',
    occur_date TEXT NOT NULL DEFAULT '',
    amount NUMERIC NOT NULL DEFAULT 0,
    reason TEXT NOT NULL DEFAULT '',
    salary_month TEXT NOT NULL DEFAULT '',
    synced INTEGER NOT NULL DEFAULT 0,
    create_by TEXT NOT NULL DEFAULT '',
    creator_name TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_reward_bill_status ON hrm_reward_bill (status, synced);
CREATE INDEX idx_hrm_reward_bill_salary_month ON hrm_reward_bill (employee_uuid, salary_month);

-- HRM 菜单目录: 奖惩管理
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000901','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000001','1','奖惩管理',0,'',7,'hrm_reward','/hrm/reward','layout.base','mdi:stamper','1','route.hrm_reward',0,'',0,0,'[]','[]',0,'');

-- 奖惩单 page
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000902','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000901','2','奖惩单',0,'',1,'hrm_reward_bill','/hrm/reward/bill','view.hrm_reward_bill','mdi:file-document-edit-outline','1','route.hrm_reward_bill',0,'',0,0,'[]','[{"code":"v1:hrm:reward:list","desc":"奖惩单列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 奖惩单 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000911','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000902','3','新增奖惩单',0,'',0,'','','','','1','button.hrm.reward.add',0,'',0,0,'[]','[{"code":"v1:hrm:reward:add","desc":"新增奖惩单"}]',0,'v1:hrm:reward:add'),
    ('aaa10000-1111-4000-8000-000000000912','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000902','3','编辑奖惩单',0,'',1,'','','','','1','button.hrm.reward.edit',0,'',0,0,'[]','[{"code":"v1:hrm:reward:edit","desc":"编辑奖惩单"}]',0,'v1:hrm:reward:edit'),
    ('aaa10000-1111-4000-8000-000000000913','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000902','3','删除奖惩单',0,'',2,'','','','','1','button.hrm.reward.delete',0,'',0,0,'[]','[{"code":"v1:hrm:reward:delete","desc":"删除奖惩单"}]',0,'v1:hrm:reward:delete'),
    ('aaa10000-1111-4000-8000-000000000921','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000902','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000000922','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000902','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000000923','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000902','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000000924','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000000902','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000901','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000901',0),
    ('aaa20000-1111-4000-8000-000000000902','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000902',0),
    ('aaa20000-1111-4000-8000-000000000903','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000911',0),
    ('aaa20000-1111-4000-8000-000000000904','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000912',0),
    ('aaa20000-1111-4000-8000-000000000905','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000913',0),
    ('aaa20000-1111-4000-8000-000000000906','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000921',0),
    ('aaa20000-1111-4000-8000-000000000907','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000922',0),
    ('aaa20000-1111-4000-8000-000000000908','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000923',0),
    ('aaa20000-1111-4000-8000-000000000909','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000924',0);
