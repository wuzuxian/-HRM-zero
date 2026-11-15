package reward

import (
	"context"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_reward_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/reward"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// Seams under test (per ticket acceptance criteria), all at the logic-method
// seam against the real migrations (svctest):
//  1. Add: 单据编号 HR215-... generated in-tx, 未提交 status, 申请人 snapshot,
//     salary_month defaults to the 发生日期 month.
//  2. 薪资数据源同步 (spec 用户故事 53): SubmitBill → ApproveBill flips
//     synced=1 inside the approval transaction; the bill is then addressable
//     by (employee_uuid, salary_month) for HRM-34/HRM-37 取数.
//  3. Lifecycle guards: edit/delete rejected outside editable statuses;
//     审批通过 single-step approval also flips synced.
//  4. 「我的单据」 default filter with admin bypass.
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

func addReward(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, mutate func(*types.AddRequest)) (string, *types.AddRequest) {
	t.Helper()
	req := &types.AddRequest{
		EmployeeNo:   "EMP001",
		EmployeeName: "张三",
		RewardType:   RewardTypeReward,
		RewardLevel:  RewardLevelCompany,
		DocNo:        "HR-2026-001",
		OccurDate:    "2026-11-01",
		Amount:       800,
		Reason:       "季度优秀员工",
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
		t.Fatal("added reward not visible in 我的单据 list")
	}
	rec := resp.Records[0]
	if !strings.HasPrefix(rec.BillNo, "HR215-") {
		t.Fatalf("billNo %q does not follow HR215- format", rec.BillNo)
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

	billUuid, req := addReward(t, svcCtx, ctx, nil)

	row := mustFindReward(t, svcCtx, billUuid)
	if row.SalaryMonth != "2026-11" {
		t.Fatalf("salary_month default = %q, want 2026-11 (发生日期所在月)", row.SalaryMonth)
	}
	if row.Synced != 0 {
		t.Fatalf("synced = %d on creation, want 0 (not yet a salary source)", row.Synced)
	}
	if row.Amount != req.Amount {
		t.Fatalf("amount = %v, want %v", row.Amount, req.Amount)
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
		{"missing employee", func(r *types.AddRequest) { r.EmployeeNo = ""; r.EmployeeName = "" }},
		{"bad reward type", func(r *types.AddRequest) { r.RewardType = "bonus" }},
		{"bad reward level", func(r *types.AddRequest) { r.RewardLevel = "national" }},
		{"negative amount", func(r *types.AddRequest) { r.Amount = -1 }},
		{"missing occur date", func(r *types.AddRequest) { r.OccurDate = "" }},
		{"bad occur date format", func(r *types.AddRequest) { r.OccurDate = "2026/11/01" }},
		{"bad salary month format", func(r *types.AddRequest) { r.SalaryMonth = "2026-1" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{
				EmployeeNo:   "EMP001",
				EmployeeName: "张三",
				RewardType:   RewardTypeReward,
				RewardLevel:  RewardLevelCompany,
				OccurDate:    "2026-11-01",
			}
			tc.mutat(req)
			if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestApproveSyncsSalarySource(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	// 处分 → 奖惩扣款 source, explicit 入账年月.
	punishUuid, _ := addReward(t, svcCtx, ctx, func(r *types.AddRequest) {
		r.RewardType = RewardTypePunishment
		r.RewardLevel = RewardLevelNotice
		r.Amount = 300
		r.OccurDate = "2026-11-05"
		r.SalaryMonth = "2026-12"
	})
	// 奖励 → 奖惩奖励 source, default 入账年月 (= 发生日期月份).
	rewardUuid, _ := addReward(t, svcCtx, ctx, nil)

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	for _, u := range []string{punishUuid, rewardUuid} {
		if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, u, op, ""); err != nil {
			t.Fatalf("SubmitBill %s: %v", u, err)
		}
		if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, u, op, "同意"); err != nil {
			t.Fatalf("ApproveBill %s: %v", u, err)
		}
	}

	punish := mustFindReward(t, svcCtx, punishUuid)
	if punish.Status != bill.StatusApproved {
		t.Fatalf("punishment status = %q, want approved", punish.Status)
	}
	if punish.Synced != 1 {
		t.Fatalf("punishment synced = %d, want 1 (奖惩扣款 data source ready)", punish.Synced)
	}
	if punish.SalaryMonth != "2026-12" {
		t.Fatalf("punishment salary_month = %q, want explicit 2026-12", punish.SalaryMonth)
	}

	reward := mustFindReward(t, svcCtx, rewardUuid)
	if reward.Synced != 1 {
		t.Fatalf("reward synced = %d, want 1 (奖惩奖励 data source ready)", reward.Synced)
	}
	if reward.SalaryMonth != "2026-11" {
		t.Fatalf("reward salary_month = %q, want 2026-11 (default from 发生日期)", reward.SalaryMonth)
	}

	// 薪资取数契约 (HRM-34/HRM-37): query the data source exactly as the
	// salary domain will — by employee + salary_month + synced.
	rewardSum, punishSum, err := salarySourceSums(context.Background(), env.SqlConn, "EMP001", "2026-11")
	if err != nil {
		t.Fatalf("salarySourceSums: %v", err)
	}
	if rewardSum != 800 || punishSum != 0 {
		t.Fatalf("2026-11 sums: reward=%v punishment=%v, want 800/0", rewardSum, punishSum)
	}
	rewardSum, punishSum, err = salarySourceSums(context.Background(), env.SqlConn, "EMP001", "2026-12")
	if err != nil {
		t.Fatalf("salarySourceSums: %v", err)
	}
	if rewardSum != 0 || punishSum != 300 {
		t.Fatalf("2026-12 sums: reward=%v punishment=%v, want 0/300", rewardSum, punishSum)
	}
}

func TestApproveIdempotentSync(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	billUuid, _ := addReward(t, svcCtx, ctx, nil)

	// Re-run the write-back: must be a no-op, never double-sync.
	row := mustFindReward(t, svcCtx, billUuid)
	if err := onApproved(context.Background(), env.SqlConn, billUuid); err != nil {
		t.Fatalf("onApproved (pre-approval replay): %v", err)
	}
	if err := markSynced(context.Background(), env.SqlConn, row); err != nil {
		t.Fatalf("markSynced replay: %v", err)
	}

	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	row = mustFindReward(t, svcCtx, billUuid)
	if row.Synced != 1 || row.Status != bill.StatusApproved {
		t.Fatalf("after approval synced=%d status=%q, want 1/approved", row.Synced, row.Status)
	}
	if row.SalaryMonth != "2026-11" {
		t.Fatalf("salary_month = %q, want 2026-11", row.SalaryMonth)
	}
}

func TestEditDeleteGuardedByStatus(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}

	billUuid, _ := addReward(t, svcCtx, ctx, nil)

	// 审批中 (approving) is locked: edit and delete must be rejected.
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid,
		EmployeeNo:   "EMP001",
		EmployeeName: "张三",
		RewardType:   RewardTypeReward,
		RewardLevel:  RewardLevelCompany,
		OccurDate:    "2026-11-01",
	}); err == nil {
		t.Fatal("edit during 审批中 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批中 should fail")
	}

	// 审批通过 (approved) is locked too — synced salary source.
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:         billUuid,
		EmployeeNo:   "EMP001",
		EmployeeName: "张三",
		RewardType:   RewardTypeReward,
		RewardLevel:  RewardLevelCompany,
		OccurDate:    "2026-11-01",
	}); err == nil {
		t.Fatal("edit during 审批通过 should fail")
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); err == nil {
		t.Fatal("delete during 审批通过 should fail")
	}

	// 已撤销 (revoked) is editable/deletable again.
	billUuid2, _ := addReward(t, svcCtx, ctx, nil)
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
		RewardType:   RewardTypePunishment,
		RewardLevel:  RewardLevelDept,
		OccurDate:    "2026-11-02",
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
	addReward(t, svcCtx, userCtx, nil)
	addReward(t, svcCtx, userCtx, func(r *types.AddRequest) {
		r.RewardType = RewardTypePunishment
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

	// 奖惩类型 filter (demo-inventory §7: 筛选单据编号/单据状态/奖惩类型).
	typeFilter, err := NewList(adminCtx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10},
		RewardType:  RewardTypePunishment,
	})
	if err != nil {
		t.Fatalf("List(rewardType filter): %v", err)
	}
	if typeFilter.Total != 1 || typeFilter.Records[0].RewardType != RewardTypePunishment {
		t.Fatalf("rewardType filter: total=%d, want 1 punishment", typeFilter.Total)
	}
}

func mustFindReward(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) *hrm_reward_bill.HrmRewardBill {
	t.Helper()
	row, err := svcCtx.Model.HrmRewardBill.FindOneByUuid(context.Background(), nil, billUuid)
	if err != nil {
		t.Fatalf("find reward %s: %v", billUuid, err)
	}
	return row
}

// salarySourceRow mirrors one synced 奖惩单 row for the 取数 contract check.
type salarySourceRow struct {
	Synced      int64   `db:"synced"`
	RewardType  string  `db:"reward_type"`
	Amount      float64 `db:"amount"`
	EmployeeNo  string  `db:"employee_no"`
	SalaryMonth string  `db:"salary_month"`
	Status      string  `db:"status"`
}

// salarySourceSums implements the HRM-34/HRM-37 取数 contract against the
// synced source rows: 奖惩奖励 (reward) and 奖惩扣款 (punishment) totals per
// employee + salary month. Reference query for the salary domain tickets.
func salarySourceSums(ctx context.Context, conn sqlx.SqlConn, employeeNo, salaryMonth string) (rewardSum, punishmentSum float64, err error) {
	var rows []salarySourceRow
	err = conn.QueryRowsCtx(ctx, &rows,
		"SELECT synced, reward_type, amount, employee_no, salary_month, status FROM hrm_reward_bill WHERE employee_no = ?",
		employeeNo)
	if err != nil {
		return 0, 0, err
	}
	for _, r := range rows {
		if r.Synced != 1 || r.SalaryMonth != salaryMonth {
			continue
		}
		switch r.RewardType {
		case RewardTypeReward:
			rewardSum += r.Amount
		case RewardTypePunishment:
			punishmentSum += r.Amount
		}
	}
	return rewardSum, punishmentSum, nil
}

func superAdminRoleUuid() string {
	// keep the seeded R_SUPER uuid local to tests; production logic uses hrmuser
	return "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
}
