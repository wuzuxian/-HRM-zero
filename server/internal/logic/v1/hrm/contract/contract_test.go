package contract

import (
	"context"
	"strings"
	"testing"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/contract"
	emptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

// newCtx DB-backed harness（同 employee 测试：contract 逻辑依赖 Model）。
func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter:   config.NewStaticConfigCenter(config.Config{Hrm: config.HrmConf{InitialUserPassword: "test-pwd-123"}}),
		Model:          env.Model,
	}
}

func addEmployee(t *testing.T, svcCtx *svc.ServiceContext, name string) emptypes.HrmEmployee {
	t.Helper()
	if _, err := employee.NewAdd(context.Background(), svcCtx, nil).Add(&emptypes.AddRequest{Name: name, PersonStatus: "1"}); err != nil {
		t.Fatalf("add employee %s: %v", name, err)
	}
	list, err := employee.NewList(context.Background(), svcCtx, nil).List(&emptypes.ListRequest{PageRequest: emptypes.PageRequest{Current: 1, Size: 10}, Name: name})
	if err != nil {
		t.Fatalf("list employees: %v", err)
	}
	if len(list.Records) != 1 {
		t.Fatalf("expected 1 employee %q, got %d", name, len(list.Records))
	}
	return list.Records[0]
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) types.ListResponse {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list contracts: %v", err)
	}
	return *resp
}

func withPage(req types.ListRequest, size int) types.ListRequest {
	req.Current = 1
	req.Size = size
	return req
}

func oneRecord(t *testing.T, list types.ListResponse) types.HrmEmployeeContract {
	t.Helper()
	if len(list.Records) != 1 {
		t.Fatalf("expected 1 record, got %d (%+v)", len(list.Records), list.Records)
	}
	return list.Records[0]
}

// TestContractAddWithEmployeeSnapshot 覆盖新增合同：员工快照带出（工号/姓名/部门）、
// 参数非法报错、员工不存在报错（spec 用户故事 17）。
func TestContractAddWithEmployeeSnapshot(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := addEmployee(t, svcCtx, "张三")

	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: emp.Uuid,
		ContractNo:   "HT-2026-001",
		StartDate:    "2026-01-01",
		EndDate:      "2026-11-16",
	}); err != nil {
		t.Fatalf("add contract: %v", err)
	}

	rec := oneRecord(t, mustList(t, svcCtx, types.ListRequest{}))
	if rec.EmployeeNo != emp.EmployeeNo || rec.EmployeeName != "张三" || rec.OrgUuid != emp.OrgUuid {
		t.Fatalf("snapshot mismatch: %+v (emp %+v)", rec, emp)
	}
	if rec.RenewCount != 0 || rec.Status != "1" {
		t.Fatalf("renewCount/status = %d/%q, want 0/1", rec.RenewCount, rec.Status)
	}
	if rec.DaysToExpiry != 0 || rec.ExpiryState != "warning" {
		t.Fatalf("days/state = %d/%q, want 0/warning", rec.DaysToExpiry, rec.ExpiryState)
	}

	// 员工不存在
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: "no-such-uuid", StartDate: "2026-01-01", EndDate: "2026-12-31",
	}); err == nil {
		t.Fatal("add contract with unknown employee should fail")
	}
	// 到期日早于生效日
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: emp.Uuid, StartDate: "2026-06-01", EndDate: "2026-01-01",
	}); err == nil {
		t.Fatal("end before start should fail")
	}
	// 日期格式非法
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: emp.Uuid, StartDate: "2026/01/01", EndDate: "2026-12-31",
	}); err == nil {
		t.Fatal("bad date format should fail")
	}
}

// TestContractListFilters 覆盖预警页筛选口径：姓名/工号模糊、到期范围区间（spec 用户故事 14）。
func TestContractListFilters(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := addEmployee(t, svcCtx, "李四")
	add := func(name, end string) {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
			EmployeeUuid: emp.Uuid, ContractNo: name, StartDate: "2026-01-01", EndDate: end,
		}); err != nil {
			t.Fatalf("add contract %s: %v", name, err)
		}
	}
	add("A-远期", "2028-06-30")
	add("B-临期", "2026-11-20")
	add("C-已过期", "2026-10-01")

	// 到期范围筛选：11-01 ~ 12-31 → 只剩 B-临期
	list := mustList(t, svcCtx, types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, StartDate: "2026-11-01", EndDate: "2026-12-31"})
	if list.Total != 1 || len(list.Records) != 1 || list.Records[0].ContractNo != "B-临期" {
		t.Fatalf("range filter total=%d records=%+v, want only B-临期", list.Total, list.Records)
	}

	// 预警状态筛选（内存口径过滤）
	warned := mustList(t, svcCtx, withPage(types.ListRequest{ExpiryState: "warning"}, 50))
	if warned.Total != 1 || warned.Records[0].ContractNo != "B-临期" {
		t.Fatalf("warning filter = %+v, want only B-临期", warned.Records)
	}
	expired := mustList(t, svcCtx, withPage(types.ListRequest{ExpiryState: "expired"}, 50))
	if expired.Total != 1 || expired.Records[0].ContractNo != "C-已过期" {
		t.Fatalf("expired filter = %+v, want only C-已过期", expired.Records)
	}
	normal := mustList(t, svcCtx, withPage(types.ListRequest{ExpiryState: "normal"}, 50))
	if normal.Total != 1 || normal.Records[0].ContractNo != "A-远期" {
		t.Fatalf("normal filter = %+v, want only A-远期", normal.Records)
	}

	// 姓名模糊
	byName := mustList(t, svcCtx, withPage(types.ListRequest{EmployeeName: "李四"}, 50))
	if byName.Total != 3 {
		t.Fatalf("name filter total=%d, want 3", byName.Total)
	}
}

// TestContractRenew 覆盖续签动作闭环（spec 用户故事 15）：
// 续签次数递增、缺省新到期日=原到期日+1 年、指定到期日生效、已解除合同不可续签。
func TestContractRenew(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := addEmployee(t, svcCtx, "王五")

	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: emp.Uuid, ContractNo: "HT-001", StartDate: "2025-11-16", EndDate: "2026-11-16",
	}); err != nil {
		t.Fatalf("add contract: %v", err)
	}
	rec := oneRecord(t, mustList(t, svcCtx, types.ListRequest{}))

	// 第一次续签（缺省 = 原到期日 + 1 年）
	if _, err := NewRenew(ctx, svcCtx, nil).Renew(&types.RenewRequest{Uuid: rec.Uuid}); err != nil {
		t.Fatalf("renew #1: %v", err)
	}
	rec = oneRecord(t, mustList(t, svcCtx, types.ListRequest{}))
	if rec.RenewCount != 1 {
		t.Fatalf("renewCount after #1 = %d, want 1", rec.RenewCount)
	}
	if rec.EndDate != "2027-11-16" {
		t.Fatalf("endDate after default renew = %q, want 2027-11-16", rec.EndDate)
	}

	// 第二次续签（指定新到期日）
	if _, err := NewRenew(ctx, svcCtx, nil).Renew(&types.RenewRequest{Uuid: rec.Uuid, EndDate: "2030-06-30"}); err != nil {
		t.Fatalf("renew #2: %v", err)
	}
	rec = oneRecord(t, mustList(t, svcCtx, types.ListRequest{}))
	if rec.RenewCount != 2 || rec.EndDate != "2030-06-30" {
		t.Fatalf("after #2 renewCount/endDate = %d/%q, want 2/2030-06-30", rec.RenewCount, rec.EndDate)
	}

	// 已解除合同不可续签
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: rec.Uuid, ContractNo: rec.ContractNo, StartDate: rec.StartDate, EndDate: rec.EndDate, Status: "3",
	}); err != nil {
		t.Fatalf("terminate contract: %v", err)
	}
	if _, err := NewRenew(ctx, svcCtx, nil).Renew(&types.RenewRequest{Uuid: rec.Uuid}); err == nil {
		t.Fatal("renew terminated contract should fail")
	}

	// 不存在的合同
	if _, err := NewRenew(ctx, svcCtx, nil).Renew(&types.RenewRequest{Uuid: "no-such"}); err == nil {
		t.Fatal("renew unknown contract should fail")
	}
}

// TestContractDelete 覆盖单删/批删。
func TestContractDelete(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := addEmployee(t, svcCtx, "赵六")

	add := func(no string) string {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
			EmployeeUuid: emp.Uuid, ContractNo: no, StartDate: "2026-01-01", EndDate: "2027-01-01",
		}); err != nil {
			t.Fatalf("add contract %s: %v", no, err)
		}
		list := mustList(t, svcCtx, withPage(types.ListRequest{EmployeeNo: emp.EmployeeNo}, 50))
		for _, rec := range list.Records {
			if rec.ContractNo == no {
				return rec.Uuid
			}
		}
		t.Fatalf("contract %s not found after add", no)
		return ""
	}
	_ = add("D-1")
	u2 := add("D-2")

	list := mustList(t, svcCtx, withPage(types.ListRequest{}, 50))
	if list.Total != 2 {
		t.Fatalf("before delete total=%d, want 2", list.Total)
	}
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{u2}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list = mustList(t, svcCtx, withPage(types.ListRequest{}, 50))
	if list.Total != 1 || !strings.Contains(list.Records[0].ContractNo, "D-1") {
		t.Fatalf("after delete total=%d records=%+v, want only D-1", list.Total, list.Records)
	}
}
