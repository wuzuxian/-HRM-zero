package leavebalance

import (
	"context"
	"testing"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	emptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	orgtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
	"github.com/jzero-io/jzero/core/status"
)

const (
	testYear         = int64(2026)
	annualLeaveCode  = "annual_leave"
)

func newSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		ServiceContext: &coresvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

// testCtxFor mimics JWT auth info (uuid/username/role_uuids keys, reward_test.go 同款).
func testCtxFor(uuid, username string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "uuid", uuid)
	ctx = context.WithValue(ctx, "username", username)
	ctx = context.WithValue(ctx, "role_uuids", []string{"some-role"})
	return ctx
}

// addOrg adds an org node through the public org seam and returns its uuid.
func addOrg(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, parentUuid, name string) string {
	t.Helper()
	req := &orgtypes.AddRequest{ParentUuid: parentUuid, OrgName: name, OrgType: "2", Status: "1"}
	if parentUuid == "" {
		req.OrgType = "1" // company root
	}
	if _, err := org.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add org %s: %v", name, err)
	}
	tree, err := org.NewList(ctx, svcCtx, nil).List(&orgtypes.ListRequest{})
	if err != nil {
		t.Fatalf("list org tree: %v", err)
	}
	u := findOrgUuid(tree.Records, name)
	if u == "" {
		t.Fatalf("org %s not found in tree", name)
	}
	return u
}

func findOrgUuid(nodes []orgtypes.HrmOrg, name string) string {
	for _, n := range nodes {
		if n.OrgName == name {
			return n.Uuid
		}
		if u := findOrgUuid(n.Children, name); u != "" {
			return u
		}
	}
	return ""
}

// addEmployee seeds an employee through the public employee seam and returns its uuid.
func addEmployee(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, employeeNo, name, orgUuid, personStatus string) string {
	t.Helper()
	req := &emptypes.AddRequest{
		EmployeeNo:   employeeNo,
		Name:         name,
		Gender:       "1",
		OrgUuid:      orgUuid,
		CompanyUuid:  "",
		Position:     "工程师",
		Post:         "后端",
		PersonStatus: personStatus,
		HireDate:     "2024-03-01",
	}
	if _, err := employee.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add employee %s: %v", name, err)
	}
	resp, err := employee.NewList(ctx, svcCtx, nil).List(&emptypes.ListRequest{PageRequest: emptypes.PageRequest{Current: 1, Size: 50}})
	if err != nil {
		t.Fatalf("list employees: %v", err)
	}
	for _, rec := range resp.Records {
		if rec.EmployeeNo == employeeNo {
			return rec.Uuid
		}
	}
	t.Fatalf("employee %s not found after add", employeeNo)
	return ""
}

// initAll initializes every managed leave type for the seeded year.
func initAll(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, year int64, employeeUuids []string) {
	t.Helper()
	req := &types.InitRequest{Year: year}
	if employeeUuids != nil {
		req.EmployeeUuids = employeeUuids
	}
	if _, err := NewInit(ctx, svcCtx, nil).Init(req); err != nil {
		t.Fatalf("Init: %v", err)
	}
}

func listBalances(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, req *types.ListRequest) *types.ListResponse {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 100
	}
	resp, err := NewList(ctx, svcCtx, nil).List(req)
	if err != nil {
		t.Fatalf("List balances: %v", err)
	}
	return resp
}

func findBalance(resp *types.ListResponse, employeeUuid, typeCode string) *types.HrmLeaveBalance {
	for i := range resp.Records {
		rec := &resp.Records[i]
		if rec.EmployeeUuid == employeeUuid && rec.LeaveTypeCode == typeCode {
			return rec
		}
	}
	return nil
}

func listLogs(t *testing.T, svcCtx *svc.ServiceContext, ctx context.Context, req *types.LogListRequest) *types.LogListResponse {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 100
	}
	resp, err := NewLogList(ctx, svcCtx, nil).LogList(req)
	if err != nil {
		t.Fatalf("LogList: %v", err)
	}
	return resp
}

func balanceErrCode(err error) int {
	if err == nil {
		return 0
	}
	if sc := status.FromError(err); sc != nil {
		return int(sc.Code())
	}
	return -1
}

// TestInitSeedsManagedTypesWithLog 初始化按默认天数落账 + init 流水 + 列表快照/可用余额。
func TestInitSeedsManagedTypesWithLog(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("op-1", "管理员")

	orgUuid := addOrg(t, svcCtx, ctx, "", "总公司")
	empUuid := addEmployee(t, svcCtx, ctx, "00000001", "张三", orgUuid, "1")

	initAll(t, svcCtx, ctx, testYear, nil)

	resp := listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear})
	bal := findBalance(resp, empUuid, annualLeaveCode)
	if bal == nil {
		t.Fatal("annual_leave balance row missing after init")
	}
	if bal.Granted != 5 || bal.Used != 0 || bal.Available != 5 {
		t.Fatalf("annual_leave granted/used/available = %v/%v/%v, want 5/0/5", bal.Granted, bal.Used, bal.Available)
	}
	if bal.EmployeeNo != "00000001" || bal.EmployeeName != "张三" || bal.OrgName != "总公司" || bal.HireDate != "2024-03-01" {
		t.Fatalf("balance snapshot wrong: %+v", bal)
	}
	// comp_leave also managed (default_days=0 → granted 0)
	comp := findBalance(resp, empUuid, "comp_leave")
	if comp == nil || comp.Granted != 0 {
		t.Fatalf("comp_leave row wrong: %+v", comp)
	}
	// 事假 (balance_enabled=0) must NOT be initialized
	if findBalance(resp, empUuid, "personal_leave") != nil {
		t.Fatal("personal_leave should not be initialized (余额管理未启用)")
	}

	logs := listLogs(t, svcCtx, ctx, &types.LogListRequest{Year: testYear, EmployeeUuid: empUuid})
	if len(logs.Records) != 2 {
		t.Fatalf("init logs = %d, want 2 (annual+comp)", len(logs.Records))
	}
	for _, lg := range logs.Records {
		if lg.ChangeType != ChangeTypeInit {
			t.Fatalf("log changeType = %q, want init", lg.ChangeType)
		}
	}
}

// TestInitIdempotentSkipsExisting 重复初始化：已有余额行跳过，不覆盖、不重复记流水。
func TestInitIdempotentSkipsExisting(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("op-1", "管理员")

	orgUuid := addOrg(t, svcCtx, ctx, "", "总公司")
	empUuid := addEmployee(t, svcCtx, ctx, "00000001", "张三", orgUuid, "1")

	initAll(t, svcCtx, ctx, testYear, nil)
	// 空 balanceUuid 调整 → 参数非法
	if _, err := NewAdjust(ctx, svcCtx, nil).Adjust(&types.AdjustRequest{
		BalanceUuid: "", Days: 1,
	}); err == nil {
		t.Fatal("adjust without balanceUuid should fail")
	}
	resp := listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear})
	bal := findBalance(resp, empUuid, annualLeaveCode)
	if _, err := NewAdjust(ctx, svcCtx, nil).Adjust(&types.AdjustRequest{
		BalanceUuid: bal.Uuid, Days: 2, Reason: "补提",
	}); err != nil {
		t.Fatalf("adjust: %v", err)
	}

	initAll(t, svcCtx, ctx, testYear, nil) // repeat init

	resp = listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear})
	bal = findBalance(resp, empUuid, annualLeaveCode)
	if bal.Granted != 7 {
		t.Fatalf("granted after re-init = %v, want 7 (skip keeps adjust)", bal.Granted)
	}
	logs := listLogs(t, svcCtx, ctx, &types.LogListRequest{Year: testYear, EmployeeUuid: empUuid})
	initCount := 0
	adjustCount := 0
	for _, lg := range logs.Records {
		switch lg.ChangeType {
		case ChangeTypeInit:
			initCount++
		case ChangeTypeAdjust:
			adjustCount++
		}
	}
	if initCount != 2 {
		t.Fatalf("init log count = %d, want 2 (annual+comp; re-init must not add more)", initCount)
	}
	if adjustCount != 1 {
		t.Fatalf("adjust log count = %d, want 1", adjustCount)
	}
}

// TestAdjustBoundaryRejectsNegativeAvailable 调整边界：扣减导致可用 < 0 → 10513。
func TestAdjustBoundaryRejectsNegativeAvailable(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("op-1", "管理员")

	orgUuid := addOrg(t, svcCtx, ctx, "", "总公司")
	empUuid := addEmployee(t, svcCtx, ctx, "00000001", "张三", orgUuid, "1")
	initAll(t, svcCtx, ctx, testYear, nil)

	resp := listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear})
	bal := findBalance(resp, empUuid, annualLeaveCode)

	// 扣 6 天（可用只有 5）→ 拒绝
	_, err := NewAdjust(ctx, svcCtx, nil).Adjust(&types.AdjustRequest{BalanceUuid: bal.Uuid, Days: -6, Reason: "超扣"})
	if balanceErrCode(err) != 10513 {
		t.Fatalf("over-deduct err code = %v, want 10513 (LeaveBalanceInsufficient), err=%v", balanceErrCode(err), err)
	}

	// 扣 5 天 → 可用 0；再扣 0.5 → 拒绝
	if _, err = NewAdjust(ctx, svcCtx, nil).Adjust(&types.AdjustRequest{BalanceUuid: bal.Uuid, Days: -5, Reason: "扣完"}); err != nil {
		t.Fatalf("deduct all: %v", err)
	}
	_, err = NewAdjust(ctx, svcCtx, nil).Adjust(&types.AdjustRequest{BalanceUuid: bal.Uuid, Days: -0.5, Reason: "再扣"})
	if balanceErrCode(err) != 10513 {
		t.Fatalf("deduct from zero err code = %v, want 10513", balanceErrCode(err))
	}

	// days=0 → 10512
	_, err = NewAdjust(ctx, svcCtx, nil).Adjust(&types.AdjustRequest{BalanceUuid: bal.Uuid, Days: 0})
	if balanceErrCode(err) != 10512 {
		t.Fatalf("zero adjust err code = %v, want 10512", balanceErrCode(err))
	}
}

// TestLogIntegrityConsumeAccrue 流水完整性 + ConsumeLeave/AccrueCompLeave 事务原子性。
func TestLogIntegrityConsumeAccrue(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("op-1", "管理员")

	orgUuid := addOrg(t, svcCtx, ctx, "", "总公司")
	empUuid := addEmployee(t, svcCtx, ctx, "00000001", "张三", orgUuid, "1")
	initAll(t, svcCtx, ctx, testYear, nil)

	// HRM-21: 调休入账 16h = 2 天 comp_leave
	if err := AccrueCompLeave(ctx, env.SqlConn, svcCtx, empUuid, 16, "bill-ot-1"); err != nil {
		t.Fatalf("AccrueCompLeave: %v", err)
	}
	resp := listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear})
	comp := findBalance(resp, empUuid, "comp_leave")
	if comp == nil || comp.Granted != 2 || comp.Available != 2 {
		t.Fatalf("comp_leave after accrual = %+v, want granted 2", comp)
	}

	// HRM-19: 年假扣减 1.5 天
	if err := ConsumeLeave(ctx, env.SqlConn, svcCtx, empUuid, annualLeaveCode, 1.5, "bill-leave-1"); err != nil {
		t.Fatalf("ConsumeLeave: %v", err)
	}
	resp = listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear})
	bal := findBalance(resp, empUuid, annualLeaveCode)
	if bal.Used != 1.5 || bal.Available != 3.5 {
		t.Fatalf("annual after consume used/available = %v/%v, want 1.5/3.5", bal.Used, bal.Available)
	}

	// 扣减超过可用 → 10513，且失败调用不产生流水（原子性）
	err := ConsumeLeave(ctx, env.SqlConn, svcCtx, empUuid, annualLeaveCode, 99, "bill-leave-2")
	if balanceErrCode(err) != 10513 {
		t.Fatalf("over-consume err code = %v, want 10513", balanceErrCode(err))
	}

	logs := listLogs(t, svcCtx, ctx, &types.LogListRequest{Year: testYear, EmployeeUuid: empUuid})
	var consumeLogs, accrueLogs []*types.HrmLeaveBalanceLog
	for i := range logs.Records {
		lg := &logs.Records[i]
		switch lg.ChangeType {
		case ChangeTypeConsume:
			consumeLogs = append(consumeLogs, lg)
		case ChangeTypeAccrue:
			accrueLogs = append(accrueLogs, lg)
		}
	}
	if len(consumeLogs) != 1 {
		t.Fatalf("consume log count = %d, want 1 (failed call must not log)", len(consumeLogs))
	}
	if len(accrueLogs) != 1 {
		t.Fatalf("accrue log count = %d, want 1", len(accrueLogs))
	}
	if consumeLogs[0].Days != -1.5 || consumeLogs[0].BillUuid != "bill-leave-1" || consumeLogs[0].UsedAfter != 1.5 {
		t.Fatalf("consume log fields wrong: %+v", consumeLogs[0])
	}
	if accrueLogs[0].Days != 2 || accrueLogs[0].BillUuid != "bill-ot-1" || accrueLogs[0].GrantedAfter != 2 {
		t.Fatalf("accrue log fields wrong: %+v", accrueLogs[0])
	}
}

// TestAccrueRequiresManagedCompLeave comp_leave 未启用余额管理时入账被拒（10514）。
func TestAccrueRequiresManagedCompLeave(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("op-1", "管理员")

	// 通过假期类型 seam 停用 comp_leave 的余额管理
	ltResp, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{})
	_ = ltResp
	_ = err
	ltRows, err := svcCtx.Model.HrmLeaveType.FindByCondition(ctx, nil,
		conditionEqualTypeCode("comp_leave").Build()...)
	if err != nil || len(ltRows) == 0 {
		t.Fatalf("comp_leave seed missing: %v", err)
	}
	if err := svcCtx.Model.HrmLeaveType.UpdateFieldsByCondition(ctx, nil,
		map[string]any{"balance_enabled": int64(0)},
		conditionEqualTypeCode("comp_leave").Build()...); err != nil {
		t.Fatalf("disable comp balance: %v", err)
	}

	orgUuid := addOrg(t, svcCtx, ctx, "", "总公司")
	empUuid := addEmployee(t, svcCtx, ctx, "00000001", "张三", orgUuid, "1")

	err = AccrueCompLeave(ctx, env.SqlConn, svcCtx, empUuid, 8, "bill-ot-2")
	if balanceErrCode(err) != 10514 {
		t.Fatalf("accrue err code = %v, want 10514 (LeaveBalanceTypeNotManaged)", balanceErrCode(err))
	}
}

// TestListOrgSubtreeFilter 部门含下级筛选：父部门过滤命中子部门员工余额。
func TestListOrgSubtreeFilter(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("op-1", "管理员")

	companyUuid := addOrg(t, svcCtx, ctx, "", "总公司")
	rdUuid := addOrg(t, svcCtx, ctx, companyUuid, "研发部")
	saleUuid := addOrg(t, svcCtx, ctx, companyUuid, "销售部")

	rdEmp := addEmployee(t, svcCtx, ctx, "00000001", "张三", rdUuid, "1")
	saleEmp := addEmployee(t, svcCtx, ctx, "00000002", "李四", saleUuid, "1")
	initAll(t, svcCtx, ctx, testYear, nil)

	// 部门=研发部 → 只出张三
	resp := listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear, OrgUuid: rdUuid})
	if findBalance(resp, rdEmp, annualLeaveCode) == nil {
		t.Fatal("rd employee balance missing with rd org filter")
	}
	if findBalance(resp, saleEmp, annualLeaveCode) != nil {
		t.Fatal("sale employee balance leaked into rd org filter")
	}

	// 部门=总公司（含下级）→ 两个都出
	resp = listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear, OrgUuid: companyUuid})
	if findBalance(resp, rdEmp, annualLeaveCode) == nil || findBalance(resp, saleEmp, annualLeaveCode) == nil {
		t.Fatalf("company-wide subtree filter missed descendants: %d records", len(resp.Records))
	}

	// 姓名模糊
	resp = listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear, Name: "张"})
	if findBalance(resp, rdEmp, annualLeaveCode) == nil || findBalance(resp, saleEmp, annualLeaveCode) != nil {
		t.Fatal("name filter wrong")
	}
}

// TestInitSkipsResignedRetired 全员初始化跳过 离职/退休 员工（在册口径）。
func TestInitSkipsResignedRetired(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("op-1", "管理员")

	orgUuid := addOrg(t, svcCtx, ctx, "", "总公司")
	active := addEmployee(t, svcCtx, ctx, "00000001", "张三", orgUuid, "1")
	resigned := addEmployee(t, svcCtx, ctx, "00000002", "王五", orgUuid, "5")

	initAll(t, svcCtx, ctx, testYear, nil)

	resp := listBalances(t, svcCtx, ctx, &types.ListRequest{Year: testYear})
	if findBalance(resp, active, annualLeaveCode) == nil {
		t.Fatal("active employee balance missing")
	}
	if findBalance(resp, resigned, annualLeaveCode) != nil {
		t.Fatal("resigned employee should be skipped by 全员初始化 (在册口径)")
	}
}

// TestInitRejectsUnknownEmployee 指定不存在的员工 → 10515。
func TestInitRejectsUnknownEmployee(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("op-1", "管理员")

	_, err := NewInit(ctx, svcCtx, nil).Init(&types.InitRequest{
		Year: testYear, EmployeeUuids: []string{"no-such-uuid"},
	})
	if balanceErrCode(err) != 10515 {
		t.Fatalf("init unknown employee err code = %v, want 10515", balanceErrCode(err))
	}
}
