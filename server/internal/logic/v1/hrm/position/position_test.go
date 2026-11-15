package position

import (
	"context"
	"testing"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/requirement"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/position"
	reqtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/requirement"
)

// hrmRecruitmentPosition aliases the generated model row for test helpers.
type hrmRecruitmentPosition = model.HrmRecruitmentPosition

// envSqlConn extracts the raw connection from a test ServiceContext.
func envSqlConn(svcCtx *svc.ServiceContext) sqlx.SqlConn {
	return svcCtx.SqlxConn
}

// Seams under test (per ticket acceptance criteria), all at the logic-method
// seam against the real migrations (svctest):
//  1. Add: 来源需求单必填且仅 status='approved' 可引用（非 approved 拒绝 1058x）；
//     默认状态 招聘中(open)，招聘人数>=1 校验.
//  2. List filters: 岗位名称（模糊）/来源需求单/岗位类别/状态（demo-inventory §5）.
//  3. Edit/Delete: 普通实体 CRUD（无审批锁），headcount 不得小于已录用人数.
//  4. UpdateStatus: closed 手工关闭终态；full/closed 不可手工互转.
//  5. IncreaseHiredCount（HRM-27 契约）: 事务化递增，达 headcount 自动置 已满(full)，
//     超招拒绝（10585），closed 职位拒绝.
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

// addApprovedRequirement drives a 需求单 through the shared 单据底座 to approved,
// returning its uuid — the only kind of requirement a position may reference.
func addApprovedRequirement(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, headcount int64) string {
	t.Helper()
	if _, err := requirement.NewAdd(ctx, svcCtx, nil).Add(&reqtypes.AddRequest{
		PostName:     "Go 工程师",
		PostCategory: requirement.PostCategoryTech,
		Headcount:    headcount,
		Urgency:      requirement.UrgencyNormal,
		OrgUuid:      "org-uuid-1",
		OrgName:      "研发部",
	}); err != nil {
		t.Fatalf("requirement Add: %v", err)
	}
	list, err := requirement.NewList(ctx, svcCtx, nil).List(&reqtypes.ListRequest{
		PageRequest: reqtypes.PageRequest{Current: 1, Size: 10},
	})
	if err != nil {
		t.Fatalf("requirement List: %v", err)
	}
	if len(list.Records) == 0 {
		t.Fatal("no requirement found after add")
	}
	// take the most recent one created by this call (list is OrderByDesc(id))
	reqUuid := list.Records[0].Uuid
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), envSqlConn(svcCtx), requirement.BillTypeCode, reqUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), envSqlConn(svcCtx), requirement.BillTypeCode, reqUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	return reqUuid
}

func addPosition(t *testing.T, svcCtx *svc.ServiceContext, reqUuid string, mutate func(*types.AddRequest)) string {
	t.Helper()
	req := &types.AddRequest{
		PostName:        "Go 工程师",
		RequirementUuid: reqUuid,
		PostCategory:    requirement.PostCategoryTech,
		OwnerUuid:       "owner-1",
		OwnerName:       "王招聘",
		Headcount:       2,
		PublishDate:     "2026-11-16",
	}
	if mutate != nil {
		mutate(req)
	}
	if _, err := NewAdd(context.Background(), svcCtx, nil).Add(req); err != nil {
		t.Fatalf("Add: %v", err)
	}
	// fetch through the public list seam, filtered to this requirement
	resp, err := NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest:     types.PageRequest{Current: 1, Size: 10},
		RequirementUuid: reqUuid,
	})
	if err != nil {
		t.Fatalf("List after Add: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("list total = %d, want 1", resp.Total)
	}
	return resp.Records[0].Uuid
}

func TestAddRejectsNonApprovedRequirement(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	// unsubmitted requirement must be rejected
	if _, err := requirement.NewAdd(ctx, svcCtx, nil).Add(&reqtypes.AddRequest{
		PostName:     "Go 工程师",
		PostCategory: requirement.PostCategoryTech,
		Headcount:    2,
		Urgency:      requirement.UrgencyNormal,
		OrgUuid:      "org-uuid-1",
	}); err != nil {
		t.Fatalf("requirement Add: %v", err)
	}
	unsubList, err := requirement.NewList(ctx, svcCtx, nil).List(&reqtypes.ListRequest{
		PageRequest: reqtypes.PageRequest{Current: 1, Size: 10},
	})
	if err != nil {
		t.Fatalf("requirement List: %v", err)
	}
	unsubUuid := unsubList.Records[0].Uuid

	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		PostName:        "Go 工程师",
		RequirementUuid: unsubUuid,
		PostCategory:    requirement.PostCategoryTech,
		Headcount:       1,
	}); err == nil {
		t.Fatal("position referencing unsubmitted requirement should fail")
	}

	// approving unlocks the reference
	if err := bill.SubmitBill(context.Background(), envSqlConn(svcCtx), requirement.BillTypeCode, unsubUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), envSqlConn(svcCtx), requirement.BillTypeCode, unsubUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		PostName:        "Go 工程师",
		RequirementUuid: unsubUuid,
		PostCategory:    requirement.PostCategoryTech,
		Headcount:       1,
	}); err != nil {
		t.Fatalf("position referencing approved requirement: %v", err)
	}
}

func TestAddValidation(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	reqUuid := addApprovedRequirement(t, svcCtx, ctx, 2)

	cases := []struct {
		name   string
		mutate func(*types.AddRequest)
	}{
		{"missing post name", func(r *types.AddRequest) { r.PostName = "" }},
		{"bad post category", func(r *types.AddRequest) { r.PostCategory = "art" }},
		{"zero headcount", func(r *types.AddRequest) { r.Headcount = 0 }},
		{"missing requirement", func(r *types.AddRequest) { r.RequirementUuid = "" }},
		{"nonexistent requirement", func(r *types.AddRequest) { r.RequirementUuid = "no-such-uuid" }},
		{"bad publish date", func(r *types.AddRequest) { r.PublishDate = "2026/11/16" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{
				PostName:        "Go 工程师",
				RequirementUuid: reqUuid,
				PostCategory:    requirement.PostCategoryTech,
				Headcount:       1,
			}
			tc.mutate(req)
			if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestAddDefaultsAndListFilters(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	reqUuid := addApprovedRequirement(t, svcCtx, ctx, 3)
	reqUuid2 := addApprovedRequirement(t, svcCtx, ctx, 1)
	posUuid := addPosition(t, svcCtx, reqUuid, nil)

	// defaults: open 招聘中, hired_count 0
	resp, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	rec := resp.Records[0]
	if rec.Uuid != posUuid {
		t.Fatalf("uuid mismatch: %s", rec.Uuid)
	}
	if rec.Status != PositionStatusOpen {
		t.Fatalf("new position status = %q, want open", rec.Status)
	}
	if rec.HiredCount != 0 {
		t.Fatalf("hiredCount = %d, want 0", rec.HiredCount)
	}
	if rec.OwnerName != "王招聘" {
		t.Fatalf("ownerName = %q, want 王招聘", rec.OwnerName)
	}
	if rec.RequirementBillNo == "" {
		t.Fatal("requirementBillNo snapshot empty")
	}

	// second position against a different requirement
	addPosition(t, svcCtx, reqUuid2, func(r *types.AddRequest) {
		r.PostName = "产品经理"
		r.PostCategory = requirement.PostCategoryProduct
	})

	// 筛选：岗位名称（模糊）
	byName, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		PostName:    "产品",
	})
	if err != nil {
		t.Fatalf("List(postName): %v", err)
	}
	if byName.Total != 1 || byName.Records[0].PostName != "产品经理" {
		t.Fatalf("postName filter: total=%d", byName.Total)
	}

	// 筛选：来源需求单
	byReq, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest:     types.PageRequest{Current: 1, Size: 10},
		RequirementUuid: reqUuid,
	})
	if err != nil {
		t.Fatalf("List(requirementUuid): %v", err)
	}
	if byReq.Total != 1 || byReq.Records[0].RequirementUuid != reqUuid {
		t.Fatalf("requirementUuid filter: total=%d", byReq.Total)
	}

	// 筛选：岗位类别
	byCat, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest:  types.PageRequest{Current: 1, Size: 10},
		PostCategory: requirement.PostCategoryProduct,
	})
	if err != nil {
		t.Fatalf("List(postCategory): %v", err)
	}
	if byCat.Total != 1 {
		t.Fatalf("postCategory filter: total=%d", byCat.Total)
	}

	// 筛选：状态
	byStatus, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		Status:      PositionStatusOpen,
	})
	if err != nil {
		t.Fatalf("List(status): %v", err)
	}
	if byStatus.Total != 2 {
		t.Fatalf("status filter: total=%d, want 2", byStatus.Total)
	}
}

func TestEditDeleteNormalEntity(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	reqUuid := addApprovedRequirement(t, svcCtx, ctx, 3)
	posUuid := addPosition(t, svcCtx, reqUuid, nil)

	// 普通实体：无审批锁，直接编辑
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:            posUuid,
		PostName:        "高级 Go 工程师",
		RequirementUuid: reqUuid,
		PostCategory:    requirement.PostCategoryTech,
		OwnerUuid:       "owner-2",
		OwnerName:       "赵招聘",
		Headcount:       3,
		PublishDate:     "2026-11-17",
	}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	resp, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	rec := resp.Records[0]
	if rec.PostName != "高级 Go 工程师" || rec.Headcount != 3 || rec.OwnerName != "赵招聘" {
		t.Fatalf("edit not applied: %+v", rec)
	}

	// headcount 不得小于已录用人数（先递增 2）
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		PostName:        "临时职位",
		RequirementUuid: reqUuid,
		PostCategory:    requirement.PostCategoryTech,
		Headcount:       2,
	}); err != nil {
		t.Fatalf("Add 2: %v", err)
	}
	list2, _ := NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	var pos2 string
	for _, r := range list2.Records {
		if r.PostName == "临时职位" {
			pos2 = r.Uuid
		}
	}
	if err := IncreaseHiredCount(context.Background(), svcCtx, pos2); err != nil {
		t.Fatalf("IncreaseHiredCount: %v", err)
	}
	if err := IncreaseHiredCount(context.Background(), svcCtx, pos2); err != nil {
		t.Fatalf("IncreaseHiredCount 2: %v", err)
	}
	// headcount < hired_count must be rejected
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:            pos2,
		PostName:        "临时职位",
		RequirementUuid: reqUuid,
		PostCategory:    requirement.PostCategoryTech,
		Headcount:       1,
	}); err == nil {
		t.Fatal("edit lowering headcount below hired_count should fail")
	}

	// delete 不限状态（普通实体）
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{pos2}}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{pos2}}); err == nil {
		t.Fatal("delete nonexistent should fail")
	}
}

func TestUpdateStatusGuards(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	reqUuid := addApprovedRequirement(t, svcCtx, ctx, 2)
	posUuid := addPosition(t, svcCtx, reqUuid, nil)

	// open→full 手工流转：允许（演示口径），closed 为终态
	if _, err := NewUpdateStatus(ctx, svcCtx, nil).UpdateStatus(&types.UpdateStatusRequest{
		Uuid: posUuid, Status: PositionStatusClosed,
	}); err != nil {
		t.Fatalf("open->closed: %v", err)
	}
	// closed 为终态，不可再流转
	if _, err := NewUpdateStatus(ctx, svcCtx, nil).UpdateStatus(&types.UpdateStatusRequest{
		Uuid: posUuid, Status: PositionStatusOpen,
	}); err == nil {
		t.Fatal("closed->open should fail")
	}
	if _, err := NewUpdateStatus(ctx, svcCtx, nil).UpdateStatus(&types.UpdateStatusRequest{
		Uuid: posUuid, Status: PositionStatusFull,
	}); err == nil {
		t.Fatal("closed->full should fail")
	}
	// 非法状态值
	if _, err := NewUpdateStatus(ctx, svcCtx, nil).UpdateStatus(&types.UpdateStatusRequest{
		Uuid: posUuid, Status: "paused",
	}); err == nil {
		t.Fatal("invalid status should fail")
	}
	// 不存在的职位
	if _, err := NewUpdateStatus(ctx, svcCtx, nil).UpdateStatus(&types.UpdateStatusRequest{
		Uuid: "no-such-uuid", Status: PositionStatusClosed,
	}); err == nil {
		t.Fatal("nonexistent position should fail")
	}
}

func TestIncreaseHiredCountAutoFull(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	reqUuid := addApprovedRequirement(t, svcCtx, ctx, 2)
	posUuid := addPosition(t, svcCtx, reqUuid, nil)

	// first increment: hired=1, still open
	if err := IncreaseHiredCount(context.Background(), svcCtx, posUuid); err != nil {
		t.Fatalf("IncreaseHiredCount 1: %v", err)
	}
	row := findPosition(t, svcCtx, posUuid)
	if row.HiredCount != 1 || row.Status != PositionStatusOpen {
		t.Fatalf("after 1st increment hired=%d status=%s, want 1/open", row.HiredCount, row.Status)
	}

	// second increment: hired=2 == headcount → auto 已满(full)
	if err := IncreaseHiredCount(context.Background(), svcCtx, posUuid); err != nil {
		t.Fatalf("IncreaseHiredCount 2: %v", err)
	}
	row = findPosition(t, svcCtx, posUuid)
	if row.HiredCount != 2 || row.Status != PositionStatusFull {
		t.Fatalf("after 2nd increment hired=%d status=%s, want 2/full", row.HiredCount, row.Status)
	}

	// 超招拒绝
	if err := IncreaseHiredCount(context.Background(), svcCtx, posUuid); err == nil {
		t.Fatal("increment beyond headcount should fail")
	}

	// closed 职位拒绝递增
	reqUuid2 := addApprovedRequirement(t, svcCtx, ctx, 1)
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		PostName:        "关闭职位",
		RequirementUuid: reqUuid2,
		PostCategory:    requirement.PostCategoryTech,
		Headcount:       1,
	}); err != nil {
		t.Fatalf("Add closed position: %v", err)
	}
	list, _ := NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	var closedUuid string
	for _, r := range list.Records {
		if r.PostName == "关闭职位" {
			closedUuid = r.Uuid
		}
	}
	if _, err := NewUpdateStatus(ctx, svcCtx, nil).UpdateStatus(&types.UpdateStatusRequest{
		Uuid: closedUuid, Status: PositionStatusClosed,
	}); err != nil {
		t.Fatalf("close position: %v", err)
	}
	if err := IncreaseHiredCount(context.Background(), svcCtx, closedUuid); err == nil {
		t.Fatal("increment on closed position should fail")
	}

	// 不存在的职位
	if err := IncreaseHiredCount(context.Background(), svcCtx, "no-such-uuid"); err == nil {
		t.Fatal("increment on nonexistent position should fail")
	}
}

func findPosition(t *testing.T, svcCtx *svc.ServiceContext, uuid string) *hrmRecruitmentPosition {
	t.Helper()
	row, err := svcCtx.Model.HrmRecruitmentPosition.FindOneByUuid(context.Background(), nil, uuid)
	if err != nil {
		t.Fatalf("find position %s: %v", uuid, err)
	}
	return row
}
