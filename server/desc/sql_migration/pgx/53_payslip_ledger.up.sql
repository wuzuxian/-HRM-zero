-- HRM-39 工资条与薪资统计台账（演示站 8.4 发薪流程，GLOSSARY「工资条」「薪资统计台账」）。
--   发布：按批次从试算明细行（hrm_salary_batch_item）聚合生成每员工工资条
--   hrm_salary_payslip（UNIQUE(batch_uuid,employee_uuid)，重发布覆盖旧明细），明细按
--   薪资项目 JSON 快照 items_json。发布守卫：批次须已确认（confirm_status='1'，
--   HRM-38 薪资确认审批写回，本票只读消费）。已读：员工本人查看（mine 端点）记
--   read_status=1 + read_time（幂等不覆盖首次）。发薪登记（pay_status）由 HRM-38
--   发薪登记写回，台账/统计卡只读透传。
-- 错误码 10695-10698（server/internal/errcodes/hrm_salary_payslip.go，
--   1069x 段 grep：10690-10694 已占 HRM-45 绩效结果，10695 起空闲）。
-- 菜单 uuid 段 39xx（grep 全部三方言迁移确认 00xx-38xx 已用、39xx 空闲）：
--   3901 工资条管理 page / 3902 薪资统计 page（薪酬管理 dir migration 25 uuid 1301），
--   3911 发布工资条按钮 / 3921 无额外按钮（统计页只读）；role_menu aaa20000 39xx。

DROP TABLE IF EXISTS "hrm_salary_payslip";

CREATE TABLE "hrm_salary_payslip" (
    id serial NOT NULL,
    uuid varchar(36) NOT NULL UNIQUE,
    create_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    batch_uuid varchar(36) NOT NULL,
    batch_no varchar(20) NOT NULL DEFAULT '',
    batch_name varchar(100) NOT NULL DEFAULT '',
    batch_year integer NOT NULL,
    batch_month integer NOT NULL,
    employee_uuid varchar(36) NOT NULL,
    employee_no varchar(20) NOT NULL DEFAULT '',
    employee_name varchar(50) NOT NULL DEFAULT '',
    org_name varchar(100) NOT NULL DEFAULT '',
    title varchar(200) NOT NULL DEFAULT '',
    earning_total numeric(18,2) NOT NULL DEFAULT 0,
    deduction_total numeric(18,2) NOT NULL DEFAULT 0,
    net_total numeric(18,2) NOT NULL DEFAULT 0,
    labor_cost_total numeric(18,2) NOT NULL DEFAULT 0,
    items_json text NOT NULL DEFAULT '[]',
    published integer NOT NULL DEFAULT 0,
    read_status integer NOT NULL DEFAULT 0,
    read_time timestamp NULL,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX uk_hrm_salary_payslip_batch_emp ON "hrm_salary_payslip" (batch_uuid, employee_uuid);
CREATE INDEX idx_hrm_salary_payslip_employee ON "hrm_salary_payslip" (employee_uuid);
CREATE INDEX idx_hrm_salary_payslip_batch ON "hrm_salary_payslip" (batch_uuid);

CREATE TRIGGER update_hrm_salary_payslip_update_time BEFORE UPDATE ON "hrm_salary_payslip"
FOR EACH ROW EXECUTE FUNCTION update_update_time_column();

-- 工资条管理 page（薪酬管理 dir，order 8：批次=7；统计=9）
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003901','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','工资条管理',0,'',8,'hrm_salary_payslip','/hrm/salary/payslip','view.hrm_salary_payslip','mdi:receipt-text','1','route.hrm_salary_payslip',0,'',0,0,'[]','[{"code":"v1:hrm:salary_payslip:list","desc":"工资条列表"}]',0,''),
    ('aaa10000-1111-4000-8000-000000003902','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','薪资统计',0,'',9,'hrm_salary_statistics','/hrm/salary/statistics','view.hrm_salary_statistics','mdi:chart-box','1','route.hrm_salary_statistics',0,'',0,0,'[]','[{"code":"v1:hrm:salary_payslip:ledger","desc":"薪资统计台账"}]',0,'');

-- 工资条管理 buttons
INSERT INTO "manage_menu" (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003911','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003901','3','发布工资条',0,'',0,'','','','','1','button.hrm.salary_payslip.publish',0,'',0,0,'[]','[{"code":"v1:hrm:salary_payslip:publish","desc":"按批次发布工资条"}]',0,'v1:hrm:salary_payslip:publish');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO "manage_role_menu" (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003901','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003901',0),
    ('aaa20000-1111-4000-8000-000000003902','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003902',0),
    ('aaa20000-1111-4000-8000-000000003911','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003911',0);
