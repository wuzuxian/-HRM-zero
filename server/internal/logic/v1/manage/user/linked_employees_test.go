package user

import (
	"context"
	"strings"
	"testing"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	hrmTypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	manageUserTypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/manage/user"
)

// HRM-51 用户管理侧只读反查: 批量按账号 uuid 返回关联员工 (工号+姓名)。
// 关联唯一写入入口在员工侧 (ADR 0007), 这里只验证查询行为。

func newManageUserCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter:   config.NewStaticConfigCenter(config.Config{Hrm: config.HrmConf{InitialUserPassword: "test-pwd-123"}}),
		Model:          env.Model,
	}
}

func addHrmEmployee(t *testing.T, svcCtx *svc.ServiceContext, req *hrmTypes.AddRequest) {
	t.Helper()
	if _, err := employee.NewAdd(context.Background(), svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add employee %s: %v", req.Name, err)
	}
}

func hrmList(t *testing.T, svcCtx *svc.ServiceContext) []hrmTypes.HrmEmployee {
	t.Helper()
	resp, err := employee.NewList(context.Background(), svcCtx, nil).List(
		&hrmTypes.ListRequest{PageRequest: hrmTypes.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("list employees: %v", err)
	}
	return resp.Records
}

// TestLinkedEmployeesQuery 按账号 uuid 批量返回工号+姓名; 未关联/不存在的账号不返回。
func TestLinkedEmployeesQuery(t *testing.T) {
	svcCtx := newManageUserCtx(t)

	addHrmEmployee(t, svcCtx, &hrmTypes.AddRequest{Name: "张三", EmployeeNo: "00000001"})
	addHrmEmployee(t, svcCtx, &hrmTypes.AddRequest{Name: "李四", EmployeeNo: "00000002"})
	records := hrmList(t, svcCtx)
	if len(records) != 2 {
		t.Fatalf("records=%d, want 2", len(records))
	}
	// 列表按 id DESC: 按工号定位, 不按下标
	var a, b hrmTypes.HrmEmployee
	for _, rec := range records {
		switch rec.EmployeeNo {
		case "00000001":
			a = rec
		case "00000002":
			b = rec
		}
	}
	if a.Uuid == "" || b.Uuid == "" {
		t.Fatalf("records missing expected employees: %+v", records)
	}

	resp, err := NewLinkedEmployees(context.Background(), svcCtx, nil).LinkedEmployees(&manageUserTypes.LinkedEmployeesRequest{
		Uuids: strings.Join([]string{a.UserUuid, b.UserUuid, "orphan-uuid", "missing-uuid"}, ","),
	})
	if err != nil {
		t.Fatalf("linkedEmployees: %v", err)
	}
	byUser := map[string]manageUserTypes.LinkedEmployee{}
	for _, rec := range resp.Records {
		byUser[rec.UserUuid] = rec
	}
	if len(byUser) != 2 {
		t.Fatalf("records=%v, want 2 (未关联/不存在的账号不返回)", byUser)
	}
	if rec := byUser[a.UserUuid]; rec.EmployeeNo != "00000001" || rec.EmployeeName != "张三" {
		t.Fatalf("A record = %+v, want 00000001/张三", rec)
	}
	if rec := byUser[b.UserUuid]; rec.EmployeeNo != "00000002" || rec.EmployeeName != "李四" {
		t.Fatalf("B record = %+v, want 00000002/李四", rec)
	}

	// 空 uuids → 空结果不报错
	empty, err := NewLinkedEmployees(context.Background(), svcCtx, nil).LinkedEmployees(&manageUserTypes.LinkedEmployeesRequest{})
	if err != nil || len(empty.Records) != 0 {
		t.Fatalf("empty uuids: resp=%+v err=%v", empty, err)
	}
}
