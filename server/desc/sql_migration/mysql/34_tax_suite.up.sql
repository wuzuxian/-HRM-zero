-- HRM-36 个税管理三件套（迁移 34）
-- 三张表 join hrm_employee.uuid（HRM-03 档案引用约定）：
--   hrm_tax_profile           员工个税档案（员工唯一；居民类型/税城市/默认专项附加/启停）
--   hrm_tax_special_deduction 专项附加扣除（员工+年度+月份+扣除项 唯一；7 类扣除项逐月金额）
--   hrm_tax_opening           个税累计期初（员工+主体+年度 唯一；累计预扣法期初数，HRM-37 试算消费）
-- 页面挂 薪酬管理 dir（...1301，迁移 25）下 order 3/4/5，菜单 uuid 段 19xx（18xx 已被 HRM-18 假期余额迁移 33 占用——合并 feat/hrm 时撞车发现，二次声明占用 19xx，grep 已确认 19xx 空闲）
-- （grep 全部迁移：10-17 段已占用；18xx 已被 HRM-18（迁移 33）占用，本票声明占用 19xx）。

CREATE TABLE `hrm_tax_profile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工（hrm_employee.uuid）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名快照',
  `id_card` varchar(18) NOT NULL DEFAULT '' COMMENT '身份证号快照',
  `resident_type` varchar(20) NOT NULL DEFAULT 'resident' COMMENT '居民类型 resident=居民个人 non_resident=非居民个人',
  `tax_city` varchar(50) NOT NULL DEFAULT '' COMMENT '税城市',
  `default_special_extra` varchar(200) NOT NULL DEFAULT '' COMMENT '默认专项附加（逗号分隔扣除项编码）',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '启停 1=启用 0=停用',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_tax_profile_employee` (`employee_uuid`),
  KEY `idx_hrm_tax_profile_name` (`employee_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_tax_special_deduction` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工（hrm_employee.uuid）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名快照',
  `deduct_year` varchar(4) NOT NULL COMMENT '扣除年度（如 2026）',
  `deduct_month` bigint NOT NULL COMMENT '扣除月份 1-12',
  `deduct_type` varchar(20) NOT NULL COMMENT '扣除项 CHILDREN_EDU/CONTINUING_EDU/CRITICAL_ILLNESS/HOUSE_LOAN/HOUSE_RENT/SUPPORT_ELDERLY/INFANT_CARE',
  `amount` decimal(12,2) NOT NULL DEFAULT 0 COMMENT '扣除金额（元/月）',
  `source` varchar(20) NOT NULL DEFAULT 'manual' COMMENT '来源 manual=手工维护 import=导入',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_tax_special_emp_year_month_type` (`employee_uuid`, `deduct_year`, `deduct_month`, `deduct_type`),
  KEY `idx_hrm_tax_special_emp_year` (`employee_uuid`, `deduct_year`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_tax_opening` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工（hrm_employee.uuid）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名快照',
  `subject_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '薪酬主体（hrm_salary_subject.uuid，空=不限主体）',
  `subject_name` varchar(100) NOT NULL DEFAULT '' COMMENT '主体名称快照',
  `tax_year` varchar(4) NOT NULL COMMENT '年度（如 2026）',
  `until_month` bigint NOT NULL COMMENT '截至月 1-12',
  `income_total` decimal(14,2) NOT NULL DEFAULT 0 COMMENT '累计收入',
  `special_deduction_total` decimal(14,2) NOT NULL DEFAULT 0 COMMENT '累计专项扣除（三险一金个人部分）',
  `special_extra_total` decimal(14,2) NOT NULL DEFAULT 0 COMMENT '累计专项附加扣除',
  `exempt_total` decimal(14,2) NOT NULL DEFAULT 0 COMMENT '累计免税收入',
  `other_deduction_total` decimal(14,2) NOT NULL DEFAULT 0 COMMENT '累计其他扣除（企业年金/商业健康险等）',
  `taxable_income_total` decimal(14,2) NOT NULL DEFAULT 0 COMMENT '累计应纳税所得额',
  `tax_paid_total` decimal(14,2) NOT NULL DEFAULT 0 COMMENT '累计已缴税额',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_tax_opening_emp_subject_year` (`employee_uuid`, `subject_uuid`, `tax_year`),
  KEY `idx_hrm_tax_opening_emp_year` (`employee_uuid`, `tax_year`),
  KEY `idx_hrm_tax_opening_subject` (`subject_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 员工个税档案 page（薪酬管理 dir ...1301 下 order 3）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001901','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','员工个税档案',0,'',3,'hrm_salary_tax_tax-profile','/hrm/salary/tax/tax-profile','view.hrm_salary_tax_tax-profile','mdi:account-card-details','1','route.hrm_salary_tax_tax-profile',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:list","desc":"员工个税档案列表"}]',0,'');

-- 员工个税档案 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001911','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001901','3','新增个税档案',0,'',0,'','','','','1','button.hrm.tax_profile.add',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:add","desc":"新增个税档案"}]',0,'v1:hrm:tax_profile:add'),
    ('aaa10000-1111-4000-8000-000000001912','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001901','3','编辑个税档案',0,'',1,'','','','','1','button.hrm.tax_profile.edit',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:edit","desc":"编辑个税档案"}]',0,'v1:hrm:tax_profile:edit'),
    ('aaa10000-1111-4000-8000-000000001913','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001901','3','删除个税档案',0,'',2,'','','','','1','button.hrm.tax_profile.delete',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:delete","desc":"删除个税档案"}]',0,'v1:hrm:tax_profile:delete'),
    ('aaa10000-1111-4000-8000-000000001914','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001901','3','导出个税档案',0,'',3,'','','','','1','button.hrm.tax_profile.export',0,'',0,0,'[]','[{"code":"v1:hrm:tax_profile:export","desc":"导出个税档案"}]',0,'v1:hrm:tax_profile:export');

-- 专项附加扣除 page（order 4）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001902','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','专项附加扣除',0,'',4,'hrm_salary_tax_special-deduction','/hrm/salary/tax/special-deduction','view.hrm_salary_tax_special-deduction','mdi:home-heart','1','route.hrm_salary_tax_special-deduction',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:list","desc":"专项附加扣除列表"}]',0,'');

-- 专项附加扣除 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001915','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','维护专项附加',0,'',0,'','','','','1','button.hrm.tax_special_deduction.maintain',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:maintain","desc":"维护专项附加"}]',0,'v1:hrm:tax_special_deduction:maintain'),
    ('aaa10000-1111-4000-8000-000000001916','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','删除专项附加',0,'',1,'','','','','1','button.hrm.tax_special_deduction.delete',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:delete","desc":"删除专项附加"}]',0,'v1:hrm:tax_special_deduction:delete'),
    ('aaa10000-1111-4000-8000-000000001917','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','年度维护专项附加',0,'',2,'','','','','1','button.hrm.tax_special_deduction.maintain_year',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:maintain_year","desc":"年度维护专项附加"}]',0,'v1:hrm:tax_special_deduction:maintain_year'),
    ('aaa10000-1111-4000-8000-000000001918','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','导入专项附加',0,'',3,'','','','','1','button.hrm.tax_special_deduction.import',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:import","desc":"导入专项附加"}]',0,'v1:hrm:tax_special_deduction:import'),
    ('aaa10000-1111-4000-8000-000000001919','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001902','3','导出专项附加',0,'',4,'','','','','1','button.hrm.tax_special_deduction.export',0,'',0,0,'[]','[{"code":"v1:hrm:tax_special_deduction:export","desc":"导出专项附加"}]',0,'v1:hrm:tax_special_deduction:export');

-- 个税累计期初 page（order 5）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001903','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','个税累计期初',0,'',5,'hrm_salary_tax_tax-opening','/hrm/salary/tax/tax-opening','view.hrm_salary_tax_tax-opening','mdi:calculator-variant','1','route.hrm_salary_tax_tax-opening',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:list","desc":"个税累计期初列表"}]',0,'');

-- 个税累计期初 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001921','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','新增累计期初',0,'',0,'','','','','1','button.hrm.tax_opening.add',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:add","desc":"新增累计期初"}]',0,'v1:hrm:tax_opening:add'),
    ('aaa10000-1111-4000-8000-000000001922','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','编辑累计期初',0,'',1,'','','','','1','button.hrm.tax_opening.edit',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:edit","desc":"编辑累计期初"}]',0,'v1:hrm:tax_opening:edit'),
    ('aaa10000-1111-4000-8000-000000001923','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','删除累计期初',0,'',2,'','','','','1','button.hrm.tax_opening.delete',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:delete","desc":"删除累计期初"}]',0,'v1:hrm:tax_opening:delete'),
    ('aaa10000-1111-4000-8000-000000001924','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','导入累计期初',0,'',3,'','','','','1','button.hrm.tax_opening.import',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:import","desc":"导入累计期初"}]',0,'v1:hrm:tax_opening:import'),
    ('aaa10000-1111-4000-8000-000000001925','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001903','3','导出累计期初',0,'',4,'','','','','1','button.hrm.tax_opening.export',0,'',0,0,'[]','[{"code":"v1:hrm:tax_opening:export","desc":"导出累计期初"}]',0,'v1:hrm:tax_opening:export');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001901','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001901',0),
    ('aaa20000-1111-4000-8000-000000001902','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001902',0),
    ('aaa20000-1111-4000-8000-000000001903','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001903',0),
    ('aaa20000-1111-4000-8000-000000001911','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001911',0),
    ('aaa20000-1111-4000-8000-000000001912','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001912',0),
    ('aaa20000-1111-4000-8000-000000001913','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001913',0),
    ('aaa20000-1111-4000-8000-000000001914','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001914',0),
    ('aaa20000-1111-4000-8000-000000001915','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001915',0),
    ('aaa20000-1111-4000-8000-000000001916','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001916',0),
    ('aaa20000-1111-4000-8000-000000001917','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001917',0),
    ('aaa20000-1111-4000-8000-000000001918','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001918',0),
    ('aaa20000-1111-4000-8000-000000001919','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001919',0),
    ('aaa20000-1111-4000-8000-000000001921','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001921',0),
    ('aaa20000-1111-4000-8000-000000001922','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001922',0),
    ('aaa20000-1111-4000-8000-000000001923','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001923',0),
    ('aaa20000-1111-4000-8000-000000001924','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001924',0),
    ('aaa20000-1111-4000-8000-000000001925','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001925',0);
