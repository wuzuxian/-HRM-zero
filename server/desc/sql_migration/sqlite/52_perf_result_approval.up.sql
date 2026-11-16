-- HRM-45 批次结果审批与结果发布/签收（演示站 9.3 结果，GLOSSARY「绩效结果」）：
--   批次到达 '4'结果已发布（HRM-43 任务链联动后）→ 提交 HR211 批次结果审批单
--   （单号 HR211-{YYYYMMDD}{5位流水}，复用 hrm_bill_sequence，bill_type_code='211'，
--   走 HRM-06 单据公共组件单步审批）。提交时从 hrm_perf_assessment 聚合快照：
--   人数 / 平均得分 / 绩效金额合计 + 公司/部门名称快照。
--   审批通过事务内（OnApproved）：生成 hrm_perf_result 结果行（每参与人一行，
--   UNIQUE(period_uuid,employee_uuid)）+ 批次 AdvancePeriod '4'→'5'已发布 + 结果行
--   published=1。系数映射（HRM-43 契约口径）：A=1.2 B=1.0 C=0.8 D=0.6 其他=1.0；
--   金额制绩效金额=final_score 直取，系数制绩效金额=系数×绩效工资基数（基数 HRM-49
--   联薪时回填，本表先落 0）。签收：员工本人对结果行 signed=1 + sign_time，幂等。
--   salary_synced（已同步薪酬）列由 HRM-49 联薪写入，本票只建列不写。
-- 错误码 10690-10694（server/internal/errcodes/hrm_perf_result.go，1069x 段 grep 空闲）。
-- 菜单 uuid 段 38xx（grep 全部 migration 确认 00xx-37xx 已用、38xx 空闲）：
--   3801 批次结果审批 page / 3802 绩效结果 page（绩效管理 dir，migration 12），
--   3811-3815 审批按钮（v1:hrm:bill:* 共享审批码）+ 3821 签收按钮；role_menu aaa20000 38xx。

DROP TABLE IF EXISTS hrm_perf_result;
DROP TABLE IF EXISTS hrm_perf_result_approval;

CREATE TABLE hrm_perf_result_approval (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bill_no TEXT NOT NULL UNIQUE,
    period_uuid TEXT NOT NULL,
    period_no TEXT NOT NULL DEFAULT '',
    period_name TEXT NOT NULL DEFAULT '',
    company_snapshot TEXT NOT NULL DEFAULT '[]',
    org_snapshot TEXT NOT NULL DEFAULT '[]',
    participant_count INTEGER NOT NULL DEFAULT 0,
    avg_score REAL NOT NULL DEFAULT 0,
    total_amount REAL NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'unsubmitted',
    submit_by TEXT NOT NULL DEFAULT '',
    submit_time TEXT NOT NULL DEFAULT '',
    approve_by TEXT NOT NULL DEFAULT '',
    approve_time TEXT NOT NULL DEFAULT '',
    opinion TEXT NOT NULL DEFAULT ''
);

CREATE TABLE hrm_perf_result (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    approval_uuid TEXT NOT NULL,
    assessment_uuid TEXT NOT NULL,
    period_uuid TEXT NOT NULL,
    period_no TEXT NOT NULL DEFAULT '',
    employee_uuid TEXT NOT NULL,
    employee_name TEXT NOT NULL DEFAULT '',
    company_uuid TEXT NOT NULL DEFAULT '',
    company_name TEXT NOT NULL DEFAULT '',
    org_uuid TEXT NOT NULL DEFAULT '',
    org_name TEXT NOT NULL DEFAULT '',
    score REAL NOT NULL DEFAULT 0,
    grade TEXT NOT NULL DEFAULT '',
    coefficient REAL NOT NULL DEFAULT 1.0,
    perf_amount REAL NOT NULL DEFAULT 0,
    bonus_amount REAL NOT NULL DEFAULT 0,
    result_status TEXT NOT NULL DEFAULT '1',
    published INTEGER NOT NULL DEFAULT 0,
    signed INTEGER NOT NULL DEFAULT 0,
    sign_time TEXT NOT NULL DEFAULT '',
    sign_by TEXT NOT NULL DEFAULT '',
    salary_synced INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX uniq_hrm_perf_result_approval_period ON hrm_perf_result_approval (period_uuid);
CREATE UNIQUE INDEX uniq_hrm_perf_result_period_employee ON hrm_perf_result (period_uuid, employee_uuid);
CREATE INDEX idx_hrm_perf_result_approval ON hrm_perf_result (approval_uuid);
CREATE INDEX idx_hrm_perf_result_assessment ON hrm_perf_result (assessment_uuid);
CREATE INDEX idx_hrm_perf_result_employee ON hrm_perf_result (employee_uuid);
CREATE INDEX idx_hrm_perf_result_period ON hrm_perf_result (period_uuid);

CREATE TRIGGER hrm_perf_result_approval_update_time
AFTER UPDATE ON hrm_perf_result_approval
FOR EACH ROW
BEGIN
    UPDATE hrm_perf_result_approval SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER hrm_perf_result_update_time
AFTER UPDATE ON hrm_perf_result
FOR EACH ROW
BEGIN
    UPDATE hrm_perf_result SET update_time = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- 绩效管理目录（migration 12 uuid f0e1d2c3-...4e5f）下新增两个结果页
-- （demo-inventory 9.3 路径：/hrm/performance/result-confirm-list + /hrm/performance/result-list）
INSERT INTO manage_menu (uuid, create_time, update_time, status, parent_uuid, menu_type, menu_name, hide_in_menu,
    active_menu, "order", route_name, route_path, component, icon, icon_type, i18n_key, keep_alive, href,
    multi_tab, fixed_index_in_tab, query, permissions, constant, button_code)
VALUES
    ('aaa10000-1111-4000-8000-000000003801','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','批次结果审批',0,'',5,'hrm_performance_result-confirm-list','/hrm/performance/result-confirm-list','view.hrm_performance_result-confirm-list','mdi:clipboard-text-check','1','route.hrm_performance_result-confirm-list',0,'',0,0,'[]','[{"code":"v1:hrm:perf_result:approvalList","desc":"批次结果审批列表"},{"code":"v1:hrm:perf_result:submit","desc":"提交结果审批"},{"code":"v1:hrm:bill:getApprovalRecordList","desc":"审批记录"}]',0,''),
    ('aaa10000-1111-4000-8000-000000003802','2026-11-16 00:00:00','2026-11-16 00:00:00','1','f0e1d2c3-a4b5-4c6d-8e7f-0a1b2c3d4e5f','2','绩效结果',0,'',6,'hrm_performance_result-list','/hrm/performance/result-list','view.hrm_performance_result-list','mdi:chart-bar','1','route.hrm_performance_result-list',0,'',0,0,'[]','[{"code":"v1:hrm:perf_result:list","desc":"绩效结果列表"}]',0,''),
    ('aaa10000-1111-4000-8000-000000003811','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003801','3','提交单据',0,'',1,'','','','','1','button.hrm.bill.submit',0,'',0,0,'[]','[{"code":"v1:hrm:bill:submit","desc":"提交单据"}]',0,'v1:hrm:bill:submit'),
    ('aaa10000-1111-4000-8000-000000003812','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003801','3','审批通过',0,'',2,'','','','','1','button.hrm.bill.approve',0,'',0,0,'[]','[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]',0,'v1:hrm:bill:approve'),
    ('aaa10000-1111-4000-8000-000000003813','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003801','3','审批驳回',0,'',3,'','','','','1','button.hrm.bill.reject',0,'',0,0,'[]','[{"code":"v1:hrm:bill:reject","desc":"审批驳回"}]',0,'v1:hrm:bill:reject'),
    ('aaa10000-1111-4000-8000-000000003814','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003801','3','撤销单据',0,'',4,'','','','','1','button.hrm.bill.revoke',0,'',0,0,'[]','[{"code":"v1:hrm:bill:revoke","desc":"撤销单据"}]',0,'v1:hrm:bill:revoke'),
    ('aaa10000-1111-4000-8000-000000003821','2026-11-16 00:00:00','2026-11-16 00:00:00','1','aaa10000-1111-4000-8000-000000003802','3','签收结果',0,'',0,'','','','','1','button.hrm.performance.result.sign',0,'',0,0,'[]','[{"code":"v1:hrm:perf_result:sign","desc":"签收结果"}]',0,'v1:hrm:perf_result:sign');

-- 超级管理员角色可见
INSERT INTO manage_role_menu (uuid, create_time, update_time, role_uuid, menu_uuid, is_home)
VALUES
    ('aaa20000-1111-4000-8000-000000003801','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003801',0),
    ('aaa20000-1111-4000-8000-000000003802','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003802',0),
    ('aaa20000-1111-4000-8000-000000003811','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003811',0),
    ('aaa20000-1111-4000-8000-000000003812','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003812',0),
    ('aaa20000-1111-4000-8000-000000003813','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003813',0),
    ('aaa20000-1111-4000-8000-000000003814','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003814',0),
    ('aaa20000-1111-4000-8000-000000003821','2026-11-16 00:00:00','2026-11-16 00:00:00','1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d','aaa10000-1111-4000-8000-000000003821',0);
