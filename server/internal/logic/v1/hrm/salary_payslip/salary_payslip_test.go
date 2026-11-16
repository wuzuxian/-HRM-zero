package salary_payslip

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_payslip"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payslip"
)

// ===== svctest harness（与 salary_batch 测试同款装配） =====

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

// ===== 布景 helpers =====

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

// seedBatch 批次行；confirmed=false 模拟未确认（HRM-38 写回前状态）。
func seedBatch(t *testing.T, svcCtx *svc.ServiceContext, subjectUuid string, year, month int64,
	confirmed bool, payStatus, payslipStatus, warnings string) string {
	t.Helper()
	confirm := "0"
	if confirmed {
		confirm = "1"
	}
	u := uuid.New().String()
	batchNo := fmt.Sprintf("SA%d03010001", year)
	if _, err := svcCtx.Model.HrmSalaryBatch.Insert(t.Context(), nil, &hrm_salary_batch.HrmSalaryBatch{
		Uuid: u, BatchNo: batchNo, BatchName: "2026-03 月度批次", BatchType: "monthly",
		SubjectUuid: subjectUuid, SubjectName: "3月主体",
		BatchYear: year, BatchMonth: month,
		TrialStatus: "1", ConfirmStatus: confirm, PayStatus: payStatus, PayslipStatus: payslipStatus,
		Warnings: nullStringOf(warnings),
	}); err != nil {
		t.Fatalf("seed batch: %v", err)
	}
	return u
}

// nullStringOf 空串 → 无效 sql.NullString（批次无告警），否则带值。
func nullStringOf(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// seedBatchItem 批次试算明细行（amount 应扣类为正数，HRM-37 契约）。
func seedBatchItem(t *testing.T, svcCtx *svc.ServiceContext, batchUuid, employeeUuid, employeeNo, employeeName, orgName,
	itemCode, itemName, itemClass string, amount float64, calcOrder int64) {
	t.Helper()
	if err := svcCtx.Model.HrmSalaryBatchItem.InsertV2(t.Context(), nil, &hrm_salary_batch_item.HrmSalaryBatchItem{
		Uuid: uuid.New().String(), BatchUuid: batchUuid,
		EmployeeUuid: employeeUuid, EmployeeNo: employeeNo, EmployeeName: employeeName, OrgName: orgName,
		ItemCode: itemCode, ItemName: itemName, ItemClass: itemClass, Amount: amount, CalcOrder: calcOrder,
	}); err != nil {
		t.Fatalf("seed batch item %s: %v", itemCode, err)
	}
}

func mustPublish(t *testing.T, svcCtx *svc.ServiceContext, batchUuid string) *types.PublishResponse {
	t.Helper()
	resp, err := NewPublish(context.Background(), svcCtx, nil).Publish(&types.PublishRequest{BatchUuid: batchUuid})
	if err != nil {
		t.Fatalf("publish %s: %v", batchUuid, err)
	}
	return resp
}

// findPayslipOf 按 批次+员工 取工资条行。
func findPayslipOf(t *testing.T, svcCtx *svc.ServiceContext, batchUuid, employeeUuid string) *hrm_salary_payslip.HrmSalaryPayslip {
	t.Helper()
	row, err := svcCtx.Model.HrmSalaryPayslip.FindOneByCondition(t.Context(), nil,
		condition.NewChain().
			Equal(hrm_salary_payslip.BatchUuid, batchUuid).
			Equal(hrm_salary_payslip.EmployeeUuid, employeeUuid).
			Build()...)
	if err != nil {
		t.Fatalf("find payslip: %v", err)
	}
	return row
}

// workedExample 布景：已确认批次（payslip_status=1、发薪=0、告警非空）+ 员工张三
// （应发 BASIC 10000 / FIXED_ALLOWANCE 500，应扣 TAX 300，成本 SI_COMPANY 1200）+
// 员工李四（应发 BASIC 8000，应扣 TAX 200）。
func workedExample(t *testing.T, svcCtx *svc.ServiceContext) (batchUuid, empA, empB string) {
	t.Helper()
	company := seedOrgRow(t, svcCtx, "甲公司", "1")
	dept := seedOrgRow(t, svcCtx, "技术部", "2")
	subject := seedSubjectRow(t, svcCtx, "3月主体")

	empA = seedEmployeeRow(t, svcCtx, "张三", "E001", company, dept)
	empB = seedEmployeeRow(t, svcCtx, "李四", "E002", company, dept)

	batchUuid = seedBatch(t, svcCtx, subject, 2026, 3, true, "0", "1", "张三: 绩效结果缺失")

	seedBatchItem(t, svcCtx, batchUuid, empA, "E001", "张三", "技术部", "BASIC_SALARY", "基本工资", "earning", 10000, 1)
	seedBatchItem(t, svcCtx, batchUuid, empA, "E001", "张三", "技术部", "FIXED_ALLOWANCE", "固定补贴", "earning", 500, 2)
	seedBatchItem(t, svcCtx, batchUuid, empA, "E001", "张三", "技术部", "TAX", "个税", "deduction", 300, 10)
	seedBatchItem(t, svcCtx, batchUuid, empA, "E001", "张三", "技术部", "SI_COMPANY", "公司社保", "cost", 1200, 13)
	seedBatchItem(t, svcCtx, batchUuid, empB, "E002", "李四", "技术部", "BASIC_SALARY", "基本工资", "earning", 8000, 1)
	seedBatchItem(t, svcCtx, batchUuid, empB, "E002", "李四", "技术部", "TAX", "个税", "deduction", 200, 10)

	return batchUuid, empA, empB
}

// ===== 发布：从批次明细行生成工资条 =====

func TestPublishGeneratesPayslipsWithItemJSON(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, empA, empB := workedExample(t, svcCtx)

	resp := mustPublish(t, svcCtx, batchUuid)
	if resp.Count != 2 {
		t.Fatalf("publish count = %d, want 2", resp.Count)
	}

	// 张三工资条聚合口径
	payA := findPayslipOf(t, svcCtx, batchUuid, empA)
	if payA.EarningTotal != 10500 {
		t.Errorf("A earningTotal = %v, want 10500", payA.EarningTotal)
	}
	if payA.DeductionTotal != 300 {
		t.Errorf("A deductionTotal = %v, want 300", payA.DeductionTotal)
	}
	if payA.NetTotal != 10200 {
		t.Errorf("A netTotal = %v, want 10200", payA.NetTotal)
	}
	if payA.LaborCostTotal != 1200 {
		t.Errorf("A laborCostTotal = %v, want 1200", payA.LaborCostTotal)
	}
	if payA.Published != 1 || payA.ReadStatus != 0 {
		t.Errorf("A published/read = %d/%d, want 1/0", payA.Published, payA.ReadStatus)
	}
	if payA.Title == "" {
		t.Errorf("A title should not be empty")
	}

	// 明细 JSON 按薪资项目快照（独立真值断言，不重算实现口径）
	var items []types.HrmSalaryPayslipItem
	if err := json.Unmarshal([]byte(payA.ItemsJson), &items); err != nil {
		t.Fatalf("unmarshal items_json: %v (raw=%s)", err, payA.ItemsJson)
	}
	if len(items) != 4 {
		t.Fatalf("A items len = %d, want 4 (raw=%s)", len(items), payA.ItemsJson)
	}
	if items[0].ItemCode != "BASIC_SALARY" || items[0].Amount != 10000 || items[0].ItemClass != "earning" {
		t.Errorf("item[0] = %+v, want BASIC_SALARY/10000/earning", items[0])
	}
	if items[2].ItemCode != "TAX" || items[2].Amount != 300 || items[2].ItemClass != "deduction" {
		t.Errorf("item[2] = %+v, want TAX/300/deduction", items[2])
	}
	if items[3].ItemCode != "SI_COMPANY" || items[3].Amount != 1200 || items[3].ItemClass != "cost" {
		t.Errorf("item[3] = %+v, want SI_COMPANY/1200/cost", items[3])
	}

	// 李四工资条 NET = 8000 - 200 = 7800
	payB := findPayslipOf(t, svcCtx, batchUuid, empB)
	if payB.NetTotal != 7800 {
		t.Errorf("B netTotal = %v, want 7800", payB.NetTotal)
	}
}

// ===== 发布守卫：批次未确认拒绝 =====

func TestPublishRejectsUnconfirmedBatch(t *testing.T) {
	svcCtx := newCtx(t)
	subject := seedSubjectRow(t, svcCtx, "3月主体")
	batchUuid := seedBatch(t, svcCtx, subject, 2026, 3, false, "0", "0", "")

	_, err := NewPublish(context.Background(), svcCtx, nil).Publish(&types.PublishRequest{BatchUuid: batchUuid})
	assertErrCode(t, err, 10697, "unconfirmed batch publish")
}

// 批次不存在 → 10695；批次无试算明细 → 10698。
func TestPublishGuards(t *testing.T) {
	svcCtx := newCtx(t)

	_, err := NewPublish(context.Background(), svcCtx, nil).Publish(&types.PublishRequest{BatchUuid: uuid.New().String()})
	assertErrCode(t, err, 10695, "missing batch")

	subject := seedSubjectRow(t, svcCtx, "3月主体")
	batchUuid := seedBatch(t, svcCtx, subject, 2026, 4, true, "0", "0", "")
	_, err = NewPublish(context.Background(), svcCtx, nil).Publish(&types.PublishRequest{BatchUuid: batchUuid})
	assertErrCode(t, err, 10698, "batch without items")
}

// 重发布幂等覆盖（不产生重复行，UNIQUE(batch_uuid,employee_uuid)）。
func TestPublishIdempotentOverwrite(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, _, _ := workedExample(t, svcCtx)

	mustPublish(t, svcCtx, batchUuid)
	mustPublish(t, svcCtx, batchUuid)

	req := types.ListRequest{BatchUuid: batchUuid}
	req.Current = 1
	req.Size = 10
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("after re-publish total = %d, want 2 (idempotent overwrite)", resp.Total)
	}
}

// ===== 已读留痕：幂等不覆盖首次 =====

func TestReadMarksFirstReadTimeIdempotent(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, empA, _ := workedExample(t, svcCtx)
	mustPublish(t, svcCtx, batchUuid)

	payA := findPayslipOf(t, svcCtx, batchUuid, empA)

	if err := MarkRead(context.Background(), svcCtx.SqlxConn, payA.Uuid); err != nil {
		t.Fatalf("first read: %v", err)
	}
	after1, err := svcCtx.Model.HrmSalaryPayslip.FindOneByUuid(t.Context(), nil, payA.Uuid)
	if err != nil {
		t.Fatalf("find after first read: %v", err)
	}
	if after1.ReadStatus != 1 || !after1.ReadTime.Valid {
		t.Fatalf("after first read: read=%d timeValid=%v, want 1/true", after1.ReadStatus, after1.ReadTime.Valid)
	}
	firstTime := after1.ReadTime.Time.Format("2006-01-02 15:04:05")

	// 第二次阅读：幂等——已读时间不覆盖
	if err := MarkRead(context.Background(), svcCtx.SqlxConn, payA.Uuid); err != nil {
		t.Fatalf("second read: %v", err)
	}
	after2, err := svcCtx.Model.HrmSalaryPayslip.FindOneByUuid(t.Context(), nil, payA.Uuid)
	if err != nil {
		t.Fatalf("find after second read: %v", err)
	}
	if after2.ReadTime.Time.Format("2006-01-02 15:04:05") != firstTime {
		t.Fatalf("second read overwrote read_time: %s -> %s", firstTime, after2.ReadTime.Time.Format("2006-01-02 15:04:05"))
	}
}

// ===== 薪资统计台账：聚合口径 worked example + 统计卡 =====

func TestLedgerAggregationAndStats(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, _, _ := workedExample(t, svcCtx)
	mustPublish(t, svcCtx, batchUuid)

	req := types.LedgerRequest{BatchUuid: batchUuid}
	req.Current = 1
	req.Size = 10
	resp, err := NewLedger(context.Background(), svcCtx, nil).Ledger(&req)
	if err != nil {
		t.Fatalf("ledger: %v", err)
	}

	// 统计卡：员工数=2 / 应发=10500+8000 / 应扣=300+200 / 实发=10200+7800 / 人工成本=1200
	if resp.EmployeeCount != 2 {
		t.Errorf("EmployeeCount = %d, want 2", resp.EmployeeCount)
	}
	if resp.EarningTotal != 18500 {
		t.Errorf("EarningTotal = %v, want 18500", resp.EarningTotal)
	}
	if resp.DeductionTotal != 500 {
		t.Errorf("DeductionTotal = %v, want 500", resp.DeductionTotal)
	}
	if resp.NetTotal != 18000 {
		t.Errorf("NetTotal = %v, want 18000", resp.NetTotal)
	}
	if resp.LaborCostTotal != 1200 {
		t.Errorf("LaborCostTotal = %v, want 1200", resp.LaborCostTotal)
	}

	// 台账行：批次/部门快照 + 异常透传（批次 warnings）+ 已发条/已发薪透传
	if len(resp.Records) != 2 {
		t.Fatalf("ledger records = %d, want 2", len(resp.Records))
	}
	var rowA types.LedgerRow
	for _, r := range resp.Records {
		if r.EmployeeName == "张三" {
			rowA = r
		}
	}
	if rowA.EmployeeName != "张三" {
		t.Fatalf("rowA (张三) missing in ledger records: %+v", resp.Records)
	}
	if rowA.OrgName != "技术部" {
		t.Errorf("rowA org = %s, want 技术部", rowA.OrgName)
	}
	if rowA.BatchNo != "SA202603010001" {
		t.Errorf("rowA batchNo = %s, want SA202603010001", rowA.BatchNo)
	}
	if rowA.Payslipped != 1 {
		t.Errorf("rowA payslipped = %d, want 1 (批次已发布工资条)", rowA.Payslipped)
	}
	if rowA.Paid != 0 {
		t.Errorf("rowA paid = %d, want 0 (发薪登记 HRM-38 未写回)", rowA.Paid)
	}
	if rowA.LaborCostTotal != 1200 {
		t.Errorf("rowA laborCost = %v, want 1200", rowA.LaborCostTotal)
	}
	if rowA.Abnormal == "" {
		t.Errorf("rowA abnormal should carry batch warnings text")
	}
}

// 台账筛选：仅看异常。
func TestLedgerOnlyAbnormalFilter(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, _, _ := workedExample(t, svcCtx)
	mustPublish(t, svcCtx, batchUuid)

	req := types.LedgerRequest{BatchUuid: batchUuid, OnlyAbnormal: true}
	req.Current = 1
	req.Size = 10
	resp, err := NewLedger(context.Background(), svcCtx, nil).Ledger(&req)
	if err != nil {
		t.Fatalf("ledger: %v", err)
	}
	if len(resp.Records) != 2 || resp.EmployeeCount != 2 {
		t.Fatalf("onlyAbnormal records = %d/%d, want 2/2 (批次告警透传全员)", len(resp.Records), resp.EmployeeCount)
	}
	for _, r := range resp.Records {
		if r.Abnormal == "" {
			t.Errorf("row %s abnormal should carry batch warnings", r.EmployeeName)
		}
	}
}

// 台账筛选：姓名。
func TestLedgerNameFilter(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, _, _ := workedExample(t, svcCtx)
	mustPublish(t, svcCtx, batchUuid)

	req := types.LedgerRequest{BatchUuid: batchUuid, EmployeeName: "李四"}
	req.Current = 1
	req.Size = 10
	resp, err := NewLedger(context.Background(), svcCtx, nil).Ledger(&req)
	if err != nil {
		t.Fatalf("ledger: %v", err)
	}
	if len(resp.Records) != 1 || resp.Records[0].EmployeeName != "李四" {
		t.Fatalf("name filter records = %+v, want 1 李四", resp.Records)
	}
}

// ===== 管理端列表筛选 =====

func TestPayslipListFilters(t *testing.T) {
	svcCtx := newCtx(t)
	batchUuid, empA, _ := workedExample(t, svcCtx)
	mustPublish(t, svcCtx, batchUuid)

	payA := findPayslipOf(t, svcCtx, batchUuid, empA)
	_ = MarkRead(context.Background(), svcCtx.SqlxConn, payA.Uuid)

	// 年月筛选
	req := types.ListRequest{YearMonth: "2026-03"}
	req.Current = 1
	req.Size = 10
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("yearMonth filter total = %d, want 2", resp.Total)
	}

	// 已读筛选
	req2 := types.ListRequest{ReadStatus: "1"}
	req2.Current = 1
	req2.Size = 10
	resp2, err := NewList(context.Background(), svcCtx, nil).List(&req2)
	if err != nil {
		t.Fatalf("list read: %v", err)
	}
	if resp2.Total != 1 {
		t.Errorf("read filter total = %d, want 1", resp2.Total)
	}

	// 姓名筛选
	req3 := types.ListRequest{EmployeeName: "李四"}
	req3.Current = 1
	req3.Size = 10
	resp3, err := NewList(context.Background(), svcCtx, nil).List(&req3)
	if err != nil {
		t.Fatalf("list name: %v", err)
	}
	if resp3.Total != 1 {
		t.Errorf("name filter total = %d, want 1", resp3.Total)
	}
}
