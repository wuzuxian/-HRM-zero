package leave

import (
	"context"
	"errors"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	emptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/leavebalance"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	lbtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
	orgtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leave"
	"github.com/jzero-io/jzero/core/status"
)

// Seams under test (issue #21 acceptance criteria), all at logic seams against
// the real migrations (svctest):
//  1. CalcLeaveDays pure function: 天数自动计算 0.5 步进边界 (issues: 天数计算纯函数单测).
//  2. Add: 单据编号 HR205-, 未提交, 预计天数落库.
//  3. 审批通过事务内扣减余额（余额表 used + consume 流水）(issues: 审批通过扣减余额 logic 测试);
//     余额不足 → 审批整体回滚（10513）; 非余额管理类型不扣减.
//  4. 销假回填实际时间 + 差额多退少补（refund/consume 流水）(issues: 销假回填与冲回 logic 测试);
//     销假幂等（settled=1 后拒绝）.
//  5. 生命周期守卫: edit/delete 仅可编辑状态; cancel 仅 审批通过.
//  6. 「我的单据」默认过滤 + 筛选（编号/状态/请假类型）.

// ---- shared helpers (leavebalance_test.go 同款最小 svcCtx + JWT ctx) ----

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

// TestCalcLeaveDaysExpectations 锚定独立期望值（非重算代码）：
// 口径 = 时长小时 / 8，向上取整到 0.5 天步进（不足部分进位，最短 0.5 天）.
func TestCalcLeaveDaysExpectations(t *testing.T) {
	cases := []struct {
		start, end string
		want       float64
	}{
		{"2026-03-02 09:00:00", "2026-03-02 11:00:00", 0.5},  // 2h
		{"2026-03-02 09:00:00", "2026-03-02 13:00:00", 0.5},  // 4h
		{"2026-03-02 09:00:00", "2026-03-02 14:00:00", 1},    // 5h → 0.625 → 1
		{"2026-03-02 09:00:00", "2026-03-02 17:00:00", 1},    // 8h
		{"2026-03-02 09:00:00", "2026-03-02 21:00:00", 1.5},  // 12h
		{"2026-03-02 09:00:00", "2026-03-03 09:00:00", 3},    // 24h
		{"2026-03-02 09:00:00", "2026-03-04 09:00:00", 6},    // 48h
		{"2026-03-02 09:00:00", "2026-03-03 13:00:00", 3.5},  // 28h → 3.5
		{"2026-03-02 09:00:00", "2026-03-03 15:00:00", 4},    // 30h → 3.75 → 4
	}
	for _, c := range cases {
		got, err := CalcLeaveDays(c.start, c.end)
		if err != nil {
			t.Fatalf("CalcLeaveDays(%s,%s): %v", c.start, c.end, err)
		}
		if got != c.want {
			t.Fatalf("CalcLeaveDays(%s,%s) = %v, want %v", c.start, c.end, got, c.want)
		}
	}
}

func TestCalcLeaveDaysRejectsBadRanges(t *testing.T) {
	if _, err := CalcLeaveDays("2026-03-02 09:00:00", "2026-03-02 09:00:00"); !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("equal times err = %v, want ErrInvalidTimeRange", err)
	}
	if _, err := CalcLeaveDays("2026-03-02 09:00:00", "2026-03-02 08:00:00"); !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("reversed times err = %v, want ErrInvalidTimeRange", err)
	}
	if _, err := CalcLeaveDays("2026-03-02", "2026-03-03"); !errors.Is(err, ErrInvalidTimeFormat) {
		t.Fatalf("date-only err = %v, want ErrInvalidTimeFormat", err)
	}
}

// ---- shared fixture: org + employee + balance init + one leave bill ----

const testYear = 2026

func addOrg(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, name string) string {
	t.Helper()
	req := &orgtypes.AddRequest{OrgName: name, OrgType: "1", Status: "1"}
	if _, err := org.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add org %s: %v", name, err)
	}
	tree, err := org.NewList(ctx, svcCtx, nil).List(&orgtypes.ListRequest{})
	if err != nil {
		t.Fatalf("list org tree: %v", err)
	}
	for _, n := range tree.Records {
		if n.OrgName == name {
			return n.Uuid
		}
	}
	t.Fatalf("org %s not found", name)
	return ""
}


// listLeaves calls the public List seam with default pagination filled in.
func listLeaves(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, req *types.ListRequest) *types.ListResponse {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 100
	}
	resp, err := NewList(ctx, svcCtx, nil).List(req)
	if err != nil {
		t.Fatalf("List leaves: %v", err)
	}
	return resp
}

// addLeave creates one bill through the public Add seam and returns its uuid.
func addLeave(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context,
	leaveTypeCode string, mutate func(*types.AddRequest)) (string, *types.AddRequest) {
	t.Helper()
	req := &types.AddRequest{
		EmployeeUuid:  testEmpUuid,
		EmployeeNo:    "00000001",
		EmployeeName:  "张三",
		LeaveTypeCode: leaveTypeCode,
		PlanStartTime: "2026-03-02 09:00:00",
		PlanEndTime:   "2026-03-03 13:00:00", // 28h → 3.5 天
		Reason:        "家庭事务",
	}
	if mutate != nil {
		mutate(req)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("Add leave: %v", err)
	}
	resp := listLeaves(t, svcCtx, ctx, &types.ListRequest{Mine: 1})
	if len(resp.Records) == 0 {
		t.Fatal("added leave bill not visible in 我的单据 list")
	}
	rec := resp.Records[0]
	if len(rec.BillNo) < 6 || rec.BillNo[:6] != "HR205-" {
		t.Fatalf("billNo %q does not follow HR205- format", rec.BillNo)
	}
	if rec.Status != bill.StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	return rec.Uuid, req
}

// submitAndApprove drives the shared lifecycle to 审批通过.
func submitAndApprove(t *testing.T, conn *svctest.Env, typeCode, billUuid string) {
	t.Helper()
	op := bill.Operator{UserUuid: "op-1", Name: "管理员"}
	if err := bill.SubmitBill(context.Background(), conn.SqlConn, typeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), conn.SqlConn, typeCode, billUuid, op, "同意"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
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

func leaveErrCode(err error) int {
	if err == nil {
		return 0
	}
	if sc := status.FromError(err); sc != nil {
		return int(sc.Code())
	}
	return -1
}

// testEmpUuid is the seeded employee's archive uuid — set by seedEmployee.
// Tests in this package run sequentially (no t.Parallel), so a package-level
// fixture variable is safe and keeps assertion queries readable.
var testEmpUuid string

// seedEmployee seeds one 在职 employee through the public employee seam so that
// leavebalance Init (全体在册口径) has a target; records its archive uuid in
// testEmpUuid for bill fixture + assertion queries.
func seedEmployee(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context) {
	t.Helper()
	req := &emptypes.AddRequest{
		EmployeeNo:   "00000001",
		Name:         "张三",
		Gender:       "1",
		OrgUuid:      "",
		CompanyUuid:  "",
		Position:     "工程师",
		Post:         "后端",
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
	t.Fatal("seeded employee not found")
}

// initBalanceYear initializes managed leave balances for the year of the bill
// (预计开始时间 2026-03-02 → year 2026).
func initBalanceYear(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context) {
	t.Helper()
	seedEmployee(t, svcCtx, ctx)
	if _, err := leavebalance.NewInit(ctx, svcCtx, nil).Init(&lbtypes.InitRequest{Year: testYear}); err != nil {
		t.Fatalf("leavebalance Init: %v", err)
	}
}
