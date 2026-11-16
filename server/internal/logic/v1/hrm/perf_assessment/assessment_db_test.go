package perf_assessment

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_period"
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_indicator"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period_participant"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_period"
)

// DB-backed logic tests for 考核任务链 (HRM-43, issue #45), run via svctest:
// sqlite by default; HRM_TEST_DB=mysql runs the same assertions against docker MySQL.
//
// Coverage (issue #45 acceptance criteria):
//   - 任务链状态机（五阶段顺序 + 前置守卫 + 越权填报拒绝）
//   - 权重汇总计分（WeightedTotal / PercentGrade / NormalizeResult 纯函数 + 端到端）
//   - 申诉发起与处理路径（发起→裁决维持/调整→关闭）
//   - 任务生成幂等（重复 Generate 不重复插行）+ 批次状态联动 advance

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
	var n int
	switch v := got.(type) {
	case int64:
		n = int(v)
	case int:
		n = v
	default:
		n = reflect.ValueOf(got).Len()
	}
	if n != want {
		t.Errorf("len got %v, want %v", n, want)
	}
}

func assertInDelta(t *testing.T, got, want, delta float64) {
	t.Helper()
	if got-want > delta || want-got > delta {
		t.Errorf("got %v, want %v (delta %v)", got, want, delta)
	}
}

func assertErrIs(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("expected error %v, got %v", target, err)
	}
}

func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func assertNotEmpty(t *testing.T, v string) {
	t.Helper()
	if v == "" {
		t.Error("expected non-empty string")
	}
}

func assertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("expected %q to contain %q", s, sub)
	}
}

var testCtx = context.Background()

func newSvc(t *testing.T) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}, env
}

// ---- seed helpers ----

var envCounter int

func dial(t *testing.T) int {
	envCounter++
	return envCounter
}

func seedIndicator(t *testing.T, env *svctest.Env, code string) string {
	t.Helper()
	uuid := fmt.Sprintf("ind-%s-%d", code, dial(t))
	mustNoErr(t, env.Model.HrmPerfIndicator.InsertV2(testCtx, nil, &hrm_perf_indicator.HrmPerfIndicator{
		Uuid: uuid, Code: code, Name: "指标" + code, IndicatorType: "kpi",
		ScoreRule: "manual", DataSource: "manual", Status: "1",
	}))
	return uuid
}

// seedTemplate 模板 with items [(ind1,60),(ind2,40)] — weights sum 100.
func seedTemplate(t *testing.T, env *svctest.Env) string {
	t.Helper()
	tplUuid := fmt.Sprintf("tpl-%d", dial(t))
	mustNoErr(t, env.Model.HrmPerfTemplate.InsertV2(testCtx, nil, &hrm_perf_template.HrmPerfTemplate{
		Uuid: tplUuid, Name: "模板", Status: "1",
	}))
	w1, w2 := 60.0, 40.0
	mustNoErr(t, env.Model.HrmPerfTemplateItem.InsertV2(testCtx, nil, &hrm_perf_template_item.HrmPerfTemplateItem{
		Uuid: fmt.Sprintf("titem-%d-1", dial(t)), TemplateUuid: tplUuid,
		IndicatorUuid: seedIndicator(t, env, "A"), Weight: w1,
	}))
	mustNoErr(t, env.Model.HrmPerfTemplateItem.InsertV2(testCtx, nil, &hrm_perf_template_item.HrmPerfTemplateItem{
		Uuid: fmt.Sprintf("titem-%d-2", dial(t)), TemplateUuid: tplUuid,
		IndicatorUuid: seedIndicator(t, env, "B"), Weight: w2,
	}))
	return tplUuid
}

// seedScheme 方案 with score_mode.
func seedScheme(t *testing.T, env *svctest.Env, tplUuid, scoreMode string) string {
	t.Helper()
	uuid := fmt.Sprintf("scheme-%d", dial(t))
	mustNoErr(t, hrm_perf_scheme.NewHrmPerfSchemeModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_perf_scheme.HrmPerfScheme{
		Uuid:           uuid,
		Name:           "方案",
		TemplateUuid:   tplUuid,
		PeriodType:     "month",
		AssessmentMode: "kpi",
		ScoreMode:      scoreMode,
		PayLinkMode:    "coefficient",
		Status:         "1",
	}))
	return uuid
}

// seedEmployee 在册员工 with optional bound user uuid.
func seedEmployee(t *testing.T, env *svctest.Env, userUuid string) string {
	t.Helper()
	uuid := fmt.Sprintf("emp-%d", dial(t))
	mustNoErr(t, hrm_employee.NewHrmEmployeeModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_employee.HrmEmployee{
		Uuid:             uuid,
		EmployeeNo:       fmt.Sprintf("E%04d", dial(t)),
		Name:             "员工" + fmt.Sprint(dial(t)),
		Gender:           "1",
		Birthday:         "1990-01-01",
		Nation:           "汉",
		PoliticalStatus:  "1",
		MaritalStatus:    "1",
		NativePlace:      "",
		BloodType:        "",
		IdCard:           "110101199001010011",
		Height:           "",
		Weight:           "",
		HouseholdAddress: "",
		Address:          "",
		EmergencyContact: "",
		EmergencyPhone:   "",
		Phone:            "",
		Email:            "",
		Avatar:           "",
		BankName:         "",
		BankAccount:      "",
		CompanyUuid:      "",
		OrgUuid:          "",
		Position:         "",
		Post:             "",
		PersonStatus:     "1",
		HireDate:         "2024-01-01",
		FormalDate:       "",
		UserUuid:         userUuid,
		Remark:           "",
	}))
	return uuid
}

// seedOngoingPeriod 批次（进行中，含参与人）。
func seedOngoingPeriod(t *testing.T, env *svctest.Env, schemeUuid string, employeeUuids ...string) string {
	t.Helper()
	periodUuid := fmt.Sprintf("period-%d", dial(t))
	scope := `{"companyUuids":[],"orgUuids":[],"positions":[],"posts":[],"personStatuses":[]}`
	mustNoErr(t, env.Model.HrmPerfPeriod.InsertV2(testCtx, nil, &hrm_perf_period.HrmPerfPeriod{
		Uuid: periodUuid, PeriodNo: fmt.Sprintf("PERF-2026-%04d", dial(t)),
		Name: "批次", PeriodType: "month", SchemeUuid: schemeUuid,
		Scope: scope, PeriodStart: "2026-01-01", PeriodEnd: "2026-01-31",
		Status: perf_period.PeriodStatusOngoing,
	}))
	for _, emp := range employeeUuids {
		mustNoErr(t, hrm_perf_period_participant.NewHrmPerfPeriodParticipantModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_perf_period_participant.HrmPerfPeriodParticipant{
			Uuid: fmt.Sprintf("pp-%d", dial(t)), PeriodUuid: periodUuid, EmployeeUuid: emp,
		}))
	}
	return periodUuid
}

// mustGenerate runs GenerateAssessments and requires success.
func mustGenerate(t *testing.T, svcCtx *svc.ServiceContext, periodUuid string) int64 {
	t.Helper()
	n, err := GenerateAssessments(testCtx, svcCtx.SqlxConn, svcCtx, periodUuid)
	mustNoErr(t, err)
	return n
}

// mustFindTask finds the single assessment of a period.
func mustFindTask(t *testing.T, svcCtx *svc.ServiceContext, periodUuid string) *hrm_perf_assessment.HrmPerfAssessment {
	t.Helper()
	rows, _, err := ListTasks(testCtx, svcCtx, periodUuid, "", "", "", 1, 10)
	mustNoErr(t, err)
	assertLen(t, rows, 1)
	return rows[0]
}

// actorFor builds an actorContext for tests.
func actorFor(employeeUuid string, codes ...string) *actorContext {
	return &actorContext{EmployeeUuid: employeeUuid, Codes: codes}
}

var reviewCodes = []string{PermAssessmentReview}
var calibCodes = []string{PermAssessmentCalibrate}

// ---- 1. 任务生成幂等 ----

func TestGenerateIdempotent(t *testing.T) {
	svcCtx, env := newSvc(t)
	tpl := seedTemplate(t, env)
	scheme := seedScheme(t, env, tpl, "percent")
	emp1 := seedEmployee(t, env, "")
	emp2 := seedEmployee(t, env, "")
	period := seedOngoingPeriod(t, env, scheme, emp1, emp2)

	n := mustGenerate(t, svcCtx, period)
	assertEq(t, int64(2), n)

	// 幂等：再次生成 0 新增
	n2 := mustGenerate(t, svcCtx, period)
	assertEq(t, int64(0), n2)

	rows, total, err := ListTasks(testCtx, svcCtx, period, "", "", "", 1, 10)
	mustNoErr(t, err)
	assertEq(t, int64(2), total)
	for _, row := range rows {
		assertNotEmpty(t, row.AssessmentNo)
		assertContains(t, row.AssessmentNo, "PA-2026-")
		assertEq(t, NodePlanConfirm, row.CurrentNode)
		assertEq(t, TaskStatusOngoing, row.TaskStatus)
		// 模板快照 items+权重固化
		snapshot, err := parseItemsSnapshot(row.ItemsSnapshot)
		mustNoErr(t, err)
		assertLen(t, snapshot, 2)
		assertEq(t, 60.0, snapshot[0].Weight)
		assertNotEmpty(t, snapshot[0].IndicatorCode)
	}

	// 非进行中批次拒绝生成
	draft := seedOngoingPeriod(t, env, scheme, emp1)
	mustNoErr(t, env.Model.HrmPerfPeriod.UpdateFieldsByCondition(testCtx, nil,
		map[string]any{"status": perf_period.PeriodStatusDraft},
		condition.NewChain().Equal(hrm_perf_period.Uuid, draft).Build()...))
	_ = draft
	_, err = GenerateAssessments(testCtx, svcCtx.SqlxConn, svcCtx, draft)
	assertErrIs(t, err, ErrPeriodNotOngoing)
}

func TestGenerateGuardNoParticipants(t *testing.T) {
	svcCtx, env := newSvc(t)
	tpl := seedTemplate(t, env)
	scheme := seedScheme(t, env, tpl, "percent")
	period := seedOngoingPeriod(t, env, scheme) // 无参与人
	_, err := GenerateAssessments(testCtx, svcCtx.SqlxConn, svcCtx, period)
	assertErrIs(t, err, ErrNoParticipants)
}

// ---- 2. 权重汇总计分（纯函数单测） ----

func TestWeightedTotal(t *testing.T) {
	snapshot := []ItemSnapshot{
		{IndicatorUuid: "i1", Weight: 60},
		{IndicatorUuid: "i2", Weight: 40},
	}
	total, err := WeightedTotal(snapshot, []ScoreItem{
		{IndicatorUuid: "i1", Score: 90},
		{IndicatorUuid: "i2", Score: 80},
	})
	mustNoErr(t, err)
	assertInDelta(t, 86.0, total, 0.001)

	// 缺项拒绝
	_, err = WeightedTotal(snapshot, []ScoreItem{{IndicatorUuid: "i1", Score: 90}})
	assertErrIs(t, err, ErrAssessmentInvalid)

	// 负分拒绝
	_, err = WeightedTotal(snapshot, []ScoreItem{
		{IndicatorUuid: "i1", Score: -1},
		{IndicatorUuid: "i2", Score: 80},
	})
	assertErrIs(t, err, ErrAssessmentInvalid)
}

func TestPercentGrade(t *testing.T) {
	assertEq(t, "A", PercentGrade(95))
	assertEq(t, "A", PercentGrade(90))
	assertEq(t, "B", PercentGrade(89.5))
	assertEq(t, "B", PercentGrade(75))
	assertEq(t, "C", PercentGrade(74))
	assertEq(t, "C", PercentGrade(60))
	assertEq(t, "D", PercentGrade(59.9))
}

func TestNormalizeResult(t *testing.T) {
	// percent: 自动映射等级
	score, grade, err := NormalizeResult("percent", 86, "")
	mustNoErr(t, err)
	assertEq(t, 86.0, score)
	assertEq(t, "B", grade)
	// grade: 直存等级，score=0
	score, grade, err = NormalizeResult("grade", 0, "A")
	mustNoErr(t, err)
	assertEq(t, 0.0, score)
	assertEq(t, "A", grade)
	_, _, err = NormalizeResult("grade", 0, "")
	assertErr(t, err)
	// amount: 直存金额
	score, grade, err = NormalizeResult("amount", 1234.56, "")
	mustNoErr(t, err)
	assertEq(t, 1234.56, score)
	assertEq(t, "", grade)
	// 非法模式
	_, _, err = NormalizeResult("bad", 1, "")
	assertErr(t, err)
}

// ---- 3. 五阶段顺序 + 前置守卫 + 越权填报拒绝 ----

func runFiveStageChain(t *testing.T, svcCtx *svc.ServiceContext, env *svctest.Env, scoreMode string) (
	periodUuid, assessmentUuid, empUuid string) {
	t.Helper()
	tpl := seedTemplate(t, env)
	scheme := seedScheme(t, env, tpl, scoreMode)
	emp := seedEmployee(t, env, "")
	period := seedOngoingPeriod(t, env, scheme, emp)
	mustGenerate(t, svcCtx, period)
	task := mustFindTask(t, svcCtx, period)
	return period, task.Uuid, emp
}

func TestFiveStageChainGuards(t *testing.T) {
	svcCtx, env := newSvc(t)
	_, taskUuid, emp := runFiveStageChain(t, svcCtx, env, "percent")

	selfItems := func(uuids ...string) []ScoreItem {
		out := make([]ScoreItem, 0, len(uuids))
		for _, u := range uuids {
			out = append(out, ScoreItem{IndicatorUuid: u, Score: 80})
		}
		return out
	}
	// 从快照取指标 uuid
	task := mustFindTask(t, svcCtx, mustPeriodOf(t, svcCtx, taskUuid))
	snapshot, err := parseItemsSnapshot(task.ItemsSnapshot)
	mustNoErr(t, err)
	uuids := []string{snapshot[0].IndicatorUuid, snapshot[1].IndicatorUuid}

	// 守卫：节点2未到，自评拒绝
	err = SubmitSelfReviewForTask(testCtx, svcCtx, taskUuid, selfItems(uuids...), "", actorFor(emp))
	assertErrIs(t, err, ErrNodeInvalid)
	// 越权：非本人自评拒绝
	err = ConfirmPlanForTask(testCtx, svcCtx.SqlxConn, taskUuid, "", actorFor("emp-other"))
	assertErrIs(t, err, ErrOperatorDenied)
	// 评审守卫：节点3未到（节点守卫先于权限判定触发）
	err = SubmitLeaderReviewForTask(testCtx, svcCtx, taskUuid, selfItems(uuids...), "", actorFor("", reviewCodes...))
	assertErrIs(t, err, ErrNodeInvalid)

	// 阶段1 目标确认（被考核人）
	mustNoErr(t, ConfirmPlanForTask(testCtx, svcCtx.SqlxConn, taskUuid, "确认", actorFor(emp)))
	task = mustFindTask(t, svcCtx, mustPeriodOf(t, svcCtx, taskUuid))
	assertEq(t, NodeSelfReview, task.CurrentNode)
	assertNotEmpty(t, task.PlanTime)

	// 阶段2 自评（权重汇总 80*0.6+80*0.4=80 → B）
	mustNoErr(t, SubmitSelfReviewForTask(testCtx, svcCtx, taskUuid, selfItems(uuids...), "自评", actorFor(emp)))
	task = mustFindTask(t, svcCtx, mustPeriodOf(t, svcCtx, taskUuid))
	assertEq(t, NodeLeaderReview, task.CurrentNode)
	assertInDelta(t, 80.0, task.SelfScore, 0.001)
	assertEq(t, "B", task.SelfGrade)

	// 评审越权：无 review 码（节点已到3）
	err = SubmitLeaderReviewForTask(testCtx, svcCtx, taskUuid, selfItems(uuids...), "", actorFor(emp))
	assertErrIs(t, err, ErrOperatorDenied)

	// 阶段3 上级评审（85 → 85*1.0 权重汇总 = 85 → B）
	mustNoErr(t, SubmitLeaderReviewForTask(testCtx, svcCtx, taskUuid, []ScoreItem{
		{IndicatorUuid: uuids[0], Score: 90},
		{IndicatorUuid: uuids[1], Score: 78},
	}, "评审", actorFor("", reviewCodes...)))
	task = mustFindTask(t, svcCtx, mustPeriodOf(t, svcCtx, taskUuid))
	assertEq(t, NodeCalibration, task.CurrentNode)
	assertInDelta(t, 85.2, task.ReviewScore, 0.001) // 90*0.6+78*0.4
	assertEq(t, "B", task.ReviewGrade)

	// 校准越权：无 calibrate 码
	err = SubmitCalibrationForTask(testCtx, svcCtx, taskUuid, 88, "", "", actorFor("", reviewCodes...))
	assertErrIs(t, err, ErrOperatorDenied)

	// 阶段4 校准（88 → B）
	mustNoErr(t, SubmitCalibrationForTask(testCtx, svcCtx, taskUuid, 88, "", "校准", actorFor("", calibCodes...)))
	task = mustFindTask(t, svcCtx, mustPeriodOf(t, svcCtx, taskUuid))
	assertEq(t, NodeResultConfirm, task.CurrentNode)
	assertInDelta(t, 88.0, task.FinalScore, 0.001)
	assertEq(t, "B", task.FinalGrade)

	// 结果确认越权
	err = ConfirmResultForTask(testCtx, svcCtx, taskUuid, actorFor("", calibCodes...))
	assertErrIs(t, err, ErrOperatorDenied)

	// 阶段5 结果确认（被考核人）
	mustNoErr(t, ConfirmResultForTask(testCtx, svcCtx, taskUuid, actorFor(emp)))
	task = mustFindTask(t, svcCtx, mustPeriodOf(t, svcCtx, taskUuid))
	assertEq(t, TaskStatusDone, task.TaskStatus)
	assertNotEmpty(t, task.ResultTime)

	// 完成后再填报 → 守卫拒绝
	err = SubmitSelfReviewForTask(testCtx, svcCtx, taskUuid, selfItems(uuids...), "", actorFor(emp))
	assertErrIs(t, err, ErrNodeInvalid)
}

// mustPeriodOf finds the period uuid of an assessment.
func mustPeriodOf(t *testing.T, svcCtx *svc.ServiceContext, assessmentUuid string) string {
	t.Helper()
	row, err := FindAssessmentByUuid(testCtx, svcCtx.SqlxConn, assessmentUuid)
	mustNoErr(t, err)
	return row.PeriodUuid
}

// ---- 4. 申诉发起与处理路径 ----

func TestAppealFlow(t *testing.T) {
	svcCtx, env := newSvc(t)
	_, taskUuid, emp := runFiveStageChain(t, svcCtx, env, "percent")

	task := mustFindTask(t, svcCtx, mustPeriodOf(t, svcCtx, taskUuid))
	snapshot, err := parseItemsSnapshot(task.ItemsSnapshot)
	mustNoErr(t, err)
	uuids := []string{snapshot[0].IndicatorUuid, snapshot[1].IndicatorUuid}

	// 走完前四阶段
	mustNoErr(t, ConfirmPlanForTask(testCtx, svcCtx.SqlxConn, taskUuid, "", actorFor(emp)))
	mustNoErr(t, SubmitSelfReviewForTask(testCtx, svcCtx, taskUuid, []ScoreItem{
		{IndicatorUuid: uuids[0], Score: 80}, {IndicatorUuid: uuids[1], Score: 80},
	}, "", actorFor(emp)))
	mustNoErr(t, SubmitLeaderReviewForTask(testCtx, svcCtx, taskUuid, []ScoreItem{
		{IndicatorUuid: uuids[0], Score: 90}, {IndicatorUuid: uuids[1], Score: 78},
	}, "", actorFor("", reviewCodes...)))
	mustNoErr(t, SubmitCalibrationForTask(testCtx, svcCtx, taskUuid, 88, "", "", actorFor("", calibCodes...)))

	// 申诉守卫：原因必填
	err = AppealForTask(testCtx, svcCtx.SqlxConn, taskUuid, "", actorFor(emp))
	assertErrIs(t, err, ErrAssessmentInvalid)
	// 申诉越权：非本人
	err = AppealForTask(testCtx, svcCtx.SqlxConn, taskUuid, "原因", actorFor("emp-other"))
	assertErrIs(t, err, ErrOperatorDenied)

	// 发起申诉
	mustNoErr(t, AppealForTask(testCtx, svcCtx.SqlxConn, taskUuid, "对评审得分有异议", actorFor(emp)))
	task = mustFindTask(t, svcCtx, mustPeriodOf(t, svcCtx, taskUuid))
	assertEq(t, TaskStatusAppealed, task.TaskStatus)

	// 待处理期间重复发起拒绝
	err = AppealForTask(testCtx, svcCtx.SqlxConn, taskUuid, "再申诉", actorFor(emp))
	assertErrIs(t, err, ErrAppealExists)

	// 裁决越权：无 calibrate 码
	appeals, err := svcCtx.Model.HrmPerfAppeal.FindByCondition(testCtx, nil, condition.NewChain().Build()...)
	mustNoErr(t, err)
	assertLen(t, appeals, 1)
	appealUuid := appeals[0].Uuid
	err = HandleAppealForTask(testCtx, svcCtx.SqlxConn, appealUuid, AppealUpheld, "", 0, "", actorFor("", reviewCodes...))
	assertErrIs(t, err, ErrOperatorDenied)

	// 裁决调整得分：92 → A
	mustNoErr(t, HandleAppealForTask(testCtx, svcCtx.SqlxConn, appealUuid, AppealAdjusted, "复核上调", 92, "", actorFor("", calibCodes...)))
	task = mustFindTask(t, svcCtx, mustPeriodOf(t, svcCtx, taskUuid))
	assertEq(t, TaskStatusDone, task.TaskStatus)
	assertInDelta(t, 92.0, task.FinalScore, 0.001)
	assertEq(t, "A", task.FinalGrade)
	appeals, err = svcCtx.Model.HrmPerfAppeal.FindByCondition(testCtx, nil, condition.NewChain().Build()...)
	mustNoErr(t, err)
	assertEq(t, AppealAdjusted, appeals[0].Status)

	// 已裁决不可重复处理
	err = HandleAppealForTask(testCtx, svcCtx.SqlxConn, appealUuid, AppealUpheld, "", 0, "", actorFor("", calibCodes...))
	assertErr(t, err)
}

// ---- 5. 批次状态联动 advance ----

func TestPeriodSyncByTasks(t *testing.T) {
	svcCtx, env := newSvc(t)
	period, task1, emp1 := runFiveStageChain(t, svcCtx, env, "percent")

	// 第二个任务
	tpl := mustTemplateOf(t, svcCtx, task1)
	scheme := mustSchemeOf(t, svcCtx, period)
	emp2 := seedEmployee(t, env, "")
	mustNoErr(t, hrm_perf_period_participant.NewHrmPerfPeriodParticipantModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_perf_period_participant.HrmPerfPeriodParticipant{
		Uuid: fmt.Sprintf("pp-%d", dial(t)), PeriodUuid: period, EmployeeUuid: emp2,
	}))
	mustGenerate(t, svcCtx, period)

	rows, _, err := ListTasks(testCtx, svcCtx, period, "", "", "", 1, 10)
	mustNoErr(t, err)
	assertLen(t, rows, 2)
	task2Uuid := rows[1].Uuid

	snapshot, err := parseItemsSnapshot(rows[0].ItemsSnapshot)
	mustNoErr(t, err)
	uuids := []string{snapshot[0].IndicatorUuid, snapshot[1].IndicatorUuid}

	complete := func(taskUuid, emp string) {
		mustNoErr(t, ConfirmPlanForTask(testCtx, svcCtx.SqlxConn, taskUuid, "", actorFor(emp)))
		mustNoErr(t, SubmitSelfReviewForTask(testCtx, svcCtx, taskUuid, []ScoreItem{
			{IndicatorUuid: uuids[0], Score: 85}, {IndicatorUuid: uuids[1], Score: 85},
		}, "", actorFor(emp)))
		mustNoErr(t, SubmitLeaderReviewForTask(testCtx, svcCtx, taskUuid, []ScoreItem{
			{IndicatorUuid: uuids[0], Score: 85}, {IndicatorUuid: uuids[1], Score: 85},
		}, "", actorFor("", reviewCodes...)))
		mustNoErr(t, SubmitCalibrationForTask(testCtx, svcCtx, taskUuid, 85, "", "", actorFor("", calibCodes...)))
	}

	// 任务1走到校准、任务2未动 → 批次仍进行中
	complete(task1, emp1)
	periodRow, err := findPeriod(testCtx, svcCtx.SqlxConn, period)
	mustNoErr(t, err)
	assertEq(t, perf_period.PeriodStatusOngoing, periodRow.Status)

	// 任务2也到校准 → 批次联动至 '3'待校准（'2'无独立停留）
	complete(task2Uuid, emp2)
	periodRow, err = findPeriod(testCtx, svcCtx.SqlxConn, period)
	mustNoErr(t, err)
	assertEq(t, perf_period.PeriodStatusCalibrating, periodRow.Status)

	// 全部结果确认 → 批次联动至 '4'结果已发布
	mustNoErr(t, ConfirmResultForTask(testCtx, svcCtx, task1, actorFor(emp1)))
	periodRow, err = findPeriod(testCtx, svcCtx.SqlxConn, period)
	mustNoErr(t, err)
	assertEq(t, perf_period.PeriodStatusCalibrating, periodRow.Status) // 任务2未确认不联动

	mustNoErr(t, ConfirmResultForTask(testCtx, svcCtx, task2Uuid, actorFor(emp2)))
	periodRow, err = findPeriod(testCtx, svcCtx.SqlxConn, period)
	mustNoErr(t, err)
	assertEq(t, perf_period.PeriodStatusResultPublished, periodRow.Status)

	_ = tpl
	_ = scheme
}

// helpers to reload scheme/template uuids
func mustTemplateOf(t *testing.T, svcCtx *svc.ServiceContext, assessmentUuid string) string {
	t.Helper()
	row, err := FindAssessmentByUuid(testCtx, svcCtx.SqlxConn, assessmentUuid)
	mustNoErr(t, err)
	return row.TemplateUuid
}

func mustSchemeOf(t *testing.T, svcCtx *svc.ServiceContext, periodUuid string) string {
	t.Helper()
	period, err := findPeriod(testCtx, svcCtx.SqlxConn, periodUuid)
	mustNoErr(t, err)
	return period.SchemeUuid
}

// silence unused import of types package in some builds
var _ = types.PageRequest{}
