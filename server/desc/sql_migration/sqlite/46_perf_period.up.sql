-- HRM-42 绩效批次（演示站 9.2 批次与任务流，GLOSSARY「绩效批次」）：
--   一次考核的执行实例（PERF- 编号），按方案圈定适用人群（公司/部门/职位/职务/人员状态
--   多维筛选条件存 JSON），生命周期 草稿→进行中→待确认→待校准→结果已发布→已发布。
-- 编号 PERF-年-流水 复用 hrm_bill_sequence（bill_type_code='PERF'，日期=年份，同 PX 班次编号模式）。
-- 参与人表 hrm_perf_period_participant 是 HRM-43 考核任务链的取数契约（GetPeriodParticipants）。
-- 错误码 10633-10638（server/internal/errcodes/hrm_perf_period.go，10630-10632 已被看板/考勤统计占用）。
-- 菜单 uuid 段 32xx（grep 全部 migration: 0xxx 基础/03xx 假勤设置/07xx 考勤组/09xx-30xx 各域；
-- 本票取 32xx: 3201 考核批次工作台 page / 3211-3214 buttons；role_menu aaa20000 32xx）。

DROP TABLE IF EXISTS hrm_perf_period;
DROP TABLE IF EXISTS hrm_perf_period_participant;

CREATE TABLE hrm_perf_period (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    period_no TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    period_type TEXT NOT NULL,
    scheme_uuid TEXT NOT NULL,
    scope TEXT NOT NULL,
    period_start TEXT NOT NULL,
    period_end TEXT NOT NULL,
    result_publish_date TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '0',
    remark TEXT NOT NULL DEFAULT ''
);

CREATE TABLE hrm_perf_period_participant (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    period_uuid TEXT NOT NULL,
    employee_uuid TEXT NOT NULL
);

CREATE UNIQUE INDEX uniq_hrm_perf_period_participant ON hrm_perf_period_participant (period_uuid, employee_uuid);
CREATE INDEX idx_hrm_perf_period_participant_employee ON hrm_perf_period_participant (employee_uuid);
CREATE INDEX idx_hrm_perf_period_scheme ON hrm_perf_period (scheme_uuid);

CREATE TRIGGER hrm_perf_period_update_time
AFTER UPDATE ON hrm_perf_period
FOR EACH ROW
BEGIN
    UPDATE hrm_perf_period SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_perf_period_participant_update_time
AFTER UPDATE ON hrm_perf_period_participant
FOR EACH ROW
BEGIN
    UPDATE hrm_perf_period_participant SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- 绩效管理目录（migration 12 已建 uuid f0e1d2c3-...4e5f）下新增「考核批次工作台」页面 + 按钮权限
-- （demo-inventory 9.2 路径 /hrm/performance/period-list）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003201','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','考核批次工作台',0,'',4,'hrm_performance_period-list','/hrm/performance/period-list','view.hrm_performance_period-list','mdi:calendar-clock','1','route.hrm_performance_period-list',0,'',0,0,'[]','[{"code":"v1:hrm:perf_period:list","desc":"考核批次列表"}]',0,''),
    ('aaa10000-1111-4000-8000-000000003211','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003201','3','新增批次',0,'',0,'','','','','1','button.hrm.performance.period.add',0,'',0,0,'[]','[{"code":"v1:hrm:perf_period:add","desc":"新增考核批次"},{"code":"v1:hrm:perf_scheme:list","desc":"绩效方案列表"}]',0,'v1:hrm:perf_period:add'),
    ('aaa10000-1111-4000-8000-000000003212','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003201','3','编辑批次',0,'',1,'','','','','1','button.hrm.performance.period.edit',0,'',0,0,'[]','[{"code":"v1:hrm:perf_period:edit","desc":"编辑考核批次"}]',0,'v1:hrm:perf_period:edit'),
    ('aaa10000-1111-4000-8000-000000003213','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003201','3','删除批次',0,'',2,'','','','','1','button.hrm.performance.period.delete',0,'',0,0,'[]','[{"code":"v1:hrm:perf_period:delete","desc":"删除考核批次"}]',0,'v1:hrm:perf_period:delete'),
    ('aaa10000-1111-4000-8000-000000003214','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003201','3','发布批次',0,'',3,'','','','','1','button.hrm.performance.period.publish',0,'',0,0,'[]','[{"code":"v1:hrm:perf_period:publish","desc":"发布考核批次"}]',0,'v1:hrm:perf_period:publish');

-- 超级管理员角色可见
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003201','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003201',0),
    ('aaa20000-1111-4000-8000-000000003211','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003211',0),
    ('aaa20000-1111-4000-8000-000000003212','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003212',0),
    ('aaa20000-1111-4000-8000-000000003213','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003213',0),
    ('aaa20000-1111-4000-8000-000000003214','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003214',0);
