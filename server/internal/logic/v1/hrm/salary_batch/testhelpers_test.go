package salary_batch

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_archive"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_opening"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_special_deduction"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// ===== 测试布景 helpers（salary_batch_test.go 同 package） =====

// listAllSalaryItems 全量薪资项目行（迁移 25 预置 16 项，calc_order ASC）。
func listAllSalaryItems(t *testing.T, svcCtx *svc.ServiceContext) []*hrm_salary_item.HrmSalaryItem {
	t.Helper()
	rows, err := svcCtx.Model.HrmSalaryItem.FindByCondition(t.Context(), nil,
		condition.NewChain().OrderByAsc(hrm_salary_item.CalcOrder).Build()...)
	if err != nil {
		t.Fatalf("list salary items: %v", err)
	}
	return rows
}

// seedSchemeRow 插方案行（返回 uuid）。
func seedSchemeRow(t *testing.T, svcCtx *svc.ServiceContext, name, subjectUuid string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalaryScheme.Insert(t.Context(), nil, &hrm_salary_scheme.HrmSalaryScheme{
		Uuid: u, SchemeName: name, SubjectUuid: subjectUuid, SubjectName: name + "-主体",
		PersonStatus: "1", SettleCycle: "monthly", PayDay: 1, Currency: "CNY", Status: "1",
	}); err != nil {
		t.Fatalf("seed scheme %s: %v", name, err)
	}
	return u
}

// seedSchemeItemRow 方案×项目关联行。
func seedSchemeItemRow(t *testing.T, svcCtx *svc.ServiceContext, schemeUuid, itemUuid string) {
	t.Helper()
	if err := svcCtx.Model.HrmSalarySchemeItem.InsertV2(t.Context(), nil, &hrm_salary_scheme_item.HrmSalarySchemeItem{
		Uuid: uuid.New().String(), SchemeUuid: schemeUuid, ItemUuid: itemUuid,
	}); err != nil {
		t.Fatalf("seed scheme item: %v", err)
	}
}

// archiveOf 构造员工薪资档案行（2026-01 起长期有效；固定薪资 10000 / 固定补贴 500；
// 快照工号/姓名由调用方覆写）。
func archiveOf(employeeUuid, subjectUuid, schemeUuid string, names ...string) *hrm_salary_archive.HrmSalaryArchive {
	subjectName := "主体"
	schemeName := "方案"
	if len(names) > 0 {
		subjectName = names[0]
	}
	if len(names) > 1 {
		schemeName = names[1]
	}
	return &hrm_salary_archive.HrmSalaryArchive{
		Uuid:           uuid.New().String(),
		EmployeeUuid:   employeeUuid,
		EmployeeNo:     "E000",
		EmployeeName:   "员工",
		SubjectUuid:    subjectUuid,
		SubjectName:    subjectName,
		SchemeUuid:     schemeUuid,
		SchemeName:     schemeName,
		FixedSalary:    10000,
		FixedAllowance: 500,
		Status:         "1",
		EffectiveDate:  "2026-01-01",
		ExpireDate:     "",
	}
}

// seedPerfChain 布景绩效链：perf_scheme(payLinkMode) + perf_period(2026-03 覆盖) +
// assessment(task_status='1', final_grade/final_score)（HRM-43/46 契约：结果在 assessment 行直读）。
func seedPerfChain(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, payLinkMode, finalGrade string, finalScore float64) {
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
	if err := svcCtx.Model.HrmPerfAssessment.InsertV2(t.Context(), nil, &hrm_perf_assessment.HrmPerfAssessment{
		Uuid: uuid.New().String(), AssessmentNo: "PA-2026-9001",
		PeriodUuid: perfPeriodUuidOf(), EmployeeUuid: employeeUuid,
		CurrentNode: "5", TaskStatus: "1",
		FinalScore: finalScore, FinalGrade: finalGrade,
	}); err != nil {
		t.Fatalf("seed perf assessment: %v", err)
	}
}

// seedMissingPunchRow 打卡日结缺卡行（settle_status='missing'，打卡兜底口径取数源，hrm-16 契约）。
func seedMissingPunchRow(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, punchDate string) {
	t.Helper()
	if err := svcCtx.Model.HrmPunchRecord.InsertV2(t.Context(), nil, &hrm_punch_record.HrmPunchRecord{
		Uuid: uuid.New().String(), PunchDate: punchDate, EmployeeUuid: employeeUuid,
		SettleStatus: "missing",
	}); err != nil {
		t.Fatalf("seed punch: %v", err)
	}
}

// seedTaxOpeningRow 个税期初行（员工+年度；untilMonth 截至月，hrm-36 契约）。
func seedTaxOpeningRow(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, subjectUuid, year string, untilMonth int64,
	income, special, extra, other, taxPaid float64) {
	t.Helper()
	if _, err := svcCtx.Model.HrmTaxOpening.Insert(t.Context(), nil, &hrm_tax_opening.HrmTaxOpening{
		Uuid: uuid.New().String(), EmployeeUuid: employeeUuid, SubjectUuid: subjectUuid, TaxYear: year,
		UntilMonth: untilMonth, IncomeTotal: income, SpecialDeductionTotal: special,
		SpecialExtraTotal: extra, OtherDeductionTotal: other, TaxPaidTotal: taxPaid,
	}); err != nil {
		t.Fatalf("seed tax opening: %v", err)
	}
}

// seedSpecialDeduction 个税专项附加单月单类型行。
func seedSpecialDeduction(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, year string, month int64, amount float64) {
	t.Helper()
	if _, err := svcCtx.Model.HrmTaxSpecialDeduction.Insert(t.Context(), nil, &hrm_tax_special_deduction.HrmTaxSpecialDeduction{
		Uuid: uuid.New().String(), EmployeeUuid: employeeUuid, DeductYear: year, DeductMonth: month,
		DeductType: "CHILDREN_EDU", Amount: amount, Source: "manual",
	}); err != nil {
		t.Fatalf("seed special deduction: %v", err)
	}
}

// perf chain uuid slots（同布景内 scheme/period 共享引用）
var (
	perfSchemeUuidVal string
	perfPeriodUuidVal string
)

func perfSchemeUuidOf() string {
	if perfSchemeUuidVal == "" {
		perfSchemeUuidVal = uuid.New().String()
	}
	return perfSchemeUuidVal
}

func perfPeriodUuidOf() string {
	if perfPeriodUuidVal == "" {
		perfPeriodUuidVal = uuid.New().String()
	}
	return perfPeriodUuidVal
}

// resetPerfUuids 每个 Test 独立布景（避免跨 Test slot 串扰）。
func resetPerfUuids() {
	perfSchemeUuidVal = ""
	perfPeriodUuidVal = ""
}
