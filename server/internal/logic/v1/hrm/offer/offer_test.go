package offer

import (
	"context"
	"strings"
	"testing"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	candidate "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/candidate"
	position "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/position"
	requirement "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/requirement"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_offer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_candidate"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	candtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/candidate"
	offertypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/offer"
	postypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/position"
	reqtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/requirement"
)

// Seams under test (HRM-27 acceptance criteria), all at the logic seam against
// the real migrations (svctest):
//  1. Add: 单据编号 HR208-、初始状态 未提交；候选人姓名/岗位/部门（需求单快照）落快照；
//     职级默认 初级(junior).
//  2. 守卫：候选人非 offer 态拒绝（10653）；职位非 open 态拒绝（10654）；必填校验（10651）.
//  3. 审批通过事务内三联动：职位已录用+1（自动 full）、candidate→hired、
//     草稿 HR201 入职单存在且字段正确（姓名=候选人、入职日期=计划入职日期、试用期、部门职位快照）.
//  4. 幂等：重放审批（直接再调 onApproved）不重复三联动.
//  5. 生命周期守卫：edit/delete 仅 可编辑状态；mine-by-default + 筛选
//     （单据编号/状态/候选人）.

func testCtxFor(uuid, username string) context.Context {
	return context.WithValue(context.WithValue(context.Background(),
		"uuid", uuid), "username", username)
}

func superCtx(uuid, username string) context.Context {
	ctx := testCtxFor(uuid, username)
	return context.WithValue(ctx, "role_uuids", []any{"1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"})
}

func newSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		ServiceContext: &coresvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

func envSqlConn(svcCtx *svc.ServiceContext) sqlx.SqlConn {
	return svcCtx.SqlxConn
}

func offerErrCode(err error) int {
	if err == nil {
		return 0
	}
	if sc := status.FromError(err); sc != nil {
		return int(sc.Code())
	}
	return -1
}

// ---- shared fixture: approved requirement → open position → offer-stage candidate → offer bill ----

func addApprovedRequirement(t *testing.T, svcCtx *svc.ServiceContext, orgName string) string {
	t.Helper()
	ctx := testCtxFor("user-1", "张三HR")
	if _, err := requirement.NewAdd(ctx, svcCtx, nil).Add(&reqtypes.AddRequest{
		PostName:     "Go 工程师",
		PostCategory: requirement.PostCategoryTech,
		Headcount:    2,
		Urgency:      requirement.UrgencyNormal,
		OrgUuid:      "org-uuid-1",
		OrgName:      orgName,
	}); err != nil {
		t.Fatalf("requirement Add: %v", err)
	}
	list, err := requirement.NewList(ctx, svcCtx, nil).List(&reqtypes.ListRequest{
		PageRequest: reqtypes.PageRequest{Current: 1, Size: 10},
	})
	if err != nil {
		t.Fatalf("requirement List: %v", err)
	}
	reqUuid := list.Records[0].Uuid
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), envSqlConn(svcCtx), requirement.BillTypeCode, reqUuid, op, ""); err != nil {
		t.Fatalf("requirement SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), envSqlConn(svcCtx), requirement.BillTypeCode, reqUuid, op, ""); err != nil {
		t.Fatalf("requirement ApproveBill: %v", err)
	}
	return reqUuid
}

func addOpenPosition(t *testing.T, svcCtx *svc.ServiceContext, reqUuid string, headcount int64) string {
	t.Helper()
	ctx := testCtxFor("user-1", "张三HR")
	if _, err := position.NewAdd(ctx, svcCtx, nil).Add(&postypes.AddRequest{
		PostName:        "Go 工程师",
		RequirementUuid: reqUuid,
		PostCategory:    "tech",
		OwnerName:       "王招聘",
		Headcount:       headcount,
	}); err != nil {
		t.Fatalf("position Add: %v", err)
	}
	list, err := position.NewList(ctx, svcCtx, nil).List(&postypes.ListRequest{
		PageRequest: postypes.PageRequest{Current: 1, Size: 10},
	})
	if err != nil {
		t.Fatalf("position List: %v", err)
	}
	return list.Records[0].Uuid
}

func positionRow(t *testing.T, svcCtx *svc.ServiceContext, positionUuid string) *hrm_recruitment_position.HrmRecruitmentPosition {
	t.Helper()
	row, err := svcCtx.Model.HrmRecruitmentPosition.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(hrm_recruitment_position.Uuid, positionUuid).Build()...)
	if err != nil {
		t.Fatalf("load position: %v", err)
	}
	return row
}

// addOfferStageCandidate seeds a candidate at offer（已发Offer）stage via the
// public Add + UpdateStatus seams (new→delivered→interviewing→offer).
func addOfferStageCandidate(t *testing.T, svcCtx *svc.ServiceContext, name string) string {
	t.Helper()
	ctx := context.Background()
	if _, err := candidate.NewAdd(ctx, svcCtx, nil).Add(&candtypes.AddRequest{
		Name:  name,
		Phone: "13800000001",
		Email: name + "@example.com",
	}); err != nil {
		t.Fatalf("candidate Add: %v", err)
	}
	list, err := candidate.NewList(ctx, svcCtx, nil).List(&candtypes.ListRequest{
		PageRequest: candtypes.PageRequest{Current: 1, Size: 50},
	})
	if err != nil {
		t.Fatalf("candidate List: %v", err)
	}
	var uuid string
	for _, rec := range list.Records {
		if rec.Name == name {
			uuid = rec.Uuid
		}
	}
	if uuid == "" {
		t.Fatal("candidate not found after add")
	}
	for _, next := range []string{candidate.CandidateStatusDelivered, candidate.CandidateStatusInterviewing, candidate.CandidateStatusOffer} {
		if _, err := candidate.NewUpdateStatus(ctx, svcCtx, nil).UpdateStatus(&candtypes.UpdateStatusRequest{
			Uuid: uuid, Status: next,
		}); err != nil {
			t.Fatalf("candidate advance to %s: %v", next, err)
		}
	}
	return uuid
}

func addOffer(t *testing.T, svcCtx *svc.ServiceContext, candidateUuid, positionUuid string, mutate func(*offertypes.AddRequest)) string {
	t.Helper()
	ctx := testCtxFor("user-1", "张三HR")
	req := &offertypes.AddRequest{
		CandidateUuid:   candidateUuid,
		PositionUuid:    positionUuid,
		PlanEntryDate:   "2026-12-01",
		ProbationMonths: 3,
		RankLevel:       RankLevelIntermediate,
		ReporterName:    "陈经理",
	}
	if mutate != nil {
		mutate(req)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("offer Add: %v", err)
	}
	resp, err := NewList(testCtxFor("user-1", "张三HR"), svcCtx, nil).List(&offertypes.ListRequest{
		PageRequest: offertypes.PageRequest{Current: 1, Size: 10},
	})
	if err != nil {
		t.Fatalf("offer List: %v", err)
	}
	if len(resp.Records) == 0 {
		t.Fatal("offer bill not visible in 我的单据 list")
	}
	rec := resp.Records[0]
	if !strings.HasPrefix(rec.BillNo, "HR208-") {
		t.Fatalf("billNo %q does not follow HR208- format", rec.BillNo)
	}
	if rec.Status != bill.StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	return rec.Uuid
}

func submitAndApprove(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) {
	t.Helper()
	op := bill.Operator{UserUuid: "op-1", Name: "管理员"}
	if err := bill.SubmitBill(context.Background(), envSqlConn(svcCtx), BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), envSqlConn(svcCtx),
		BillTypeCode, billUuid, op, "同意"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
}

func candidateStatusOf(t *testing.T, svcCtx *svc.ServiceContext, candidateUuid string) string {
	t.Helper()
	row, err := svcCtx.Model.HrmRecruitmentCandidate.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(hrm_recruitment_candidate.Uuid, candidateUuid).Build()...)
	if err != nil {
		t.Fatalf("load candidate: %v", err)
	}
	return row.Status
}

func entryBillsOf(t *testing.T, svcCtx *svc.ServiceContext, name string) []*hrm_entry_bill.HrmEntryBill {
	t.Helper()
	rows, err := svcCtx.Model.HrmEntryBill.FindByCondition(context.Background(), nil,
		condition.NewChain().Equal(hrm_entry_bill.Name, name).Build()...)
	if err != nil {
		t.Fatalf("load entry bills: %v", err)
	}
	return rows
}

func offerRowOf(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) *hrm_offer_bill.HrmOfferBill {
	t.Helper()
	row, err := svcCtx.Model.HrmOfferBill.FindOneByUuid(context.Background(), nil, billUuid)
	if err != nil {
		t.Fatalf("load offer bill: %v", err)
	}
	return row
}

// ---- tests ----

func TestAddSnapshotsAndDefaults(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	reqUuid := addApprovedRequirement(t, svcCtx, "研发部")
	posUuid := addOpenPosition(t, svcCtx, reqUuid, 2)
	candUuid := addOfferStageCandidate(t, svcCtx, "王五")

	billUuid := addOffer(t, svcCtx, candUuid, posUuid, nil)

	row := offerRowOf(t, svcCtx, billUuid)
	if row.CandidateName != "王五" {
		t.Fatalf("candidateName snapshot = %q, want 王五", row.CandidateName)
	}
	if row.PostName != "Go 工程师" {
		t.Fatalf("postName snapshot = %q, want Go 工程师", row.PostName)
	}
	if row.OrgUuid != "org-uuid-1" || row.OrgName != "研发部" {
		t.Fatalf("org snapshot wrong: uuid=%q name=%q", row.OrgUuid, row.OrgName)
	}
	if row.RankLevel != RankLevelIntermediate || row.ReporterName != "陈经理" ||
		row.PlanEntryDate != "2026-12-01" || row.ProbationMonths != 3 {
		t.Fatalf("offer fields wrong: %+v", row)
	}
	if row.CreateBy != "user-1" || row.CreatorName != "张三HR" {
		t.Fatalf("creator snapshot wrong: %+v", row)
	}
}

func TestAddRankLevelDefaultsToJunior(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	reqUuid := addApprovedRequirement(t, svcCtx, "研发部")
	posUuid := addOpenPosition(t, svcCtx, reqUuid, 2)
	candUuid := addOfferStageCandidate(t, svcCtx, "王五")

	billUuid := addOffer(t, svcCtx, candUuid, posUuid, func(req *offertypes.AddRequest) {
		req.RankLevel = ""
	})
	if got := offerRowOf(t, svcCtx, billUuid).RankLevel; got != RankLevelJunior {
		t.Fatalf("rankLevel = %q, want junior", got)
	}
}

func TestAddGuards(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	reqUuid := addApprovedRequirement(t, svcCtx, "研发部")
	posUuid := addOpenPosition(t, svcCtx, reqUuid, 2)
	candUuid := addOfferStageCandidate(t, svcCtx, "王五")

	// 候选人非 offer 态拒绝：new 阶段候选人直接发起录用 → 10653
	newCandUuid := func() string {
		ctx := context.Background()
		if _, err := candidate.NewAdd(ctx, svcCtx, nil).Add(&candtypes.AddRequest{Name: "赵六"}); err != nil {
			t.Fatalf("candidate Add: %v", err)
		}
		list, _ := candidate.NewList(ctx, svcCtx, nil).List(&candtypes.ListRequest{
			PageRequest: candtypes.PageRequest{Current: 1, Size: 50}})
		for _, rec := range list.Records {
			if rec.Name == "赵六" {
				return rec.Uuid
			}
		}
		t.Fatal("new candidate not found")
		return ""
	}()
	_, err := NewAdd(testCtxFor("user-1", "张三HR"), svcCtx, nil).Add(&offertypes.AddRequest{
		CandidateUuid: newCandUuid, PositionUuid: posUuid, PlanEntryDate: "2026-12-01",
	})
	if code := offerErrCode(err); code != 10653 {
		t.Fatalf("non-offer candidate err code = %d (%v), want 10653", code, err)
	}

	// 职位非 open 态拒绝：closed 职位 → 10654
	if _, err := position.NewUpdateStatus(testCtxFor("user-1", "张三HR"), svcCtx, nil).UpdateStatus(&postypes.UpdateStatusRequest{
		Uuid: posUuid, Status: position.PositionStatusClosed,
	}); err != nil {
		t.Fatalf("close position: %v", err)
	}
	_, err = NewAdd(testCtxFor("user-1", "张三HR"), svcCtx, nil).Add(&offertypes.AddRequest{
		CandidateUuid: candUuid, PositionUuid: posUuid, PlanEntryDate: "2026-12-01",
	})
	if code := offerErrCode(err); code != 10654 {
		t.Fatalf("non-open position err code = %d (%v), want 10654", code, err)
	}

	// 必填校验（候选人/职位/计划入职日期）→ 10651
	for _, req := range []*offertypes.AddRequest{
		{PositionUuid: posUuid, PlanEntryDate: "2026-12-01"},
		{CandidateUuid: candUuid, PlanEntryDate: "2026-12-01"},
		{CandidateUuid: candUuid, PositionUuid: posUuid},
	} {
		if _, err = NewAdd(testCtxFor("user-1", "张三HR"), svcCtx, nil).Add(req); offerErrCode(err) != 10651 {
			t.Fatalf("missing-field err code = %d (%v), want 10651", offerErrCode(err), err)
		}
	}
}

func TestApproveThreeLinks(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	reqUuid := addApprovedRequirement(t, svcCtx, "研发部")
	posUuid := addOpenPosition(t, svcCtx, reqUuid, 2)
	candUuid := addOfferStageCandidate(t, svcCtx, "王五")

	billUuid := addOffer(t, svcCtx, candUuid, posUuid, nil)
	submitAndApprove(t, svcCtx, billUuid)

	// 1) 职位已录用 +1（headcount=2，仍 open）
	pos := positionRow(t, svcCtx, posUuid)
	if pos.HiredCount != 1 {
		t.Fatalf("hiredCount = %d, want 1", pos.HiredCount)
	}
	if pos.Status != position.PositionStatusOpen {
		t.Fatalf("position status = %q, want open", pos.Status)
	}

	// 2) candidate → hired
	if got := candidateStatusOf(t, svcCtx, candUuid); got != candidate.CandidateStatusHired {
		t.Fatalf("candidate status = %q, want hired", got)
	}

	// 3) 草稿 HR201 入职单存在且字段正确
	rows := entryBillsOf(t, svcCtx, "王五")
	if len(rows) != 1 {
		t.Fatalf("entry bills for 王五 = %d, want 1", len(rows))
	}
	entry := rows[0]
	if !strings.HasPrefix(entry.BillNo, "HR201-") {
		t.Fatalf("entry billNo %q does not follow HR201- format", entry.BillNo)
	}
	if entry.Status != bill.StatusUnsubmitted {
		t.Fatalf("entry status = %q, want unsubmitted (草稿)", entry.Status)
	}
	if entry.HireDate != "2026-12-01" {
		t.Fatalf("entry hireDate = %q, want plan_entry_date 2026-12-01", entry.HireDate)
	}
	if entry.ProbationMonths != 3 {
		t.Fatalf("entry probationMonths = %d, want 3", entry.ProbationMonths)
	}
	if entry.Position != "Go 工程师" {
		t.Fatalf("entry position = %q, want 职位单快照 Go 工程师", entry.Position)
	}
	if entry.OrgUuid != "org-uuid-1" {
		t.Fatalf("entry orgUuid = %q, want org-uuid-1", entry.OrgUuid)
	}
	if entry.Phone != "13800000001" {
		t.Fatalf("entry phone = %q, want candidate phone", entry.Phone)
	}

	// 回写标记
	row := offerRowOf(t, svcCtx, billUuid)
	if row.Applied != 1 || row.EntryBillUuid != entry.Uuid {
		t.Fatalf("offer write-back wrong: applied=%d entryUuid=%q", row.Applied, row.EntryBillUuid)
	}
}

func TestApproveAutoFullWhenHeadcountReached(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	reqUuid := addApprovedRequirement(t, svcCtx, "研发部")
	posUuid := addOpenPosition(t, svcCtx, reqUuid, 1) // headcount=1

	candUuid := addOfferStageCandidate(t, svcCtx, "王五")
	billUuid := addOffer(t, svcCtx, candUuid, posUuid, nil)
	submitAndApprove(t, svcCtx, billUuid)

	pos := positionRow(t, svcCtx, posUuid)
	if pos.HiredCount != 1 || pos.Status != position.PositionStatusFull {
		t.Fatalf("position after approve = hired %d / %q, want 1/full", pos.HiredCount, pos.Status)
	}
}

func TestApproveIdempotentReplay(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	reqUuid := addApprovedRequirement(t, svcCtx, "研发部")
	posUuid := addOpenPosition(t, svcCtx, reqUuid, 2)
	candUuid := addOfferStageCandidate(t, svcCtx, "王五")

	billUuid := addOffer(t, svcCtx, candUuid, posUuid, nil)
	submitAndApprove(t, svcCtx, billUuid)

	// 重放审批事务回调（模拟重复回写）：applied=1 直接 return nil
	if err := onApproved(context.Background(), env.SqlConn, billUuid); err != nil {
		t.Fatalf("onApproved replay: %v", err)
	}

	if got := positionRow(t, svcCtx, posUuid).HiredCount; got != 1 {
		t.Fatalf("hiredCount after replay = %d, want 1 (不重复递增)", got)
	}
	if rows := entryBillsOf(t, svcCtx, "王五"); len(rows) != 1 {
		t.Fatalf("entry bills after replay = %d, want 1 (不重复建单)", len(rows))
	}
	if got := candidateStatusOf(t, svcCtx, candUuid); got != candidate.CandidateStatusHired {
		t.Fatalf("candidate status after replay = %q, want hired", got)
	}
}

func TestEditDeleteLifecycleGuards(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	reqUuid := addApprovedRequirement(t, svcCtx, "研发部")
	posUuid := addOpenPosition(t, svcCtx, reqUuid, 2)
	candUuid := addOfferStageCandidate(t, svcCtx, "王五")

	billUuid := addOffer(t, svcCtx, candUuid, posUuid, nil)

	// 未提交可编辑
	if _, err := NewEdit(testCtxFor("user-1", "张三HR"), svcCtx, nil).Edit(&offertypes.EditRequest{
		Uuid: billUuid, CandidateUuid: candUuid, PositionUuid: posUuid,
		PlanEntryDate: "2027-01-05", ProbationMonths: 6, RankLevel: RankLevelSenior, ReporterName: "刘总",
	}); err != nil {
		t.Fatalf("Edit unsubmitted: %v", err)
	}
	row := offerRowOf(t, svcCtx, billUuid)
	if row.PlanEntryDate != "2027-01-05" || row.ProbationMonths != 6 || row.RankLevel != RankLevelSenior {
		t.Fatalf("edit fields wrong: %+v", row)
	}

	// 提交后（审批中）不可编辑/删除
	op := bill.Operator{UserUuid: "op-1", Name: "管理员"}
	if err := bill.SubmitBill(context.Background(), envSqlConn(svcCtx), BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(testCtxFor("user-1", "张三HR"), svcCtx, nil).Edit(&offertypes.EditRequest{
		Uuid: billUuid, CandidateUuid: candUuid, PositionUuid: posUuid, PlanEntryDate: "2027-02-01",
	}); offerErrCode(err) != 10652 {
		t.Fatalf("edit approving err code = %d (%v), want 10652", offerErrCode(err), err)
	}
	if _, err := NewDelete(testCtxFor("user-1", "张三HR"), svcCtx, nil).Delete(&offertypes.DeleteRequest{
		Uuids: []string{billUuid},
	}); offerErrCode(err) != 10652 {
		t.Fatalf("delete approving err code = %d (%v), want 10652", offerErrCode(err), err)
	}

	// 审批不通过后可删除
	if err := bill.RejectBill(context.Background(), envSqlConn(svcCtx), BillTypeCode, billUuid, op, "不合适"); err != nil {
		t.Fatalf("RejectBill: %v", err)
	}
	if _, err := NewDelete(testCtxFor("user-1", "张三HR"), svcCtx, nil).Delete(&offertypes.DeleteRequest{
		Uuids: []string{billUuid},
	}); err != nil {
		t.Fatalf("delete rejected: %v", err)
	}
}

func TestListMineDefaultAndFilters(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	reqUuid := addApprovedRequirement(t, svcCtx, "研发部")
	posUuid := addOpenPosition(t, svcCtx, reqUuid, 2)
	candUuid := addOfferStageCandidate(t, svcCtx, "王五")

	billUuid := addOffer(t, svcCtx, candUuid, posUuid, nil)

	// 默认（无 mine）：非管理员只看自己的；管理员看全部
	userList, err := NewList(testCtxFor("user-1", "张三HR"), svcCtx, nil).List(&offertypes.ListRequest{
		PageRequest: offertypes.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("user List: %v", err)
	}
	if len(userList.Records) != 1 {
		t.Fatalf("user list len = %d, want 1 (mine-by-default)", len(userList.Records))
	}
	adminList, err := NewList(superCtx("admin-1", "管理员"), svcCtx, nil).List(&offertypes.ListRequest{
		PageRequest: offertypes.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("admin List: %v", err)
	}
	if len(adminList.Records) != 1 {
		t.Fatalf("admin list len = %d, want 1", len(adminList.Records))
	}

	// 单据编号模糊
	filtered, err := NewList(superCtx("admin-1", "管理员"), svcCtx, nil).List(&offertypes.ListRequest{
		PageRequest: offertypes.PageRequest{Current: 1, Size: 10}, BillNo: "HR208-"})
	if err != nil {
		t.Fatalf("filter List: %v", err)
	}
	if len(filtered.Records) != 1 {
		t.Fatalf("billNo filter len = %d, want 1", len(filtered.Records))
	}
	// 状态筛选
	filtered, err = NewList(superCtx("admin-1", "管理员"), svcCtx, nil).List(&offertypes.ListRequest{
		PageRequest: offertypes.PageRequest{Current: 1, Size: 10}, Status: bill.StatusUnsubmitted})
	if err != nil {
		t.Fatalf("status filter List: %v", err)
	}
	if len(filtered.Records) != 1 {
		t.Fatalf("status filter len = %d, want 1", len(filtered.Records))
	}
	// 候选人筛选
	filtered, err = NewList(superCtx("admin-1", "管理员"), svcCtx, nil).List(&offertypes.ListRequest{
		PageRequest: offertypes.PageRequest{Current: 1, Size: 10}, CandidateUuid: candUuid})
	if err != nil {
		t.Fatalf("candidate filter List: %v", err)
	}
	if len(filtered.Records) != 1 || filtered.Records[0].Uuid != billUuid {
		t.Fatalf("candidate filter wrong: len=%d", len(filtered.Records))
	}
	// 候选人筛选不匹配
	filtered, err = NewList(superCtx("admin-1", "管理员"), svcCtx, nil).List(&offertypes.ListRequest{
		PageRequest: offertypes.PageRequest{Current: 1, Size: 10}, CandidateUuid: "no-such-candidate"})
	if err != nil {
		t.Fatalf("candidate miss List: %v", err)
	}
	if len(filtered.Records) != 0 {
		t.Fatalf("candidate miss len = %d, want 0", len(filtered.Records))
	}
}
