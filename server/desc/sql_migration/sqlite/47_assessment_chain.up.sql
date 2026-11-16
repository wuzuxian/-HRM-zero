-- HRM-43 考核任务链（演示站 9.2 批次与任务流，GLOSSARY「考核任务」）：
--   批次发布后为每个参与人生成考核任务（考核单号 PA-年-流水，复用 hrm_bill_sequence，
--   bill_type_code='PA'，日期=批次周期开始年份）。五阶段流转（前置守卫：前一节点完成
--   才可下一节点）：目标确认 → 自评填报 → 上级评审 → 校准填报 → 结果确认（可发起申诉）。
--   评分按模板快照 items 权重汇总（percent 百分制→等级 A/B/C/D；grade/amount 直存）。
--   申诉表 hrm_perf_appeal：发起→校准人裁决（维持/调整得分）→关闭。
-- 批次状态联动（联动点，HRM-42 状态机不改动）：
--   - 任务生成后批次仍为'1'进行中；全部任务到达校准填报（current_node='4'）→ 由
--     GenerateAndSyncPeriod 联动 AdvancePeriod 至'2'待确认再至'3'待校准（待确认节点
--     在任务链内联，无独立停留）；本次实现：全部任务 current_node>='4' 时推至'3'待校准。
--   - 全部任务 result_time 非空（结果确认完成）→ 推至'4'结果已发布。
--   HRM-45 审批通过后推'4'→'5'已发布（不改本表）。
-- 错误码 10640-10647（server/internal/errcodes/hrm_perf_assessment.go，1063x 已被占用）。
-- 菜单 uuid 段 33xx（grep 全部 migration 确认 00xx-32xx 已用、33xx 空闲）：
--   3301 目标确认 / 3302 自评填报 / 3303 评审填报 / 3304 结果确认 page，
--   3311-3320 buttons，3341 申诉裁决 button；role_menu aaa20000 33xx。

DROP TABLE IF EXISTS hrm_perf_appeal;
DROP TABLE IF EXISTS hrm_perf_assessment;

CREATE TABLE hrm_perf_assessment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    assessment_no TEXT NOT NULL UNIQUE,
    period_uuid TEXT NOT NULL,
    employee_uuid TEXT NOT NULL,
    template_uuid TEXT NOT NULL,
    items_snapshot TEXT NOT NULL,
    current_node TEXT NOT NULL DEFAULT '1',
    task_status TEXT NOT NULL DEFAULT '0',
    plan_time TEXT NOT NULL DEFAULT '',
    plan_by TEXT NOT NULL DEFAULT '',
    self_score REAL NOT NULL DEFAULT 0,
    self_grade TEXT NOT NULL DEFAULT '',
    self_items TEXT NOT NULL DEFAULT '[]',
    self_comment TEXT NOT NULL DEFAULT '',
    self_time TEXT NOT NULL DEFAULT '',
    self_by TEXT NOT NULL DEFAULT '',
    review_score REAL NOT NULL DEFAULT 0,
    review_grade TEXT NOT NULL DEFAULT '',
    review_items TEXT NOT NULL DEFAULT '[]',
    review_comment TEXT NOT NULL DEFAULT '',
    review_time TEXT NOT NULL DEFAULT '',
    review_by TEXT NOT NULL DEFAULT '',
    final_score REAL NOT NULL DEFAULT 0,
    final_grade TEXT NOT NULL DEFAULT '',
    calib_comment TEXT NOT NULL DEFAULT '',
    calib_time TEXT NOT NULL DEFAULT '',
    calib_by TEXT NOT NULL DEFAULT '',
    result_time TEXT NOT NULL DEFAULT '',
    result_by TEXT NOT NULL DEFAULT ''
);

CREATE TABLE hrm_perf_appeal (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    assessment_uuid TEXT NOT NULL,
    period_uuid TEXT NOT NULL,
    employee_uuid TEXT NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT '0',
    decision_comment TEXT NOT NULL DEFAULT '',
    adjusted_score REAL NOT NULL DEFAULT 0,
    adjusted_grade TEXT NOT NULL DEFAULT '',
    handled_by TEXT NOT NULL DEFAULT '',
    handled_time TEXT NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX uniq_hrm_perf_assessment_period_employee ON hrm_perf_assessment (period_uuid, employee_uuid);
CREATE INDEX idx_hrm_perf_assessment_period ON hrm_perf_assessment (period_uuid);
CREATE INDEX idx_hrm_perf_assessment_employee ON hrm_perf_assessment (employee_uuid);
CREATE INDEX idx_hrm_perf_assessment_node ON hrm_perf_assessment (current_node, task_status);
CREATE INDEX idx_hrm_perf_appeal_assessment ON hrm_perf_appeal (assessment_uuid);
CREATE INDEX idx_hrm_perf_appeal_period ON hrm_perf_appeal (period_uuid);
CREATE INDEX idx_hrm_perf_appeal_status ON hrm_perf_appeal (status);

CREATE TRIGGER hrm_perf_assessment_update_time
AFTER UPDATE ON hrm_perf_assessment
FOR EACH ROW
BEGIN
    UPDATE hrm_perf_assessment SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_perf_appeal_update_time
AFTER UPDATE ON hrm_perf_appeal
FOR EACH ROW
BEGIN
    UPDATE hrm_perf_appeal SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- 绩效管理目录（migration 12 uuid f0e1d2c3-...4e5f）下新增四个填报列表页
-- （demo-inventory 9.2 路径：/hrm/performance/assessment-task/{plan,self-review,leader-review,result-confirm}-list）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu, active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href, multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003300','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','1','考核任务',0,'',4,'hrm_performance_assessment-task','/hrm/performance/assessment-task','view.hrm_performance_assessment-task','mdi:format-list-checks','1','route.hrm_performance_assessment-task',0,'',0,0,'[]','[]',0,''),
    ('aaa10000-1111-4000-8000-000000003301','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003300','2','目标确认',0,'',0,'hrm_performance_assessment-task_plan-list','/hrm/performance/assessment-task/plan-list','view.hrm_performance_assessment-task_plan-list','mdi:target','1','route.hrm_performance_assessment-task_plan-list',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:list","desc":"考核任务列表"},{"code":"v1:hrm:perf_assessment:detail","desc":"考核任务详情"}]',0,''),
    ('aaa10000-1111-4000-8000-000000003302','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003300','2','自评填报',0,'',1,'hrm_performance_assessment-task_self-review-list','/hrm/performance/assessment-task/self-review-list','view.hrm_performance_assessment-task_self-review-list','mdi:pencil-box','1','route.hrm_performance_assessment-task_self-review-list',0,'',0,0,'[]','[]',0,''),
    ('aaa10000-1111-4000-8000-000000003303','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003300','2','评审填报',0,'',2,'hrm_performance_assessment-task_leader-review-list','/hrm/performance/assessment-task/leader-review-list','view.hrm_performance_assessment-task_leader-review-list','mdi:account-check','1','route.hrm_performance_assessment-task_leader-review-list',0,'',0,0,'[]','[]',0,''),
    ('aaa10000-1111-4000-8000-000000003304','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003300','2','结果确认',0,'',3,'hrm_performance_assessment-task_result-confirm-task-list','/hrm/performance/assessment-task/result-confirm-task-list','view.hrm_performance_assessment-task_result-confirm-task-list','mdi:clipboard-check','1','route.hrm_performance_assessment-task_result-confirm-task-list',0,'',0,0,'[]','[]',0,''),
    ('aaa10000-1111-4000-8000-000000003311','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003301','3','生成考核任务',0,'',0,'','','','','1','button.hrm.performance.assessment.generate',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:generate","desc":"生成考核任务"}]',0,'v1:hrm:perf_assessment:generate'),
    ('aaa10000-1111-4000-8000-000000003312','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003301','3','确认目标',0,'',1,'','','','','1','button.hrm.performance.assessment.plan',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:confirmPlan","desc":"目标确认"}]',0,'v1:hrm:perf_assessment:confirmPlan'),
    ('aaa10000-1111-4000-8000-000000003313','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003302','3','自评提交',0,'',0,'','','','','1','button.hrm.performance.assessment.selfReview',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:submitSelfReview","desc":"自评填报"}]',0,'v1:hrm:perf_assessment:submitSelfReview'),
    ('aaa10000-1111-4000-8000-000000003314','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003303','3','评审提交',0,'',0,'','','','','1','button.hrm.performance.assessment.leaderReview',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:review","desc":"上级评审"}]',0,'v1:hrm:perf_assessment:review'),
    ('aaa10000-1111-4000-8000-000000003315','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003303','3','校准提交',0,'',1,'','','','','1','button.hrm.performance.assessment.calibrate',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:calibrate","desc":"校准填报"}]',0,'v1:hrm:perf_assessment:calibrate'),
    ('aaa10000-1111-4000-8000-000000003316','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003304','3','确认结果',0,'',0,'','','','','1','button.hrm.performance.assessment.confirmResult',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:confirmResult","desc":"结果确认"}]',0,'v1:hrm:perf_assessment:confirmResult'),
    ('aaa10000-1111-4000-8000-000000003317','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003304','3','发起申诉',0,'',1,'','','','','1','button.hrm.performance.assessment.appeal',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:appeal","desc":"发起申诉"}]',0,'v1:hrm:perf_assessment:appeal'),
    ('aaa10000-1111-4000-8000-000000003318','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003304','3','申诉裁决',0,'',2,'','','','','1','button.hrm.performance.assessment.handleAppeal',0,'',0,0,'[]','[{"code":"v1:hrm:perf_assessment:handleAppeal","desc":"申诉裁决"}]',0,'v1:hrm:perf_assessment:handleAppeal');

-- 超级管理员角色可见
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003300','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003300',0),
    ('aaa20000-1111-4000-8000-000000003301','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003301',0),
    ('aaa20000-1111-4000-8000-000000003302','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003302',0),
    ('aaa20000-1111-4000-8000-000000003303','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003303',0),
    ('aaa20000-1111-4000-8000-000000003304','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003304',0),
    ('aaa20000-1111-4000-8000-000000003311','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003311',0),
    ('aaa20000-1111-4000-8000-000000003312','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003312',0),
    ('aaa20000-1111-4000-8000-000000003313','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003313',0),
    ('aaa20000-1111-4000-8000-000000003314','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003314',0),
    ('aaa20000-1111-4000-8000-000000003315','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003315',0),
    ('aaa20000-1111-4000-8000-000000003316','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003316',0),
    ('aaa20000-1111-4000-8000-000000003317','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003317',0),
    ('aaa20000-1111-4000-8000-000000003318','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003318',0);
