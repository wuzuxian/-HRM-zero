package transfer

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_transfer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/transfer"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// Seams under test (per ticket acceptance criteria), all at the logic-method
// seam against the real migrations (svctest):
//  1. Add: 单据编号 HR203-... generated in-tx, 未提交 status, 申请人 snapshot.
//  2. 立即生效路径: SubmitBill → ApproveBill runs the HRM-03 seam write-back
//     and flips applied=1 inside the approval transaction.
//  3. 生效日定时任务: ScanPendingEffective applies only due (生效日期<=今天)
//     approved delayed bills, is idempotent, and self-heals bills whose
//     write-back was skipped while the HRM-03 seam was not wired.
//  4. Lifecycle guards: edit/delete rejected outside editable statuses;
//     「我的单据」 default filter with admin bypass.
//
// The fake applier stands in for HRM-03 员工档案 (not merged yet) via the
// exported SetEffectApplier seam.

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

func addTransfer(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, mutate func(*types.AddRequest)) (string, *types.AddRequest) {
	t.Helper()
	req := &types.AddRequest{
		EmployeeNo:           "EMP001",
		EmployeeName:         "张三",
		ChangeType:           ChangeTypeOrg,
		OriginalOrg:          "org-old-uuid",
		NewOrg:               "org-new-uuid",
		TransferReason:       "组织调整",
		EffectiveImmediately: 0,
		EffectiveDate:        "2099-01-01",
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
		t.Fatal("added transfer not visible in 我的单据 list")
	}
	rec := resp.Records[0]
	if !strings.HasPrefix(rec.BillNo, "HR203-") {
		t.Fatalf("billNo %q does not follow HR203- format", rec.BillNo)
	}
	if rec.Status != bill.StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	if rec.CreatorName != "张三HR" || rec.CreatorUuid == "" {
		t.Fatalf("creator snapshot wrong: uuid=%q name=%q", rec.CreatorUuid, rec.CreatorName)
	}
	return rec.Uuid, req
}

func TestAddGeneratesBillNoAndDefaults(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	addTransfer(t, svcCtx, ctx, nil)
}

func TestAddValidation(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	cases := []struct {
		name  string
		mutat func(*types.AddRequest)
	}{
		{"missing employee", func(r *types.AddRequest) { r.EmployeeNo = ""; r.EmployeeName = "" }},
		{"bad change type", func(r *types.AddRequest) { r.ChangeType = "demotion" }},
		{"delayed without date", func(r *types.AddRequest) { r.EffectiveDate = "" }},
		{"bad date format", func(r *types.AddRequest) { r.EffectiveDate = "2026/01/02" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{
				EmployeeNo:           "EMP001",
				EmployeeName:         "张三",
				ChangeType:           ChangeTypePost,
				EffectiveImmediately: 0,
				EffectiveDate:        "2099-01-01",
			}
			tc.mutat(req)
			if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestApproveImmediateEffectAppliesWriteBack(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	applied := make([]Effect, 0, 1)
	SetEffectApplier(func(ctx context.Context, session sqlx.Session, e Effect) error {
		applied = append(applied, e)
		return nil
	})
	t.Cleanup(func() { SetEffectApplier(nil) })

	billUuid, req := addTransfer(t, svcCtx, ctx, func(r *types.AddRequest) {
		r.EffectiveImmediately = 1
		r.EffectiveDate = ""
	})

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, "同意"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	if len(applied) != 1 {
		t.Fatalf("HRM-03 seam called %d times, want 1 (immediate effect inside approval tx)", len(applied))
	}
	if applied[0].NewOrg != req.NewOrg || applied[0].EmployeeNo != req.EmployeeNo {
		t.Fatalf("effect payload mismatch: %+v", applied[0])
	}

	row := mustFindTransfer(t, svcCtx, billUuid)
	if row.Status != bill.StatusApproved {
		t.Fatalf("status = %q, want approved", row.Status)
	}
	if row.Applied != 1 {
		t.Fatalf("applied = %d, want 1 (immediate effect)", row.Applied)
	}
}

func TestApproveDelayedEffectWaitsForCron(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	var seamCalls int
	SetEffectApplier(func(ctx context.Context, session sqlx.Session, e Effect) error {
		seamCalls++
		return nil
	})
	t.Cleanup(func() { SetEffectApplier(nil) })

	billUuid, _ := addTransfer(t, svcCtx, ctx, nil) // delayed, 生效日期 2099-01-01

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	row := mustFindTransfer(t, svcCtx, billUuid)
	if row.Applied != 0 {
		t.Fatalf("delayed bill applied too early: applied=%d", row.Applied)
	}

	// Cron scan before 生效日 must not touch it.
	now := time.Date(2026, 11, 15, 8, 0, 0, 0, time.Local)
	count, err := ScanPendingEffective(context.Background(), env.SqlConn, now)
	if err != nil {
		t.Fatalf("ScanPendingEffective: %v", err)
	}
	if count != 0 || seamCalls != 0 {
		t.Fatalf("future-dated bill applied: count=%d seamCalls=%d", count, seamCalls)
	}

	// Cron scan on/after 生效日 applies it (inject past date via direct row update
	// to keep the test independent of wall-clock).
	if _, err := env.SqlConn.ExecCtx(context.Background(),
		"UPDATE hrm_transfer_bill SET effective_date = ? WHERE uuid = ?",
		"2026-11-15", billUuid); err != nil {
		t.Fatalf("set effective_date: %v", err)
	}
	count, err = ScanPendingEffective(context.Background(), env.SqlConn, now)
	if err != nil {
		t.Fatalf("ScanPendingEffective (due): %v", err)
	}
	if count != 1 || seamCalls != 1 {
		t.Fatalf("due bill not applied: count=%d seamCalls=%d", count, seamCalls)
	}
	row = mustFindTransfer(t, svcCtx, billUuid)
	if row.Applied != 1 {
		t.Fatalf("applied = %d, want 1 after due scan", row.Applied)
	}

	// Idempotency: a second scan must not re-apply.
	count, err = ScanPendingEffective(context.Background(), env.SqlConn, now)
	if err != nil {
		t.Fatalf("ScanPendingEffective (repeat): %v", err)
	}
	if count != 0 || seamCalls != 1 {
		t.Fatalf("scan is not idempotent: count=%d seamCalls=%d", count, seamCalls)
	}
}

func TestScanHealsBillsWithoutSeam(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	// No applier registered (HRM-03 not merged): approval must succeed and
	// the bill must stay pending-applied rather than being marked done.
	SetEffectApplier(nil)

	billUuid, _ := addTransfer(t, svcCtx, ctx, func(r *types.AddRequest) {
		r.EffectiveImmediately = 1
		r.EffectiveDate = ""
	})

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill with nil seam (must degrade, not fail): %v", err)
	}

	row := mustFindTransfer(t, svcCtx, billUuid)
	if row.Applied != 0 {
		t.Fatalf("bill marked applied although HRM-03 seam absent; applied=%d", row.Applied)
	}

	// HRM-03 lands later: registering the seam heals the bill on the next scan
	// (immediate-effect bills are also picked up by the scan while applied=0).
	var healed Effect
	SetEffectApplier(func(ctx context.Context, session sqlx.Session, e Effect) error {
		healed = e
		return nil
	})
	t.Cleanup(func() { SetEffectApplier(nil) })

	count, err := ScanPendingEffective(context.Background(), env.SqlConn, time.Now())
	if err != nil {
		t.Fatalf("ScanPendingEffective (heal): %v", err)
	}
	if count != 1 {
		t.Fatalf("heal scan count=%d, want 1", count)
	}
	if healed.EmployeeNo != "EMP001" {
		t.Fatalf("healed effect payload: %+v", healed)
	}
	if row = mustFindTransfer(t, svcCtx, billUuid); row.Applied != 1 {
		t.Fatalf("applied = %d, want 1 after heal", row.Applied)
	}
}

func TestEditDeleteGuardedByStatus(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	billUuid, _ := addTransfer(t, svcCtx, ctx, nil)

	// 审批中 (approving) is locked: edit and delete must be rejected.
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:                 billUuid,
		EmployeeNo:           "EMP001",
		EmployeeName:         "张三",
		ChangeType:           ChangeTypeOther,
		EffectiveImmediately: 1,
	}); err == nil {
		t.Fatal("edit during 审批中 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批中 should fail")
	}

	// 已撤销 (revoked) is editable/deletable again.
	if err := bill.RevokeBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, "撤回"); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:                 billUuid,
		EmployeeNo:           "EMP002",
		EmployeeName:         "张三",
		ChangeType:           ChangeTypeOther,
		EffectiveImmediately: 1,
	}); err != nil {
		t.Fatalf("edit after revoke: %v", err)
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err != nil {
		t.Fatalf("delete after revoke: %v", err)
	}
}

func TestListMineFilterAndAdminBypass(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	userCtx := testCtxFor("user-1", "张三HR")
	addTransfer(t, svcCtx, userCtx, nil)
	adminCtx := testCtxFor("admin-1", "管理员", hrmuser_SuperAdminRoleUuid())

	mineResp, err := NewList(userCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List(mine): %v", err)
	}
	if mineResp.Total != 1 {
		t.Fatalf("non-admin sees %d bills, want only own 1", mineResp.Total)
	}

	allResp, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List(admin): %v", err)
	}
	if allResp.Total != 1 {
		t.Fatalf("admin should see all bills: total=%d", allResp.Total)
	}

	forcedMine, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, Mine: 1})
	if err != nil {
		t.Fatalf("List(admin, mine=1): %v", err)
	}
	if forcedMine.Total != 0 {
		t.Fatalf("admin with mine=1 should see none of user-1's bills: total=%d", forcedMine.Total)
	}
}

func mustFindTransfer(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) *hrm_transfer_bill.HrmTransferBill {
	t.Helper()
	row, err := svcCtx.Model.HrmTransferBill.FindOneByUuid(context.Background(), nil, billUuid)
	if err != nil {
		t.Fatalf("find transfer %s: %v", billUuid, err)
	}
	return row
}

func hrmuser_SuperAdminRoleUuid() string {
	// keep the seeded R_SUPER uuid local to tests; production logic uses hrmuser
	return "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
}
