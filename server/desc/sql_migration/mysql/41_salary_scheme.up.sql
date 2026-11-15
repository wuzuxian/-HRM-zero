-- HRM-34 薪酬方案（薪酬管理域 薪酬设置）
-- 菜单 uuid 段：27xx（开工时已 grep 确认 22xx/26xx/27xx 中 27xx 空闲；26xx 在途加班已占，
-- 25xx 职位、24xx 补卡 —— 声明占用 27xx：dir 复用 13xx 薪酬管理目录，方案 page ...2702/buttons ...2711-2712）。

CREATE TABLE `hrm_salary_scheme` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `scheme_name` varchar(100) NOT NULL COMMENT '方案名称',
  `subject_uuid` varchar(36) NOT NULL COMMENT '关联薪酬主体（hrm_salary_subject.uuid）',
  `subject_name` varchar(100) NOT NULL DEFAULT '' COMMENT '关联主体名称快照',
  `person_status` varchar(100) NOT NULL DEFAULT '' COMMENT '适用人员状态（逗号分隔，1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休）',
  `settle_cycle` varchar(20) NOT NULL DEFAULT 'monthly' COMMENT '结算周期 monthly=月度',
  `pay_day` bigint NOT NULL DEFAULT 1 COMMENT '发薪日（几号，1-31）',
  `currency` varchar(10) NOT NULL DEFAULT 'CNY' COMMENT '币种（ISO 4217）',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '启停 1=启用 0=停用',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_salary_scheme_subject` (`subject_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `hrm_salary_scheme_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `scheme_uuid` varchar(36) NOT NULL COMMENT '薪酬方案 uuid',
  `item_uuid` varchar(36) NOT NULL COMMENT '薪资项目 uuid（不复制快照）',
  `override_amount` decimal(18,2) NULL COMMENT '覆盖金额（NULL=继承项目默认）',
  `override_formula` varchar(500) NOT NULL DEFAULT '' COMMENT '覆盖公式（空=继承项目默认）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_salary_scheme_item_scheme` (`scheme_uuid`),
  KEY `idx_hrm_salary_scheme_item_item` (`item_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 薪酬方案 page（挂在 13xx 薪酬管理目录下，order 3：主体=1 薪资项目=2）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002702','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','薪酬方案',0,'',3,'hrm_salary_setting_scheme','/hrm/salary/setting/scheme','view.hrm_salary_setting_scheme','mdi:file-cog-outline','1','route.hrm_salary_setting_scheme',0,'',0,0,'[]','[{"code":"v1:hrm:salary_scheme:list","desc":"薪酬方案列表"},{"code":"v1:hrm:salary_scheme:getSchemeDetail","desc":"薪酬方案详情"},{"code":"v1:hrm:salary_scheme:getSchemeItems","desc":"方案项目清单"}]',0,'');

-- 薪酬方案 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002711','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002702','3','新增薪酬方案',0,'',0,'','','','','1','button.hrm.salary_scheme.add',0,'',0,0,'[]','[{"code":"v1:hrm:salary_scheme:add","desc":"新增薪酬方案"}]',0,'v1:hrm:salary_scheme:add'),
    ('aaa10000-1111-4000-8000-000000002712','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002702','3','编辑薪酬方案',0,'',1,'','','','','1','button.hrm.salary_scheme.edit',0,'',0,0,'[]','[{"code":"v1:hrm:salary_scheme:edit","desc":"编辑薪酬方案"}]',0,'v1:hrm:salary_scheme:edit'),
    ('aaa10000-1111-4000-8000-000000002713','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002702','3','删除薪酬方案',0,'',2,'','','','','1','button.hrm.salary_scheme.delete',0,'',0,0,'[]','[{"code":"v1:hrm:salary_scheme:delete","desc":"删除薪酬方案"}]',0,'v1:hrm:salary_scheme:delete');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002701','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002702',0),
    ('aaa20000-1111-4000-8000-000000002711','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002711',0),
    ('aaa20000-1111-4000-8000-000000002712','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002712',0),
    ('aaa20000-1111-4000-8000-000000002713','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002713',0);
