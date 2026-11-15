package leavetype

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavetype"
)

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{Model: env.Model}
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmLeaveType {
	t.Helper()
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list leave types: %v", err)
	}
	return resp.Records
}

func findByCode(records []types.HrmLeaveType, code string) *types.HrmLeaveType {
	for i := range records {
		if records[i].TypeCode == code {
			return &records[i]
		}
	}
	return nil
}

// TestLeaveTypePresets verifies the 预置假期类型 seeded by migration 13
// (年假/调休假/事假/病假/婚假/产假/陪产假/丧假/其他), including 余额管理/需要附件/
// 默认天数/排序 defaults, and the order_num-based listing.
func TestLeaveTypePresets(t *testing.T) {
	svcCtx := newCtx(t)

	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 9 {
		t.Fatalf("expected 9 preset leave types, got %d: %+v", len(records), records)
	}

	wantOrder := []string{
		"annual_leave", "comp_leave", "personal_leave", "sick_leave", "marriage_leave",
		"maternity_leave", "paternity_leave", "bereavement_leave", "other_leave",
	}
	for i, code := range wantOrder {
		if records[i].TypeCode != code {
			t.Fatalf("preset order: records[%d].TypeCode = %q, want %q", i, records[i].TypeCode, code)
		}
	}

	annual := findByCode(records, "annual_leave")
	if annual == nil || annual.TypeName != "年假" || annual.BalanceEnabled != 1 ||
		annual.AttachmentRequired != 0 || annual.DefaultDays != 5 || annual.TypeValue != 1 || annual.Status != "1" {
		t.Fatalf("unexpected 年假 preset: %+v", annual)
	}
	comp := findByCode(records, "comp_leave")
	if comp == nil || comp.TypeName != "调休假" || comp.BalanceEnabled != 1 {
		t.Fatalf("unexpected 调休假 preset: %+v", comp)
	}
	sick := findByCode(records, "sick_leave")
	if sick == nil || sick.AttachmentRequired != 1 || sick.BalanceEnabled != 0 {
		t.Fatalf("unexpected 病假 preset (需要附件): %+v", sick)
	}
	maternity := findByCode(records, "maternity_leave")
	if maternity == nil || maternity.DefaultDays != 98 {
		t.Fatalf("unexpected 产假 preset: %+v", maternity)
	}
	paternity := findByCode(records, "paternity_leave")
	if paternity == nil || paternity.DefaultDays != 15 {
		t.Fatalf("unexpected 陪产假 preset: %+v", paternity)
	}
}

// TestLeaveTypeCrud walks the CRUD seam: 新增/编辑/删除 + 编码唯一性 + 名称筛选.
func TestLeaveTypeCrud(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	// 新增自定义假期类型
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		TypeCode: "welfare_leave", TypeName: "福利假", BalanceEnabled: 1,
		DefaultDays: 2, OrderNum: 10,
	}); err != nil {
		t.Fatalf("add leave type: %v", err)
	}
	if records := mustList(t, svcCtx, types.ListRequest{}); len(records) != 10 {
		t.Fatalf("expected 10 leave types after add, got %d", len(records))
	}

	// 类型编码唯一
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		TypeCode: "annual_leave", TypeName: "重复年假",
	}); err == nil {
		t.Fatalf("duplicate type code should fail")
	}

	// 类型编码非法 (大写/横线/中文)
	for _, bad := range []string{"Welfare", "welfare-leave", "福利假", ""} {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
			TypeCode: bad, TypeName: "非法编码",
		}); err == nil {
			t.Fatalf("invalid type code %q should fail", bad)
		}
	}

	// 编辑: 调整默认天数/余额开关/排序
	annual := findByCode(mustList(t, svcCtx, types.ListRequest{}), "annual_leave")
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: annual.Uuid, TypeCode: "annual_leave", TypeName: "年假",
		BalanceEnabled: 0, DefaultDays: 10, OrderNum: 1, Status: "1",
	}); err != nil {
		t.Fatalf("edit leave type: %v", err)
	}
	annual = findByCode(mustList(t, svcCtx, types.ListRequest{}), "annual_leave")
	if annual.BalanceEnabled != 0 || annual.DefaultDays != 10 {
		t.Fatalf("edit not applied: %+v", annual)
	}

	// 编辑不存在的假期类型
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: "no-such-uuid", TypeCode: "ghost", TypeName: "幽灵假", Status: "1",
	}); err == nil {
		t.Fatalf("editing unknown leave type should fail")
	}

	// 编辑成别人的编码也不行 (唯一性)
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: annual.Uuid, TypeCode: "welfare_leave", TypeName: "年假", Status: "1",
	}); err == nil {
		t.Fatalf("editing type code onto an existing one should fail")
	}

	// 名称筛选
	if named := mustList(t, svcCtx, types.ListRequest{TypeName: "福利假"}); len(named) != 1 || named[0].TypeCode != "welfare_leave" {
		t.Fatalf("name filter failed: %+v", named)
	}

	// 删除自定义类型; 预置类型保留
	welfare := findByCode(mustList(t, svcCtx, types.ListRequest{}), "welfare_leave")
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{welfare.Uuid}}); err != nil {
		t.Fatalf("delete leave type: %v", err)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 9 || findByCode(records, "welfare_leave") != nil {
		t.Fatalf("expected presets only after delete, got %+v", records)
	}
}
