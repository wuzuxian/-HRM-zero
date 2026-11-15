package cardreplacement

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	employeelogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	orglogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_card_replacement_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	emptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	orgtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cardreplacement"
)

// Seams under test (issue acceptance criteria), all at logic seams against the
// real migrations (svctest):
//  1. Add: HR206 编号 / 未提交 / 员工+原打卡快照列（原上班打卡/原下班打卡）.
//  2. 审批通过事务内: 补卡写入当日 hrm_punch_record 对应列 + re-run SettleDay（迟到→正常）;
//     无打卡行建行; 幂等（重复审批被状态机拒绝）.
//  3. 生命周期守卫: edit/delete 仅 未提交/审批不通过/已撤销.
//  4. 「我的单据」默认过滤 + 筛选（编号/状态/工号/补卡类型）.

// ---- shared helpers (leave_test.go 同款最小 svcCtx + JWT ctx) ----

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

// approvalCtx attaches the test svcCtx so OnApproved (补卡写打卡行 + SettleDay)
// can resolve it — mirrors the production handler context wiring via WithSvcCtx.
func approvalCtx(svcCtx *svc.ServiceContext) context.Context {
	return WithSvcCtx(context.Background(), svcCtx)
}

func crErrCode(err error) int {
	if err == nil {
		return 0
	}
	if sc := status.FromError(err); sc != nil {
		return int(sc.Code())
	}
	return -1
}

// listBills calls the public List seam with default pagination filled in.
func listBills(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, req *types.ListRequest) *types.ListResponse {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 100
	}
	resp, err := NewList(ctx, svcCtx, nil).List(req)
	if err != nil {
		t.Fatalf("List card replacements: %v", err)
	}
	return resp
}

// findBill locates one bill in the admin list by uuid.
func findBill(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) *types.HrmCardReplacement {
	t.Helper()
	resp := listBills(t, svcCtx, superCtx("admin", "管理员"), &types.ListRequest{})
	for i := range resp.Records {
		if resp.Records[i].Uuid == billUuid {
			return &resp.Records[i]
		}
	}
	return nil
}

// addBill creates one bill through the public Add seam and returns its uuid +
// the request used (独立期望值来源: 断言以 req 字面量为基准, 非重算代码).
func addBill(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context,
	mutate func(*types.AddRequest)) (string, *types.AddRequest) {
	t.Helper()
	req := &types.AddRequest{
		EmployeeUuid: testEmpUuid,
		EmployeeNo:   "E0001",
		EmployeeName: "张三",
		PunchDate:    "2026-11-16",
		CardType:     CardTypeClockIn,
		PunchTime:    "2026-11-16 09:10:00",
		Reason:       "忘记打卡",
	}
	if mutate != nil {
		mutate(req)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("Add card replacement: %v", err)
	}
	rec := findBill(t, svcCtx, latestUuid(t, svcCtx))
	if rec == nil {
		t.Fatal("added bill not visible in admin list")
	}
	if len(rec.BillNo) < 6 || rec.BillNo[:6] != "HR206-" {
		t.Fatalf("billNo %q does not follow HR206- format", rec.BillNo)
	}
	if rec.Status != bill.StatusUnsubmitted {
		t.Fatalf("new bill status = %q, want unsubmitted", rec.Status)
	}
	return rec.Uuid, req
}

// latestUuid returns the newest bill's uuid (tests add exactly one bill per call).
func latestUuid(t *testing.T, svcCtx *svc.ServiceContext) string {
	t.Helper()
	resp, err := svcCtx.Model.HrmCardReplacementBill.FindByCondition(context.Background(), nil)
	if err != nil || len(resp) == 0 {
		t.Fatalf("read bills: %v (n=%d)", err, len(resp))
	}
	var best *hrm_card_replacement_bill.HrmCardReplacementBill
	for _, row := range resp {
		if best == nil || row.Id > best.Id {
			best = row
		}
	}
	return best.Uuid
}

// findPunchRowOf 当日打卡行（测试断言用，一人一天一行）.
func findPunchRowOf(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, date string) *hrm_punch_record.HrmPunchRecord {
	t.Helper()
	row, err := svcCtx.Model.HrmPunchRecord.FindOneByCondition(context.Background(), nil,
		condition.NewChain().
			Equal(hrm_punch_record.PunchDate, date).
			Equal(hrm_punch_record.EmployeeUuid, employeeUuid).
			Build()...)
	if err != nil {
		if err == hrm_punch_record.ErrNotFound {
			return nil
		}
		t.Fatalf("find punch row: %v", err)
	}
	return row
}

// submitAndApprove drives the shared lifecycle to 审批通过 (svcCtx wired for OnApproved).
func submitAndApprove(t *testing.T, env *svctest.Env, svcCtx *svc.ServiceContext, billUuid string) {
	t.Helper()
	op := bill.Operator{UserUuid: "op-1", Name: "管理员"}
	if err := bill.SubmitBill(approvalCtx(svcCtx), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(approvalCtx(svcCtx), env.SqlConn, BillTypeCode, billUuid, op, "同意"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
}

// ---- shared fixture: org + employee（同 punch.settle_test 口径: 公司A→技术部）----

var testEmpUuid string

func seedEmployee(t *testing.T, svcCtx *svc.ServiceContext, orgUuid string) {
	t.Helper()
	ctx := context.Background()
	req := &emptypes.AddRequest{Name: "张三", EmployeeNo: "E0001", OrgUuid: orgUuid, PersonStatus: "1"}
	if _, err := employeelogic.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add employee: %v", err)
	}
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, "E0001").Build()...)
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	testEmpUuid = row.Uuid
}

func seedOrgTree(t *testing.T, svcCtx *svc.ServiceContext) string {
	t.Helper()
	ctx := context.Background()
	if _, err := orglogic.NewAdd(ctx, svcCtx, nil).Add(&orgtypes.AddRequest{OrgName: "公司A", OrgType: "1", Status: "1"}); err != nil {
		t.Fatalf("add org A: %v", err)
	}
	tree, err := orglogic.NewList(ctx, svcCtx, nil).List(&orgtypes.ListRequest{})
	if err != nil {
		t.Fatalf("list org: %v", err)
	}
	return tree.Records[0].Uuid
}

// nullTime helper
func nullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}
