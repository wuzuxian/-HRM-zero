-- HRM-20 补卡单（HR206）（假勤管理域 假勤申请）
-- 表:
--   - hrm_card_replacement_bill 补卡单：补卡日期/补卡类型（clockIn 上班卡 / clockOut 下班卡）/
--     补卡时间/补卡原因；原上班打卡/原下班打卡 = 提交时 hrm_punch_record 当日行的快照列。
--   - 审批通过事务内（bill.OnApproved, 类型码 206）：补卡时间写入当日 hrm_punch_record 对应列
--     （上班卡→clock_in_time/status/location，下班卡→clock_out_*；当日无打卡行则建行，
--     source=card_replacement），并 re-run punch.SettleDay（HRM-16 消费契约）重算该员工该日日结。
-- 生命周期（提交/审批/撤销/审批记录）走共享单据底座 v1/hrm/bill（迁移 11）。
-- 错误码 1057x 段（server/internal/errcodes/hrm_card_replacement.go；
-- grep errcodes/: 1032x-1056x 已占用，1057x 空闲，本票取 10570-10574）。
-- 菜单 uuid 段 24xx（grep 全部 migration: 10xx-21xx+23xx 已占用；注册表指定下票取 24xx:
-- 2401 page / 2411-2417 buttons（含共享单据底座 hrm:bill:* 按钮）/ role_menu aaa20000 24xx）。

CREATE TABLE `hrm_card_replacement_bill` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `bill_no` varchar(32) NOT NULL COMMENT '单据编号 HR206-YYYYMMDD+5位流水',
  `status` varchar(16) NOT NULL DEFAULT 'unsubmitted' COMMENT '单据状态 unsubmitted/approving/approved/rejected/revoked',
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工档案 uuid（HRM-03）',
  `employee_no` varchar(50) NOT NULL DEFAULT '' COMMENT '员工工号（快照）',
  `employee_name` varchar(100) NOT NULL DEFAULT '' COMMENT '员工姓名（快照）',
  `org_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '员工所属部门（组织 uuid，快照）',
  `org_name` varchar(100) NOT NULL DEFAULT '' COMMENT '部门名称（快照）',
  `punch_date` varchar(10) NOT NULL COMMENT '补卡日期 yyyy-MM-dd',
  `card_type` varchar(10) NOT NULL COMMENT '补卡类型 clockIn 上班卡 / clockOut 下班卡',
  `punch_time` varchar(19) NOT NULL COMMENT '补卡时间 YYYY-MM-DD HH:MM:SS',
  `reason` varchar(500) NOT NULL DEFAULT '' COMMENT '补卡原因',
  `orig_clock_in_time` varchar(19) NOT NULL DEFAULT '' COMMENT '原上班打卡（提交时 hrm_punch_record 快照，空=缺卡）',
  `orig_clock_out_time` varchar(19) NOT NULL DEFAULT '' COMMENT '原下班打卡（提交时 hrm_punch_record 快照，空=缺卡）',
  `create_by` varchar(36) NOT NULL DEFAULT '' COMMENT '创建人（申请人）用户 uuid',
  `creator_name` varchar(100) NOT NULL DEFAULT '' COMMENT '申请人姓名（快照）',
  PRIMARY KEY (`id`),
  KEY `idx_hrm_crb_status` (`status`),
  KEY `idx_hrm_crb_employee` (`employee_uuid`),
  KEY `idx_hrm_crb_date_type` (`punch_date`, `card_type`),
  KEY `idx_hrm_crb_employee_no` (`employee_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 补卡管理 page (假勤管理 dir aaa...000101 下, order 10; 已用: 设置 1-5/周期 6/余额 6/打卡 7/请假 7/记录 8/排班表 9)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002401','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','补卡管理',0,'',10,'hrm_attendance_apply_card-replacement-list','/hrm/attendance/apply/card-replacement-list','view.hrm_attendance_apply_card-replacement-list','mdi:card-account-details-clock-outline','1','route.hrm_attendance_apply_card-replacement-list',0,'',0,0,'[]','[{"code":"v1:hrm:cardreplacement:list","desc":"补卡单列表"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,'');

-- 补卡管理 buttons（页面操作 + 共享单据底座按钮码 v1:hrm:bill:*）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002411','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002401','3','新增补卡单',0,'',0,'','','','','1','button.hrm.cardreplacement.add',0,'',0,0,'[]','[{"code":"v1:hrm:cardreplacement:add","desc":"新增补卡单"}]',0,'v1:hrm:cardreplacement:add'),
    ('aaa10000-1111-4000-8000-000000002412','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002401','3','编辑补卡单',0,'',1,'','','','','1','button.hrm.cardreplacement.edit',0,'',0,0,'[]','[{"code":"v1:hrm:cardreplacement:edit","desc":"编辑补卡单"}]',0,'v1:hrm:cardreplacement:edit'),
    ('aaa10000-1111-4000-8000-000000002413','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002401','3','删除补卡单',0,'',2,'','','','','1','button.hrm.cardreplacement.delete',0,'',0,0,'[]','[{"code":"v1:hrm:cardreplacement:delete","desc":"删除补卡单"}]',0,'v1:hrm:cardreplacement:delete'),
    ('aaa10000-1111-4000-8000-000000002414','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002401','3','提交单据',0,'',3,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000002415','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002401','3','审批通过',0,'',4,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000002416','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002401','3','审批驳回',0,'',5,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000002417','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002401','3','撤销单据',0,'',6,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002401','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002401',0),
    ('aaa20000-1111-4000-8000-000000002402','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002411',0),
    ('aaa20000-1111-4000-8000-000000002403','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002412',0),
    ('aaa20000-1111-4000-8000-000000002404','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002413',0),
    ('aaa20000-1111-4000-8000-000000002405','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002414',0),
    ('aaa20000-1111-4000-8000-000000002406','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002415',0),
    ('aaa20000-1111-4000-8000-000000002407','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002416',0),
    ('aaa20000-1111-4000-8000-000000002408','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002417',0);
