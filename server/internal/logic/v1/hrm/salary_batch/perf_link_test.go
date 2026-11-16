package salary_batch

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// ===== HRM-47 绩效联薪接入试算（issue #49）svctest 测试 =====
//
// Coverage:
//   - 回填幂等：系数制结果行首次试算回填 perf_amount=系数×基数+salary_synced=1；重算只读不变金额
//   - 未发布不回填：published=0 结果行不回填、引擎走考核任务兜底 + 告警
//   - 引擎取数：MERIT_SALARY = 结果行回填值 = 系数×基数（worked example 系数制）
//   - amount 制不回填：金额制结果行 perf_amount=score 直读，salary_synced 保持 0
//   - 结果行缺失告警：有已完成任务但无结果行 → 未发布告警（绩效缺失提示已有）

// seedPerfSchemeAndPeriod 只插方案+批次（无考核任务）：结果行直挂 period 场景（HRM-47）。
func seedPerfSchemeAndPeriod(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, payLinkMode string) {
	t.Helper()
	if _, err := svcCtx.Model.HrmPerfScheme.Insert(t.Context(), nil, &hrm_perf_scheme.HrmPerfScheme{
		Uuid: perfSchemeUuidOf(), Name: "绩效方案", PeriodType: "month", ScoreMode: "percent", PayLinkMode: payLinkMode, Status: "1",
	}); err != nil {
		t.Fatalf("seed perf scheme: %v", err)
	}
	if err := svcCtx.Model.HrmPerfPeriod.InsertV2(t.Context(), nil, &hrm_perf_period.HrmPerfPeriod{
		Uuid: perfPeriodUuidOf(), PeriodNo: "PERF-2026-9001", Name: "2026-03 批次", PeriodType: "month",
		SchemeUuid: perfSchemeUuidOf(), Scope: "{}",
		PeriodStart: "2026-03-01", PeriodEnd: "2026-03-31", Status: "5",
	}); err != nil {
		t.Fatalf("seed perf period: %v", err)
	}
}

// seedPerfResultRow 插一条绩效结果行（HRM-45 契约：UNIQUE(period_uuid,employee_uuid)）。
func seedPerfResultRow(t *testing.T, svcCtx *svc.ServiceContext, periodUuid, employeeUuid, periodNo, grade string,
	coefficient, perfAmount float64, published, salarySynced int64) string {
	t.Helper()
	u := uuid.New().String()
	if err := svcCtx.Model.HrmPerfResult.InsertV2(t.Context(), nil, &hrm_perf_result.HrmPerfResult{
		Uuid: u, ApprovalUuid: uuid.New().String(), AssessmentUuid: uuid.New().String(),
		PeriodUuid: periodUuid, PeriodNo: periodNo, EmployeeUuid: employeeUuid,
		EmployeeName: "员工", Score: 88, Grade: grade,
		Coefficient: coefficient, PerfAmount: perfAmount,
		ResultStatus: "1", Published: published, SalarySynced: salarySynced,
	}); err != nil {
		t.Fatalf("seed perf result: %v", err)
	}
	return u
}

// findPerfResultRow 按 uuid 读结果行（断言回填后状态）。
func findPerfResultRow(t *testing.T, svcCtx *svc.ServiceContext, resultUuid string) *hrm_perf_result.HrmPerfResult {
	t.Helper()
	row, err := svcCtx.Model.HrmPerfResult.FindOneByCondition(t.Context(), nil,
		condition.NewChain().Equal(hrm_perf_result.Uuid, resultUuid).Build()...)
	if err != nil {
		t.Fatalf("find perf result: %v", err)
	}
	return row
}

// setupLinkExample 布景：员工固定薪资 10000（系数制基数口径=档案 fixed_salary 全额，
// 方案无 MERIT_SALARY 覆盖配置）+ 绩效方案 coefficient + 批次（2026-03 覆盖）+ 结果行。
func setupLinkExample(t *testing.T, svcCtx *svc.ServiceContext, published, salarySynced int64,
	coefficient float64, seedAssessment bool) (batchUuid, employeeUuid, resultUuid, periodUuid string) {
	t.Helper()
	resetPerfUuids()
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subject := seedSubjectRow(t, svcCtx, "联薪主体")
	scheme := seedSchemeRow(t, svcCtx, "联薪方案", subject)
	items := listAllSalaryItems(t, svcCtx)
	for _, it := range items {
		seedSchemeItemRow(t, svcCtx, scheme, it.Uuid)
	}
	employeeUuid = seedEmployeeRow(t, svcCtx, "联薪员", "E101", company, dept)
	archive := archiveOf(employeeUuid, subject, scheme, "联薪主体", "联薪方案")
	archive.EmployeeNo = "E101"
	archive.EmployeeName = "联薪员"
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(t.Context(), nil, archive); err != nil {
		t.Fatalf("seed archive: %v", err)
	}
	// 绩效链（seedPerfChain：period 2026-03 覆盖 + scheme coefficient + 已完成考核任务）。
	// seedAssessment=false 时用 seedPerfChainKeepAssessment 不插任务行（结果行直挂 period）。
	if seedAssessment {
		seedPerfChain(t, svcCtx, employeeUuid, "coefficient", "B", 88)
	} else {
		seedPerfSchemeAndPeriod(t, svcCtx, employeeUuid, "coefficient")
	}
	resultUuid = seedPerfResultRow(t, svcCtx, perfPeriodUuidOf(), employeeUuid, "PERF-2026-9001", "B",
		coefficient, 0, published, salarySynced)
	batchUuid = mustAddBatch(t, svcCtx, "联薪批次", subject, 2026, 3)
	return batchUuid, employeeUuid, resultUuid, perfPeriodUuidOf()
}

// TestTrialPerfLinkBackfill 核心验收（票面 worked example）：系数制员工（等级 B 系数 1.0）
// 已发布结果行 perf_amount=0 → 试算 MERIT_SALARY = 系数×基数 = 1.0×10000 = 10000，
// 且结果行回填 perf_amount=10000 + salary_synced=1；重算幂等（金额不变、不重复翻倍）。
func TestTrialPerfLinkBackfill(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, employeeUuid, resultUuid, _ := setupLinkExample(t, svcCtx, 1, 0, 1.0, false)

	// 第一次试算：MERIT_SALARY = 1.0 × 10000
	mustTrial(t, svcCtx, batchUuid)
	m := itemAmountsOf(t, svcCtx, batchUuid, employeeUuid)
	if got := amountOf(t, m, "MERIT_SALARY"); got != 10000 {
		t.Errorf("MERIT_SALARY = %v, want 10000 (1.0 × 10000)", got)
	}
	if m["MERIT_SALARY"].Source != SourcePerf {
		t.Errorf("MERIT_SALARY source = %s, want %s", m["MERIT_SALARY"].Source, SourcePerf)
	}
	row := findPerfResultRow(t, svcCtx, resultUuid)
	if row.PerfAmount != 10000 {
		t.Errorf("result perf_amount = %v, want 10000 (回填)", row.PerfAmount)
	}
	if row.SalarySynced != 1 {
		t.Errorf("result salary_synced = %d, want 1", row.SalarySynced)
	}

	// 重算幂等：回填后金额不变（勿重算 GradeCoefficient/勿二次叠加）
	mustTrial(t, svcCtx, batchUuid)
	m = itemAmountsOf(t, svcCtx, batchUuid, employeeUuid)
	if got := amountOf(t, m, "MERIT_SALARY"); got != 10000 {
		t.Errorf("MERIT_SALARY after re-trial = %v, want 10000 (幂等)", got)
	}
	row = findPerfResultRow(t, svcCtx, resultUuid)
	if row.PerfAmount != 10000 {
		t.Errorf("result perf_amount after re-trial = %v, want 10000", row.PerfAmount)
	}
	if row.SalarySynced != 1 {
		t.Errorf("result salary_synced after re-trial = %d, want 1", row.SalarySynced)
	}
}

// TestTrialPerfLinkCoefficientA 系数 1.2（等级 A）：MERIT_SALARY = 1.2 × 10000 = 12000。
func TestTrialPerfLinkCoefficientA(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, employeeUuid, resultUuid, _ := setupLinkExample(t, svcCtx, 1, 0, 1.2, false)
	mustTrial(t, svcCtx, batchUuid)
	m := itemAmountsOf(t, svcCtx, batchUuid, employeeUuid)
	if got := amountOf(t, m, "MERIT_SALARY"); got != 12000 {
		t.Errorf("MERIT_SALARY = %v, want 12000 (1.2 × 10000)", got)
	}
	row := findPerfResultRow(t, svcCtx, resultUuid)
	if row.PerfAmount != 12000 || row.SalarySynced != 1 {
		t.Errorf("backfill = (%v, %d), want (12000, 1)", row.PerfAmount, row.SalarySynced)
	}
}

// TestTrialPerfLinkUnpublishedNotBackfilled 未发布（published=0）结果行不回填：
// 引擎走考核任务兜底口径（等级 B 系数 1.0 × 10000 = 10000）+ 未发布告警；salary_synced 保持 0。
func TestTrialPerfLinkUnpublishedNotBackfilled(t *testing.T) {
	svcCtx := newCtx(t)
	// seedAssessment=true：保留已完成考核任务供兜底
	batchUuid, employeeUuid, resultUuid, _ := setupLinkExample(t, svcCtx, 0, 0, 1.0, true)
	resp := mustTrial(t, svcCtx, batchUuid)
	m := itemAmountsOf(t, svcCtx, batchUuid, employeeUuid)
	// 兜底口径金额同为 1.0×10000（考核任务 final_grade=B），但结果行不回填
	if got := amountOf(t, m, "MERIT_SALARY"); got != 10000 {
		t.Errorf("MERIT_SALARY = %v, want 10000 (考核兜底)", got)
	}
	row := findPerfResultRow(t, svcCtx, resultUuid)
	if row.SalarySynced != 0 || row.PerfAmount != 0 {
		t.Errorf("unpublished result backfilled: perf_amount=%v salary_synced=%d, want (0, 0)",
			row.PerfAmount, row.SalarySynced)
	}
	if !containsWarning(resp.Warnings, "未发布") {
		t.Errorf("warnings should mention 结果未发布: %q", resp.Warnings)
	}
}

// TestTrialPerfLinkAmountModeNoBackfill 金额制结果行不回填：perf_amount=score 已有值直读，
// salary_synced 保持 0（HRM-45 契约：金额制无需 HRM-49 回填）。
func TestTrialPerfLinkAmountModeNoBackfill(t *testing.T) {
	svcCtx := newCtx(t)
	resetPerfUuids()
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subject := seedSubjectRow(t, svcCtx, "金额联薪主体")
	scheme := seedSchemeRow(t, svcCtx, "金额联薪方案", subject)
	items := listAllSalaryItems(t, svcCtx)
	for _, it := range items {
		seedSchemeItemRow(t, svcCtx, scheme, it.Uuid)
	}
	emp := seedEmployeeRow(t, svcCtx, "金额员", "E102", company, dept)
	archive := archiveOf(emp, subject, scheme, "金额联薪主体", "金额联薪方案")
	archive.EmployeeNo = "E102"
	archive.EmployeeName = "金额员"
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(t.Context(), nil, archive); err != nil {
		t.Fatalf("seed archive: %v", err)
	}
	seedPerfChain(t, svcCtx, emp, "amount", "A", 666.66)
	resultUuid := seedPerfResultRow(t, svcCtx, perfPeriodUuidOf(), emp, "PERF-2026-9001", "A",
		1.2, 666.66, 1, 0)
	batchUuid := mustAddBatch(t, svcCtx, "金额联薪批次", subject, 2026, 3)
	mustTrial(t, svcCtx, batchUuid)
	m := itemAmountsOf(t, svcCtx, batchUuid, emp)
	if got := amountOf(t, m, "MERIT_SALARY"); got != 666.66 {
		t.Errorf("MERIT_SALARY = %v, want 666.66 (结果行直读)", got)
	}
	row := findPerfResultRow(t, svcCtx, resultUuid)
	if row.SalarySynced != 0 {
		t.Errorf("amount-mode result salary_synced = %d, want 0 (金额制不回填)", row.SalarySynced)
	}
}

// TestTrialPerfLinkMissingResultWarning 有已完成考核任务但无任何结果行 → 考核兜底 + 未发布告警
// （结果行缺失路径；无绩效提示已有，见 TestTrialPerfMissingFallback）。
func TestTrialPerfLinkMissingResultWarning(t *testing.T) {
	svcCtx := newCtx(t)
	resetPerfUuids()
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subject := seedSubjectRow(t, svcCtx, "缺失主体")
	scheme := seedSchemeRow(t, svcCtx, "缺失方案", subject)
	items := listAllSalaryItems(t, svcCtx)
	for _, it := range items {
		seedSchemeItemRow(t, svcCtx, scheme, it.Uuid)
	}
	emp := seedEmployeeRow(t, svcCtx, "缺失员", "E103", company, dept)
	archive := archiveOf(emp, subject, scheme, "缺失主体", "缺失方案")
	archive.EmployeeNo = "E103"
	archive.EmployeeName = "缺失员"
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(t.Context(), nil, archive); err != nil {
		t.Fatalf("seed archive: %v", err)
	}
	// 只有考核任务（无结果行）：seedPerfChain 布景 period+scheme+assessment
	seedPerfChain(t, svcCtx, emp, "coefficient", "B", 88)
	batchUuid := mustAddBatch(t, svcCtx, "缺失批次", subject, 2026, 3)
	resp := mustTrial(t, svcCtx, batchUuid)
	m := itemAmountsOf(t, svcCtx, batchUuid, emp)
	if got := amountOf(t, m, "MERIT_SALARY"); got != 10000 {
		t.Errorf("MERIT_SALARY = %v, want 10000 (考核兜底 1.0×10000)", got)
	}
	if !containsWarning(resp.Warnings, "未发布") {
		t.Errorf("warnings should mention 结果未发布: %q", resp.Warnings)
	}
}

// containsWarning 告警文本按行包含判定。
func containsWarning(warnings, substr string) bool {
	for _, line := range splitLines(warnings) {
		if len(line) >= len(substr) && indexOf(line, substr) >= 0 {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	out := []string{}
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	return append(out, cur)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
