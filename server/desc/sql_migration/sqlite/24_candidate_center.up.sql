-- HRM-26 候选人中心（Excel 导入）纵切（迁移 24）
-- 候选人为普通实体（非单据底座）：无 bill_no/审批状态，状态字段按
-- 新线索→已投递→面试中→已发Offer→录用通过 流转（不可跳级/不可回退）。
-- AI 入库通道（邮箱线索入库/文本识别补录）按 ADR-0002 裁剪，仅保留 Excel 导入 + 手工新增。
-- 录用衔接接缝（HRM-25 录用审批 HR208）：录用审批单通过 candidate_uuid 引用本表。
-- 菜单 uuid 段：11xx（候选人 page/buttons，挂在 10xx 招聘管理 dir 下）。

CREATE TABLE hrm_recruitment_candidate (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    source_channel TEXT NOT NULL DEFAULT 'other',
    education TEXT NOT NULL DEFAULT '',
    current_company TEXT NOT NULL DEFAULT '',
    current_position TEXT NOT NULL DEFAULT '',
    expect_salary TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'new',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_recruitment_candidate_status ON hrm_recruitment_candidate (status);
CREATE INDEX idx_hrm_recruitment_candidate_phone ON hrm_recruitment_candidate (phone);

-- 候选人中心 page（挂在 招聘管理 dir aaa10000-...1001 下，order 2）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001101','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001001','2','候选人中心',0,'',2,'hrm_recruitment_candidate-list','/hrm/recruitment/candidate-list','view.hrm_recruitment_candidate-list','mdi:account-multiple-outline','1','route.hrm_recruitment_candidate-list',0,'',0,0,'[]','[{"code":"v1:hrm:candidate:list","desc":"候选人列表"}]',0,'');

-- 候选人中心 buttons（页面操作）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001111','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001101','3','新增候选人',0,'',0,'','','','','1','button.hrm.candidate.add',0,'',0,0,'[]','[{"code":"v1:hrm:candidate:add","desc":"新增候选人"}]',0,'v1:hrm:candidate:add'),
    ('aaa10000-1111-4000-8000-000000001112','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001101','3','编辑候选人',0,'',1,'','','','','1','button.hrm.candidate.edit',0,'',0,0,'[]','[{"code":"v1:hrm:candidate:edit","desc":"编辑候选人"}]',0,'v1:hrm:candidate:edit'),
    ('aaa10000-1111-4000-8000-000000001113','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001101','3','删除候选人',0,'',2,'','','','','1','button.hrm.candidate.delete',0,'',0,0,'[]','[{"code":"v1:hrm:candidate:delete","desc":"删除候选人"}]',0,'v1:hrm:candidate:delete'),
    ('aaa10000-1111-4000-8000-000000001114','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001101','3','导入候选人',0,'',3,'','','','','1','button.hrm.candidate.import',0,'',0,0,'[]','[{"code":"v1:hrm:candidate:import","desc":"导入候选人"}]',0,'v1:hrm:candidate:import'),
    ('aaa10000-1111-4000-8000-000000001115','2026-11-15 00:00:00','2026-11-15 00:00:00','1','aaa10000-1111-4000-8000-000000001101','3','候选人状态流转',0,'',4,'','','','','1','button.hrm.candidate.updatestatus',0,'',0,0,'[]','[{"code":"v1:hrm:candidate:updateStatus","desc":"候选人状态流转"}]',0,'v1:hrm:candidate:updateStatus');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001101','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001101',0),
    ('aaa20000-1111-4000-8000-000000001102','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001111',0),
    ('aaa20000-1111-4000-8000-000000001103','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001112',0),
    ('aaa20000-1111-4000-8000-000000001104','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001113',0),
    ('aaa20000-1111-4000-8000-000000001105','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001114',0),
    ('aaa20000-1111-4000-8000-000000001106','2026-11-15 00:00:00','2026-11-15 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001115',0);
