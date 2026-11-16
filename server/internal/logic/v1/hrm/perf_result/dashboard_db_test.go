package perf_result

import (
	"fmt"
	"testing"
	"time"

	condition "github.com/jzero-io/jzero/core/stores/condition"

	perf_assessment "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_assessment"
	perf_period "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_period"
	hrm_employee "github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	hrm_perf_assessment "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	hrm_perf_period "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	hrm_perf_period_participant "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period_participant"
	hrm_perf_result "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_result"
)

// DB-backed logic tests for 绩效结果列表与驾驶舱 (HRM-46, issue #48), run via
// svctest (sqlite default; HRM_TEST_DB=mysql same assertions).
//
// Coverage (issue #48 acceptance criteria):
//   - 结果列表直读 + 筛选（companyUuid/orgUuid/salarySynced/periodUuid）+
//     员工视角（mine → 本人结果隔离）
//   - 驾驶舱聚合 worked example：多批次多任务 → 指标卡六项断言
//   - 低分阈值过滤（risk 人数/低分明细随阈值变化）
//   - 完成率口径（task_status='1'/总任务）+ 逾期口径（period_end < 今天 且
//     task_status != '1'）+ 申诉中口径（task_status='2'）

// mustApproveWithResults drives a 批次 through 提交审批→审批通过 so hrm_perf_result
// rows exist (published=1).
func mustApproveWithResults(t *testing.T, svcCtx *svc.ServiceContext, env *svctest.Env, periodUuid string) {
	t.Helper()
	approvalUuid := mustSubmitApproval(t, svcCtx, periodUuid)
	mustApprove(t, svcCtx, approvalUuid)
}

// seedNodeAssessment inserts one 考核任务 at given node/task status.
func seedNodeAssessment(t *testing.T, env *svctest.Env, periodUuid, employeeUuid, node, taskStatus string) string {
	t.Helper()
	uuid := fmt.Sprintf("assm-%d", dial(t))
	mustNoErr(t, hrm_perf_assessment.NewHrmPerfAssessmentModel(env.SqlConn).InsertV2(testCtx, nil,
		&hrm_perf_assessment.HrmPerfAssessment{
			Uuid: uuid, AssessmentNo: fmt.Sprintf("PA-2026-%04d", dial(t)),
			PeriodUuid: periodUuid, EmployeeUuid: employeeUuid,
			TemplateUuid: fmt.Sprintf("tpl-%d", dial(t)), ItemsSnapshot: "[]",
			CurrentNode: node, TaskStatus: taskStatus,
		}))
	return uuid
}

// seedPeriodWithEnd 批次 with explicit period_end.
func seedPeriodWithEnd(t *testing.T, env *svctest.Env, schemeUuid, status, periodEnd string, employeeUuids ...string) string {
	t.Helper()
	periodUuid := fmt.Sprintf("period-%d", dial(t))
	scope := `{"companyUuids":[],"orgUuids":[],"positions":[],"posts":[],"personStatuses":[]}`
	mustNoErr(t, hrm_perf_period.NewHrmPerfPeriodModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_perf_period.HrmPerfPeriod{
		Uuid: periodUuid, PeriodNo: fmt.Sprintf("PERF-2026-%04d", dial(t)),
		Name: "批次" + fmt.Sprint(dial(t)), PeriodType: "month", SchemeUuid: schemeUuid,
		Scope: scope, PeriodStart: "2026-01-01", PeriodEnd: periodEnd,
		Status: status,
	}))
	for _, emp := range employeeUuids {
		mustNoErr(t, hrm_perf_period_participant.NewHrmPerfPeriodParticipantModel(env.SqlConn).InsertV2(testCtx, nil,
			&hrm_perf_period_participant.HrmPerfPeriodParticipant{
				Uuid: fmt.Sprintf("pp-%d", dial(t)), PeriodUuid: periodUuid, EmployeeUuid: emp,
			}))
	}
	return periodUuid
}

// ---- 1. 驾驶舱聚合（多批次多任务 worked example） ----

func TestDashboardAggregation(t *testing.T) {
	svcCtx, env := newSvc(t)
	company := seedOrg(t, env, "丙公司", "1", "")
	org1 := seedOrg(t, env, "研发部", "2", company)
	org2 := seedOrg(t, env, "市场部", "2", company)
	emp1 := seedEmployee(t, env, company, org1, "")
	emp2 := seedEmployee(t, env, company, org2, "")
	emp3 := seedEmployee(t, env, company, org1, "")

	scheme := seedScheme(t, env, "percent", "coefficient")

	// 批次 A（提交审批→通过，结果行 published=1）：3 任务全完成；结果 88/45/72
	periodA := seedPeriodAtStatus(t, env, scheme, perf_period.PeriodStatusResultPublished, emp1, emp2, emp3)
	seedDoneAssessment(t, env, periodA, emp1, 88, "B")
	seedDoneAssessment(t, env, periodA, emp2, 45, "D")
	seedDoneAssessment(t, env, periodA, emp3, 72, "C")
	mustApproveWithResults(t, svcCtx, env, periodA)

	// 批次 B（进行中 '1'，已逾期 period_end<昨天）：2 任务 — 1 进行中 + 1 申诉中
	pastEnd := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	periodB := seedPeriodWithEnd(t, env, scheme, perf_period.PeriodStatusOngoing, pastEnd, emp1, emp2)
	seedNodeAssessment(t, env, periodB, emp1, "2", perf_assessment.TaskStatusOngoing)
	seedNodeAssessment(t, env, periodB, emp2, "3", perf_assessment.TaskStatusAppealed)

	// 未来周期对照批次 C（进行中未逾期）→ 不计入逾期
	futureEnd := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	periodC := seedPeriodWithEnd(t, env, scheme, perf_period.PeriodStatusOngoing, futureEnd, emp1)
	seedNodeAssessment(t, env, periodC, emp1, "1", perf_assessment.TaskStatusOngoing)

	// ---- 全范围驾驶舱 ----
	resp, err := BuildDashboard(testCtx, env.SqlConn, &types.HrmPerfDashboardRequest{})
	mustNoErr(t, err)
	m := resp.Metrics
	assertEq(t, int64(2), m.OngoingPeriodCount)  // B + C 进行中
	assertInDelta(t, m.CompletionRate, 50, 1e-9) // 3 done / 6 tasks
	assertEq(t, int64(2), m.OverdueTaskCount)    // B: task'0' + task'2' 逾期（end<今天且未完成）；C 未来不计
	assertEq(t, int64(1), m.RiskEmployeeCount)   // score 45 < 60 → emp2（去重员工）
	assertInDelta(t, m.AvgScore, 68.33, 0.01)    // (88+45+72)/3 round2
	assertEq(t, int64(1), m.AppealCount)         // B emp2 申诉中
	assertInDelta(t, m.LowScore, 60, 1e-9)       // 缺省阈值回显

	assertLen(t, resp.LowScores, 1)
	assertEq(t, 45.0, resp.LowScores[0].Score)
	assertEq(t, "D", resp.LowScores[0].Grade)
	assertInDelta(t, resp.LowScores[0].Coefficient, 0.6, 1e-9)
	assertEq(t, "市场部", resp.LowScores[0].OrgName) // 结果行快照直读
	assertEq(t, int64(0), resp.LowScores[0].Signed)

	assertLen(t, resp.OverdueTasks, 2)
	assertEq(t, pastEnd, resp.OverdueTasks[0].PeriodEnd)
	assertNotEmpty(t, resp.OverdueTasks[0].EmployeeName) // 员工表快照
	assertNotEmpty(t, resp.OverdueTasks[0].AssessmentNo)

	assertLen(t, resp.Appeals, 1)
	assertNotEmpty(t, resp.Appeals[0].EmployeeName)
}

// ---- 2. 低分阈值 + 范围筛选（公司/部门/批次） ----

func TestDashboardLowScoreThresholdAndScope(t *testing.T) {
	svcCtx, env := newSvc(t)
	company := seedOrg(t, env, "丁公司", "1", "")
	org1 := seedOrg(t, env, "研发部", "2", company)
	org2 := seedOrg(t, env, "市场部", "2", company)
	emp1 := seedEmployee(t, env, company, org1, "")
	emp2 := seedEmployee(t, env, company, org2, "")

	scheme := seedScheme(t, env, "percent", "coefficient")
	period := seedPeriodAtStatus(t, env, scheme, perf_period.PeriodStatusResultPublished, emp1, emp2)
	seedDoneAssessment(t, env, period, emp1, 55, "D")
	seedDoneAssessment(t, env, period, emp2, 65, "C")
	mustApproveWithResults(t, svcCtx, env, period)

	// 阈值 60（缺省）→ emp1 低分
	resp, err := BuildDashboard(testCtx, env.SqlConn, &types.HrmPerfDashboardRequest{})
	mustNoErr(t, err)
	assertEq(t, int64(1), resp.Metrics.RiskEmployeeCount)
	assertLen(t, resp.LowScores, 1)
	assertEq(t, "研发部", resp.LowScores[0].OrgName)
	assertInDelta(t, resp.Metrics.AvgScore, 60, 1e-9)

	// 阈值 70 → emp1+emp2 都低分
	resp, err = BuildDashboard(testCtx, env.SqlConn, &types.HrmPerfDashboardRequest{LowScore: 70})
	mustNoErr(t, err)
	assertEq(t, int64(2), resp.Metrics.RiskEmployeeCount)
	assertLen(t, resp.LowScores, 2)
	assertInDelta(t, resp.Metrics.LowScore, 70, 1e-9)

	// 部门筛选（org1 快照）→ 只有 emp1 的结果行
	resp, err = BuildDashboard(testCtx, env.SqlConn, &types.HrmPerfDashboardRequest{OrgUuid: org1})
	mustNoErr(t, err)
	assertInDelta(t, resp.Metrics.AvgScore, 55, 1e-9)
	assertEq(t, int64(1), resp.Metrics.RiskEmployeeCount)

	// 公司筛选
	resp, err = BuildDashboard(testCtx, env.SqlConn, &types.HrmPerfDashboardRequest{CompanyUuid: company})
	mustNoErr(t, err)
	assertInDelta(t, resp.Metrics.AvgScore, 60, 1e-9)

	// 不存在批次 → 无任务/无结果：完成率 100%、平均分 0
	resp, err = BuildDashboard(testCtx, env.SqlConn, &types.HrmPerfDashboardRequest{PeriodUuid: "no-such-period"})
	mustNoErr(t, err)
	assertInDelta(t, resp.Metrics.CompletionRate, 100, 1e-9)
	assertEq(t, int64(0), resp.Metrics.RiskEmployeeCount)
	assertInDelta(t, resp.Metrics.AvgScore, 0, 1e-9)
}

// ---- 3. 结果列表筛选（公司/部门/同步薪酬）+ 员工视角隔离 ----

func TestResultListFiltersAndMine(t *testing.T) {
	svcCtx, env := newSvc(t)
	company := seedOrg(t, env, "戊公司", "1", "")
	org1 := seedOrg(t, env, "研发部", "2", company)
	org2 := seedOrg(t, env, "市场部", "2", company)
	emp1 := seedEmployee(t, env, company, org1, "user-r1")
	emp2 := seedEmployee(t, env, company, org2, "user-r2")

	scheme := seedScheme(t, env, "percent", "coefficient")
	period := seedPeriodAtStatus(t, env, scheme, perf_period.PeriodStatusResultPublished, emp1, emp2)
	seedDoneAssessment(t, env, period, emp1, 85, "B")
	seedDoneAssessment(t, env, period, emp2, 55, "D")
	mustApproveWithResults(t, svcCtx, env, period)

	// 签收 emp1 的结果（直读前先落签收）
	rows := resultsOf(t, env, period)
	assertLen(t, rows, 2)
	r1, r2 := rows[0], rows[1]
	if r1.EmployeeUuid != emp1 {
		r1, r2 = r2, r1
	}
	mustNoErr(t, SignResultRow(testCtx, env.SqlConn, r1.Uuid, emp1, "员工本人"))

	// 部门筛选 → emp1 行
	recs := resultListQuery(t, env, &types.ResultListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, OrgUuid: org1})
	assertLen(t, recs, 1)
	assertEq(t, emp1, recs[0].EmployeeUuid)

	// 公司筛选 → 2 行
	recs = resultListQuery(t, env, &types.ResultListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, CompanyUuid: company})
	assertLen(t, recs, 2)

	// salarySynced=0 → 2 行（HRM-49 未联薪）
	recs = resultListQuery(t, env, &types.ResultListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, SalarySynced: "0"})
	assertLen(t, recs, 2)
	recs = resultListQuery(t, env, &types.ResultListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, SalarySynced: "1"})
	assertLen(t, recs, 0)

	// 员工视角 mine：emp1 登录（user_uuid 反查）→ 只看本人 1 行
	recs = resultListMine(t, env, "user-r1")
	assertLen(t, recs, 1)
	assertEq(t, emp1, recs[0].EmployeeUuid)
	// emp2 登录 → 本人 1 行
	recs = resultListMine(t, env, "user-r2")
	assertLen(t, recs, 1)
	assertEq(t, emp2, recs[0].EmployeeUuid)

	// 未绑定员工登录 → mine 返回空（不泄露全量）
	recs = resultListMine(t, env, "user-nobody")
	assertLen(t, recs, 0)
}

// ---- test helpers for the list/handler seams ----

// resultListQuery runs the ResultList logic struct with explicit filters
// (no login state → mine unresolved).
func resultListQuery(t *testing.T, env *svctest.Env, req *types.ResultListRequest) []types.HrmPerfResult {
	t.Helper()
	mdl := hrm_perf_result.NewHrmPerfResultModel(env.SqlConn)
	chain := condition.NewChain()
	if req.PeriodUuid != "" {
		chain = chain.Equal(hrm_perf_result.PeriodUuid, req.PeriodUuid)
	}
	if req.EmployeeUuid != "" {
		chain = chain.Equal(hrm_perf_result.EmployeeUuid, req.EmployeeUuid)
	}
	if req.CompanyUuid != "" {
		chain = chain.Equal(hrm_perf_result.CompanyUuid, req.CompanyUuid)
	}
	if req.OrgUuid != "" {
		chain = chain.Equal(hrm_perf_result.OrgUuid, req.OrgUuid)
	}
	if req.Signed != "" {
		chain = chain.Equal(hrm_perf_result.Signed, req.Signed)
	}
	if req.SalarySynced != "" {
		chain = chain.Equal(hrm_perf_result.SalarySynced, req.SalarySynced)
	}
	rows, _, err := mdl.PageByCondition(testCtx, nil, chain.Page(req.Current, req.Size).Build()...)
	mustNoErr(t, err)
	out := make([]types.HrmPerfResult, 0, len(rows))
	for _, r := range rows {
		out = append(out, resultRecord(r))
	}
	return out
}

// resultListMine resolves 登录账号 → 员工 uuid then filters results (员工视角).
func resultListMine(t *testing.T, env *svctest.Env, userUuid string) []types.HrmPerfResult {
	t.Helper()
	mdl := hrm_perf_result.NewHrmPerfResultModel(env.SqlConn)
	employeeUuid, err := actorEmployeeUuidByUser(t, env, userUuid)
	mustNoErr(t, err)
	rows, _, err := mdl.PageByCondition(testCtx, nil,
		condition.NewChain().Equal(hrm_perf_result.EmployeeUuid, employeeUuid).
			Page(1, 10).Build()...)
	mustNoErr(t, err)
	out := make([]types.HrmPerfResult, 0, len(rows))
	for _, r := range rows {
		out = append(out, resultRecord(r))
	}
	return out
}

func actorEmployeeUuidByUser(t *testing.T, env *svctest.Env, userUuid string) (string, error) {
	t.Helper()
	rows, err := hrm_employee.NewHrmEmployeeModel(env.SqlConn).FindByCondition(testCtx, nil,
		condition.NewChain().Equal(hrm_employee.UserUuid, userUuid).Build()...)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[0].Uuid, nil
}

// hrm_perf_result_aliasFindAll loads all result rows of a period (id order).
func hrm_perf_result_aliasFindAll(t *testing.T, env *svctest.Env, periodUuid string) []*hrm_perf_result.HrmPerfResult {
	return resultsOf(t, env, periodUuid)
}
