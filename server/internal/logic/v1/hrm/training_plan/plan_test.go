package training_plan

import (
	"context"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// Seams under test (per ticket acceptance criteria), all at the logic-method
// seam against the real migrations (svctest):
//  1. Add: 单据编号 HR216-... generated in-tx, 未提交 status, 申请人 snapshot.
//  2. 审批与发布流转: SubmitBill → ApproveBill flips published 0→1 (发布状态
//     草稿/已发布) inside the approval transaction; RejectBill/RevokeBill leave
//     it 草稿. Approval record trail is queryable through the shared seam.
//  3. Lifecycle guards: edit/delete rejected outside editable statuses and
//     after 发布 (已发布计划不允许修改/删除).
//  4. Validation: missing name/year/type/date range rejected.

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

func validAddReq() *types.AddRequest {
	return &types.AddRequest{
		PlanName:  "2026年度新员工培养计划",
		Theme:     "新员工入职培养",
		PlanYear:  "2026",
		PlanType:  PlanTypeNewHire,
		StartDate: "2026-01-01",
		EndDate:   "2026-12-31",
	}
}

func addPlan(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, mutate func(*types.AddRequest)) (string, *types.AddRequest) {
	t.Helper()
	req := validAddReq()
	if mutate != nil {
		mutate(req)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("Add: %v", err)
	}
	// fetch through the public list seam
	resp, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List after Add: %v", err)
	}
	if len(resp.Records) == 0 {
		t.Fatal("added plan not visible in list")
	}
	rec := resp.Records[0]
	if rec.BillNo[:5] != "HR216" {
		t.Fatalf("billNo %q does not follow HR216 format", rec.BillNo)
	}
	if rec.Status != bill.StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	if rec.Published != 0 {
		t.Fatalf("new plan published = %d, want 0 (草稿)", rec.Published)
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

	addPlan(t, svcCtx, ctx, nil)
}

func TestAddValidation(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	cases := []struct {
		name   string
		mutate func(*types.AddRequest)
	}{
		{"missing name", func(r *types.AddRequest) { r.PlanName = "" }},
		{"missing year", func(r *types.AddRequest) { r.PlanYear = "" }},
		{"bad year", func(r *types.AddRequest) { r.PlanYear = "26" }},
		{"bad plan type", func(r *types.AddRequest) { r.PlanType = "vacation" }},
		{"missing start date", func(r *types.AddRequest) { r.StartDate = "" }},
		{"end before start", func(r *types.AddRequest) { r.StartDate = "2026-12-01"; r.EndDate = "2026-01-01" }},
		{"bad date format", func(r *types.AddRequest) { r.StartDate = "2026/01/01" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validAddReq()
			tc.mutate(req)
			if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestApprovePublishesPlan(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	billUuid, _ := addPlan(t, svcCtx, ctx, nil)

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, "同意"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	row := mustFindPlan(t, env, billUuid)
	if row.Status != bill.StatusApproved {
		t.Fatalf("status = %q, want approved", row.Status)
	}
	if row.Published != 1 {
		t.Fatalf("published = %d, want 1 (审批通过即发布)", row.Published)
	}
}

func TestRejectRevokeKeepsDraft(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	// 驳回: stays 草稿
	billUuid, _ := addPlan(t, svcCtx, ctx, nil)
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.RejectBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, "不符合年度安排"); err != nil {
		t.Fatalf("RejectBill: %v", err)
	}
	if row := mustFindPlan(t, env, billUuid); row.Published != 0 {
		t.Fatalf("rejected plan published = %d, want 0", row.Published)
	}

	// 撤销: stays 草稿
	billUuid2, _ := addPlan(t, svcCtx, ctx, nil)
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.RevokeBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, "撤回"); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	if row := mustFindPlan(t, env, billUuid2); row.Published != 0 {
		t.Fatalf("revoked plan published = %d, want 0", row.Published)
	}

	// 审批记录 trail queryable through the shared seam
	records, err := bill.Records(context.Background(), env.SqlConn, BillTypeCode, billUuid2)
	if err != nil {
		t.Fatalf("bill.Records: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("approval records = %d, want 2 (submit+revoke)", len(records))
	}
}

func TestEditDeleteGuardedByStatus(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	billUuid, _ := addPlan(t, svcCtx, ctx, nil)

	// 审批中 (approving) is locked: edit and delete must be rejected.
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(editReqFor(billUuid)); err == nil {
		t.Fatal("edit during 审批中 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批中 should fail")
	}

	// 已撤销 (revoked) is editable/deletable again.
	if err := bill.RevokeBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, "撤回"); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(editReqFor(billUuid)); err != nil {
		t.Fatalf("edit after revoke: %v", err)
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err != nil {
		t.Fatalf("delete after revoke: %v", err)
	}
}

func TestPublishedPlanLocked(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	billUuid, _ := addPlan(t, svcCtx, ctx, nil)
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	// 已发布 (approved → published=1) 不允许修改/删除
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(editReqFor(billUuid)); err == nil {
		t.Fatal("edit published plan should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete published plan should fail")
	}
}

func editReqFor(uuid string) *types.EditRequest {
	req := &types.EditRequest{
		Uuid: uuid,
	}
	req.PlanName = "2026年度新员工培养计划(修订)"
	req.PlanYear = "2026"
	req.PlanType = PlanTypeNewHire
	req.StartDate = "2026-01-01"
	req.EndDate = "2026-12-31"
	return req
}

func mustFindPlan(t *testing.T, env *svctest.Env, billUuid string) *hrm_training_plan.HrmTrainingPlan {
	t.Helper()
	row, err := hrm_training_plan.NewHrmTrainingPlanModel(env.SqlConn).FindOneByUuid(context.Background(), nil, billUuid)
	if err != nil {
		t.Fatalf("find plan %s: %v", billUuid, err)
	}
	return row
}

var _ = sqlx.ErrNotFound // keep import stable across edits
