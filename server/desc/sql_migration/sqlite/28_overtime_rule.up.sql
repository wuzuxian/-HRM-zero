-- HRM-13 加班规则（假勤管理域 假勤设置）
-- 表: hrm_overtime_rule 加班规则
--   - 工作日/休息日/法定节假日 三种当天类型各自的 需审批/免审批、需打卡/免打卡、调休倍数
--   - 最小计算单位（分钟, 默认30）、月上限（小时, 0=不限制）
--   - group_uuids 使用的考勤组（空=适用于所有考勤组的兜底规则）; org_uuid 组织默认规则归属（空=全局）
--   - 消费契约（HRM-21 加班申请单）: ResolveOvertimeRule 按 加班日期当天类型 查规则判定审批/折算
-- 菜单 uuid 段: 15xx（10xx=假勤已用, 12xx/13xx=绩效/薪酬, 14xx=证书模板, 15xx 本工单独占）

CREATE TABLE hrm_overtime_rule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rule_name TEXT NOT NULL,
    workday_approval_required INTEGER NOT NULL DEFAULT 1,
    workday_punch_required INTEGER NOT NULL DEFAULT 1,
    workday_comp_multiplier REAL NOT NULL DEFAULT 1,
    restday_approval_required INTEGER NOT NULL DEFAULT 1,
    restday_punch_required INTEGER NOT NULL DEFAULT 1,
    restday_comp_multiplier REAL NOT NULL DEFAULT 1,
    holiday_approval_required INTEGER NOT NULL DEFAULT 1,
    holiday_punch_required INTEGER NOT NULL DEFAULT 1,
    holiday_comp_multiplier REAL NOT NULL DEFAULT 1,
    min_unit_minutes INTEGER NOT NULL DEFAULT 30,
    monthly_limit_hours REAL NOT NULL DEFAULT 0,
    group_uuids TEXT NOT NULL DEFAULT '',
    org_uuid TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '1',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_hrm_overtime_rule_org_uuid ON hrm_overtime_rule (org_uuid);
CREATE INDEX idx_hrm_overtime_rule_status ON hrm_overtime_rule (status);

-- 加班规则 page (under 假勤设置 dir aaa...000101)
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001501','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000000101','2','加班规则',0,'',5,'hrm_attendance_setting_overtime-rule','/hrm/attendance/setting/overtime-rule','view.hrm_attendance_setting_overtime-rule','mdi:clock-fast','1','route.hrm_attendance_setting_overtime-rule',0,'',0,0,'[]','[{"code":"v1:hrm:overtimerule:list","desc":"加班规则列表"},{"code":"v1:hrm:overtimerule:resolve","desc":"加班规则解析查询"}]',0,'');

-- 加班规则 buttons
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000001511','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001501','3','新增加班规则',0,'',0,'','','','','1','button.hrm.overtimerule.add',0,'',0,0,'[]','[{"code":"v1:hrm:overtimerule:add","desc":"新增加班规则"}]',0,'v1:hrm:overtimerule:add'),
    ('aaa10000-1111-4000-8000-000000001512','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001501','3','编辑加班规则',0,'',1,'','','','','1','button.hrm.overtimerule.edit',0,'',0,0,'[]','[{"code":"v1:hrm:overtimerule:edit","desc":"编辑加班规则"}]',0,'v1:hrm:overtimerule:edit'),
    ('aaa10000-1111-4000-8000-000000001513','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001501','3','删除加班规则',0,'',2,'','','','','1','button.hrm.overtimerule.delete',0,'',0,0,'[]','[{"code":"v1:hrm:overtimerule:delete","desc":"删除加班规则"}]',0,'v1:hrm:overtimerule:delete'),
    ('aaa10000-1111-4000-8000-000000001514','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000001501','3','加班规则启停',0,'',3,'','','','','1','button.hrm.overtimerule.togglestatus',0,'',0,0,'[]','[{"code":"v1:hrm:overtimerule:toggleStatus","desc":"加班规则启用/停用"}]',0,'v1:hrm:overtimerule:toggleStatus');

-- 超级管理员 gets the new menus (casbin rules are derived from these at startup)
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000001501','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001501',0),
    ('aaa20000-1111-4000-8000-000000001502','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001511',0),
    ('aaa20000-1111-4000-8000-000000001503','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001512',0),
    ('aaa20000-1111-4000-8000-000000001504','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001513',0),
    ('aaa20000-1111-4000-8000-000000001505','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000001514',0);

CREATE TRIGGER hrm_overtime_rule_update_time
AFTER UPDATE ON hrm_overtime_rule
FOR EACH ROW
BEGIN
    UPDATE hrm_overtime_rule SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
