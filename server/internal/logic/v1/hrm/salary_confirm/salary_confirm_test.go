package salary_confirm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	salary_batch "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_archive"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_confirm"
	types_batch "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
)

// ===== 布景（org/subject/scheme/item/employee/archive + 批次 + 试算） =====

func seedOrgRow(t *testing.T, name, orgType string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmOrg.Insert(testCtx, nil, &hrm_org.HrmOrg{
		Uuid: u, OrgName: name, OrgType: orgType, OrderNum: 1, Status: "1",
	}); err != nil {
		t.Fatalf("seed org %s: %v", name, err)
	}
	return u
}

func seedSubjectRow(t *testing.T, name string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalarySubject.Insert(testCtx, nil, &hrm_salary_subject.HrmSalarySubject{
		Uuid: u, SubjectName: name, OrgUuid: uuid.New().String(), OrgName: "甲公司", Status: "1",
	}); err != nil {
		t.Fatalf("seed subject %s: %v", name, err)
	}
	return u
}

func seedSchemeAllItems(t *testing.T, subjectUuid string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalaryScheme.Insert(testCtx, nil, &hrm_salary_scheme.HrmSalaryScheme{
		Uuid: u, SchemeName: "标准方案", SubjectUuid: subjectUuid, SubjectName: "主体",
		PersonStatus: "1", SettleCycle: "monthly", PayDay: 1, Currency: "CNY", Status: "1",
	}); err != nil {
		t.Fatalf("seed scheme: %v", err)
	}
	items, err := svcCtx.Model.HrmSalaryItem.FindByCondition(testCtx, nil,
		condition.NewChain().Build()...)
	if err != nil {
		t.Fatalf("list salary items: %v", err)
	}
	for _, it := range items {
		if err := svcCtx.Model.HrmSalarySchemeItem.InsertV2(testCtx, nil, &hrm_salary_scheme_item.HrmSalarySchemeItem{
			Uuid: uuid.New().String(), SchemeUuid: u, ItemUuid: it.Uuid,
		}); err != nil {
			t.Fatalf("seed scheme item: %v", err)
		}
	}
	return u
}

func seedEmployeeWithArchive(t *testing.T, subjectUuid, schemeUuid string) string {
	t.Helper()
	empUuid := uuid.New().String()
	if err := svcCtx.Model.HrmEmployee.InsertV2(testCtx, nil, &hrm_employee.HrmEmployee{
		Uuid: empUuid, EmployeeNo: "E001", Name: "张三",
		CompanyUuid: seedOrgRow(t, "甲公司"+uuid.New().String()[:8], "1"),
		OrgUuid:     seedOrgRow(t, "技术部"+uuid.New().String()[:8], "2"),
		PersonStatus: "1",
	}); err != nil {
		t.Fatalf("seed employee: %v", err)
	}
	if _, err := svcCtx.Model.HrmSalaryArchive.Insert(testCtx, nil, &hrm_salary_archive.HrmSalaryArchive{
		Uuid:           uuid.New().String(),
		EmployeeUuid:   empUuid,
		EmployeeNo:     "E001",
		EmployeeName:   "张三",
		SubjectUuid:    subjectUuid,
		SubjectName:    "主体",
		SchemeUuid:     schemeUuid,
		SchemeName:     "标准方案",
		FixedSalary:    10000,
		FixedAllowance: 500,
		Status:         "1",
		EffectiveDate:  "2026-01-01",
	}); err != nil {
		t.Fatalf("seed archive: %v", err)
	}
	return empUuid
}

// setupTrialledBatch 最小布景：员工档案 + 批次 + 试算完成（trial_status='1'）。
func setupTrialledBatch(t *testing.T) (batchUuid string) {
	t.Helper()
	subject := seedSubjectRow(t, "3月主体")
	scheme := seedSchemeAllItems(t, subject)
	seedEmployeeWithArchive(t, subject, scheme)
	resp, err := salary_batch.NewAdd(testCtx, svcCtx, nil).Add(&types_batch.AddRequest{
		BatchName: "2026-03 月度批次", SubjectUuid: subject, BatchYear: 2026, BatchMonth: 3,
	})
	if err != nil {
		t.Fatalf("add batch: %v", err)
	}
	if _, err := salary_batch.RunPayrollTrial(testCtx, svcCtx, resp.Uuid); err != nil {
		t.Fatalf("trial: %v", err)
	}
	return resp.Uuid
}

// ===== 守卫 + 事务 + 幂等 + 快照 =====

// TestSubmitGuardUnTrialledBatch 守卫：批次未试算拒绝（确认审批门槛=已试算）。
func TestSubmitGuardUnTrialledBatch(t *testing.T) {
	svcCtx := newSvcCtx(t)
	subject := seedSubjectRow(t, "未试算主体")
	resp, err := salary_batch.NewAdd(testCtx, svcCtx, nil).Add(&types_batch.AddRequest{
		BatchName: "未试算批次", SubjectUuid: subject, BatchYear: 2026, BatchMonth: 4,
	})
	if err != nil {
		t.Fatalf("add batch: %v", err)
	}
	_, err = NewSubmit(testCtx, svcCtx, nil).Submit(&types.SubmitRequest{BatchUuid: resp.Uuid})
	if err == nil {
		t.Fatalf("expected 未试算拒绝, got nil")
	}
	// 未试算守卫走本域 10711（门槛语义=未试算不可确认审批）。
	assertErrCode(t, err, 10711, "未试算批次提交确认审批")
}

// TestSubmitGuardAlreadyConfirmed 守卫：批次已确认（confirm_status='1'）拒绝重复提交。
func TestSubmitGuardAlreadyConfirmed(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupTrialledBatch(t)

	billUuid := mustSubmit(t, svcCtx, batchUuid)
	mustApprove(t, svcCtx, billUuid)

	_, err := NewSubmit(testCtx, svcCtx, nil).Submit(&types.SubmitRequest{BatchUuid: batchUuid})
	assertErrCode(t, err, 10711, "已确认批次重复提交")
}

// TestApproveSetsConfirmFlagAtomic 审批通过事务内：批次 confirm_status='1' +
// approve_time 快照；驳回不推进 confirm 标记。
func TestApproveSetsConfirmFlagAtomic(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupTrialledBatch(t)

	billUuid := mustSubmit(t, svcCtx, batchUuid)

	// 审批中：confirm 标记不动（HRM-37 未推进批次落 ''，与 '0' 同义）
	if b := batchOf(t, svcCtx, batchUuid); b.ConfirmStatus != "" && b.ConfirmStatus != "0" {
		t.Fatalf("confirm_status during approving = %s, want ''/0", b.ConfirmStatus)
	}

	mustApprove(t, svcCtx, billUuid)

	batch := batchOf(t, svcCtx, batchUuid)
	if batch.ConfirmStatus != "1" {
		t.Fatalf("confirm_status after approve = %s, want 1", batch.ConfirmStatus)
	}
	row := billOf(t, svcCtx, billUuid)
	if row.Status != bill.StatusApproved {
		t.Fatalf("bill status = %s, want approved", row.Status)
	}
	if row.ApproveTime == "" {
		t.Fatalf("approve_time empty after approve")
	}
	if row.BillNo == "" || row.BillNo[:5] != "HR209" {
		t.Fatalf("bill_no = %q, want HR209- prefix", row.BillNo)
	}
}

// TestRejectDoesNotConfirm 驳回：批次 confirm_status 保持 '0'，可重新提交。
func TestRejectDoesNotConfirm(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupTrialledBatch(t)

	billUuid := mustSubmit(t, svcCtx, batchUuid)
	if err := bill.RejectBill(testCtx, svcCtx.SqlxConn, BillTypeCode, billUuid,
		bill.Operator{Name: "审批人"}, "驳回"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	if b := batchOf(t, svcCtx, batchUuid); b.ConfirmStatus != "" && b.ConfirmStatus != "0" {
		t.Fatalf("confirm_status after reject = %s, want ''/0", b.ConfirmStatus)
	}
	// 驳回后可重新提交（重新聚合快照）
	reUuid := mustSubmit(t, svcCtx, batchUuid)
	if reUuid != billUuid {
		t.Fatalf("re-submit should reuse same bill row, got %s want %s", reUuid, billUuid)
	}
}

// TestSubmitIdempotentOneBillPerBatch 一批次一张审批单：审批中重复提交拒绝；
// 撤销后可重新提交；审批通过（applied）后 confirm 幂等。
func TestSubmitIdempotentOneBillPerBatch(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupTrialledBatch(t)

	billUuid := mustSubmit(t, svcCtx, batchUuid)

	// 审批中重复提交拒绝
	_, err := NewSubmit(testCtx, svcCtx, nil).Submit(&types.SubmitRequest{BatchUuid: batchUuid})
	assertErrCode(t, err, 10711, "审批中重复提交")

	// 撤销 → 重新提交（同单复用）
	if err := bill.RevokeBill(testCtx, svcCtx.SqlxConn, BillTypeCode, billUuid,
		bill.Operator{Name: "提交人"}, "撤回"); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	reUuid := mustSubmit(t, svcCtx, batchUuid)
	if reUuid != billUuid {
		t.Fatalf("re-submit after revoke should reuse bill, got %s want %s", reUuid, billUuid)
	}

	// 审批通过（幂等 applied：confirm 已 '1' 再走一遍写回无害）
	mustApprove(t, svcCtx, reUuid)
	if b := batchOf(t, svcCtx, batchUuid); b.ConfirmStatus != "1" {
		t.Fatalf("confirm_status = %s, want 1", b.ConfirmStatus)
	}
}

// TestSubmitSnapshotFixed 提交时快照批次 编号/名称/主体/年月/人数/实发合计；
// 提交后批次改名不影响单据快照（金额快照口径）。
func TestSubmitSnapshotFixed(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupTrialledBatch(t)

	batch := batchOf(t, svcCtx, batchUuid)
	billUuid := mustSubmit(t, svcCtx, batchUuid)
	row := billOf(t, svcCtx, billUuid)

	if row.BatchNo != batch.BatchNo {
		t.Errorf("snapshot batch_no = %q, want %q", row.BatchNo, batch.BatchNo)
	}
	if row.BatchName != batch.BatchName {
		t.Errorf("snapshot batch_name = %q, want %q", row.BatchName, batch.BatchName)
	}
	if row.SubjectUuid != batch.SubjectUuid || row.SubjectName != batch.SubjectName {
		t.Errorf("snapshot subject mismatch: %s/%s", row.SubjectUuid, row.SubjectName)
	}
	if row.BatchYear != batch.BatchYear || row.BatchMonth != batch.BatchMonth {
		t.Errorf("snapshot year/month = %d-%d, want %d-%d", row.BatchYear, row.BatchMonth, batch.BatchYear, batch.BatchMonth)
	}
	if row.Headcount != batch.Headcount {
		t.Errorf("snapshot headcount = %d, want %d", row.Headcount, batch.Headcount)
	}
	if row.NetTotal != batch.NetTotal {
		t.Errorf("snapshot net_total = %v, want %v（实发合计快照）", row.NetTotal, batch.NetTotal)
	}

	// 提交后批次重命名 → 单据快照不变
	if err := svcCtx.Model.HrmSalaryBatch.UpdateFieldsByCondition(testCtx, nil,
		map[string]any{"batch_name": "改名后的批次"},
		condition.NewChain().Equal(hrm_salary_batch.Uuid, batchUuid).Build()...); err != nil {
		t.Fatalf("rename batch: %v", err)
	}
	if got := billOf(t, svcCtx, billUuid).BatchName; got != batch.BatchName {
		t.Errorf("snapshot batch_name after rename = %q, want original %q", got, batch.BatchName)
	}
}

// TestListMineScope mine 口径：超级管理员（测试路径无 JWT=非管理员）只看本人创建单据。
func TestListMineScope(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupTrialledBatch(t)

	mustSubmit(t, svcCtx, batchUuid)

	listReq := types.ListRequest{}
	listReq.Current = 1
	listReq.Size = 10
	resp, err := NewList(testCtx, svcCtx, nil).List(&listReq)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// 测试路径无 JWT：非超级管理员 → 只看本人（uuid 空 → create_by='' 匹配测试布景创建行）
	if resp.Total < 1 {
		t.Fatalf("list total = %d, want >= 1", resp.Total)
	}
	if resp.Records[0].CreatorName != "" {
		t.Logf("creator name = %q (测试路径匿名)", resp.Records[0].CreatorName)
	}
}
