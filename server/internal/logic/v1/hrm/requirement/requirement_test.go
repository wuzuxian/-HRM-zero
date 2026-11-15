package requirement

import (
	"context"
	"strings"
	"testing"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_requirement"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/requirement"
)

// Seams under test (per ticket acceptance criteria), all at the logic-method
// seam against the real migrations (svctest):
//  1. Add: 单据编号 HR207-... generated in-tx, 未提交 status, 申请人 snapshot.
//  2. Validation: 岗位名称/类别/人数/紧急度/所属部门/日期 guards (1043x).
//  3. Lifecycle: SubmitBill → ApproveBill via the shared 单据底座 type code 207;
//     edit/delete rejected outside editable statuses.
//  4. 「我的单据」 default filter with admin bypass + 筛选 (岗位名称/状态).
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

func addRequirement(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, mutate func(*types.AddRequest)) string {
	t.Helper()
	req := &types.AddRequest{
		PostName:        "Go 工程师",
		PostCategory:    PostCategoryTech,
		Headcount:       2,
		Urgency:         UrgencyNormal,
		ExpectDate:      "2026-12-01",
		OrgUuid:         "org-uuid-1",
		OrgName:         "研发部",
		RequirementDesc: "负责核心系统开发",
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
		t.Fatal("added requirement not visible in 我的单据 list")
	}
	rec := resp.Records[0]
	if !strings.HasPrefix(rec.BillNo, "HR207-") {
		t.Fatalf("billNo %q does not follow HR207- format", rec.BillNo)
	}
	if rec.Status != bill.StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	if rec.CreatorName == "" || rec.CreatorUuid == "" {
		t.Fatalf("creator snapshot wrong: uuid=%q name=%q", rec.CreatorUuid, rec.CreatorName)
	}
	return rec.Uuid
}

func TestAddGeneratesBillNoAndSnapshot(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	billUuid := addRequirement(t, svcCtx, ctx, nil)

	row := mustFindRequirement(t, svcCtx, billUuid)
	if row.PostName != "Go 工程师" || row.PostCategory != PostCategoryTech {
		t.Fatalf("post snapshot wrong: %q/%q", row.PostName, row.PostCategory)
	}
	if row.Headcount != 2 {
		t.Fatalf("headcount = %d, want 2", row.Headcount)
	}
	if row.Urgency != UrgencyNormal {
		t.Fatalf("urgency = %q, want normal", row.Urgency)
	}
	if row.ExpectDate != "2026-12-01" {
		t.Fatalf("expect_date = %q, want 2026-12-01", row.ExpectDate)
	}
	if row.OrgUuid != "org-uuid-1" || row.OrgName != "研发部" {
		t.Fatalf("org snapshot wrong: uuid=%q name=%q", row.OrgUuid, row.OrgName)
	}
	if row.CreateBy != "user-1" || row.CreatorName != "张三HR" {
		t.Fatalf("creator snapshot wrong: uuid=%q name=%q", row.CreateBy, row.CreatorName)
	}
}

func TestAddValidation(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	cases := []struct {
		name  string
		mutat func(*types.AddRequest)
	}{
		{"missing post name", func(r *types.AddRequest) { r.PostName = "" }},
		{"bad post category", func(r *types.AddRequest) { r.PostCategory = "art" }},
		{"zero headcount", func(r *types.AddRequest) { r.Headcount = 0 }},
		{"negative headcount", func(r *types.AddRequest) { r.Headcount = -1 }},
		{"bad urgency", func(r *types.AddRequest) { r.Urgency = "whenever" }},
		{"missing org", func(r *types.AddRequest) { r.OrgUuid = "" }},
		{"bad expect date format", func(r *types.AddRequest) { r.ExpectDate = "2026/12/01" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{
				PostName:     "Go 工程师",
				PostCategory: PostCategoryTech,
				Headcount:    1,
				Urgency:      UrgencyNormal,
				OrgUuid:      "org-uuid-1",
			}
			tc.mutat(req)
			if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestLifecycleSubmitApproveViaBillBase(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	billUuid := addRequirement(t, svcCtx, ctx, nil)

	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	row := mustFindRequirement(t, svcCtx, billUuid)
	if row.Status != bill.StatusApproving {
		t.Fatalf("after submit status = %q, want approving", row.Status)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, "同意"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	row = mustFindRequirement(t, svcCtx, billUuid)
	if row.Status != bill.StatusApproved {
		t.Fatalf("after approve status = %q, want approved (可被招聘职位 HRM-25 引用)", row.Status)
	}

	// 审批记录留痕 via the shared 审批记录 seam
	records, err := bill.Records(context.Background(), env.SqlConn, BillTypeCode, billUuid)
	if err != nil {
		t.Fatalf("Records: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("approval records = %d, want 2 (submit+approve)", len(records))
	}
}

func TestEditDeleteGuardedByStatus(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	billUuid := addRequirement(t, svcCtx, ctx, nil)

	// 审批中 (approving) is locked: edit and delete must be rejected.
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid,
		PostName:     "Go 工程师",
		PostCategory: PostCategoryTech,
		Headcount:    3,
		Urgency:      UrgencyNormal,
		OrgUuid:      "org-uuid-1",
	}); err == nil {
		t.Fatal("edit during 审批中 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批中 should fail")
	}

	// 审批通过 (approved) is locked too — the 招聘职位 reference source.
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid,
		PostName:     "Go 工程师",
		PostCategory: PostCategoryTech,
		Headcount:    3,
		Urgency:      UrgencyNormal,
		OrgUuid:      "org-uuid-1",
	}); err == nil {
		t.Fatal("edit during 审批通过 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批通过 should fail")
	}

	// 已撤销 (revoked) is editable/deletable again.
	billUuid2 := addRequirement(t, svcCtx, ctx, nil)
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, ""); err != nil {
		t.Fatalf("SubmitBill 2: %v", err)
	}
	if err := bill.RevokeBill(context.Background(), env.SqlConn, BillTypeCode, billUuid2, op, "撤回"); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid2,
		PostName:     "产品经理",
		PostCategory: PostCategoryProduct,
		Headcount:    1,
		Urgency:      UrgencyUrgent,
		OrgUuid:      "org-uuid-2",
		OrgName:      "产品部",
	}); err != nil {
		t.Fatalf("edit after revoke: %v", err)
	}
	row := mustFindRequirement(t, svcCtx, billUuid2)
	if row.PostName != "产品经理" || row.Urgency != UrgencyUrgent || row.Headcount != 1 {
		t.Fatalf("edit after revoke did not apply: %+v", row)
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid2}}); err != nil {
		t.Fatalf("delete after revoke: %v", err)
	}
}

func TestListMineFilterAndAdminBypass(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	userCtx := testCtxFor("user-1", "张三HR")
	addRequirement(t, svcCtx, userCtx, nil)
	addRequirement(t, svcCtx, userCtx, func(r *types.AddRequest) {
		r.PostName = "产品经理"
		r.PostCategory = PostCategoryProduct
		r.Urgency = UrgencyUrgent
	})
	// another user's bill — must be invisible to user-1
	otherCtx := testCtxFor("user-2", "李四")
	addRequirement(t, svcCtx, otherCtx, nil)

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
	if allResp.Total != 3 {
		t.Fatalf("admin should see all bills: total=%d, want 3", allResp.Total)
	}

	forcedMine, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, Mine: 1})
	if err != nil {
		t.Fatalf("List(admin, mine=1): %v", err)
	}
	if forcedMine.Total != 0 {
		t.Fatalf("admin with mine=1 should see none of user-1's bills: total=%d", forcedMine.Total)
	}

	// 筛选：岗位名称 (demo-inventory §5)
	nameFilter, err := NewList(userCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		PostName:    "产品",
	})
	if err != nil {
		t.Fatalf("List(postName filter): %v", err)
	}
	if nameFilter.Total != 1 || nameFilter.Records[0].PostName != "产品经理" {
		t.Fatalf("postName filter: total=%d, want 1 产品经理", nameFilter.Total)
	}

	// 筛选：单据状态 (demo-inventory §5)
	statusFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		Status:      bill.StatusUnsubmitted,
	})
	if err != nil {
		t.Fatalf("List(status filter): %v", err)
	}
	if statusFilter.Total != 3 {
		t.Fatalf("status filter: total=%d, want 3 unsubmitted", statusFilter.Total)
	}

	// 筛选：紧急度
	urgencyFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		Urgency:     UrgencyUrgent,
	})
	if err != nil {
		t.Fatalf("List(urgency filter): %v", err)
	}
	if urgencyFilter.Total != 1 || urgencyFilter.Records[0].Urgency != UrgencyUrgent {
		t.Fatalf("urgency filter: total=%d, want 1 urgent", urgencyFilter.Total)
	}
}

func mustFindRequirement(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) *hrm_recruitment_requirement.HrmRecruitmentRequirement {
	t.Helper()
	row, err := svcCtx.Model.HrmRecruitmentRequirement.FindOneByUuid(context.Background(), nil, billUuid)
	if err != nil {
		t.Fatalf("find requirement %s: %v", billUuid, err)
	}
	return row
}

func superAdminRoleUuid() string {
	// keep the seeded R_SUPER uuid local to tests; production logic uses hrmuser
	return "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
}
