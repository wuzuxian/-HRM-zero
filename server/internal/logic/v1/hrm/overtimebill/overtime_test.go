package overtimebill

import (
	"context"
	"errors"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	emptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/leavebalance"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/overtimerule"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	lbtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimebill"
	ortypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimerule"
	"github.com/jzero-io/jzero/core/status"
)

// Seams under test (issue #23 acceptance criteria), all at logic seams against
// the real migrations (svctest):
//  1. CalcOvertimeHours / ConvertAccrueHours / RoundHoursToUnit pure functions:
//     申请时长自动计算（最小计算单位向下取整）+ 倍数折算 (issues: 规则折算纯函数单测).
//  2. Add: 单据编号 HR217-, 未提交, 申请时长落库.
//  3. SubmitOvertimeRule 判定: 命中规则 需审批 → 审批中; 免审批 → 自动通过
//     (issues: 规则判定 需审批 vs 免审批 logic 测试).
//  4. 核定: confirmHours ≤ applyHours 校验（>→10573）; 审批通过事务内按 CompMultiplier
//     折算调休入账（余额表 granted+ + accrue 流水 with multiplier）(issues: 调休入账流水 logic 测试).
//  5. 月上限守卫: 超月上限 → 10574 审批整体回滚（单据状态保持 审批中，无流水）.
//  6. 幂等: settled=1 后重复 ApproveBill 不重复入账.
//  7. 生命周期守卫: edit/delete 仅可编辑状态; confirm 仅 审批中.
//  8. 「我的单据」默认过滤 + 筛选（编号/状态/当天类型）.

// ---- shared helpers (leave/leavebalance 测试同款最小 svcCtx + JWT ctx) ----

func newSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		ServiceContext: &coresvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
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

// superCtx returns a context carrying the 超级管理员 role (sees all bills).
func superCtx(uuid, username string) context.Context {
	return testCtxFor(uuid, username, "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
}

func errCode(err error) int {
	if err == nil {
		return 0
	}
	if sc := status.FromError(err); sc != nil {
		return int(sc.Code())
	}
	return -1
}

// ---- pure function expectations (issues: 规则折算纯函数单测) ----

func TestCalcOvertimeHoursExpectations(t *testing.T) {
	cases := []struct {
		name       string
		start, end string
		minUnit    int64
		want       float64
	}{
		{"2h05m unit 30min → 2h", "2026-03-02 18:00:00", "2026-03-02 20:05:00", 30, 2},
		{"2h29m unit 60min → 2h", "2026-03-02 18:00:00", "2026-03-02 20:29:00", 60, 2},
		{"45m unit 30min → 0.5h (1 个最小单位)", "2026-03-02 18:00:00", "2026-03-02 18:45:00", 30, 0.5},
		{"29m unit 30min → 0 (不足最小单位不计)", "2026-03-02 18:00:00", "2026-03-02 18:29:00", 30, 0},
		{"3h unit 0 → 默认30min → 3h", "2026-03-02 18:00:00", "2026-03-02 21:00:00", 0, 3},
	}
	for _, c := range cases {
		got, err := CalcOvertimeHours(c.start, c.end, c.minUnit)
		if err != nil {
			t.Fatalf("CalcOvertimeHours(%s): %v", c.name, err)
		}
		if got != c.want {
			t.Fatalf("CalcOvertimeHours(%s) = %v, want %v", c.name, got, c.want)
		}
	}
	if _, err := CalcOvertimeHours("2026-03-02 18:00:00", "2026-03-02 18:00:00", 30); !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("equal times err = %v, want ErrInvalidTimeRange", err)
	}
}

func TestConvertAccrueHoursExpectations(t *testing.T) {
	cases := []struct {
		base, multiplier, want float64
	}{
		{2, 1.5, 3},      // 工作日 1.5 倍
		{2, 2, 4},        // 休息日 2 倍
		{2, 3, 6},        // 法定节假日 3 倍
		{1.5, 2, 3},      // 0.75 单位 × 2
		{0, 3, 0},        // 无基准不计
		{2, 0, 2},        // 倍数缺失按 1
		{0.75, 1.5, 1.1}, // 1.125 → 保留一位小数 1.1
	}
	for _, c := range cases {
		if got := ConvertAccrueHours(c.base, c.multiplier); got != c.want {
			t.Fatalf("ConvertAccrueHours(%v,%v) = %v, want %v", c.base, c.multiplier, got, c.want)
		}
	}
}

func TestRoundHoursToUnitExpectations(t *testing.T) {
	cases := []struct {
		hours   float64
		minUnit int64
		want    float64
	}{
		{1.9, 30, 1.5},
		{2.0, 60, 2},
		{0.4, 30, 0},
		{1.0, 0, 1},
	}
	for _, c := range cases {
		if got := RoundHoursToUnit(c.hours, c.minUnit); got != c.want {
			t.Fatalf("RoundHoursToUnit(%v,%v) = %v, want %v", c.hours, c.minUnit, got, c.want)
		}
	}
}

// ---- shared fixtures ----

var testEmpUuid string

func seedEmployee(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context) {
	t.Helper()
	req := &emptypes.AddRequest{
		EmployeeNo:   "00000001",
		Name:         "张三",
		Gender:       "1",
		PersonStatus: "1",
		HireDate:     "2024-03-01",
	}
	if _, err := employee.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("seed employee: %v", err)
	}
	list, err := employee.NewList(ctx, svcCtx, nil).List(&emptypes.ListRequest{PageRequest: emptypes.PageRequest{Current: 1, Size: 50}})
	if err != nil {
		t.Fatalf("list employees: %v", err)
	}
	for _, rec := range list.Records {
		if rec.EmployeeNo == "00000001" {
			testEmpUuid = rec.Uuid
			return
		}
	}
	panic("seeded employee not found")
}

// seedOvertimeRule 建一条全局兜底加班规则（不绑考勤组/组织 → 全局兜底），
// policy 按 dayType 前缀区分: work*  / rest* / holiday*.
func seedOvertimeRule(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, name string,
	workApproval, restApproval, holidayApproval bool,
	workMul, restMul, holidayMul float64, monthlyLimit float64) string {
	t.Helper()
	req := &ortypes.AddRequest{
		RuleName:          name,
		Workday:           ortypes.DayOvertimePolicy{ApprovalRequired: b2i(workApproval), CompMultiplier: workMul},
		Restday:           ortypes.DayOvertimePolicy{ApprovalRequired: b2i(restApproval), CompMultiplier: restMul},
		Holiday:           ortypes.DayOvertimePolicy{ApprovalRequired: b2i(holidayApproval), CompMultiplier: holidayMul},
		MinUnitMinutes:    30,
		MonthlyLimitHours: monthlyLimit,
		Status:            "1",
	}
	if _, err := overtimerule.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("seed overtime rule %s: %v", name, err)
	}
	list, err := overtimerule.NewList(ctx, svcCtx, nil).List(&ortypes.ListRequest{PageRequest: ortypes.PageRequest{Current: 1, Size: 50}})
	if err != nil {
		t.Fatalf("list overtime rules: %v", err)
	}
	for _, rec := range list.Records {
		if rec.RuleName == name {
			return rec.Uuid
		}
	}
	panic("seeded rule not found")
}

func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// listOvertimes calls the public List seam with default pagination filled in.
func listOvertimes(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, req *types.ListRequest) *types.ListResponse {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 100
	}
	resp, err := NewList(ctx, svcCtx, nil).List(req)
	if err != nil {
		t.Fatalf("List overtimes: %v", err)
	}
	return resp
}

// addOvertime creates one bill through the public Add seam and returns its uuid.
func addOvertime(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context,
	overtimeDate, dayType, start, end string, mutate func(*types.AddRequest)) string {
	t.Helper()
	req := &types.AddRequest{
		EmployeeUuid: testEmpUuid,
		EmployeeNo:   "00000001",
		EmployeeName: "张三",
		OvertimeDate: overtimeDate,
		DayType:      dayType,
		StartTime:    start,
		EndTime:      end,
		Reason:       "项目上线",
	}
	if mutate != nil {
		mutate(req)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("Add overtime: %v", err)
	}
	resp := listOvertimes(t, svcCtx, ctx, &types.ListRequest{Mine: 1})
	if len(resp.Records) == 0 {
		t.Fatal("added overtime bill not visible in 我的单据 list")
	}
	rec := resp.Records[0]
	if len(rec.BillNo) < 6 || rec.BillNo[:6] != "HR217-" {
		t.Fatalf("billNo %q does not follow HR217- format", rec.BillNo)
	}
	if rec.Status != StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	if rec.Source != SourceEmployee {
		t.Fatalf("source = %q, want employee", rec.Source)
	}
	return rec.Uuid
}

// submitOvertime drives SubmitOvertimeBill (rule 判定 + 底座提交[+免审批自动通过]).
func submitOvertime(t *testing.T, env *svctest.Env, svcCtx *svc.ServiceContext, billUuid string) string {
	t.Helper()
	ctx := WithSvcCtx(context.Background(), svcCtx)
	st, err := SubmitOvertimeBill(ctx, env.SqlConn, svcCtx, billUuid,
		bill.Operator{UserUuid: "op-1", Name: "管理员"}, "")
	if err != nil {
		t.Fatalf("SubmitOvertimeBill: %v", err)
	}
	return st
}

// confirmAndApprove 核定 + 底座审批通过.
func confirmAndApprove(t *testing.T, env *svctest.Env, svcCtx *svc.ServiceContext,
	billUuid string, confirmHours float64, compType string) error {
	t.Helper()
	if confirmHours > 0 {
		if _, err := NewConfirm(context.Background(), svcCtx, nil).Confirm(&types.ConfirmRequest{
			Uuid:         billUuid,
			ConfirmHours: confirmHours,
			CompType:     compType,
		}); err != nil {
			return err
		}
	}
	return bill.ApproveBill(WithSvcCtx(context.Background(), svcCtx), env.SqlConn, BillTypeCode,
		billUuid, bill.Operator{UserUuid: "op-1", Name: "管理员"}, "同意")
}

func listBalances(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, req *lbtypes.ListRequest) *lbtypes.ListResponse {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 100
	}
	resp, err := leavebalance.NewList(ctx, svcCtx, nil).List(req)
	if err != nil {
		t.Fatalf("leavebalance List: %v", err)
	}
	return resp
}

func findBalance(resp *lbtypes.ListResponse, employeeUuid, typeCode string) *lbtypes.HrmLeaveBalance {
	for i := range resp.Records {
		rec := &resp.Records[i]
		if rec.EmployeeUuid == employeeUuid && rec.LeaveTypeCode == typeCode {
			return rec
		}
	}
	return nil
}

func listLogs(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, req *lbtypes.LogListRequest) *lbtypes.LogListResponse {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 100
	}
	resp, err := leavebalance.NewLogList(ctx, svcCtx, nil).LogList(req)
	if err != nil {
		t.Fatalf("leavebalance LogList: %v", err)
	}
	return resp
}
