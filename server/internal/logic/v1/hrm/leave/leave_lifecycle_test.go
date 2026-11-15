package leave

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/leavebalance"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leave"
	lbtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
)

// TestAddGeneratesBillNoAndPlanDays 新增落库：HR205- 编号、未提交、预计天数自动计算快照.
func TestAddGeneratesBillNoAndPlanDays(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	billUuid, req := addLeave(t, svcCtx, ctx, "annual_leave", nil)

	resp := listLeaves(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	var rec *types.HrmLeave
	for i := range resp.Records {
		if resp.Records[i].Uuid == billUuid {
			rec = &resp.Records[i]
		}
	}
	if rec == nil {
		t.Fatal("bill not found in list")
	}
	if rec.PlanDays != 3.5 { // 28h/8 = 3.5
		t.Fatalf("planDays = %v, want 3.5", rec.PlanDays)
	}
	if rec.LeaveTypeName != "年假" {
		t.Fatalf("leaveTypeName snapshot = %q, want 年假", rec.LeaveTypeName)
	}
	if rec.CreatorName != "张三HR" || rec.CreatorUuid == "" {
		t.Fatalf("creator snapshot wrong: uuid=%q name=%q", rec.CreatorUuid, rec.CreatorName)
	}
	_ = req
}

// TestApproveConsumesBalanceAndLogs 审批通过事务内扣减余额（余额表 used + consume 流水）——
// issues 验收: 审批通过扣减余额（流水）logic 测试.
func TestApproveConsumesBalanceAndLogs(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	initBalanceYear(t, svcCtx, ctx)
	billUuid, _ := addLeave(t, svcCtx, ctx, "annual_leave", nil)

	// 审批前余额未动
	balances := listBalances(t, svcCtx, ctx, &lbtypes.ListRequest{Year: testYear})
	bal := findBalance(balances, testEmpUuid, "annual_leave")
	if bal == nil || bal.Used != 0 {
		t.Fatalf("before approval used = %+v, want 0", bal)
	}

	submitAndApprove(t, env, BillTypeCode, billUuid)

	// 审批后: used = 3.5（预计天数），available 相应减少
	balances = listBalances(t, svcCtx, ctx, &lbtypes.ListRequest{Year: testYear})
	bal = findBalance(balances, testEmpUuid, "annual_leave")
	if bal == nil {
		t.Fatal("annual_leave balance row missing after approval")
	}
	if bal.Used != 3.5 {
		t.Fatalf("after approval used = %v, want 3.5", bal.Used)
	}
	if bal.Available != bal.Granted-3.5 {
		t.Fatalf("after approval available = %v, want granted-3.5 = %v", bal.Available, bal.Granted-3.5)
	}

	// consume 流水: days=-3.5, billUuid 关联请假单
	logs := listLogs(t, svcCtx, ctx, &lbtypes.LogListRequest{Year: testYear, EmployeeUuid: testEmpUuid})
	var consumeLogs []*lbtypes.HrmLeaveBalanceLog
	for i := range logs.Records {
		if logs.Records[i].ChangeType == leavebalance.ChangeTypeConsume {
			consumeLogs = append(consumeLogs, &logs.Records[i])
		}
	}
	if len(consumeLogs) != 1 {
		t.Fatalf("consume log count = %d, want 1", len(consumeLogs))
	}
	if consumeLogs[0].Days != -3.5 || consumeLogs[0].BillUuid != billUuid || consumeLogs[0].UsedAfter != 3.5 {
		t.Fatalf("consume log fields wrong: %+v", consumeLogs[0])
	}
}

// TestApproveInsufficientBalanceRollsBack 余额不足 → 审批整体失败（单据状态不落 审批通过）.
func TestApproveInsufficientBalanceRollsBack(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	// 不初始化余额 → 自动建空行，可用 0 < 3.5 → 10513
	billUuid, _ := addLeave(t, svcCtx, ctx, "annual_leave", nil)

	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid,
		bill.Operator{UserUuid: "op-1", Name: "管理员"}, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid,
		bill.Operator{UserUuid: "op-1", Name: "管理员"}, "同意")
	if leaveErrCode(err) != 10513 {
		t.Fatalf("approve with insufficient balance err code = %v, want 10513", leaveErrCode(err))
	}

	// 单据状态保持 审批中（审批事务回滚）
	resp := listLeaves(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	for _, rec := range resp.Records {
		if rec.Uuid == billUuid && rec.Status != bill.StatusApproving {
			t.Fatalf("status after failed approval = %q, want approving (rolled back)", rec.Status)
		}
	}

	// 失败扣减不落流水
	logs := listLogs(t, svcCtx, ctx, &lbtypes.LogListRequest{Year: testYear, EmployeeUuid: testEmpUuid})
	if len(logs.Records) != 0 {
		t.Fatalf("failed approval must not write logs, got %d", len(logs.Records))
	}
}

// TestApproveNonBalanceTypeSkipsDeduction 非余额管理类型（事假 balance_enabled=0）审批通过不扣减.
func TestApproveNonBalanceTypeSkipsDeduction(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	initBalanceYear(t, svcCtx, ctx)
	billUuid, _ := addLeave(t, svcCtx, ctx, "personal_leave", nil) // 事假: balance_enabled=0

	submitAndApprove(t, env, BillTypeCode, billUuid) // 不应因余额缺失失败

	balances := listBalances(t, svcCtx, ctx, &lbtypes.ListRequest{Year: testYear})
	if bal := findBalance(balances, testEmpUuid, "personal_leave"); bal != nil && bal.Used != 0 {
		t.Fatalf("personal_leave used = %v, want 0 (non-balance type must not consume)", bal.Used)
	}
	logs := listLogs(t, svcCtx, ctx, &lbtypes.LogListRequest{Year: testYear, EmployeeUuid: testEmpUuid})
	for _, lg := range logs.Records {
		if lg.LeaveTypeCode == "personal_leave" {
			t.Fatalf("non-balance type must not write balance logs, got %+v", lg)
		}
	}
}

// TestCancelRefundsDifference 实际 < 预计 → 销假退回差额（refund 流水, used 回落）——
// issues 验收: 销假回填与冲回 logic 测试.
func TestCancelRefundsDifference(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	initBalanceYear(t, svcCtx, ctx)
	billUuid, _ := addLeave(t, svcCtx, ctx, "annual_leave", nil) // 预计 3.5 天
	submitAndApprove(t, env, BillTypeCode, billUuid)

	// 销假: 实际只休了 2 天（16h）
	if _, err := NewCancel(ctx, svcCtx, nil).Cancel(&types.CancelRequest{
		Uuid:            billUuid,
		ActualStartTime: "2026-03-02 09:00:00",
		ActualEndTime:   "2026-03-03 09:00:00", // 24h = 3 天 → 仍多休? no: 3 < 3.5 → 退 0.5
	}); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	resp := listLeaves(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	for _, rec := range resp.Records {
		if rec.Uuid == billUuid {
			if rec.ActualDays != 3 {
				t.Fatalf("actualDays = %v, want 3", rec.ActualDays)
			}
			if rec.Settled != 1 {
				t.Fatalf("settled = %v, want 1", rec.Settled)
			}
		}
	}

	balances := listBalances(t, svcCtx, ctx, &lbtypes.ListRequest{Year: testYear})
	bal := findBalance(balances, testEmpUuid, "annual_leave")
	if bal.Used != 3 { // 3.5 扣减 − 0.5 退回
		t.Fatalf("after cancel used = %v, want 3", bal.Used)
	}

	logs := listLogs(t, svcCtx, ctx, &lbtypes.LogListRequest{Year: testYear, EmployeeUuid: testEmpUuid})
	var refundLogs []*lbtypes.HrmLeaveBalanceLog
	for i := range logs.Records {
		if logs.Records[i].ChangeType == leavebalance.ChangeTypeRefund {
			refundLogs = append(refundLogs, &logs.Records[i])
		}
	}
	if len(refundLogs) != 1 {
		t.Fatalf("refund log count = %d, want 1", len(refundLogs))
	}
	if refundLogs[0].Days != 0.5 || refundLogs[0].BillUuid != billUuid || refundLogs[0].UsedAfter != 3 {
		t.Fatalf("refund log fields wrong: %+v", refundLogs[0])
	}
}

// TestCancelConsumesDifference 实际 > 预计 → 销假补扣差额（consume 流水, used 上升）.
func TestCancelConsumesDifference(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	initBalanceYear(t, svcCtx, ctx)
	// 预计 2 天 (16h)，实际休 3 天 (24h) → 补扣 1 天
	billUuid, _ := addLeave(t, svcCtx, ctx, "annual_leave", func(req *types.AddRequest) {
		req.PlanStartTime = "2026-03-02 09:00:00"
		req.PlanEndTime = "2026-03-03 09:00:00" // 24h = 3 天…改小: 用 09:00→17:00 8h=1 天
		req.PlanStartTime = "2026-03-02 09:00:00"
		req.PlanEndTime = "2026-03-02 17:00:00" // 8h = 1 天
	})
	submitAndApprove(t, env, BillTypeCode, billUuid)

	if _, err := NewCancel(ctx, svcCtx, nil).Cancel(&types.CancelRequest{
		Uuid:            billUuid,
		ActualStartTime: "2026-03-02 09:00:00",
		ActualEndTime:   "2026-03-03 13:00:00", // 28h = 3.5 天 → 补扣 2.5 天
	}); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	balances := listBalances(t, svcCtx, ctx, &lbtypes.ListRequest{Year: testYear})
	bal := findBalance(balances, testEmpUuid, "annual_leave")
	if bal.Used != 3.5 { // 1 扣减 + 2.5 补扣
		t.Fatalf("after cancel used = %v, want 3.5", bal.Used)
	}
}

// TestCancelConsumesDifferenceRejectsOnInsufficient 补扣差额超可用 → 销假整体回滚（不回填）.
func TestCancelConsumesDifferenceRejectsOnInsufficient(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	initBalanceYear(t, svcCtx, ctx)
	// 预计 1 天；把余额可用调到只剩 0.5：init 5 天(annual default) − 已扣 1 = 4 可用，
	// 先调低应发到 1.5 使补扣 2.5 天时可用 0.5 < 2.5 → 10513
	billUuid, _ := addLeave(t, svcCtx, ctx, "annual_leave", func(req *types.AddRequest) {
		req.PlanStartTime = "2026-03-02 09:00:00"
		req.PlanEndTime = "2026-03-02 17:00:00"
	})
	submitAndApprove(t, env, BillTypeCode, billUuid)

	// 把余额应发压到 1（used 1 → 可用 0）→ 补扣任何差额都失败
	balances := listBalances(t, svcCtx, ctx, &lbtypes.ListRequest{Year: testYear})
	bal := findBalance(balances, testEmpUuid, "annual_leave")
	if _, err := leavebalance.NewAdjust(ctx, svcCtx, nil).Adjust(&lbtypes.AdjustRequest{
		BalanceUuid: bal.Uuid,
		Days:        -(bal.Granted - 1),
		Reason:      "测试压低可用",
	}); err != nil {
		t.Fatalf("Adjust: %v", err)
	}

	err := cancelExpectError(t, svcCtx, ctx, billUuid, "2026-03-02 09:00:00", "2026-03-03 13:00:00")
	if leaveErrCode(err) != 10513 {
		t.Fatalf("cancel over-balance err code = %v, want 10513", leaveErrCode(err))
	}

	// 回滚断言: settled 仍 0, actual 未回填
	resp := listLeaves(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	for _, rec := range resp.Records {
		if rec.Uuid == billUuid {
			if rec.Settled != 0 || rec.ActualDays != 0 {
				t.Fatalf("rolled-back cancel persisted: settled=%d actualDays=%v", rec.Settled, rec.ActualDays)
			}
		}
	}
}

func cancelExpectError(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, uuid, start, end string) error {
	t.Helper()
	_, err := NewCancel(ctx, svcCtx, nil).Cancel(&types.CancelRequest{Uuid: uuid, ActualStartTime: start, ActualEndTime: end})
	return err
}

// TestCancelIdempotentAndGuards 重复销假拒绝 + 非 审批通过 状态不可销假.
func TestCancelIdempotentAndGuards(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	initBalanceYear(t, svcCtx, ctx)
	billUuid, _ := addLeave(t, svcCtx, ctx, "annual_leave", nil)
	submitAndApprove(t, env, BillTypeCode, billUuid)

	req := &types.CancelRequest{Uuid: billUuid, ActualStartTime: "2026-03-02 09:00:00", ActualEndTime: "2026-03-02 17:00:00"}
	if _, err := NewCancel(ctx, svcCtx, nil).Cancel(req); err != nil {
		t.Fatalf("first Cancel: %v", err)
	}
	if _, err := NewCancel(ctx, svcCtx, nil).Cancel(req); leaveErrCode(err) != 10542 {
		t.Fatalf("second Cancel err code = %v, want 10542 (idempotent guard)", leaveErrCode(err))
	}

	// 未审批的单据不可销假
	billUuid2, _ := addLeave(t, svcCtx, ctx, "annual_leave", func(r *types.AddRequest) {
		r.EmployeeNo = "EMP002"
		r.EmployeeName = "李四"
	})
	if _, err := NewCancel(ctx, svcCtx, nil).Cancel(&types.CancelRequest{
		Uuid: billUuid2, ActualStartTime: "2026-03-02 09:00:00", ActualEndTime: "2026-03-02 17:00:00",
	}); leaveErrCode(err) != 10542 {
		t.Fatalf("cancel unapproved err code = %v, want 10542", leaveErrCode(err))
	}
}

// TestLifecycleGuardsAndFilters 生命周期守卫（edit/delete 仅可编辑状态）+ 筛选 + 我的单据.
func TestLifecycleGuardsAndFilters(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	billUuid, _ := addLeave(t, svcCtx, ctx, "annual_leave", nil)

	// 审批中不可编辑/删除
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid,
		bill.Operator{UserUuid: "op-1", Name: "管理员"}, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: billUuid, EmployeeNo: "EMP001", EmployeeName: "张三", LeaveTypeCode: "annual_leave",
		PlanStartTime: "2026-03-02 09:00:00", PlanEndTime: "2026-03-02 17:00:00",
	}); leaveErrCode(err) != 10542 {
		t.Fatalf("edit while approving err code = %v, want 10542", leaveErrCode(err))
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{billUuid}}); leaveErrCode(err) != 10542 {
		t.Fatalf("delete while approving err code = %v, want 10542", leaveErrCode(err))
	}

	// 撤销 → 已撤销可编辑/删除
	if err := bill.RevokeBill(context.Background(), env.SqlConn, BillTypeCode, billUuid,
		bill.Operator{UserUuid: "op-1", Name: "管理员"}, ""); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: billUuid, EmployeeNo: "EMP001", EmployeeName: "张三", LeaveTypeCode: "annual_leave",
		PlanStartTime: "2026-03-02 09:00:00", PlanEndTime: "2026-03-02 17:00:00",
	}); err != nil {
		t.Fatalf("edit after revoke: %v", err)
	}

	// 筛选: leaveTypeCode
	filtered := listLeaves(t, svcCtx, ctx, &types.ListRequest{LeaveTypeCode: "annual_leave", Mine: 1})
	if len(filtered.Records) != 1 {
		t.Fatalf("filtered records = %d, want 1", len(filtered.Records))
	}
	other := listLeaves(t, svcCtx, ctx, &types.ListRequest{LeaveTypeCode: "sick_leave", Mine: 1})
	if len(other.Records) != 0 {
		t.Fatalf("sick filter records = %d, want 0", len(other.Records))
	}

	// 我的单据: 其他用户看不到 user-1 的单据
	mine2 := listLeaves(t, svcCtx, testCtxFor("user-2", "李四"), &types.ListRequest{})
	if len(mine2.Records) != 0 {
		t.Fatalf("other user sees %d bills, want 0 (mine-by-default)", len(mine2.Records))
	}
	// 管理员可见
	admin := listLeaves(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	if len(admin.Records) != 1 {
		t.Fatalf("admin sees %d bills, want 1", len(admin.Records))
	}
}

// TestBillTypeRegistered205 共享审批底座可驱动类型码 205（Submit 状态迁移）.
func TestBillTypeRegistered205(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR")

	billUuid, _ := addLeave(t, svcCtx, ctx, "annual_leave", nil)
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid,
		bill.Operator{UserUuid: "op-1", Name: "管理员"}, ""); err != nil {
		t.Fatalf("SubmitBill type 205: %v", err)
	}
	records, err := bill.Records(context.Background(), env.SqlConn, BillTypeCode, billUuid)
	if err != nil {
		t.Fatalf("Records: %v", err)
	}
	if len(records) != 1 || records[0].Action != bill.ActionSubmit {
		t.Fatalf("approval records = %+v, want 1 submit record", records)
	}
}
