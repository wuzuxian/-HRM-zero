-- HRM-16 考勤打卡与打卡记录（日结基础）（假勤管理域 我的假勤/假勤数据）
-- 表:
--   - hrm_punch_record 打卡记录：(punch_date, employee_uuid) 唯一 —— 一人一天一行，
--     上/下班卡写入同一行（幂等：同日同类型重复打卡按拒绝策略，见 issue 正文）。
--     签到/签退状态 normal 正常 / field 外勤 / 空=缺卡；source app/admin/import/card_replacement。
--   - 日结字段直接落在打卡记录行上（settle_status/settle_late_minutes/settle_early_minutes/
--     settle_work_hours/settle_time），日结语义 = 行内快照，HRM-17 看板 / HRM-22 考勤统计 /
--     HRM-23 月报直接消费本表日结列（消费契约见 hrm-notes/hrm-16-notes.md）。
-- 消费契约:
--   - 日结函数 SettleDay(ctx, conn, employeeUuid, date) 返回 DaySettleResult{Status,LateMinutes,EarlyMinutes,WorkHours,...}
--     （server/internal/logic/v1/hrm/punch/settle.go）；规则解析复用 attendancegroup.ResolveEffectiveRule。
--   - HRM-17 个人看板：按 employee_uuid + settle_status / settle_work_hours 汇总。
--   - HRM-22 考勤统计 / HRM-23 考勤月报：按日期区间聚合 settle_* 列；导入外部考勤数据走
--     punch.ImportPunch（CSV: date,employee_no,clock_in_time,clock_in_status,clock_in_location,
--     clock_out_time,clock_out_status,clock_out_location,overtime），HRM-22 可直接复用该 seam。
-- 错误码 1052x 段（server/internal/errcodes/hrm_punch.go）。
-- 菜单 uuid 段 19xx（grep 全部 migration: 0xxx 基础/03xx 假勤设置/07xx 考勤组/09xx-16xx 各域/
-- 18xx 假期余额；本票取 20xx（19xx 已被 HRM-36 个税占用）: 2001 我的打卡 page / 2011-2012 buttons；2002 打卡记录 page /
-- 2021-2026 buttons；role_menu aaa20000 20xx）。

CREATE TABLE `hrm_punch_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL UNIQUE,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `punch_date` varchar(10) NOT NULL COMMENT '打卡日期 yyyy-MM-dd（按规则归属日，跨天班下班卡归属上班日）',
  `employee_uuid` varchar(36) NOT NULL COMMENT '员工档案 uuid (HRM-03 hrm_employee.uuid)',
  `user_uuid` varchar(36) NOT NULL DEFAULT '' COMMENT '登录账号 uuid (manage_user.uuid 快照)',
  `username` varchar(50) NOT NULL DEFAULT '' COMMENT '登录账号（快照）',
  `employee_no` varchar(20) NOT NULL DEFAULT '' COMMENT '工号（快照）',
  `employee_name` varchar(100) NOT NULL DEFAULT '' COMMENT '姓名（快照）',
  `clock_in_time` datetime DEFAULT NULL COMMENT '签到时间（NULL=缺上班卡）',
  `clock_in_status` varchar(10) NOT NULL DEFAULT '' COMMENT '签到状态：normal 正常 / field 外勤 / 空=缺卡',
  `clock_in_location` varchar(200) NOT NULL DEFAULT '' COMMENT '签到地点（GPS 地址文本，外勤时填）',
  `clock_out_time` datetime DEFAULT NULL COMMENT '签退时间（NULL=缺下班卡）',
  `clock_out_status` varchar(10) NOT NULL DEFAULT '' COMMENT '签退状态：normal 正常 / field 外勤 / 空=缺卡',
  `clock_out_location` varchar(200) NOT NULL DEFAULT '' COMMENT '签退地点（GPS 地址文本，外勤时填）',
  `overtime` varchar(1) NOT NULL DEFAULT '0' COMMENT '加班（是/否）：下班卡晚于规则下班时间弹性阈值',
  `source` varchar(20) NOT NULL DEFAULT 'app' COMMENT '记录来源：app 员工打卡 / admin 管理端新增 / import 外部导入 / card_replacement 补卡',
  `rule_source` varchar(20) NOT NULL DEFAULT '' COMMENT '当日生效规则来源（group/org/空=无规则），日结快照',
  `settle_status` varchar(20) NOT NULL DEFAULT '' COMMENT '日结结果：normal 正常 / late 迟到 / early 早退 / missing 缺卡 / rest 休息日 / 无规则=未判定，空=未日结',
  `settle_late_minutes` bigint NOT NULL DEFAULT 0 COMMENT '日结迟到分钟数',
  `settle_early_minutes` bigint NOT NULL DEFAULT 0 COMMENT '日结早退分钟数',
  `settle_work_hours` decimal(6,2) NOT NULL DEFAULT 0 COMMENT '日结工时（小时，扣除休息时段，跨天班含次日时段）',
  `settle_time` datetime DEFAULT NULL COMMENT '日结时间（NULL=未日结）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_hrm_punch_record` (`punch_date`, `employee_uuid`),
  KEY `idx_hrm_punch_record_employee` (`employee_uuid`),
  KEY `idx_hrm_punch_record_date` (`punch_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 我的假勤-考勤打卡 page (假勤管理 dir aaa...000101 下, order 7)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002001','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','考勤打卡',0,'',7,'hrm_attendance_my_attendance-punch','/hrm/attendance/my/attendance-punch','view.hrm_attendance_my_attendance-punch','mdi:fingerprint','1','route.hrm_attendance_my_attendance-punch',0,'',0,0,'[]','[{"code":"v1:hrm:punch:myList","desc":"我的打卡记录列表"},{"code":"v1:hrm:punch:punch","desc":"打卡"}]',0,'');

-- 考勤打卡 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002011','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002001','3','上/下班打卡',0,'',0,'','','','','1','button.hrm.punch.punch',0,'',0,0,'[]','[{"code":"v1:hrm:punch:punch","desc":"上/下班打卡"}]',0,'v1:hrm:punch:punch'),
    ('aaa10000-1111-4000-8000-000000002012','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002001','3','刷新我的打卡',0,'',1,'','','','','1','button.hrm.punch.myList',0,'',0,0,'[]','[{"code":"v1:hrm:punch:myList","desc":"我的打卡记录列表"}]',0,'v1:hrm:punch:myList');

-- 假勤数据-打卡记录 page (管理端, order 8)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002002','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','打卡记录',0,'',8,'hrm_attendance_data_punch-record','/hrm/attendance/data/punch-record','view.hrm_attendance_data_punch-record','mdi:clipboard-text-clock-outline','1','route.hrm_attendance_data_punch-record',0,'',0,0,'[]','[{"code":"v1:hrm:punch:list","desc":"打卡记录列表"},{"code":"v1:hrm:punch:settle","desc":"手动日结"}]',0,'');

-- 打卡记录 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, `order`, route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000002021','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002002','3','新增打卡记录',0,'',0,'','','','','1','button.hrm.punch.add',0,'',0,0,'[]','[{"code":"v1:hrm:punch:add","desc":"新增打卡记录"}]',0,'v1:hrm:punch:add'),
    ('aaa10000-1111-4000-8000-000000002022','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002002','3','修改打卡记录',0,'',1,'','','','','1','button.hrm.punch.edit',0,'',0,0,'[]','[{"code":"v1:hrm:punch:edit","desc":"修改打卡记录"}]',0,'v1:hrm:punch:edit'),
    ('aaa10000-1111-4000-8000-000000002023','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002002','3','删除打卡记录',0,'',2,'','','','','1','button.hrm.punch.delete',0,'',0,0,'[]','[{"code":"v1:hrm:punch:delete","desc":"删除打卡记录"}]',0,'v1:hrm:punch:delete'),
    ('aaa10000-1111-4000-8000-000000002024','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002002','3','导出打卡记录',0,'',3,'','','','','1','button.hrm.punch.export',0,'',0,0,'[]','[{"code":"v1:hrm:punch:export","desc":"导出打卡记录"}]',0,'v1:hrm:punch:export'),
    ('aaa10000-1111-4000-8000-000000002025','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002002','3','导入外部考勤数据',0,'',4,'','','','','1','button.hrm.punch.import',0,'',0,0,'[]','[{"code":"v1:hrm:punch:import","desc":"导入外部考勤数据"}]',0,'v1:hrm:punch:import'),
    ('aaa10000-1111-4000-8000-000000002026','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000002002','3','手动日结',0,'',5,'','','','','1','button.hrm.punch.settle',0,'',0,0,'[]','[{"code":"v1:hrm:punch:settle","desc":"手动日结"}]',0,'v1:hrm:punch:settle');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000002001','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002001',0),
    ('aaa20000-1111-4000-8000-000000002002','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002002',0),
    ('aaa20000-1111-4000-8000-000000002011','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002011',0),
    ('aaa20000-1111-4000-8000-000000002012','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002012',0),
    ('aaa20000-1111-4000-8000-000000002021','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002021',0),
    ('aaa20000-1111-4000-8000-000000002022','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002022',0),
    ('aaa20000-1111-4000-8000-000000002023','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002023',0),
    ('aaa20000-1111-4000-8000-000000002024','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002024',0),
    ('aaa20000-1111-4000-8000-000000002025','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002025',0),
    ('aaa20000-1111-4000-8000-000000002026','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000002026',0);
