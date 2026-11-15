package regularization

import (
	"context"
	"strings"
	"testing"


	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_regularization_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/regularization"
)

// Seams under test (per ticket acceptance criteria), all at the logic-method
// seam against the real migrations (svctest):
//  1. Add: 选择员工带出的档案快照 persisted; 单据编号 HR202-... generated in-tx;
//     未提交 status; 申请人 snapshot.
//  2. 审批通过回写 (spec 用户故事 20): SubmitBill → ApproveBill updates
//     hrm_employee person_status='1'(正式) + formal_date inside the approval
//     transaction (优先单据实际转正日期，否则预计转正日期).
//  3. Idempotent re-approval / replayed write-back: no double effect.
//  4. Lifecycle guards: edit/delete rejected outside editable statuses.
//  5. 「我的单据」 default filter with admin bypass + 筛选 (编号/状态/工号).

// seedEmployee inserts one 试用期 employee row for the write-back assertions.
func seedEmployee(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, employeeNo, name string) {
	t.Helper()
	err := svcCtx.Model.HrmEmployee.InsertV2(context.Background(), nil, &hrm_employee.HrmEmployee{
		Uuid:         employeeUuid,
		EmployeeNo:   employeeNo,
		Name:         name,
		OrgUuid:      "org-1",
		CompanyUuid:  "company-1",
		Position:     "软件工程师",
		Post:         "技术岗",
		PersonStatus: "2", // 试用期
		HireDate:     "2026-05-01",
		FormalDate:   "2026-08-01", // 预计转正日期 (HRM-03 列语义)
	})
	if err != nil {
		t.Fatalf("seedEmployee: %v", err)
	}
}

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

// addRegularization creates one bill through the public Add seam and returns
// its uuid. Defaults mirror a 带出档案快照 flow; mutate to override.
func addRegularization(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, mutate func(*types.AddRequest)) (string, *types.AddRequest) {
	t.Helper()
	req := &types.AddRequest{
		EmployeeUuid:       "emp-uuid-1",
		EmployeeNo:         "EMP001",
		EmployeeName:       "张三",
		OrgUuid:            "org-1",
		CompanyUuid:        "company-1",
		Position:           "软件工程师",
		Post:               "技术岗",
		PersonStatus:       "2",
		HireDate:           "2026-05-01",
		ExpectedFormalDate: "2026-08-01",
		FormalDate:         "2026-11-16",
		Remark:             "试用期表现良好",
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
		t.Fatal("added regularization not visible in 我的单据 list")
	}
	rec := resp.Records[0]
	if !strings.HasPrefix(rec.BillNo, "HR202-") {
		t.Fatalf("billNo %q does not follow HR202- format", rec.BillNo)
	}
	if rec.Status != bill.StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	if rec.CreatorName != "张三HR" || rec.CreatorUuid == "" {
		t.Fatalf("creator snapshot wrong: uuid=%q name=%q", rec.CreatorUuid, rec.CreatorName)
	}
	if rec.EmployeeNo != req.EmployeeNo || rec.EmployeeName != req.EmployeeName ||
		rec.OrgUuid != req.OrgUuid || rec.CompanyUuid != req.CompanyUuid ||
		rec.Position != req.Position || rec.Post != req.Post ||
		rec.HireDate != req.HireDate || rec.ExpectedFormalDate != req.ExpectedFormalDate {
		t.Fatalf("employee snapshot not persisted: %+v", rec)
	}
	return rec.Uuid, req
}

func TestAddGeneratesBillNoAndSnapshot(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	billUuid, req := addRegularization(t, svcCtx, ctx, nil)

	row := mustFindRegularization(t, svcCtx, billUuid)
	if row.EmployeeUuid != req.EmployeeUuid {
		t.Fatalf("employee_uuid = %q, want %q", row.EmployeeUuid, req.EmployeeUuid)
	}
	if row.Applied != 0 {
		t.Fatalf("applied = %d on creation, want 0 (write-back not yet run)", row.Applied)
	}
	if row.FormalDate != req.FormalDate {
		t.Fatalf("formal_date = %q, want %q (实际转正日期)", row.FormalDate, req.FormalDate)
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
		{"bad formal date format", func(r *types.AddRequest) { r.FormalDate = "2026/11/16" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{
				EmployeeUuid: "emp-uuid-1",
				EmployeeNo:   "EMP001",
				EmployeeName: "张三",
			}
			tc.mutate(req)
			if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestApproveWritesBackPersonStatusAndFormalDate(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	seedEmployee(t, svcCtx, "emp-uuid-1", "EMP001", "张三")
	billUuid, _ := addRegularization(t, svcCtx, ctx, nil)

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, "同意转正"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	row := mustFindRegularization(t, svcCtx, billUuid)
	if row.Status != bill.StatusApproved {
		t.Fatalf("bill status = %q, want approved", row.Status)
	}
	if row.Applied != 1 {
		t.Fatalf("applied = %d, want 1 (审批回写已完成)", row.Applied)
	}

	// 回写断言（spec 用户故事 20）: person_status='1'(正式) + 转正日期=单据实际转正日期
	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, "emp-uuid-1")
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if emp.PersonStatus != "1" {
		t.Fatalf("person_status = %q, want 1 (正式)", emp.PersonStatus)
	}
	if emp.FormalDate != "2026-11-16" {
		t.Fatalf("formal_date = %q, want 2026-11-16 (优先单据实际转正日期)", emp.FormalDate)
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

func TestApproveFallsBackToExpectedFormalDate(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	seedEmployee(t, svcCtx, "emp-uuid-1", "EMP001", "张三")
	billUuid, _ := addRegularization(t, svcCtx, ctx, func(r *types.AddRequest) {
		r.FormalDate = "" // 未填实际转正日期 → 回退预计转正日期
	})

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, "emp-uuid-1")
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if emp.PersonStatus != "1" {
		t.Fatalf("person_status = %q, want 1 (正式)", emp.PersonStatus)
	}
	if emp.FormalDate != "2026-08-01" {
		t.Fatalf("formal_date = %q, want 2026-08-01 (回退预计转正日期)", emp.FormalDate)
	}
}

func TestApproveIdempotentWriteBack(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	seedEmployee(t, svcCtx, "emp-uuid-1", "EMP001", "张三")
	billUuid, _ := addRegularization(t, svcCtx, ctx, nil)

	// Re-run the write-back before approval: must be a safe replay, no status
	// change on the bill (still unsubmitted), archive untouched apart from
	// what the replay itself wrote — and the later approval still succeeds.
	row := mustFindRegularization(t, svcCtx, billUuid)
	if err := applyRegularization(context.Background(), env.SqlConn, row); err != nil {
		t.Fatalf("applyRegularization replay: %v", err)
	}
	if err := applyRegularization(context.Background(), env.SqlConn, row); err != nil {
		t.Fatalf("applyRegularization second replay: %v", err)
	}

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	// Re-approval of an already-applied bill: state machine rejects (terminal
	// status) and — for any replay path that reaches onApproved again — the
	// applied flag keeps it a no-op.
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	row = mustFindRegularization(t, svcCtx, billUuid)
	if row.Applied != 1 || row.Status != bill.StatusApproved {
		t.Fatalf("after approval applied=%d status=%q, want 1/approved", row.Applied, row.Status)
	}
	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, "emp-uuid-1")
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if emp.PersonStatus != "1" || emp.FormalDate != "2026-11-16" {
		t.Fatalf("employee after idempotent replay: person_status=%q formal_date=%q, want 1/2026-11-16",
			emp.PersonStatus, emp.FormalDate)
	}
}

func TestEditDeleteGuardedByStatus(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	billUuid, _ := addRegularization(t, svcCtx, ctx, nil)

	// 审批中 (approving) is locked: edit and delete must be rejected.
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid,
		EmployeeNo:   "EMP001",
		EmployeeName: "张三",
		FormalDate:   "2026-11-16",
	}); err == nil {
		t.Fatal("edit during 审批中 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批中 should fail")
	}

	// 审批通过 (approved) is locked too — archive write-back already done.
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid,
		EmployeeNo:   "EMP001",
		EmployeeName: "张三",
		FormalDate:   "2026-11-16",
	}); err == nil {
		t.Fatal("edit during 审批通过 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批通过 should fail")
	}

	// 已撤销 (revoked) is editable/deletable again.
	billUuid2, _ := addRegularization(t, svcCtx, ctx, nil)
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, ""); err != nil {
		t.Fatalf("SubmitBill 2: %v", err)
	}
	if err := bill.RevokeBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, "撤回"); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid2,
		EmployeeNo:   "EMP002",
		EmployeeName: "李四",
		FormalDate:   "2026-12-01",
		Remark:       "改备注",
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
	addRegularization(t, svcCtx, userCtx, nil)
	addRegularization(t, svcCtx, userCtx, func(r *types.AddRequest) {
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

	// 筛选 (demo-inventory §3.2): 单据编号 / 单据状态 / 员工工号.
	noFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		EmployeeNo:  "EMP002",
	})
	if err != nil {
		t.Fatalf("List(employeeNo filter): %v", err)
	}
	if noFilter.Total != 1 || noFilter.Records[0].EmployeeNo != "EMP002" {
		t.Fatalf("employeeNo filter: total=%d, want 1 EMP002", noFilter.Total)
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
		BillNo:      noFilter.Records[0].BillNo,
	})
	if err != nil {
		t.Fatalf("List(billNo filter): %v", err)
	}
	if billNoFilter.Total != 1 {
		t.Fatalf("billNo filter: total=%d, want 1", billNoFilter.Total)
	}
}

func mustFindRegularization(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) *hrm_regularization_bill.HrmRegularizationBill {
	t.Helper()
	row, err := svcCtx.Model.HrmRegularizationBill.FindOneByUuid(context.Background(), nil, billUuid)
	if err != nil {
		t.Fatalf("find regularization %s: %v", billUuid, err)
	}
	return row
}

func superAdminRoleUuid() string {
	// keep the seeded R_SUPER uuid local to tests; production logic uses hrmuser
	return "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
}

