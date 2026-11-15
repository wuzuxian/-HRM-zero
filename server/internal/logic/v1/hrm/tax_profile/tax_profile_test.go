package tax_profile

import (
	"context"
	"testing"

	"github.com/google/uuid"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_profile"
)

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

func seedEmployee(t *testing.T, svcCtx *svc.ServiceContext, name, idCard string) string {
	t.Helper()
	u := uuid.New().String()
	if err := svcCtx.Model.HrmEmployee.InsertV2(t.Context(), nil, &hrm_employee.HrmEmployee{
		Uuid: u, EmployeeNo: name + "-no", Name: name, IdCard: idCard, PersonStatus: "1",
	}); err != nil {
		t.Fatalf("seed employee %s: %v", name, err)
	}
	return u
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmTaxProfile {
	t.Helper()
	req.Current = 1
	req.Size = 100
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list tax profiles: %v", err)
	}
	return resp.Records
}

func findByName(records []types.HrmTaxProfile, name string) *types.HrmTaxProfile {
	for i := range records {
		if records[i].EmployeeName == name {
			return &records[i]
		}
	}
	return nil
}

// TestTaxProfileCrud walks the CRUD seam (HRM-36 user story 58):
// 新增（员工姓名/身份证快照）→ 员工唯一守卫 → 编辑（换员工/启停）→ 筛选 → 删除。
func TestTaxProfileCrud(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	empA := seedEmployee(t, svcCtx, "张三", "420106199001011234")
	empB := seedEmployee(t, svcCtx, "李四", "420106199202025678")

	// 新增: 姓名/身份证快照 + 默认启停
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: empA, ResidentType: ResidentTypeResident,
		TaxCity: "武汉", DefaultSpecialExtra: "CHILDREN_EDU,SUPPORT_ELDERLY",
	}); err != nil {
		t.Fatalf("add tax profile: %v", err)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 1 {
		t.Fatalf("expected 1 profile after add, got %d", len(records))
	}
	if records[0].EmployeeName != "张三" || records[0].IdCard != "420106199001011234" ||
		records[0].Status != "1" || records[0].DefaultSpecialExtra != "CHILDREN_EDU,SUPPORT_ELDERLY" {
		t.Fatalf("add snapshot/defaults not applied: %+v", records[0])
	}

	// 员工唯一守卫: 同员工再建档拒绝
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: empA, ResidentType: ResidentTypeResident,
	}); err == nil {
		t.Fatalf("duplicate employee profile should fail")
	}

	// 员工不存在
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: "no-such", ResidentType: ResidentTypeResident,
	}); err == nil {
		t.Fatalf("unknown employee should fail")
	}

	// 居民类型枚举守卫
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: empB, ResidentType: "alien",
	}); err == nil {
		t.Fatalf("invalid resident type should fail")
	}

	// 编辑: 换员工 + 居民类型 + 停用
	row := findByName(records, "张三")
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: row.Uuid, EmployeeUuid: empB, ResidentType: ResidentTypeNonResident,
		TaxCity: "上海", Status: "0",
	}); err != nil {
		t.Fatalf("edit tax profile: %v", err)
	}
	row = findByName(mustList(t, svcCtx, types.ListRequest{}), "李四")
	if row == nil || row.ResidentType != ResidentTypeNonResident || row.Status != "0" ||
		row.IdCard != "420106199202025678" || row.TaxCity != "上海" {
		t.Fatalf("edit not applied: %+v", row)
	}

	// 编辑换到已建档员工（ empA 无档案了？有 —— 张三档案刚改成李四，empA 现在空闲）
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: row.Uuid, EmployeeUuid: empA, ResidentType: ResidentTypeNonResident, Status: "1",
	}); err != nil {
		t.Fatalf("edit swap employee: %v", err)
	}

	// 编辑不存在的档案
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: "no-such-uuid", EmployeeUuid: empB, ResidentType: ResidentTypeResident, Status: "1",
	}); err == nil {
		t.Fatalf("editing unknown profile should fail")
	}

	// 筛选: 姓名 / 居民类型
	if named := mustList(t, svcCtx, types.ListRequest{EmployeeName: "张三"}); len(named) != 1 {
		t.Fatalf("name filter failed: %+v", named)
	}
	if typed := mustList(t, svcCtx, types.ListRequest{ResidentType: ResidentTypeNonResident}); len(typed) != 1 {
		t.Fatalf("resident type filter failed: %+v", typed)
	}

	// 删除
	row = findByName(mustList(t, svcCtx, types.ListRequest{}), "张三")
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
		t.Fatalf("delete tax profile: %v", err)
	}
	if got := len(mustList(t, svcCtx, types.ListRequest{})); got != 0 {
		t.Fatalf("expected 0 profiles after delete, got %d", got)
	}
}
