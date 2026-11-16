package perf_result

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	perf_assessment "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_assessment"
	perf_period "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period_participant"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result_approval"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// DB-backed logic tests for 批次结果审批与结果发布/签收 (HRM-45, issue #47), run via
// svctest: sqlite by default; HRM_TEST_DB=mysql runs the same assertions against
// docker MySQL.
//
// Coverage (issue #47 acceptance criteria):
//   - HR211 审批单聚合断言：人数 / 平均得分 / 绩效金额合计 / 公司部门快照
//   - 审批三联动（审批通过事务内）：结果行生成 + 批次 AdvancePeriod '4'→'5' + published=1
//   - 签收：员工本人签收 signed=1+sign_time；幂等（重复签收不报错不重复记录）
//   - 守卫：批次状态不符（非 '4'）拒绝提交；审批中/审批通过重复提交拒绝；
//     签收越权（非本人）拒绝；未发布/已作废不可签收

var testCtx = context.Background()

var envCounter int

func dial(t *testing.T) int {
	envCounter++
	return envCounter
}

// ---- tiny assertion helpers (repo has no testify; keep tests stdlib-only) ----

func mustNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertEq(t *testing.T, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertLen(t *testing.T, got any, want int) {
	t.Helper()
	n := reflect.ValueOf(got).Len()
	if n != want {
		t.Errorf("len got %v, want %v", n, want)
	}
}

func assertErrIs(t *testing.T, err, target error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error %v, got nil", target)
		return
	}
	if !containsErr(err, target) {
		t.Errorf("expected error %v, got %v", target, err)
	}
}

func containsErr(err, target error) bool {
	for e := err; e != nil; {
		if e == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := e.(unwrapper)
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}

func assertNotEmpty(t *testing.T, v string) {
	t.Helper()
	if v == "" {
		t.Error("expected non-empty string")
	}
}

func assertInDelta(t *testing.T, got, want, delta float64) {
	t.Helper()
	if got-want > delta || want-got > delta {
		t.Errorf("got %v, want %v (delta %v)", got, want, delta)
	}
}

func newSvc(t *testing.T) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}, env
}

// ---- seed helpers ----

// seedOrg inserts one org (公司/部门).
func seedOrg(t *testing.T, env *svctest.Env, name, orgType, parentUuid string) string {
	t.Helper()
	uuid := fmt.Sprintf("org-%d", dial(t))
	mustNoErr(t, hrm_org.NewHrmOrgModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_org.HrmOrg{
		Uuid: uuid, ParentUuid: parentUuid, OrgName: name, OrgType: orgType, Status: "1",
	}))
	return uuid
}

// seedEmployee 在册员工 bound to a company/org.
func seedEmployee(t *testing.T, env *svctest.Env, companyUuid, orgUuid, userUuid string) string {
	t.Helper()
	uuid := fmt.Sprintf("emp-%d", dial(t))
	mustNoErr(t, hrm_employee.NewHrmEmployeeModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_employee.HrmEmployee{
		Uuid: uuid, EmployeeNo: fmt.Sprintf("E%04d", dial(t)),
		Name: "员工" + fmt.Sprint(dial(t)), Gender: "1", Birthday: "1990-01-01",
		Nation: "汉", PoliticalStatus: "1", MaritalStatus: "1", IdCard: "110101199001010011",
		CompanyUuid: companyUuid, OrgUuid: orgUuid, PersonStatus: "1",
		HireDate: "2024-01-01", UserUuid: userUuid,
	}))
	return uuid
}

// seedScheme 方案 with score/pay link mode.
func seedScheme(t *testing.T, env *svctest.Env, scoreMode, payLinkMode string) string {
	t.Helper()
	uuid := fmt.Sprintf("scheme-%d", dial(t))
	mustNoErr(t, hrm_perf_scheme.NewHrmPerfSchemeModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_perf_scheme.HrmPerfScheme{
		Uuid: uuid, Name: "方案", TemplateUuid: fmt.Sprintf("tpl-%d", dial(t)),
		PeriodType: "month", AssessmentMode: "kpi",
		ScoreMode: scoreMode, PayLinkMode: payLinkMode, Status: "1",
	}))
	return uuid
}

// seedPeriodAtStatus 批次 driven to the given lifecycle status (参与人已圈定).
func seedPeriodAtStatus(t *testing.T, env *svctest.Env, schemeUuid, status string, employeeUuids ...string) string {
	t.Helper()
	periodUuid := fmt.Sprintf("period-%d", dial(t))
	scope := `{"companyUuids":[],"orgUuids":[],"positions":[],"posts":[],"personStatuses":[]}`
	mustNoErr(t, hrm_perf_period.NewHrmPerfPeriodModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_perf_period.HrmPerfPeriod{
		Uuid: periodUuid, PeriodNo: fmt.Sprintf("PERF-2026-%04d", dial(t)),
		Name: "批次" + fmt.Sprint(dial(t)), PeriodType: "month", SchemeUuid: schemeUuid,
		Scope: scope, PeriodStart: "2026-01-01", PeriodEnd: "2026-01-31",
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

// seedDoneAssessment inserts one 已完成考核任务 row with final score/grade
// (HRM-45 消费口径：final_score/final_grade 已校准/回写).
func seedDoneAssessment(t *testing.T, env *svctest.Env, periodUuid, employeeUuid string, score float64, grade string) string {
	t.Helper()
	uuid := fmt.Sprintf("assm-%d", dial(t))
	mustNoErr(t, hrm_perf_assessment.NewHrmPerfAssessmentModel(env.SqlConn).InsertV2(testCtx, nil,
		&hrm_perf_assessment.HrmPerfAssessment{
			Uuid: uuid, AssessmentNo: fmt.Sprintf("PA-2026-%04d", dial(t)),
			PeriodUuid: periodUuid, EmployeeUuid: employeeUuid,
			TemplateUuid: fmt.Sprintf("tpl-%d", dial(t)), ItemsSnapshot: "[]",
			CurrentNode: perf_assessment.NodeResultConfirm, TaskStatus: perf_assessment.TaskStatusDone,
			FinalScore: score, FinalGrade: grade,
		}))
	return uuid
}

// mustSubmitApproval runs SubmitPerfResultApproval and requires success.
func mustSubmitApproval(t *testing.T, svcCtx *svc.ServiceContext, periodUuid string) string {
	t.Helper()
	uuid, err := SubmitPerfResultApproval(testCtx, svcCtx.SqlxConn, svcCtx, periodUuid, "")
	mustNoErr(t, err)
	return uuid
}

// mustApprove runs the shared base approve (审批通过 → OnApproved 事务内三联动).
func mustApprove(t *testing.T, svcCtx *svc.ServiceContext, approvalUuid string) {
	t.Helper()
	mustNoErr(t, bill.ApproveBill(testCtx, svcCtx.SqlxConn, BillTypeCode, approvalUuid,
		bill.Operator{UserUuid: "approver-1", Name: "李审批"}, "同意"))
}

func resultsOf(t *testing.T, env *svctest.Env, periodUuid string) []*hrm_perf_result.HrmPerfResult {
	t.Helper()
	rows, err := hrm_perf_result.NewHrmPerfResultModel(env.SqlConn).FindByCondition(testCtx, nil,
		condition.NewChain().Equal(hrm_perf_result.PeriodUuid, periodUuid).
			OrderBy(hrm_perf_result.Id).Build()...)
	mustNoErr(t, err)
	return rows
}

func decodeNames(t *testing.T, raw string) []string {
	t.Helper()
	var out []string
	mustNoErr(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// ---- 1. 系数映射（HRM-43 契约口径） ----

func TestGradeCoefficient(t *testing.T) {
	assertInDelta(t, GradeCoefficient("A"), 1.2, 1e-9)
	assertInDelta(t, GradeCoefficient("B"), 1.0, 1e-9)
	assertInDelta(t, GradeCoefficient("C"), 0.8, 1e-9)
	assertInDelta(t, GradeCoefficient("D"), 0.6, 1e-9)
	assertInDelta(t, GradeCoefficient("S"), 1.0, 1e-9) // 等级制原值 → 默认 1.0
	assertInDelta(t, GradeCoefficient(""), 1.0, 1e-9)
}

// ---- 2. 提交聚合 + 守卫 ----

func TestSubmitApprovalAggregateAndGuards(t *testing.T) {
	svcCtx, env := newSvc(t)
	company := seedOrg(t, env, "总公司", "1", "")
	org1 := seedOrg(t, env, "研发部", "2", company)
	org2 := seedOrg(t, env, "市场部", "2", company)
	emp1 := seedEmployee(t, env, company, org1, "")
	emp2 := seedEmployee(t, env, company, org2, "")

	// 金额制方案：绩效金额 = final_score 直取
	scheme := seedScheme(t, env, "amount", "amount")
	period := seedPeriodAtStatus(t, env, scheme, perf_period.PeriodStatusOngoing, emp1, emp2)

	// 守卫1：批次状态不符（'1'进行中 ≠ '4'结果已发布）→ 拒绝提交
	_, err := SubmitPerfResultApproval(testCtx, svcCtx.SqlxConn, svcCtx, period, "")
	assertErrIs(t, err, ErrResultApprovalInvalid)

	// 推进到 '4'结果已发布（HRM-43 联动后的待审批口径）
	mustNoErr(t, hrm_perf_period.NewHrmPerfPeriodModel(env.SqlConn).UpdateFieldsByCondition(testCtx, nil,
		map[string]any{"status": perf_period.PeriodStatusResultPublished},
		condition.NewChain().Equal(hrm_perf_period.Uuid, period).Build()...))

	// 守卫2：无考核任务 → 拒绝
	_, err = SubmitPerfResultApproval(testCtx, svcCtx.SqlxConn, svcCtx, period, "")
	assertErrIs(t, err, ErrResultApprovalInvalid)

	seedDoneAssessment(t, env, period, emp1, 92.5, "A")
	seedDoneAssessment(t, env, period, emp2, 78, "B")

	// 提交成功：HR211 单号 + 聚合快照
	approvalUuid := mustSubmitApproval(t, svcCtx, period)
	row, err := hrm_perf_result_approval.NewHrmPerfResultApprovalModel(env.SqlConn).FindOneByUuid(testCtx, nil, approvalUuid)
	mustNoErr(t, err)
	assertContainsPrefix(t, row.BillNo, "HR211-")
	assertEq(t, int64(2), row.ParticipantCount)
	assertInDelta(t, row.AvgScore, 85.25, 1e-9)         // (92.5+78)/2
	assertInDelta(t, row.TotalAmount, 170.5, 1e-9)      // 金额制直取 92.5+78
	assertEq(t, bill.StatusApproving, row.Status)
	assertLen(t, decodeNames(t, row.CompanySnapshot), 1) // 总公司
	assertLen(t, decodeNames(t, row.OrgSnapshot), 2)     // 研发部/市场部
	if decodeNames(t, row.CompanySnapshot)[0] != "总公司" {
		t.Errorf("company snapshot got %q", row.CompanySnapshot)
	}

	// 守卫3：审批中重复提交 → 拒绝
	_, err = SubmitPerfResultApproval(testCtx, svcCtx.SqlxConn, svcCtx, period, "")
	assertErrIs(t, err, ErrResultApprovalInvalid)

	// 审批驳回 → 可重新提交（快照重新聚合，单据状态回审批中）
	mustNoErr(t, bill.RejectBill(testCtx, svcCtx.SqlxConn, BillTypeCode, approvalUuid,
		bill.Operator{UserUuid: "approver-1", Name: "李审批"}, "材料不全"))
	approvalUuid2 := mustSubmitApproval(t, svcCtx, period)
	assertEq(t, approvalUuid, approvalUuid2) // 一批次一张单
	row, err = hrm_perf_result_approval.NewHrmPerfResultApprovalModel(env.SqlConn).FindOneByUuid(testCtx, nil, approvalUuid)
	mustNoErr(t, err)
	assertEq(t, bill.StatusApproving, row.Status)
}

func assertContainsPrefix(t *testing.T, s, prefix string) {
	t.Helper()
	if len(s) < len(prefix) || s[:len(prefix)] != prefix {
		t.Errorf("expected %q to start with %q", s, prefix)
	}
}

// ---- 3. 审批三联动（结果行 + 批次 advance + published） ----

func TestApprovalTripleWriteback(t *testing.T) {
	svcCtx, env := newSvc(t)
	company := seedOrg(t, env, "甲公司", "1", "")
	org := seedOrg(t, env, "研发部", "2", company)
	emp1 := seedEmployee(t, env, company, org, "")
	emp2 := seedEmployee(t, env, company, org, "")

	// 金额制方案
	scheme := seedScheme(t, env, "amount", "amount")
	period := seedPeriodAtStatus(t, env, scheme, perf_period.PeriodStatusResultPublished, emp1, emp2)
	seedDoneAssessment(t, env, period, emp1, 92.5, "A")
	seedDoneAssessment(t, env, period, emp2, 58, "D")

	approvalUuid := mustSubmitApproval(t, svcCtx, period)

	// 审批通过：事务内三联动
	mustApprove(t, svcCtx, approvalUuid)

	// 联动1：结果行生成（每参与人一行：得分/等级/系数/绩效金额/结果状态/published）
	rows := resultsOf(t, env, period)
	assertLen(t, rows, 2)
	assertInDelta(t, rows[0].Score, 92.5, 1e-9)
	assertEq(t, "A", rows[0].Grade)
	assertInDelta(t, rows[0].Coefficient, 1.2, 1e-9)
	assertInDelta(t, rows[0].PerfAmount, 92.5, 1e-9) // 金额制直取
	assertEq(t, ResultStatusNormal, rows[0].ResultStatus)
	assertEq(t, int64(1), rows[0].Published)
	assertEq(t, int64(0), rows[0].Signed)      // 未签收
	assertEq(t, int64(0), rows[0].SalarySynced) // HRM-49 写入
	assertEq(t, "甲公司", rows[0].CompanyName)
	assertEq(t, "研发部", rows[0].OrgName)
	assertNotEmpty(t, rows[0].EmployeeName)
	assertInDelta(t, rows[1].Coefficient, 0.6, 1e-9) // D → 0.6

	// 联动2：批次 AdvancePeriod '4'→'5'已发布
	p, err := hrm_perf_period.NewHrmPerfPeriodModel(env.SqlConn).FindOneByUuid(testCtx, nil, period)
	mustNoErr(t, err)
	assertEq(t, perf_period.PeriodStatusPublished, p.Status)

	// 联动3：审批单回写通过信息（审批人姓名由底座审批记录/审批快照承担；
	// 审批单行 approve_by 取 JWT 操作人 — 逻辑测试无登录态为空）
	apRow, err := hrm_perf_result_approval.NewHrmPerfResultApprovalModel(env.SqlConn).FindOneByUuid(testCtx, nil, approvalUuid)
	mustNoErr(t, err)
	assertEq(t, bill.StatusApproved, apRow.Status)
	assertNotEmpty(t, apRow.ApproveTime)

	// 幂等：重复 ApproveBill（状态机拒绝，不重复插结果行）
	err = bill.ApproveBill(testCtx, svcCtx.SqlxConn, BillTypeCode, approvalUuid,
		bill.Operator{UserUuid: "approver-1", Name: "李审批"}, "")
	assertErrIs(t, err, bill.ErrInvalidTransition)
	assertLen(t, resultsOf(t, env, period), 2)
}

// ---- 4. 签收（本人 + 幂等 + 守卫） ----

func TestSignResultFlow(t *testing.T) {
	svcCtx, env := newSvc(t)
	company := seedOrg(t, env, "乙公司", "1", "")
	org := seedOrg(t, env, "研发部", "2", company)
	emp1 := seedEmployee(t, env, company, org, "user-1")
	emp2 := seedEmployee(t, env, company, org, "user-2")

	scheme := seedScheme(t, env, "percent", "coefficient")
	period := seedPeriodAtStatus(t, env, scheme, perf_period.PeriodStatusResultPublished, emp1, emp2)
	seedDoneAssessment(t, env, period, emp1, 86, "B") // 百分制系数制
	seedDoneAssessment(t, env, period, emp2, 95, "A")

	approvalUuid := mustSubmitApproval(t, svcCtx, period)
	mustApprove(t, svcCtx, approvalUuid)

	rows := resultsOf(t, env, period)
	r1 := rows[0]
	// 系数制绩效金额落 0（基数 HRM-49 联薪回填），系数已映射
	assertInDelta(t, r1.PerfAmount, 0, 1e-9)
	assertInDelta(t, r1.Coefficient, 1.0, 1e-9)

	// 越权守卫：非本人（emp2 的员工 uuid 签 emp1 的结果行）→ 拒绝
	err := SignResultRow(testCtx, env.SqlConn, r1.Uuid, emp2, "其他人")
	assertErrIs(t, err, ErrResultSignDenied)

	// 空操作人（未绑定员工）拒绝
	err = SignResultRow(testCtx, env.SqlConn, r1.Uuid, "", "")
	assertErrIs(t, err, ErrResultSignDenied)

	// 本人签收：signed=1 + sign_time + sign_by
	mustNoErr(t, SignResultRow(testCtx, env.SqlConn, r1.Uuid, emp1, "员工本人"))
	row, err := hrm_perf_result.NewHrmPerfResultModel(env.SqlConn).FindOneByUuid(testCtx, nil, r1.Uuid)
	mustNoErr(t, err)
	assertEq(t, int64(1), row.Signed)
	assertNotEmpty(t, row.SignTime)
	assertEq(t, "员工本人", row.SignBy)

	// 幂等：重复签收不报错、不覆盖签收时间
	mustNoErr(t, SignResultRow(testCtx, env.SqlConn, r1.Uuid, emp1, "员工本人"))
	row2, err := hrm_perf_result.NewHrmPerfResultModel(env.SqlConn).FindOneByUuid(testCtx, nil, r1.Uuid)
	mustNoErr(t, err)
	assertEq(t, row2.SignTime, row.SignTime)
	assertEq(t, "员工本人", row2.SignBy)

	// 结果行不存在 → ErrResultNotFound
	err = SignResultRow(testCtx, env.SqlConn, "no-such-uuid", emp1, "员工本人")
	assertErrIs(t, err, ErrResultNotFound)

	// 结果来源契约：GetPeriodParticipants 返回 2 个参与人（HRM-42 契约复跑）
	participants, err := perf_period.GetPeriodParticipants(testCtx, svcCtx, period)
	mustNoErr(t, err)
	assertLen(t, participants, 2)
}
