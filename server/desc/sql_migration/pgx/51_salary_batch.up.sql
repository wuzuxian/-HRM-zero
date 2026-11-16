-- HRM-37 薪资批次与试算引擎（薪酬管理域）。
-- 菜单 uuid 段：37xx（开工 grep 三方言迁移：34xx offer/35xx month-report 预留不冲突已占报告、
-- 33xx 考核/32xx 签到/30xx+31xx 已占；37xx 空闲，声明占用）。
-- 错误码段：1068x（1067x=考勤月报、1066x=考核任务链、1065x 已占，grep 确认 1068x 空闲）。
-- 批次编号：SA{批次月份首日 yyyyMMdd}{4位流水}，复用共享 hrm_bill_sequence（bill_type_code='SA',
-- bill_date=批次月份首日，per (主体不区分) 月度流水）。

CREATE TABLE "hrm_salary_batch" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    batch_no varchar(20) NOT NULL,
    batch_name varchar(100) NOT NULL,
    batch_type varchar(16) NOT NULL DEFAULT 'monthly',
    subject_uuid varchar(36) NOT NULL,
    subject_name varchar(100) NOT NULL DEFAULT '',
    batch_year int NOT NULL,
    batch_month int NOT NULL,
    headcount int NOT NULL DEFAULT 0,
    earning_total decimal(18,2) NOT NULL DEFAULT 0,
    deduction_total decimal(18,2) NOT NULL DEFAULT 0,
    net_total decimal(18,2) NOT NULL DEFAULT 0,
    trial_status varchar(1) NOT NULL DEFAULT '0',
    confirm_status varchar(1) NOT NULL DEFAULT '0',
    pay_status varchar(1) NOT NULL DEFAULT '0',
    payslip_status varchar(1) NOT NULL DEFAULT '0',
    trial_time timestamp NULL,
    version int NOT NULL DEFAULT 0,
    warnings text NULL,
    remark varchar(500) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX uk_hrm_salary_batch_scope ON "hrm_salary_batch" (subject_uuid, batch_year, batch_month, batch_type);
CREATE INDEX idx_hrm_salary_batch_subject ON "hrm_salary_batch" (subject_uuid);
CREATE INDEX idx_hrm_salary_batch_no ON "hrm_salary_batch" (batch_no);

CREATE TABLE "hrm_salary_batch_item" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    batch_uuid varchar(36) NOT NULL,
    employee_uuid varchar(36) NOT NULL,
    employee_no varchar(20) NOT NULL DEFAULT '',
    employee_name varchar(50) NOT NULL DEFAULT '',
    org_name varchar(100) NOT NULL DEFAULT '',
    item_uuid varchar(36) NOT NULL DEFAULT '',
    item_code varchar(50) NOT NULL,
    item_name varchar(100) NOT NULL DEFAULT '',
    item_class varchar(20) NOT NULL DEFAULT '',
    item_type varchar(20) NOT NULL DEFAULT '',
    amount decimal(18,2) NOT NULL DEFAULT 0,
    source varchar(30) NOT NULL DEFAULT '',
    source_detail varchar(255) NOT NULL DEFAULT '',
    calc_order int NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);

CREATE INDEX idx_hrm_salary_batch_item_batch ON "hrm_salary_batch_item" (batch_uuid);
CREATE INDEX idx_hrm_salary_batch_item_emp ON "hrm_salary_batch_item" (batch_uuid, employee_uuid);
CREATE INDEX idx_hrm_salary_batch_item_code ON "hrm_salary_batch_item" (batch_uuid, item_code);

CREATE TRIGGER update_hrm_salary_batch_update_time BEFORE UPDATE ON "hrm_salary_batch"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

CREATE TRIGGER update_hrm_salary_batch_item_update_time BEFORE UPDATE ON "hrm_salary_batch_item"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

-- 薪资批次 page（挂在 13xx 薪酬管理目录下，order 7：主体=1 项目=2 方案=3 个税=3/4/5 档案=6）
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003701','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','薪资批次',0,'',7,'hrm_salary_batch','/hrm/salary/batch','view.hrm_salary_batch','mdi:cash-multiple','1','route.hrm_salary_batch',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:list","desc":"薪资批次列表"},{"code":"v1:hrm:salary_batch:getBatchResult","desc":"查看试算结果"}]',0,'');

-- 薪资批次 buttons
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003711','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003701','3','新增薪资批次',0,'',0,'','','','','1','button.hrm.salary_batch.add',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:add","desc":"新增薪资批次"}]',0,'v1:hrm:salary_batch:add'),
    ('aaa10000-1111-4000-8000-000000003712','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003701','3','编辑薪资批次',0,'',1,'','','','','1','button.hrm.salary_batch.edit',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:edit","desc":"编辑薪资批次"}]',0,'v1:hrm:salary_batch:edit'),
    ('aaa10000-1111-4000-8000-000000003713','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003701','3','删除薪资批次',0,'',2,'','','','','1','button.hrm.salary_batch.delete',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:delete","desc":"删除薪资批次"}]',0,'v1:hrm:salary_batch:delete'),
    ('aaa10000-1111-4000-8000-000000003714','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003701','3','薪资试算',0,'',3,'','','','','1','button.hrm.salary_batch.trial',0,'',0,0,'[]','[{"code":"v1:hrm:salary_batch:trial","desc":"薪资试算"}]',0,'v1:hrm:salary_batch:trial');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003701','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003701',0),
    ('aaa20000-1111-4000-8000-000000003711','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003711',0),
    ('aaa20000-1111-4000-8000-000000003712','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003712',0),
    ('aaa20000-1111-4000-8000-000000003713','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003713',0),
    ('aaa20000-1111-4000-8000-000000003714','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003714',0);
