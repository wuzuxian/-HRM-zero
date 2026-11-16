package salary_batch

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_reward_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
)

// ===== svctest harness（与 salary_archive 测试同款装配） =====

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

func assertErrCode(t *testing.T, err error, wantCode int, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error %d, got nil", what, wantCode)
	}
	if sc := status.FromError(err); sc == nil || sc.Code() != status.Code(wantCode) {
		t.Fatalf("%s: error = %v, want code %d", what, err, wantCode)
	}
}

func seedOrgRow(t *testing.T, svcCtx *svc.ServiceContext, name, orgType string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmOrg.Insert(t.Context(), nil, &hrm_org.HrmOrg{
		Uuid: u, OrgName: name, OrgType: orgType, OrderNum: 1, Status: "1",
	}); err != nil {
		t.Fatalf("seed org %s: %v", name, err)
	}
	return u
}

func seedEmployeeRow(t *testing.T, svcCtx *svc.ServiceContext, name, employeeNo, companyUuid, orgUuid string) string {
	t.Helper()
	u := uuid.New().String()
	if err := svcCtx.Model.HrmEmployee.InsertV2(t.Context(), nil, &hrm_employee.HrmEmployee{
		Uuid: u, EmployeeNo: employeeNo, Name: name,
		CompanyUuid: companyUuid, OrgUuid: orgUuid, PersonStatus: "1",
	}); err != nil {
		t.Fatalf("seed employee %s: %v", name, err)
	}
	return u
}

func seedSubjectRow(t *testing.T, svcCtx *svc.ServiceContext, name string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalarySubject.Insert(t.Context(), nil, &hrm_salary_subject.HrmSalarySubject{
		Uuid: u, SubjectName: name, OrgUuid: uuid.New().String(), OrgName: "甲公司", Status: "1",
	}); err != nil {
		t.Fatalf("seed subject %s: %v", name, err)
	}
	return u
}

func mustAddBatch(t *testing.T, svcCtx *svc.ServiceContext, name, subjectUuid string, year, month int64) string {
	t.Helper()
	resp, err := NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{
		BatchName: name, SubjectUuid: subjectUuid, BatchYear: year, BatchMonth: month,
	})
	if err != nil {
		t.Fatalf("add batch %s: %v", name, err)
	}
	return resp.Uuid
}

func mustTrial(t *testing.T, svcCtx *svc.ServiceContext, batchUuid string) *types.TrialResponse {
	t.Helper()
	resp, err := RunPayrollTrial(context.Background(), svcCtx, batchUuid)
	if err != nil {
		t.Fatalf("trial %s: %v", batchUuid, err)
	}
	return resp
}

// itemAmountsOf 按员工取明细行 map[itemCode]record（查看结果端点口径）。
func itemAmountsOf(t *testing.T, svcCtx *svc.ServiceContext, batchUuid, employeeUuid string) map[string]types.HrmSalaryBatchItem {
	t.Helper()
	resp, err := NewGetBatchResult(context.Background(), svcCtx, nil).GetBatchResult(&types.GetBatchResultRequest{BatchUuid: batchUuid})
	if err != nil {
		t.Fatalf("get batch result: %v", err)
	}
	out := map[string]types.HrmSalaryBatchItem{}
	for _, it := range resp.Items {
		if it.EmployeeUuid == employeeUuid {
			out[it.ItemCode] = it
		}
	}
	return out
}

func amountOf(t *testing.T, m map[string]types.HrmSalaryBatchItem, code string) float64 {
	t.Helper()
	it, ok := m[code]
	if !ok {
		t.Fatalf("item %s missing in trial result", code)
	}
	return it.Amount
}

// ===== 试算 worked example（票面指定核心算例） =====

// setupWorkedExample 布景：员工固定薪资 10000 + 固定补贴 500 + 方案（全 16 预置项目），
// 奖励 300 + 处分 100 + 缺勤 1 天（打卡 missing 兜底）+ 无绩效无个税期初。
func setupWorkedExample(t *testing.T, svcCtx *svc.ServiceContext) (batchUuid, employeeUuid string) {
	t.Helper()
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subject := seedSubjectRow(t, svcCtx, "3月主体")
	scheme := seedSchemeRow(t, svcCtx, "标准方案", subject)

	items := listAllSalaryItems(t, svcCtx)
	if len(items) < 16 {
		t.Fatalf("preset salary items missing: got %d", len(items))
	}
	for _, it := range items {
		seedSchemeItemRow(t, svcCtx, scheme, it.Uuid)
	}

	employeeUuid = seedEmployeeRow(t, svcCtx, "张三", "E001", company, dept)
	archive := archiveOf(employeeUuid, subject, scheme, "3月主体", "标准方案")
	archive.EmployeeNo = "E001"
	archive.EmployeeName = "张三"
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(t.Context(), nil, archive); err != nil {
		t.Fatalf("seed archive: %v", err)
	}

	seedReward(t, svcCtx, employeeUuid, "reward", 300)
	seedReward(t, svcCtx, employeeUuid, "punishment", 100)
	seedMissingPunchRow(t, svcCtx, employeeUuid, "2026-03-05")

	batchUuid = mustAddBatch(t, svcCtx, "2026-03 月度批次", subject, 2026, 3)
	return batchUuid, employeeUuid
}

func TestTrialWorkedExample(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, employeeUuid := setupWorkedExample(t, svcCtx)
	resp := mustTrial(t, svcCtx, batchUuid)

	if resp.Headcount != 1 {
		t.Fatalf("headcount = %d, want 1", resp.Headcount)
	}
	m := itemAmountsOf(t, svcCtx, batchUuid, employeeUuid)

	// 固定项：BASIC_SALARY←档案固定薪资 10000；FIXED_ALLOWANCE←档案固定补贴 500
	if got := amountOf(t, m, "BASIC_SALARY"); got != 10000 {
		t.Errorf("BASIC_SALARY = %v, want 10000", got)
	}
	if got := amountOf(t, m, "FIXED_ALLOWANCE"); got != 500 {
		t.Errorf("FIXED_ALLOWANCE = %v, want 500", got)
	}
	// 取数项：REWARD_BONUS=300 / REWARD_PENALTY=100（synced+approved+2026-03）
	if got := amountOf(t, m, "REWARD_BONUS"); got != 300 {
		t.Errorf("REWARD_BONUS = %v, want 300", got)
	}
	if got := amountOf(t, m, "REWARD_PENALTY"); got != 100 {
		t.Errorf("REWARD_PENALTY = %v, want 100", got)
	}
	// 缺勤扣款：缺勤项前应发累计 = 10000+10000+0(MERIT兜底)+500+500+300 = 21300
	// 日薪 = 21300/31 = 687.096774…，缺勤 1 天 = 687.10（round2）
	if got := amountOf(t, m, "ABSENCE_DEDUCTION"); got != 687.10 {
		t.Errorf("ABSENCE_DEDUCTION = %v, want 687.10", got)
	}
	// SI/HF 社保未建兜底 0
	for _, code := range []string{"SI_PERSONAL", "HF_PERSONAL", "SI_COMPANY", "HF_COMPANY"} {
		if got := amountOf(t, m, code); got != 0 {
			t.Errorf("%s = %v, want 0 (社保未建兜底)", code, got)
		}
	}
	// TAX：无个税期初兜底 0
	if got := amountOf(t, m, "TAX"); got != 0 {
		t.Errorf("TAX = %v, want 0", got)
	}

	// 公式项（独立断言见 TestTrialFormulaItems）：NET = 21300 − 787.10 = 20512.90
	if got := amountOf(t, m, "NET_SALARY"); got != 20512.90 {
		t.Errorf("NET_SALARY = %v, want 20512.90", got)
	}

	// 汇总（NET 是 earning 类公式项，计入应发合计）
	if got := resp.EarningTotal; got != 41812.90 {
		t.Errorf("EarningTotal = %v, want 41812.90", got)
	}
	if got := resp.DeductionTotal; got != 787.10 {
		t.Errorf("DeductionTotal = %v, want 787.10", got)
	}
	if got := resp.NetTotal; got != 20512.90 {
		t.Errorf("NetTotal = %v, want 20512.90", got)
	}

	// 批次行状态：trial_status=1、trial_time 非空、version=1、告警非空
	batch, err := svcCtx.Model.HrmSalaryBatch.FindOneByUuid(t.Context(), nil, batchUuid)
	if err != nil {
		t.Fatalf("find batch: %v", err)
	}
	if batch.TrialStatus != "1" || batch.Version != 1 || !batch.TrialTime.Valid {
		t.Errorf("batch state = %s/%d/%v", batch.TrialStatus, batch.Version, batch.TrialTime.Valid)
	}
	if !batch.Warnings.Valid || batch.Warnings.String == "" {
		t.Errorf("warnings should not be empty")
	}
	// 试算时间回传到 record（列表口径）
	listReq := types.ListRequest{SubjectUuid: batch.SubjectUuid}
	listReq.Current = 1
	listReq.Size = 10
	listResp, err := NewList(context.Background(), svcCtx, nil).List(&listReq)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if listResp.Records[0].TrialTime == "" {
		t.Errorf("list TrialTime should be formatted, got empty")
	}
}

func TestTrialUnlockedMonthWarning(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, _ := setupWorkedExample(t, svcCtx)
	resp := mustTrial(t, svcCtx, batchUuid)
	if !strings.Contains(resp.Warnings, "考勤月报未锁定") {
		t.Errorf("warnings missing 月报未锁定 hint: %q", resp.Warnings)
	}
	if !strings.Contains(resp.Warnings, "社保模块未接入") {
		t.Errorf("warnings missing 社保未配置 hint: %q", resp.Warnings)
	}
}

func TestTrialFormulaItems(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, employeeUuid := setupWorkedExample(t, svcCtx)
	mustTrial(t, svcCtx, batchUuid)
	m := itemAmountsOf(t, svcCtx, batchUuid, employeeUuid)

	// NET_SALARY 公式项 = 应发小计 − 应扣小计 = 21300 − 787.10
	if got := amountOf(t, m, "NET_SALARY"); got != 20512.90 {
		t.Errorf("NET_SALARY = %v, want 20512.90", got)
	}
	// LABOR_COST 公式项 = 应发合计（含 NET_SALARY，与批次应发合计口径一致） + 公司社保0 + 公司公积金0
	// = 41812.90 + 0 + 0
	if got := amountOf(t, m, "LABOR_COST"); got != 41812.90 {
		t.Errorf("LABOR_COST = %v, want 41812.90", got)
	}
	// 来源标记
	if m["NET_SALARY"].Source != SourceFormula {
		t.Errorf("NET_SALARY source = %s, want %s", m["NET_SALARY"].Source, SourceFormula)
	}
	if m["REWARD_BONUS"].Source != SourceReward {
		t.Errorf("REWARD_BONUS source = %s", m["REWARD_BONUS"].Source)
	}
	if m["BASIC_SALARY"].Source != SourceArchive {
		t.Errorf("BASIC_SALARY source = %s", m["BASIC_SALARY"].Source)
	}
	if m["ABSENCE_DEDUCTION"].Source != SourceAttendance {
		t.Errorf("ABSENCE_DEDUCTION source = %s", m["ABSENCE_DEDUCTION"].Source)
	}
	// 取数说明非空（来源详情展示口径）
	if m["ABSENCE_DEDUCTION"].SourceDetail == "" {
		t.Errorf("ABSENCE_DEDUCTION source_detail should explain 取数")
	}
}

func TestTrialIdempotentVersionIncrements(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, employeeUuid := setupWorkedExample(t, svcCtx)

	first := mustTrial(t, svcCtx, batchUuid)
	second := mustTrial(t, svcCtx, batchUuid)
	if first.Version != 1 || second.Version != 2 {
		t.Fatalf("version = %d then %d, want 1 then 2", first.Version, second.Version)
	}
	// 重算覆盖：同一项目只出现一次（全删全插）
	resp, err := NewGetBatchResult(context.Background(), svcCtx, nil).GetBatchResult(&types.GetBatchResultRequest{BatchUuid: batchUuid})
	if err != nil {
		t.Fatalf("get result: %v", err)
	}
	byCode := map[string]int{}
	for _, it := range resp.Items {
		if it.EmployeeUuid == employeeUuid {
			byCode[it.ItemCode]++
		}
	}
	for code, n := range byCode {
		if n != 1 {
			t.Errorf("item %s appears %d times after re-trial, want 1 (覆盖语义)", code, n)
		}
	}
	// 金额与首算一致（幂等重算结果稳定）
	if second.NetTotal != first.NetTotal {
		t.Errorf("re-trial net total drifted: %v vs %v", second.NetTotal, first.NetTotal)
	}
}

func TestTrialSubjectIsolation(t *testing.T) {
	svcCtx := newCtx(t)
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subjectA := seedSubjectRow(t, svcCtx, "主体A")
	subjectB := seedSubjectRow(t, svcCtx, "主体B")
	schemeA := seedSchemeRow(t, svcCtx, "方案A", subjectA)

	items := listAllSalaryItems(t, svcCtx)
	for _, it := range items {
		seedSchemeItemRow(t, svcCtx, schemeA, it.Uuid)
	}

	empA := seedEmployeeRow(t, svcCtx, "张三", "E001", company, dept)
	archive := archiveOf(empA, subjectA, schemeA, "主体A", "方案A")
	archive.EmployeeNo = "E001"
	archive.EmployeeName = "张三"
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(t.Context(), nil, archive); err != nil {
		t.Fatalf("seed archive: %v", err)
	}

	batchA := mustAddBatch(t, svcCtx, "A-2026-03", subjectA, 2026, 3)
	batchB := mustAddBatch(t, svcCtx, "B-2026-03", subjectB, 2026, 3)

	respA := mustTrial(t, svcCtx, batchA)
	if respA.Headcount != 1 {
		t.Errorf("batch A headcount = %d, want 1", respA.Headcount)
	}
	// 主体B 无档案员工 → 试算失败 10684（主体隔离：B 不得沾 A 的员工）
	_, err := RunPayrollTrial(context.Background(), svcCtx, batchB)
	assertErrCode(t, err, 10684, "batch B trial should fail with no archives")
}

func TestTrialDuplicateScopeGuard(t *testing.T) {
	svcCtx := newCtx(t)
	subject := seedSubjectRow(t, svcCtx, "唯一主体")
	mustAddBatch(t, svcCtx, "首批次", subject, 2026, 3)
	_, err := NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{
		BatchName: "重复批次", SubjectUuid: subject, BatchYear: 2026, BatchMonth: 3,
	})
	assertErrCode(t, err, 10683, "duplicate scope should reject")
}

func TestTrialCrudAndValidation(t *testing.T) {
	svcCtx := newCtx(t)
	subject := seedSubjectRow(t, svcCtx, "CRUD主体")

	// 参数校验
	_, err := NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{BatchName: "", SubjectUuid: subject, BatchYear: 2026, BatchMonth: 3})
	assertErrCode(t, err, 10681, "empty name")
	_, err = NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{BatchName: "x", SubjectUuid: "no-such", BatchYear: 2026, BatchMonth: 3})
	assertErrCode(t, err, 10682, "unknown subject")
	_, err = NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{BatchName: "x", SubjectUuid: subject, BatchYear: 2026, BatchMonth: 13})
	assertErrCode(t, err, 10681, "bad month")

	batchUuid := mustAddBatch(t, svcCtx, "编辑前", subject, 2026, 3)
	// 编号格式 SA{批次月份首日yyyyMMdd}{4位流水}（复用 hrm_bill_sequence bill_type_code=SA）
	batch, err := svcCtx.Model.HrmSalaryBatch.FindOneByUuid(t.Context(), nil, batchUuid)
	if err != nil {
		t.Fatalf("find batch: %v", err)
	}
	if batch.BatchNo != "SA202603010001" {
		t.Errorf("batch no = %s, want SA202603010001", batch.BatchNo)
	}
	if batch.SubjectName != "CRUD主体" {
		t.Errorf("subject name snapshot = %s", batch.SubjectName)
	}

	// 编辑（名称+备注）
	if _, err := NewEdit(context.Background(), svcCtx, nil).Edit(&types.EditRequest{
		Uuid: batchUuid, BatchName: "编辑后", BatchType: "monthly", BatchYear: 2026, BatchMonth: 3, Remark: "备注",
	}); err != nil {
		t.Fatalf("edit: %v", err)
	}
	batch, _ = svcCtx.Model.HrmSalaryBatch.FindOneByUuid(t.Context(), nil, batchUuid)
	if batch.BatchName != "编辑后" || batch.Remark != "备注" {
		t.Errorf("edit not applied: %s %s", batch.BatchName, batch.Remark)
	}

	// 删除（明细一并删）
	if _, err := NewDelete(context.Background(), svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{batchUuid}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svcCtx.Model.HrmSalaryBatch.FindOneByUuid(t.Context(), nil, batchUuid); err != hrm_salary_batch.ErrNotFound {
		t.Errorf("batch should be deleted, got %v", err)
	}
	_, err = NewGetBatchResult(context.Background(), svcCtx, nil).GetBatchResult(&types.GetBatchResultRequest{BatchUuid: batchUuid})
	assertErrCode(t, err, 10680, "result after delete")
}

func TestTrialNoResultGuard(t *testing.T) {
	svcCtx := newCtx(t)
	subject := seedSubjectRow(t, svcCtx, "未试算主体")
	batchUuid := mustAddBatch(t, svcCtx, "未试算批次", subject, 2026, 3)
	_, err := NewGetBatchResult(context.Background(), svcCtx, nil).GetBatchResult(&types.GetBatchResultRequest{BatchUuid: batchUuid})
	assertErrCode(t, err, 10685, "result before trial")
}

func TestTrialListFilters(t *testing.T) {
	svcCtx := newCtx(t)
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subjectA := seedSubjectRow(t, svcCtx, "筛选A")
	subjectB := seedSubjectRow(t, svcCtx, "筛选B")
	schemeA := seedSchemeRow(t, svcCtx, "筛选方案", subjectA)
	items := listAllSalaryItems(t, svcCtx)
	for _, it := range items {
		seedSchemeItemRow(t, svcCtx, schemeA, it.Uuid)
	}
	emp := seedEmployeeRow(t, svcCtx, "筛选员", "E010", company, dept)
	archive := archiveOf(emp, subjectA, schemeA, "筛选A", "筛选方案")
	archive.EmployeeNo = "E010"
	archive.EmployeeName = "筛选员"
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(t.Context(), nil, archive); err != nil {
		t.Fatalf("seed archive: %v", err)
	}
	b1 := mustAddBatch(t, svcCtx, "三月批次", subjectA, 2026, 3)
	b2 := mustAddBatch(t, svcCtx, "四月批次", subjectB, 2026, 4)
	mustTrial(t, svcCtx, b1)

	// name Like
	filterReq := types.ListRequest{BatchName: "三月"}
	filterReq.Current = 1
	filterReq.Size = 10
	resp, err := NewList(context.Background(), svcCtx, nil).List(&filterReq)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.Total != 1 || resp.Records[0].Uuid != b1 {
		t.Errorf("name filter: total=%d", resp.Total)
	}
	// subject
	subjectReq := types.ListRequest{SubjectUuid: subjectB}
	subjectReq.Current = 1
	subjectReq.Size = 10
	resp, _ = NewList(context.Background(), svcCtx, nil).List(&subjectReq)
	if resp.Total != 1 || resp.Records[0].Uuid != b2 {
		t.Errorf("subject filter: total=%d", resp.Total)
	}
	// trialStatus
	trialReq := types.ListRequest{TrialStatus: "1"}
	trialReq.Current = 1
	trialReq.Size = 10
	resp, _ = NewList(context.Background(), svcCtx, nil).List(&trialReq)
	if resp.Total != 1 {
		t.Errorf("trialStatus filter: total=%d", resp.Total)
	}
	// 年月升序（时间线）
	allReq := types.ListRequest{}
	allReq.Current = 1
	allReq.Size = 10
	resp, _ = NewList(context.Background(), svcCtx, nil).List(&allReq)
	if len(resp.Records) != 2 || resp.Records[0].BatchNo >= resp.Records[1].BatchNo {
		t.Errorf("order by year-month: %v", resp.Records)
	}
}

func TestTrialPerfCoefficient(t *testing.T) {
	svcCtx := newCtx(t)
	resetPerfUuids()
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subject := seedSubjectRow(t, svcCtx, "绩效主体")
	scheme := seedSchemeRow(t, svcCtx, "绩效方案", subject)
	items := listAllSalaryItems(t, svcCtx)
	for _, it := range items {
		seedSchemeItemRow(t, svcCtx, scheme, it.Uuid)
	}
	emp := seedEmployeeRow(t, svcCtx, "李四", "E002", company, dept)
	archive := archiveOf(emp, subject, scheme, "绩效主体", "绩效方案")
	archive.EmployeeNo = "E002"
	archive.EmployeeName = "李四"
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(t.Context(), nil, archive); err != nil {
		t.Fatalf("seed archive: %v", err)
	}
	// 系数制 + 等级 B（final_score 88 仅作展示）→ 系数 1.0 × 档案固定薪资 10000
	seedPerfChain(t, svcCtx, emp, "coefficient", "B", 88)
	batchUuid := mustAddBatch(t, svcCtx, "绩效批次", subject, 2026, 3)
	mustTrial(t, svcCtx, batchUuid)
	m := itemAmountsOf(t, svcCtx, batchUuid, emp)

	if got := amountOf(t, m, "MERIT_SALARY"); got != 10000 {
		t.Errorf("MERIT_SALARY coefficient = %v, want 10000 (1.0 × 10000)", got)
	}
	if m["MERIT_SALARY"].Source != SourcePerf {
		t.Errorf("MERIT_SALARY source = %s", m["MERIT_SALARY"].Source)
	}
	if !strings.Contains(m["MERIT_SALARY"].SourceDetail, "B") {
		t.Errorf("MERIT_SALARY detail should mention grade: %s", m["MERIT_SALARY"].SourceDetail)
	}
}

func TestTrialPerfAmountMode(t *testing.T) {
	svcCtx := newCtx(t)
	resetPerfUuids()
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subject := seedSubjectRow(t, svcCtx, "金额制主体")
	scheme := seedSchemeRow(t, svcCtx, "金额制方案", subject)
	items := listAllSalaryItems(t, svcCtx)
	for _, it := range items {
		seedSchemeItemRow(t, svcCtx, scheme, it.Uuid)
	}
	emp := seedEmployeeRow(t, svcCtx, "王五", "E003", company, dept)
	archive := archiveOf(emp, subject, scheme, "金额制主体", "金额制方案")
	archive.EmployeeNo = "E003"
	archive.EmployeeName = "王五"
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(t.Context(), nil, archive); err != nil {
		t.Fatalf("seed archive: %v", err)
	}
	// 金额制：final_score 直取 = 1234.56
	seedPerfChain(t, svcCtx, emp, "amount", "A", 1234.56)
	batchUuid := mustAddBatch(t, svcCtx, "金额制批次", subject, 2026, 3)
	mustTrial(t, svcCtx, batchUuid)
	m := itemAmountsOf(t, svcCtx, batchUuid, emp)
	if got := amountOf(t, m, "MERIT_SALARY"); got != 1234.56 {
		t.Errorf("MERIT_SALARY amount mode = %v, want 1234.56", got)
	}
}

func TestTrialPerfMissingFallback(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, employeeUuid := setupWorkedExample(t, svcCtx)
	resp := mustTrial(t, svcCtx, batchUuid)
	_ = resp
	m := itemAmountsOf(t, svcCtx, batchUuid, employeeUuid)
	// 无绩效结果：MERIT_SALARY 兜底 0（assertErrCode 不适用——引擎不失败，告警兜底）
	if got := amountOf(t, m, "MERIT_SALARY"); got != 0 {
		t.Errorf("MERIT_SALARY missing perf = %v, want 0", got)
	}
}

func TestTrialTaxWithOpening(t *testing.T) {
	svcCtx := newCtx(t)
	resetPerfUuids()
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subject := seedSubjectRow(t, svcCtx, "个税主体")
	scheme := seedSchemeRow(t, svcCtx, "个税方案", subject)
	items := listAllSalaryItems(t, svcCtx)
	for _, it := range items {
		seedSchemeItemRow(t, svcCtx, scheme, it.Uuid)
	}
	emp := seedEmployeeRow(t, svcCtx, "赵六", "E004", company, dept)
	archive := archiveOf(emp, subject, scheme, "个税主体", "个税方案")
	archive.EmployeeNo = "E004"
	archive.EmployeeName = "赵六"
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(t.Context(), nil, archive); err != nil {
		t.Fatalf("seed archive: %v", err)
	}
	// 期初：截至 2 月（months=2），累计收入 30000、已缴 0
	seedTaxOpeningRow(t, svcCtx, emp, subject, "2026", 2, 30000, 0, 0, 0, 0)
	batchUuid := mustAddBatch(t, svcCtx, "个税批次", subject, 2026, 3)
	mustTrial(t, svcCtx, batchUuid)
	m := itemAmountsOf(t, svcCtx, batchUuid, emp)

	// 累计口径（months=3 含本月；本布景无奖惩/缺勤 → 应发小计 = 10000+10000+0+500+500 = 21000）：
	//   累计收入 = 21000 + 30000(期初) = 51000；减除 5000×3=15000 → 应税 36000
	//   税额 = 36000×0.10 − 2520 = 1080；TAX 行 = 1080
	if got := amountOf(t, m, "TAX"); got != 1080 {
		t.Errorf("TAX = %v, want 1080", got)
	}
	if m["TAX"].Source != SourceTax {
		t.Errorf("TAX source = %s", m["TAX"].Source)
	}
}

func seedReward(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, rewardType string, amount float64) {
	t.Helper()
	if err := svcCtx.Model.HrmRewardBill.InsertV2(t.Context(), nil, &hrm_reward_bill.HrmRewardBill{
		Uuid: uuid.New().String(), BillNo: "HR215-TEST" + rewardType, Status: "approved",
		EmployeeUuid: employeeUuid, RewardType: rewardType, Amount: amount,
		SalaryMonth: "2026-03", Synced: 1,
	}); err != nil {
		t.Fatalf("seed reward: %v", err)
	}
}
