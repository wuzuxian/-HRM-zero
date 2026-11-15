-- HRM-35 员工薪资档案（薪酬管理域）
-- 菜单 uuid 段：29xx（开工时已 grep 三方言迁移：22xx/28xx 空闲但 28xx 已被并行 dashboard 票登记预留，
-- 本票取 29xx，声明占用）。

CREATE TABLE `hrm_salary_archive` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工（hrm_employee.uuid）',
  `employee_no` varchar(20) NOT NULL DEFAULT '' COMMENT '员工ID（工号快照，展示用）',
  `employee_name` varchar(50) NOT NULL DEFAULT '' COMMENT '员工姓名快照（列表/试算展示不回查员工表）',
  `company_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '所属公司（选员工时由 hrm_employee.company_uuid 带出）',
  `company_name` varchar(100) NOT NULL DEFAULT '' COMMENT '所属公司名称快照',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '所属部门（选员工时由 hrm_employee.org_uuid 带出）',
  `org_name` varchar(100) NOT NULL DEFAULT '' COMMENT '所属部门名称快照',
  `subject_uuid` varchar(36) NOT NULL COMMENT '薪酬主体（hrm_salary_subject.uuid）',
  `subject_name` varchar(100) NOT NULL DEFAULT '' COMMENT '薪酬主体名称快照',
  `scheme_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '薪酬方案（hrm_salary_scheme.uuid；可为空=尚未绑定方案）',
  `scheme_name` varchar(100) NOT NULL DEFAULT '' COMMENT '薪酬方案名称快照',
  `fixed_salary` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '固定薪资',
  `fixed_allowance` decimal(18,2) NOT NULL DEFAULT 0 COMMENT '固定补贴',
  `status` varchar(1) NOT NULL DEFAULT '1' COMMENT '启停 1=启用 0=停用（停用档案不参与试算取档）',
  `effective_date` varchar(10) NOT NULL COMMENT '生效日期 YYYY-MM-DD',
  `expire_date` varchar(10) NOT NULL DEFAULT '' COMMENT '失效日期 YYYY-MM-DD（空=长期有效；必须>生效日期）',
  `remark` varchar(500) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_salary_archive_employee` (`employee_uuid`),
  KEY `idx_hrm_salary_archive_subject` (`subject_uuid`),
  KEY `idx_hrm_salary_archive_scheme` (`scheme_uuid`),
  KEY `idx_hrm_salary_archive_effective` (`effective_date`),
  KEY `idx_hrm_salary_archive_expire` (`expire_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 员工薪资档案 page（挂在 13xx 薪酬管理目录下，order 6：主体=1 项目=2 方案=3 个税=3/4/5 冲突取 6）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002901','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001301','2','员工薪资档案',0,'',6,'hrm_salary_archive-list','/hrm/salary/archive-list','view.hrm_salary_archive-list','mdi:account-cash-outline','1','route.hrm_salary_archive-list',0,'',0,0,'[]','[{"code":"v1:hrm:salary_archive:list","desc":"员工薪资档案列表"},{"code":"v1:hrm:salary_archive:getEffectiveArchive","desc":"取员工生效档案"}]',0,'');

-- 员工薪资档案 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002911','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002901','3','新增薪资档案',0,'',0,'','','','','1','button.hrm.salary_archive.add',0,'',0,0,'[]','[{"code":"v1:hrm:salary_archive:add","desc":"新增薪资档案"}]',0,'v1:hrm:salary_archive:add'),
    ('aaa10000-1111-4000-8000-000000002912','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002901','3','编辑薪资档案',0,'',1,'','','','','1','button.hrm.salary_archive.edit',0,'',0,0,'[]','[{"code":"v1:hrm:salary_archive:edit","desc":"编辑薪资档案"}]',0,'v1:hrm:salary_archive:edit'),
    ('aaa10000-1111-4000-8000-000000002913','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002901','3','删除薪资档案',0,'',2,'','','','','1','button.hrm.salary_archive.delete',0,'',0,0,'[]','[{"code":"v1:hrm:salary_archive:delete","desc":"删除薪资档案"}]',0,'v1:hrm:salary_archive:delete');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002901','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002901',0),
    ('aaa20000-1111-4000-8000-000000002911','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002911',0),
    ('aaa20000-1111-4000-8000-000000002912','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002912',0),
    ('aaa20000-1111-4000-8000-000000002913','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002913',0);
