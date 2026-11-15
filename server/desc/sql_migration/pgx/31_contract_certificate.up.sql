-- HRM-05 劳动合同与证照（迁移 31）
-- 两张表 join hrm_employee.uuid（HRM-03 档案引用，hrm-03-notes 契约 #7）：
--   hrm_employee_contract   劳动合同（生效日/到期日/续签次数/合同状态）
--   hrm_employee_certificate 员工证照（证书类型/证号/到期日；证号明文存储，接口出参统一脱敏）
-- 三个页面挂在 人事档案 dir（...0401，迁移 17）下，菜单 uuid 段 17xx（grep 全部迁移取空闲百位段）。
-- 预警口径：距到期天数 = end/expiry - today（纯函数）；状态列：normal / warning(≤90天) / expired / terminated(合同解除)。

CREATE TABLE "hrm_employee_contract" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid varchar(36) NOT NULL,
    employee_no varchar(20) NOT NULL DEFAULT '',
    employee_name varchar(50) NOT NULL DEFAULT '',
    org_uuid varchar(36) NOT NULL DEFAULT '',
    contract_no varchar(50) NOT NULL DEFAULT '',
    start_date varchar(10) NOT NULL DEFAULT '',
    end_date varchar(10) NOT NULL DEFAULT '',
    renew_count bigint NOT NULL DEFAULT 0,
    status varchar(1) NOT NULL DEFAULT '1',
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_employee_contract_employee ON "hrm_employee_contract" (employee_uuid);
CREATE INDEX idx_hrm_employee_contract_end_date ON "hrm_employee_contract" (end_date);

CREATE TRIGGER update_hrm_employee_contract_update_time BEFORE UPDATE ON "hrm_employee_contract"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

CREATE TABLE "hrm_employee_certificate" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_uuid varchar(36) NOT NULL,
    employee_no varchar(20) NOT NULL DEFAULT '',
    employee_name varchar(50) NOT NULL DEFAULT '',
    org_uuid varchar(36) NOT NULL DEFAULT '',
    cert_type varchar(50) NOT NULL DEFAULT '',
    cert_no varchar(100) NOT NULL DEFAULT '',
    issue_date varchar(10) NOT NULL DEFAULT '',
    expiry_date varchar(10) NOT NULL DEFAULT '',
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_employee_certificate_employee ON "hrm_employee_certificate" (employee_uuid);
CREATE INDEX idx_hrm_employee_certificate_expiry ON "hrm_employee_certificate" (expiry_date);

CREATE TRIGGER update_hrm_employee_certificate_update_time BEFORE UPDATE ON "hrm_employee_certificate"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

-- 劳动合同到期预警 page（人事档案 dir 下，uuid 段 17xx；route/路径见 demo-inventory §2.2）
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001701','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000401','2','劳动合同到期预警',0,'',2,'hrm_employee_employee-contract-expiring','/hrm/employee/employee-contract-expiring','view.hrm_employee_employee-contract-expiring','mdi:file-document-alert','1','route.hrm_employee_employee-contract-expiring',0,'',0,0,'[]','[{"code":"v1:hrm:contract:list","desc":"劳动合同到期预警列表"},{"code":"v1:hrm:contract:renew","desc":"合同续签"}]',0,'');

-- 待续签合同 page（同结构，筛选口径=即将到期/已到期；行操作含续签）
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001702','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000401','2','待续签合同',0,'',3,'hrm_employee_employee-contract-renewable','/hrm/employee/employee-contract-renewable','view.hrm_employee_employee-contract-renewable','mdi:file-document-edit','1','route.hrm_employee_employee-contract-renewable',0,'',0,0,'[]','[{"code":"v1:hrm:contract:list","desc":"待续签合同列表"},{"code":"v1:hrm:contract:renew","desc":"合同续签"},{"code":"v1:hrm:contract:add","desc":"新增劳动合同"},{"code":"v1:hrm:contract:edit","desc":"编辑劳动合同"},{"code":"v1:hrm:contract:delete","desc":"删除劳动合同"}]',0,'');

-- 证照到期预警 page（demo-inventory §2.4；证号脱敏展示）
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001703','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000401','2','证照到期预警',0,'',4,'hrm_employee_certificate-expiring','/hrm/employee/certificate-expiring','view.hrm_employee_certificate-expiring','mdi:card-account-details-star','1','route.hrm_employee_certificate-expiring',0,'',0,0,'[]','[{"code":"v1:hrm:empcert:list","desc":"证照到期预警列表"},{"code":"v1:hrm:empcert:add","desc":"新增证照"},{"code":"v1:hrm:empcert:edit","desc":"编辑证照"},{"code":"v1:hrm:empcert:delete","desc":"删除证照"}]',0,'');

-- 合同维护 buttons（挂 劳动合同到期预警 page；员工档案上下文的维护入口）
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001711','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001701','3','新增劳动合同',0,'',0,'','','','','1','button.hrm.contract.add',0,'',0,0,'[]','[{"code":"v1:hrm:contract:add","desc":"新增劳动合同"}]',0,'v1:hrm:contract:add'),
    ('aaa10000-1111-4000-8000-000000001712','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001701','3','编辑劳动合同',0,'',1,'','','','','1','button.hrm.contract.edit',0,'',0,0,'[]','[{"code":"v1:hrm:contract:edit","desc":"编辑劳动合同"}]',0,'v1:hrm:contract:edit'),
    ('aaa10000-1111-4000-8000-000000001713','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001701','3','删除劳动合同',0,'',2,'','','','','1','button.hrm.contract.delete',0,'',0,0,'[]','[{"code":"v1:hrm:contract:delete","desc":"删除劳动合同"}]',0,'v1:hrm:contract:delete'),
    ('aaa10000-1111-4000-8000-000000001714','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001701','3','合同续签',0,'',3,'','','','','1','button.hrm.contract.renew',0,'',0,0,'[]','[{"code":"v1:hrm:contract:renew","desc":"合同续签"}]',0,'v1:hrm:contract:renew'),
    ('aaa10000-1111-4000-8000-000000001715','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001703','3','新增证照',0,'',4,'','','','','1','button.hrm.empcert.add',0,'',0,0,'[]','[{"code":"v1:hrm:empcert:add","desc":"新增证照"}]',0,'v1:hrm:empcert:add'),
    ('aaa10000-1111-4000-8000-000000001716','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001703','3','编辑证照',0,'',5,'','','','','1','button.hrm.empcert.edit',0,'',0,0,'[]','[{"code":"v1:hrm:empcert:edit","desc":"编辑证照"}]',0,'v1:hrm:empcert:edit'),
    ('aaa10000-1111-4000-8000-000000001717','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001703','3','删除证照',0,'',6,'','','','','1','button.hrm.empcert.delete',0,'',0,0,'[]','[{"code":"v1:hrm:empcert:delete","desc":"删除证照"}]',0,'v1:hrm:empcert:delete');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001701','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001701',0),
    ('aaa20000-1111-4000-8000-000000001702','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001702',0),
    ('aaa20000-1111-4000-8000-000000001703','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001703',0),
    ('aaa20000-1111-4000-8000-000000001711','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001711',0),
    ('aaa20000-1111-4000-8000-000000001712','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001712',0),
    ('aaa20000-1111-4000-8000-000000001713','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001713',0),
    ('aaa20000-1111-4000-8000-000000001714','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001714',0),
    ('aaa20000-1111-4000-8000-000000001715','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001715',0),
    ('aaa20000-1111-4000-8000-000000001716','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001716',0),
    ('aaa20000-1111-4000-8000-000000001717','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001717',0);
