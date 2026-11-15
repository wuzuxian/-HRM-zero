package overtimebill

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/leavebalance"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	lbtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimebill"
)

// ---- rule 判定: 需审批 vs 免审批 + 调休入账（issues: 规则判定/折算入账流水 logic 测试） ----

const testYear = 2026

// TestSubmitRequiresApprovalThenConfirmAccrues 全流程（需审批）:
// 工作日规则 需审批 ×1.5、月上限不限 → 提交后 审批中 → 核定 2h/调休 → 审批通过 →
// 余额 granted += 3h 折算 (2×1.5=3h=0.375 天→round1 0.4 天), accrue 流水 billUuid 关联.
func TestSubmitRequiresApprovalThenConfirmAccrues(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	seedEmployee(t, svcCtx, ctx)
	seedOvertimeRule(t, svcCtx, ctx, "工作日需审批", true, true, true, 1.5, 2, 3, 0)

	billUuid := addOvertime(t, svcCtx, ctx, "2026-03-02", "1",
		"2026-03-02 18:00:00", "2026-03-02 20:00:00", nil)

	st := submitOvertime(t, env, svcCtx, billUuid)
	if st != StatusApproving {
		t.Fatalf("status after submit = %q, want approving (需审批)", st)
	}

	// 审批前核定校验: 核定 > 申请 → 10603
	if err := confirmAndApprove(t, env, svcCtx, billUuid, 3, CompTypeLeave); errCode(err) != 10603 {
		t.Fatalf("confirm > apply err code = %v, want 10603", errCode(err))
	}
	// 非审批中不可核定: 单据仍在 审批中，正常核定 2h
	if err := confirmAndApprove(t, env, svcCtx, billUuid, 2, CompTypeLeave); err != nil {
		t.Fatalf("confirm 2h + approve: %v", err)
	}

	// 单据状态 审批通过, 折算 accrue_hours = 2×1.5 = 3
	resp := listOvertimes(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	for _, rec := range resp.Records {
		if rec.Uuid == billUuid {
			if rec.Status != StatusApproved {
				t.Fatalf("status = %q, want approved", rec.Status)
			}
			if rec.AccrueHours != 3 {
				t.Fatalf("accrueHours = %v, want 3 (2h×1.5)", rec.AccrueHours)
			}
			if rec.ConfirmHours != 2 || rec.CompType != CompTypeLeave {
				t.Fatalf("confirm snapshot wrong: %+v", rec)
			}
			if rec.RuleSource != "global" || rec.CompMultiplier != 1.5 {
				t.Fatalf("rule snapshot wrong: source=%q mult=%v", rec.RuleSource, rec.CompMultiplier)
			}
		}
	}

	// 余额表 granted += 3h/8 天 = 0.4（round1），accrue 流水
	balances := listBalances(t, svcCtx, ctx, lbListReqPtr(testYear))
	bal := findBalance(balances, testEmpUuid, leavebalance.CompLeaveCode)
	if bal == nil {
		t.Fatal("comp_leave balance row missing after approval")
	}
	if bal.Granted != 0.4 {
		t.Fatalf("comp_leave granted = %v, want 0.4 (3h/8=0.375→round1)", bal.Granted)
	}
	logs := listLogs(t, svcCtx, ctx, lbLogReqPtr(testYear))
	var accrueLogs []*lbLog
	for i := range logs.Records {
		if logs.Records[i].ChangeType == leavebalance.ChangeTypeAccrue {
			accrueLogs = append(accrueLogs, &logs.Records[i])
		}
	}
	if len(accrueLogs) != 1 {
		t.Fatalf("accrue log count = %d, want 1", len(accrueLogs))
	}
	if accrueLogs[0].BillUuid != billUuid || accrueLogs[0].Days != 0.4 {
		t.Fatalf("accrue log fields wrong: %+v", accrueLogs[0])
	}
}

// TestSubmitFreeApprovalAutoApproves 免审批口径: 休息日规则 免审批 ×2 →
// 提交即 底座自动通过（同一事务 submit+approve），不落核定，按申请时长×倍数折算入账.
func TestSubmitFreeApprovalAutoApproves(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	seedEmployee(t, svcCtx, ctx)
	seedOvertimeRule(t, svcCtx, ctx, "休息日免审批", true, false, true, 1.5, 2, 3, 0)

	billUuid := addOvertime(t, svcCtx, ctx, "2026-03-07", "2",
		"2026-03-07 09:00:00", "2026-03-07 13:00:00", nil) // 4h

	st := submitOvertime(t, env, svcCtx, billUuid)
	if st != StatusApproved {
		t.Fatalf("status after submit = %q, want approved (免审批自动通过)", st)
	}

	resp := listOvertimes(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	for _, rec := range resp.Records {
		if rec.Uuid == billUuid {
			if rec.ConfirmHours != 0 {
				t.Fatalf("免审批单不应落核定, confirmHours = %v", rec.ConfirmHours)
			}
			if rec.AccrueHours != 8 { // 4h × 2
				t.Fatalf("accrueHours = %v, want 8 (4h×2)", rec.AccrueHours)
			}
		}
	}
	// 入账 8h = 1 天
	balances := listBalances(t, svcCtx, ctx, lbListReqPtr(testYear))
	bal := findBalance(balances, testEmpUuid, leavebalance.CompLeaveCode)
	if bal == nil || bal.Granted != 1 {
		t.Fatalf("comp_leave granted = %+v, want 1 (8h=1天)", bal)
	}
}

// TestMonthlyLimitGuardRollsBack 月上限守卫: 上限 6h，第一单折算 6h 通过，
// 第二单折算 3h 超限 → 10604 审批整体回滚（状态保持 审批中，无第二条 accrue 流水）.
func TestMonthlyLimitGuardRollsBack(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	seedEmployee(t, svcCtx, ctx)
	seedOvertimeRule(t, svcCtx, ctx, "月上限6h", true, true, true, 1, 1, 1, 6)

	// 第一单: 3/2 工作日 6h → 折算 6h = 0.75 天, 恰好贴上限
	bill1 := addOvertime(t, svcCtx, ctx, "2026-03-02", "1",
		"2026-03-02 18:00:00", "2026-03-03 00:00:00", nil)
	if st := submitOvertime(t, env, svcCtx, bill1); st != StatusApproving {
		t.Fatalf("bill1 submit status = %q", st)
	}
	if err := confirmAndApprove(t, env, svcCtx, bill1, 6, CompTypeLeave); err != nil {
		t.Fatalf("bill1 approve: %v", err)
	}

	// 第二单: 3/9 工作日 3h → 折算 3h, 6+3 > 6 → 10604
	bill2 := addOvertime(t, svcCtx, ctx, "2026-03-09", "1",
		"2026-03-09 18:00:00", "2026-03-09 21:00:00", nil)
	if st := submitOvertime(t, env, svcCtx, bill2); st != StatusApproving {
		t.Fatalf("bill2 submit status = %q", st)
	}
	if err := confirmAndApprove(t, env, svcCtx, bill2, 3, CompTypeLeave); errCode(err) != 10604 {
		t.Fatalf("bill2 approve err code = %v, want 10604", errCode(err))
	}

	// 状态保持 审批中（回滚），流水只有 1 条 accrue
	resp := listOvertimes(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	for _, rec := range resp.Records {
		if rec.Uuid == bill2 && rec.Status != StatusApproving {
			t.Fatalf("bill2 status after failed approve = %q, want approving (rolled back)", rec.Status)
		}
	}
	logs := listLogs(t, svcCtx, ctx, lbLogReqPtr(testYear))
	count := 0
	for i := range logs.Records {
		if logs.Records[i].ChangeType == leavebalance.ChangeTypeAccrue {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("accrue log count = %d, want 1 (failed approval must not log)", count)
	}
}

// TestOtherCompTypeSkipsAccrual 补偿方式=其他: 审批通过只落 accrue_hours 记录，不入余额.
func TestOtherCompTypeSkipsAccrual(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	seedEmployee(t, svcCtx, ctx)
	seedOvertimeRule(t, svcCtx, ctx, "其他补偿", true, true, true, 1.5, 2, 3, 0)

	billUuid := addOvertime(t, svcCtx, ctx, "2026-03-02", "1",
		"2026-03-02 18:00:00", "2026-03-02 20:00:00", nil)
	if st := submitOvertime(t, env, svcCtx, billUuid); st != StatusApproving {
		t.Fatalf("submit status = %q", st)
	}
	if err := confirmAndApprove(t, env, svcCtx, billUuid, 2, CompTypeOther); err != nil {
		t.Fatalf("approve: %v", err)
	}

	resp := listOvertimes(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	for _, rec := range resp.Records {
		if rec.Uuid == billUuid && rec.AccrueHours != 3 {
			t.Fatalf("accrueHours = %v, want 3 (折算落定但未入账)", rec.AccrueHours)
		}
	}
	balances := listBalances(t, svcCtx, ctx, lbListReqPtr(testYear))
	if bal := findBalance(balances, testEmpUuid, leavebalance.CompLeaveCode); bal != nil {
		t.Fatalf("comp_leave balance should not exist for comp_type=other, got %+v", bal)
	}
}

// TestApproveIdempotentSettled 审批幂等: settled=1 后重复 ApproveBill 不重复入账
// （底座状态机从 审批通过 不允许再 approve，直接验证守卫）.
func TestApproveIdempotentSettled(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	seedEmployee(t, svcCtx, ctx)
	seedOvertimeRule(t, svcCtx, ctx, "幂等", true, true, true, 1.5, 2, 3, 0)

	billUuid := addOvertime(t, svcCtx, ctx, "2026-03-02", "1",
		"2026-03-02 18:00:00", "2026-03-02 20:00:00", nil)
	_ = submitOvertime(t, env, svcCtx, billUuid)
	if err := confirmAndApprove(t, env, svcCtx, billUuid, 2, CompTypeLeave); err != nil {
		t.Fatalf("approve: %v", err)
	}
	// 重复审批 → 状态机拒绝（审批通过 是终态）
	err := bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), env.SqlConn, BillTypeCode,
		billUuid, bill.Operator{UserUuid: "op-1", Name: "管理员"}, "again")
	if err == nil {
		t.Fatal("second approve should fail (approved is terminal)")
	}
	// 余额仍只有一次入账
	balances := listBalances(t, svcCtx, ctx, lbListReqPtr(testYear))
	bal := findBalance(balances, testEmpUuid, leavebalance.CompLeaveCode)
	if bal == nil || bal.Granted != 0.4 {
		t.Fatalf("granted = %+v, want 0.4 (single accrual)", bal)
	}
}

// TestLifecycleGuardsAndFilters 生命周期守卫 + mine/筛选:
// edit/delete 仅 可编辑状态; confirm 仅 审批中; 非管理员默认只看自己的单据;
// 筛选 单据编号/单据状态/当天类型.
func TestLifecycleGuardsAndFilters(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")
	seedEmployee(t, svcCtx, ctx)
	seedOvertimeRule(t, svcCtx, ctx, "守卫规则", true, true, true, 1.5, 2, 3, 0)

	billUuid := addOvertime(t, svcCtx, ctx, "2026-03-02", "1",
		"2026-03-02 18:00:00", "2026-03-02 20:00:00", nil)

	// 未提交可编辑/删除; 提交后（审批中）不可
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: billUuid, EmployeeNo: "00000001", EmployeeName: "张三",
		OvertimeDate: "2026-03-02", DayType: "1",
		StartTime: "2026-03-02 18:00:00", EndTime: "2026-03-02 21:00:00",
	}); err != nil {
		t.Fatalf("edit unsubmitted: %v", err)
	}
	_ = submitOvertime(t, env, svcCtx, billUuid)
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: billUuid, EmployeeNo: "00000001", EmployeeName: "张三",
		OvertimeDate: "2026-03-02", DayType: "1",
		StartTime: "2026-03-02 18:00:00", EndTime: "2026-03-02 21:00:00",
	}); errCode(err) != 10602 {
		t.Fatalf("edit approving err code = %v, want 10602", errCode(err))
	}
	// 核定 0 小时 → 10601
	if _, err := NewConfirm(context.Background(), svcCtx, nil).Confirm(&types.ConfirmRequest{
		Uuid: billUuid, ConfirmHours: 0, CompType: CompTypeLeave,
	}); errCode(err) != 10601 {
		t.Fatalf("confirm 0 hours err code = %v, want 10601", errCode(err))
	}
	if err := confirmAndApprove(t, env, svcCtx, billUuid, 2, CompTypeLeave); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); errCode(err) != 10602 {
		t.Fatalf("delete approved err code = %v, want 10602", errCode(err))
	}

	// 第二个普通用户看不到 user-1 的单据（mine-by-default）
	other := testCtxFor("user-2", "李四HR")
	mine := listOvertimes(t, svcCtx, other, &types.ListRequest{})
	if len(mine.Records) != 0 {
		t.Fatalf("other user sees %d bills, want 0 (mine-by-default)", len(mine.Records))
	}
	admin := listOvertimes(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	if len(admin.Records) != 1 {
		t.Fatalf("admin sees %d bills, want 1", len(admin.Records))
	}

	// 筛选: 编号/状态/当天类型
	first := admin.Records[0]
	byNo := listOvertimes(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{BillNo: first.BillNo[:10]})
	if len(byNo.Records) != 1 {
		t.Fatalf("filter by billNo got %d, want 1", len(byNo.Records))
	}
	byStatus := listOvertimes(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{Status: StatusApproved})
	if len(byStatus.Records) != 1 {
		t.Fatalf("filter by status got %d, want 1", len(byStatus.Records))
	}
	byDay := listOvertimes(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{DayType: "2"})
	if len(byDay.Records) != 0 {
		t.Fatalf("filter by dayType=2 got %d, want 0", len(byDay.Records))
	}

}

// ---- leavebalance types shims (avoid importing lbtypes twice in struct literals) ----

func lbListReq(year int64) lbtypes.ListRequest {
	return lbtypes.ListRequest{Year: year}
}

func lbLogReq(year int64) lbtypes.LogListRequest {
	return lbtypes.LogListRequest{Year: year, EmployeeUuid: testEmpUuid}
}

type lbLog = lbtypes.HrmLeaveBalanceLog

func lbListReqPtr(year int64) *lbtypes.ListRequest {
	req := lbListReq(year)
	return &req
}

func lbLogReqPtr(year int64) *lbtypes.LogListRequest {
	req := lbLogReq(year)
	return &req
}
