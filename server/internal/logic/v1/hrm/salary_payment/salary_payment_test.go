package salary_payment

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	salary_batch "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/salary_batch"
	salary_confirm "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/salary_confirm"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_archive"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types_batch "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payment"
)

// DB-backed logic tests for 发薪登记 HR210 (HRM-38, issue #40), run via
// svctest (sqlite default; HRM_TEST_DB=mysql for docker MySQL).
//
// Coverage (issue #40 审批门禁 acceptance + ticket 票面):
//   - 审批门禁：批次未确认（confirm_status 未推进）→ 发薪登记拒绝（10714）
//   - 已发薪（pay_status='1'）重复登记拒绝（10714）
//   - 金额快照：发薪金额留空默认=批次实发合计；显式金额可改
//   - 登记事务内：批次 pay_status='1'；登记行快照 批次编号/名称/主体
//   - 发薪方式枚举守卫（bank/cash/other）

var testCtx = context.Background()

// svcCtx package-level slot so 布景 helpers read like the salary_confirm ones.
var svcCtx *svc.ServiceContext

func newSvcCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	svcCtx = &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
	return svcCtx
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

// ===== 布景（org/subject/scheme/item/employee/archive + 批次 + 试算 + 审批确认） =====

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

func seedEmployeeWithArchive(t *testing.T, subjectUuid, schemeUuid string) {
	t.Helper()
	empUuid := uuid.New().String()
	if err := svcCtx.Model.HrmEmployee.InsertV2(testCtx, nil, &hrm_employee.HrmEmployee{
		Uuid:         empUuid,
		EmployeeNo:   "E001",
		Name:         "张三",
		CompanyUuid:  seedOrgRow(t, "甲公司"+uuid.New().String()[:8], "1"),
		OrgUuid:      seedOrgRow(t, "技术部"+uuid.New().String()[:8], "2"),
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
}

func mustAddBatch(t *testing.T, name, subjectUuid string, year, month int64) string {
	t.Helper()
	resp, err := salary_batch.NewAdd(testCtx, svcCtx, nil).Add(&types_batch.AddRequest{
		BatchName: name, SubjectUuid: subjectUuid, BatchYear: year, BatchMonth: month,
	})
	if err != nil {
		t.Fatalf("add batch %s: %v", name, err)
	}
	return resp.Uuid
}

// setupConfirmedBatch 布景：批次试算 + HR209 审批通过（confirm_status='1'）。
func setupConfirmedBatch(t *testing.T) string {
	t.Helper()
	subject := seedSubjectRow(t, "3月主体")
	scheme := seedSchemeAllItems(t, subject)
	seedEmployeeWithArchive(t, subject, scheme)
	batchUuid := mustAddBatch(t, "2026-03 月度批次", subject, 2026, 3)
	if _, err := salary_batch.RunPayrollTrial(testCtx, svcCtx, batchUuid); err != nil {
		t.Fatalf("trial: %v", err)
	}
	billUuid, err := salary_confirm.SubmitSalaryConfirm(testCtx, svcCtx.SqlxConn, svcCtx, batchUuid, "")
	if err != nil {
		t.Fatalf("submit salary confirm: %v", err)
	}
	if err := bill.ApproveBill(salary_confirm.WithSvcCtx(testCtx, svcCtx), svcCtx.SqlxConn,
		salary_confirm.BillTypeCode, billUuid, bill.Operator{Name: "审批人"}, "同意"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	return batchUuid
}

func batchOf(t *testing.T, batchUuid string) *hrm_salary_batch.HrmSalaryBatch {
	t.Helper()
	row, err := svcCtx.Model.HrmSalaryBatch.FindOneByUuid(testCtx, nil, batchUuid)
	if err != nil {
		t.Fatalf("find batch %s: %v", batchUuid, err)
	}
	return row
}

func mustRegister(t *testing.T, req *types.RegisterRequest) *types.RegisterResponse {
	t.Helper()
	resp, err := NewRegister(testCtx, svcCtx, nil).Register(req)
	if err != nil {
		t.Fatalf("register salary payment for %s: %v", req.BatchUuid, err)
	}
	return resp
}

// ===== 守卫 + 事务 + 金额快照 =====

// TestRegisterGuardUnconfirmed 审批门禁：批次未确认 → 发薪登记拒绝（10714）。
func TestRegisterGuardUnconfirmed(t *testing.T) {
	svcCtx := newSvcCtx(t)
	subject := seedSubjectRow(t, "未确认主体")
	scheme := seedSchemeAllItems(t, subject)
	seedEmployeeWithArchive(t, subject, scheme)
	batchUuid := mustAddBatch(t, "未确认批次", subject, 2026, 4)
	if _, err := salary_batch.RunPayrollTrial(testCtx, svcCtx, batchUuid); err != nil {
		t.Fatalf("trial: %v", err)
	}

	_, err := NewRegister(testCtx, svcCtx, nil).Register(&types.RegisterRequest{BatchUuid: batchUuid})
	assertErrCode(t, err, 10714, "未确认批次发薪登记")
	if b := batchOf(t, batchUuid); b.PayStatus != "" && b.PayStatus != "0" {
		t.Fatalf("pay_status after denied register = %s, want ''/0", b.PayStatus)
	}
}

// TestRegisterGuardAlreadyPaid 已发薪（pay_status='1'）后重复登记拒绝（10714）。
func TestRegisterGuardAlreadyPaid(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupConfirmedBatch(t)

	mustRegister(t, &types.RegisterRequest{BatchUuid: batchUuid, PayMode: PayModeBank})

	_, err := NewRegister(testCtx, svcCtx, nil).Register(&types.RegisterRequest{BatchUuid: batchUuid})
	assertErrCode(t, err, 10714, "已发薪批次重复登记")
}

// TestRegisterDefaultAmountSnapshot 金额快照：发薪金额留空默认=批次实发合计；
// 显式金额可改；事务内批次 pay_status='1'；登记行快照 批次编号/名称/主体。
func TestRegisterDefaultAmountSnapshot(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupConfirmedBatch(t)
	batch := batchOf(t, batchUuid)

	// 默认金额 = 实发合计
	resp := mustRegister(t, &types.RegisterRequest{BatchUuid: batchUuid})
	row, err := svcCtx.Model.HrmSalaryPayment.FindOneByUuid(testCtx, nil, resp.Uuid)
	if err != nil {
		t.Fatalf("find payment: %v", err)
	}
	if row.PayAmount != batch.NetTotal {
		t.Errorf("default pay_amount = %v, want net_total %v", row.PayAmount, batch.NetTotal)
	}
	if row.PayMode != PayModeBank {
		t.Errorf("default pay_mode = %s, want bank", row.PayMode)
	}
	if row.BatchNo != batch.BatchNo || row.BatchName != batch.BatchName {
		t.Errorf("payment snapshot batch mismatch: %s/%s", row.BatchNo, row.BatchName)
	}
	if row.SubjectUuid != batch.SubjectUuid || row.SubjectName != batch.SubjectName {
		t.Errorf("payment snapshot subject mismatch: %s/%s", row.SubjectUuid, row.SubjectName)
	}
}

// TestRegisterCustomAmountAndFields 显式金额/发薪方式/回执编号/到账时间可登记。
func TestRegisterCustomAmountAndFields(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupConfirmedBatch(t)

	resp := mustRegister(t, &types.RegisterRequest{
		BatchUuid:  batchUuid,
		PayAmount:  20000.5,
		PayMode:    PayModeCash,
		ReceiptNo:  "RCPT-2026-001",
		ArriveTime: "2026-04-10 09:30:00",
	})
	row, err := svcCtx.Model.HrmSalaryPayment.FindOneByUuid(testCtx, nil, resp.Uuid)
	if err != nil {
		t.Fatalf("find payment: %v", err)
	}
	if row.PayAmount != 20000.5 {
		t.Errorf("pay_amount = %v, want 20000.5", row.PayAmount)
	}
	if row.PayMode != PayModeCash || row.ReceiptNo != "RCPT-2026-001" || row.ArriveTime != "2026-04-10 09:30:00" {
		t.Errorf("payment fields mismatch: %s/%s/%s", row.PayMode, row.ReceiptNo, row.ArriveTime)
	}
	// 事务内批次 pay 标记
	if b := batchOf(t, batchUuid); b.PayStatus != "1" {
		t.Errorf("pay_status after register = %s, want 1", b.PayStatus)
	}
}

// TestRegisterInvalidPayMode 发薪方式枚举守卫。
func TestRegisterInvalidPayMode(t *testing.T) {
	svcCtx := newSvcCtx(t)
	batchUuid := setupConfirmedBatch(t)

	_, err := NewRegister(testCtx, svcCtx, nil).Register(&types.RegisterRequest{
		BatchUuid: batchUuid, PayMode: "wechat",
	})
	assertErrCode(t, err, 10713, "非法发薪方式")
}
