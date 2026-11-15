-- HRM-10 员工离职单（HR204）纵切（迁移 32）
-- 单据底座（hrm_approval_record/hrm_bill_sequence）见迁移 11。
-- 审批通过仅记录（applied=0），定时任务在离职日当天统一将员工 person_status→'5'(离职)
-- （GLOSSARY/HRM-03 人员状态枚举），幂等标记 applied；复用内嵌调度骨架（迁移 16 起的
-- scheduler + ScanPendingEffective 模式，参照 transfer）。

CREATE TABLE `hrm_resignation_bill` (
    id bigint NOT NULL AUTO_INCREMENT,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    bill_no varchar(32) NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'unsubmitted',
    employee_uuid varchar(36) NOT NULL DEFAULT '',
    employee_no varchar(50) NOT NULL DEFAULT '',
    employee_name varchar(100) NOT NULL DEFAULT '',
    org_uuid varchar(36) NOT NULL DEFAULT '',
    company_uuid varchar(36) NOT NULL DEFAULT '',
    position varchar(100) NOT NULL DEFAULT '',
    post varchar(100) NOT NULL DEFAULT '',
    resign_type varchar(16) NOT NULL DEFAULT 'voluntary',
    resign_reason varchar(500) NOT NULL DEFAULT '',
    apply_date varchar(10) NOT NULL DEFAULT '',
    resign_date varchar(10) NOT NULL DEFAULT '',
    last_work_date varchar(10) NOT NULL DEFAULT '',
    handover_person varchar(100) NOT NULL DEFAULT '',
    salary_settle_date varchar(10) NOT NULL DEFAULT '',
    applied bigint NOT NULL DEFAULT 0,
    create_by varchar(36) NOT NULL DEFAULT '',
    creator_name varchar(100) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_resignation_bill_status ON `hrm_resignation_bill` (status, applied);
CREATE INDEX idx_hrm_resignation_bill_employee ON `hrm_resignation_bill` (employee_uuid);

-- 员工离职 page（人事管理 dir 下，uuid 段 17xx：迁移 16 调动 60x、迁移 29 转正 63x/64x/65x、迁移 27 入职 14xx，本票取 66x/67x/68x）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000661','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000601','2','员工离职',0,'',3,'hrm_employee-relation_resignation-list','/hrm/employee-relation/resignation-list','view.hrm_employee-relation_resignation-list','mdi:account-arrow-right','1','route.hrm_employee-relation_resignation-list',0,'',0,0,'[]','[{"code":"v1:hrm:resignation:list","desc":"员工离职列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 员工离职 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000000671','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000661','3','新增离职单',0,'',0,'','','','','1','button.hrm.resignation.add',0,'',0,0,'[]','[{"code":"v1:hrm:resignation:add","desc":"新增离职单"}]',0,'v1:hrm:resignation:add'),
    ('aaa10000-1111-4000-8000-000000000672','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000661','3','编辑离职单',0,'',1,'','','','','1','button.hrm.resignation.edit',0,'',0,0,'[]','[{"code":"v1:hrm:resignation:edit","desc":"编辑离职单"}]',0,'v1:hrm:resignation:edit'),
    ('aaa10000-1111-4000-8000-000000000673','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000661','3','删除离职单',0,'',2,'','','','','1','button.hrm.resignation.delete',0,'',0,0,'[]','[{"code":"v1:hrm:resignation:delete","desc":"删除离职单"}]',0,'v1:hrm:resignation:delete'),
    ('aaa10000-1111-4000-8000-000000000681','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000661','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000000682','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000661','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000000683','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000661','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000000684','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000661','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000000661','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000661',0),
    ('aaa20000-1111-4000-8000-000000000671','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000671',0),
    ('aaa20000-1111-4000-8000-000000000672','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000672',0),
    ('aaa20000-1111-4000-8000-000000000673','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000673',0),
    ('aaa20000-1111-4000-8000-000000000681','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000681',0),
    ('aaa20000-1111-4000-8000-000000000682','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000682',0),
    ('aaa20000-1111-4000-8000-000000000683','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000683',0),
    ('aaa20000-1111-4000-8000-000000000684','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000000684',0);
