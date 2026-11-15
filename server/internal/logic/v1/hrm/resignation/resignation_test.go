package resignation

import (
	"context"
	"strings"
	"time"
	"testing"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_resignation_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/resignation"
)

// Seams under test (per ticket acceptance criteria), all at the logic-method
// seam against the real migrations (svctest):
//  1. Add: 单据编号 HR204-... generated in-tx, 未提交 status, 申请人 snapshot,
//     申请日期 defaults to today.
//  2. 审批通过仅记录: SubmitBill → ApproveBill leaves the archive untouched and
//     applied=0 (write-back belongs to the cron, not the approval tx).
//  3. 离职日定时任务: ScanPendingResigned applies only bills whose 离职日期
//     <= today, flipping person_status → '5'(离职); idempotent (applied flag).
//  4. Lifecycle guards: edit/delete rejected outside editable statuses.
//  5. 「我的单据」 default filter with admin bypass + 筛选 (编号/状态/工号).

func testCtxFor(uuid, username string, roles ...string) context.Context {
	roleUuids := make([]any, len(roles))
	for i, r := range roles {
		roleUuids[i] = r
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "uuid", uuid)
	ctx = context.WithValue(ctx, "username", username)
	ctx = context.WithValue(ctx, "role_uuids", roleUuids)
	return ctx
}

func newSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		ServiceContext: &coresvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

// seedEmployee inserts one 在职 employee row for the write-back assertions.
func seedEmployee(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, employeeNo, name, personStatus string) {
	t.Helper()
	err := svcCtx.Model.HrmEmployee.InsertV2(context.Background(), nil, &hrm_employee.HrmEmployee{
		Uuid:         employeeUuid,
		EmployeeNo:   employeeNo,
		Name:         name,
		OrgUuid:      "org-1",
		CompanyUuid:  "company-1",
		Position:     "软件工程师",
		Post:         "技术岗",
		PersonStatus: personStatus,
		HireDate:     "2026-05-01",
	})
	if err != nil {
		t.Fatalf("seedEmployee: %v", err)
	}
}

// addResignation creates one bill through the public Add seam and returns its
// uuid. Defaults mirror a 带出档案快照 flow; mutate to override.
func addResignation(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, mutate func(*types.AddRequest)) (string, *types.AddRequest) {
	t.Helper()
	req := &types.AddRequest{
		EmployeeUuid:     "emp-uuid-1",
		EmployeeNo:       "EMP001",
		EmployeeName:     "张三",
		OrgUuid:          "org-1",
		CompanyUuid:      "company-1",
		Position:         "软件工程师",
		Post:             "技术岗",
		ResignType:       ResignTypeVoluntary,
		ResignReason:     "个人发展",
		ResignDate:       "2099-01-01", // future by default so scans never apply it accidentally
		LastWorkDate:     "2099-01-01",
		HandoverPerson:   "李交接",
		SalarySettleDate: "2099-02-01",
	}
	if mutate != nil {
		mutate(req)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("Add: %v", err)
	}
	// fetch through the public list seam
	resp, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, Mine: 1})
	if err != nil {
		t.Fatalf("List after Add: %v", err)
	}
	if len(resp.Records) == 0 {
		t.Fatal("added resignation not visible in 我的单据 list")
	}
	rec := resp.Records[0]
	if !strings.HasPrefix(rec.BillNo, "HR204-") {
		t.Fatalf("billNo %q does not follow HR204- format", rec.BillNo)
	}
	if rec.Status != bill.StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	if rec.CreatorName != "张三HR" || rec.CreatorUuid == "" {
		t.Fatalf("creator snapshot wrong: uuid=%q name=%q", rec.CreatorUuid, rec.CreatorName)
	}
	return rec.Uuid, req
}

func TestAddGeneratesBillNoAndSnapshot(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	billUuid, req := addResignation(t, svcCtx, ctx, nil)

	row := mustFindResignation(t, svcCtx, billUuid)
	if row.EmployeeUuid != req.EmployeeUuid {
		t.Fatalf("employee_uuid = %q, want %q", row.EmployeeUuid, req.EmployeeUuid)
	}
	if row.Applied != 0 {
		t.Fatalf("applied = %d on creation, want 0", row.Applied)
	}
	if row.ResignDate != req.ResignDate {
		t.Fatalf("resign_date = %q, want %q", row.ResignDate, req.ResignDate)
	}
	if row.HandoverPerson != req.HandoverPerson || row.SalarySettleDate != req.SalarySettleDate {
		t.Fatalf("handover/salarySettle not persisted: %q/%q", row.HandoverPerson, row.SalarySettleDate)
	}
	if row.ApplyDate == "" {
		t.Fatal("apply_date empty, want defaulted to creation day")
	}
}

func TestAddValidation(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	cases := []struct {
		name   string
		mutate func(*types.AddRequest)
	}{
		{"missing employee", func(r *types.AddRequest) { r.EmployeeNo = ""; r.EmployeeName = "" }},
		{"bad resign type", func(r *types.AddRequest) { r.ResignType = "foo" }},
		{"missing resign date", func(r *types.AddRequest) { r.ResignDate = "" }},
		{"bad resign date format", func(r *types.AddRequest) { r.ResignDate = "2026/01/02" }},
		{"bad last work date format", func(r *types.AddRequest) { r.LastWorkDate = "01-02-2026" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{
				EmployeeNo:   "EMP001",
				EmployeeName: "张三",
				ResignType:   ResignTypeVoluntary,
				ResignDate:   "2099-01-01",
			}
			tc.mutate(req)
			if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestApproveOnlyRecordsNoImmediateWriteBack(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	seedEmployee(t, svcCtx, "emp-uuid-1", "EMP001", "张三", "2") // 试用期
	billUuid, _ := addResignation(t, svcCtx, ctx, nil)          // 离职日 2099-01-01

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, "同意离职"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	row := mustFindResignation(t, svcCtx, billUuid)
	if row.Status != bill.StatusApproved {
		t.Fatalf("bill status = %q, want approved", row.Status)
	}
	if row.Applied != 0 {
		t.Fatalf("applied = %d, want 0 (审批通过仅记录，置离职归定时任务)", row.Applied)
	}

	// 审批通过不改档案（离职日未到）
	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, "emp-uuid-1")
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if emp.PersonStatus == "5" {
		t.Fatal("person_status flipped to 离职 at approval time, want untouched until 离职日")
	}

	// 审批记录留痕 (共享底座, spec 用户故事 24)
	records, err := bill.Records(context.Background(), env.SqlConn, BillTypeCode, billUuid)
	if err != nil {
		t.Fatalf("Records: %v", err)
	}
	if len(records) < 2 {
		t.Fatalf("approval records = %d, want at least submit+approve", len(records))
	}
}

func TestScanAppliesOnlyOnResignDate(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	seedEmployee(t, svcCtx, "emp-uuid-1", "EMP001", "张三", "2")
	billUuid, _ := addResignation(t, svcCtx, ctx, nil) // 离职日 2099-01-01

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	// Cron scan before 离职日 must not touch it.
	now := time.Date(2026, 11, 16, 8, 0, 0, 0, time.Local)
	count, err := ScanPendingResigned(context.Background(), env.SqlConn, now)
	if err != nil {
		t.Fatalf("ScanPendingResigned: %v", err)
	}
	if count != 0 {
		t.Fatalf("future-dated bill applied: count=%d", count)
	}
	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, "emp-uuid-1")
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if emp.PersonStatus == "5" {
		t.Fatal("person_status flipped before 离职日")
	}

	// Cron scan on 离职日 applies it (inject past date via direct row update to
	// keep the test independent of wall-clock).
	if _, err := env.SqlConn.ExecCtx(context.Background(),
		"UPDATE hrm_resignation_bill SET resign_date = ? WHERE uuid = ?",
		"2026-11-16", billUuid); err != nil {
		t.Fatalf("set resign_date: %v", err)
	}
	count, err = ScanPendingResigned(context.Background(), env.SqlConn, now)
	if err != nil {
		t.Fatalf("ScanPendingResigned (due): %v", err)
	}
	if count != 1 {
		t.Fatalf("due bill not applied: count=%d", count)
	}
	row := mustFindResignation(t, svcCtx, billUuid)
	if row.Applied != 1 {
		t.Fatalf("applied = %d, want 1 after due scan", row.Applied)
	}
	emp, err = svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, "emp-uuid-1")
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if emp.PersonStatus != "5" {
		t.Fatalf("person_status = %q, want 5 (离职)", emp.PersonStatus)
	}

	// Idempotency: a second scan must not re-apply.
	count, err = ScanPendingResigned(context.Background(), env.SqlConn, now)
	if err != nil {
		t.Fatalf("ScanPendingResigned (repeat): %v", err)
	}
	if count != 0 {
		t.Fatalf("scan is not idempotent: count=%d", count)
	}
}

func TestScanSkipsUnapprovedBills(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	seedEmployee(t, svcCtx, "emp-uuid-1", "EMP001", "张三", "2")
	billUuid, _ := addResignation(t, svcCtx, ctx, func(r *types.AddRequest) {
		r.ResignDate = "2026-11-16" // due today but NOT approved yet
	})

	now := time.Date(2026, 11, 16, 8, 0, 0, 0, time.Local)
	count, err := ScanPendingResigned(context.Background(), env.SqlConn, now)
	if err != nil {
		t.Fatalf("ScanPendingResigned: %v", err)
	}
	if count != 0 {
		t.Fatalf("unapproved bill applied: count=%d", count)
	}
	row := mustFindResignation(t, svcCtx, billUuid)
	if row.Applied != 0 {
		t.Fatalf("applied = %d, want 0 (审批通过才进入定时任务生效)", row.Applied)
	}
	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, "emp-uuid-1")
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if emp.PersonStatus == "5" {
		t.Fatal("unapproved bill flipped person_status")
	}
}

func TestEditDeleteGuardedByStatus(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	billUuid, _ := addResignation(t, svcCtx, ctx, nil)

	// 审批中 (approving) is locked: edit and delete must be rejected.
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid,
		EmployeeNo:   "EMP001",
		EmployeeName: "张三",
		ResignType:   ResignTypeInvoluntary,
		ResignDate:   "2099-01-01",
	}); err == nil {
		t.Fatal("edit during 审批中 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批中 should fail")
	}

	// 审批通过 (approved) is locked too.
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid,
		EmployeeNo:   "EMP001",
		EmployeeName: "张三",
		ResignType:   ResignTypeInvoluntary,
		ResignDate:   "2099-01-01",
	}); err == nil {
		t.Fatal("edit during 审批通过 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批通过 should fail")
	}

	// 已撤销 (revoked) is editable/deletable again.
	billUuid2, _ := addResignation(t, svcCtx, ctx, nil)
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, ""); err != nil {
		t.Fatalf("SubmitBill 2: %v", err)
	}
	if err := bill.RevokeBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, "撤回"); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:           billUuid2,
		EmployeeNo:     "EMP002",
		EmployeeName:   "李四",
		ResignType:     ResignTypeContractEnd,
		ResignDate:     "2099-03-01",
		HandoverPerson: "王交接",
	}); err != nil {
		t.Fatalf("edit after revoke: %v", err)
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid2}}); err != nil {
		t.Fatalf("delete after revoke: %v", err)
	}
}

func TestListMineFilterAndAdminBypass(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	userCtx := testCtxFor("user-1", "张三HR")
	addResignation(t, svcCtx, userCtx, nil)
	addResignation(t, svcCtx, userCtx, func(r *types.AddRequest) {
		r.EmployeeNo = "EMP002"
		r.EmployeeName = "李四"
	})
	adminCtx := testCtxFor("admin-1", "管理员", superAdminRoleUuid())

	mineResp, err := NewList(userCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List(mine): %v", err)
	}
	if mineResp.Total != 2 {
		t.Fatalf("non-admin sees %d bills, want own 2", mineResp.Total)
	}

	allResp, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List(admin): %v", err)
	}
	if allResp.Total != 2 {
		t.Fatalf("admin should see all bills: total=%d", allResp.Total)
	}

	forcedMine, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, Mine: 1})
	if err != nil {
		t.Fatalf("List(admin, mine=1): %v", err)
	}
	if forcedMine.Total != 0 {
		t.Fatalf("admin with mine=1 should see none of user-1's bills: total=%d", forcedMine.Total)
	}

	// 筛选 (demo-inventory §3.4): 单据编号 / 单据状态 / 员工工号.
	empFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		EmployeeNo:  "EMP002",
	})
	if err != nil {
		t.Fatalf("List(employeeNo filter): %v", err)
	}
	if empFilter.Total != 1 || empFilter.Records[0].EmployeeNo != "EMP002" {
		t.Fatalf("employeeNo filter: total=%d, want 1 EMP002", empFilter.Total)
	}

	statusFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		Status:      bill.StatusUnsubmitted,
	})
	if err != nil {
		t.Fatalf("List(status filter): %v", err)
	}
	if statusFilter.Total != 2 {
		t.Fatalf("status filter: total=%d, want 2", statusFilter.Total)
	}

	billNoFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		BillNo:      empFilter.Records[0].BillNo,
	})
	if err != nil {
		t.Fatalf("List(billNo filter): %v", err)
	}
	if billNoFilter.Total != 1 {
		t.Fatalf("billNo filter: total=%d, want 1", billNoFilter.Total)
	}
}

func mustFindResignation(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) *hrm_resignation_bill.HrmResignationBill {
	t.Helper()
	row, err := svcCtx.Model.HrmResignationBill.FindOneByUuid(context.Background(), nil, billUuid)
	if err != nil {
		t.Fatalf("find resignation %s: %v", billUuid, err)
	}
	return row
}

func superAdminRoleUuid() string {
	// keep the seeded R_SUPER uuid local to tests; production logic uses hrmuser
	return "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
}
