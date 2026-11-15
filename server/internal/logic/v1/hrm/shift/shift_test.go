package shift

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/shift"
)

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{Model: env.Model}
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmShift {
	t.Helper()
	req.Current = 1
	req.Size = 100
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list shifts: %v", err)
	}
	return resp.Records
}

func findByName(records []types.HrmShift, name string) *types.HrmShift {
	for i := range records {
		if records[i].ShiftName == name {
			return &records[i]
		}
	}
	return nil
}

// TestShiftCrudAndCrossDay walks the primary user story 38: 维护班次（上下班时间
// 支持跨天、休息时段、标准工时、弹性分钟数）. Cross-day shifts (下班时间早于上班
// 时间) and 标准工时 (= 上下班时长－休息时长) are computed by the logic layer.
func TestShiftCrudAndCrossDay(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	// 白班 09:00-18:00 午休 1h → 标准工时 8h, 不跨天
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		ShiftName: "标准班", StartTime: "09:00", EndTime: "18:00",
		RestStart: "12:00", RestEnd: "13:00", FlexibleMinutes: 30,
	}); err != nil {
		t.Fatalf("add day shift: %v", err)
	}

	// 夜班 22:00-次日06:00 休息 00:00-00:30 → 跨天, 标准工时 7.5h
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		ShiftName: "夜班", StartTime: "22:00", EndTime: "06:00",
		RestStart: "00:00", RestEnd: "00:30",
	}); err != nil {
		t.Fatalf("add overnight shift: %v", err)
	}

	// 通宵班 23:30-07:30 无休息 → 跨天, 标准工时 8h
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		ShiftName: "通宵班", StartTime: "23:30", EndTime: "07:30",
	}); err != nil {
		t.Fatalf("add overnight shift without rest: %v", err)
	}

	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 3 {
		t.Fatalf("expected 3 shifts, got %d: %+v", len(records), records)
	}

	dayShift := findByName(records, "标准班")
	if dayShift == nil || dayShift.CrossDay != 0 || dayShift.StandardHours != 8 || dayShift.FlexibleMinutes != 30 {
		t.Fatalf("unexpected 标准班: %+v", dayShift)
	}
	nightShift := findByName(records, "夜班")
	if nightShift == nil || nightShift.CrossDay != 1 || nightShift.StandardHours != 7.5 {
		t.Fatalf("unexpected 夜班 (cross-day/rest math): %+v", nightShift)
	}
	allNight := findByName(records, "通宵班")
	if allNight == nil || allNight.CrossDay != 1 || allNight.StandardHours != 8 {
		t.Fatalf("unexpected 通宵班: %+v", allNight)
	}

	// 校验: 上下班时间格式非法 (不是 HH:mm / 越界)
	for _, bad := range []struct{ start, end string }{
		{"9:00", "18:00"},  // 缺前导零
		{"25:00", "18:00"}, // 小时越界
		{"09:60", "18:00"}, // 分钟越界
		{"0900", "18:00"},  // 缺冒号
		{"09:00", "09:00"}, // 上下班时间相同
	} {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
			ShiftName: "非法班次", StartTime: bad.start, EndTime: bad.end,
		}); err == nil {
			t.Fatalf("add shift with times %q-%q should fail", bad.start, bad.end)
		}
	}

	// 校验: 休息时段必须成对且开始早于结束
	for _, badRest := range []struct{ start, end string }{
		{"12:00", ""},      // 只填开始
		{"", "13:00"},      // 只填结束
		{"13:00", "12:00"}, // 开始晚于结束
		{"12:00", "12:00"}, // 相等
	} {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
			ShiftName: "非法休息", StartTime: "09:00", EndTime: "18:00",
			RestStart: badRest.start, RestEnd: badRest.end,
		}); err == nil {
			t.Fatalf("add shift with rest %q-%q should fail", badRest.start, badRest.end)
		}
	}

	if records := mustList(t, svcCtx, types.ListRequest{}); len(records) != 3 {
		t.Fatalf("invalid shifts must not be persisted, got %d", len(records))
	}

	// 编辑: 夜班改为 20:00-次日04:00 无休息 → 仍跨天, 标准工时 8h
	edited, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: nightShift.Uuid, ShiftName: "夜班(调整)", StartTime: "20:00", EndTime: "04:00",
		Status: "1",
	})
	if err != nil {
		t.Fatalf("edit shift: %v", err)
	}
	_ = edited
	night := findByName(mustList(t, svcCtx, types.ListRequest{}), "夜班(调整)")
	if night == nil || night.CrossDay != 1 || night.StandardHours != 8 || night.RestStart != "" {
		t.Fatalf("edit not recomputed: %+v", night)
	}

	// 编辑不存在的班次
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: "no-such-uuid", ShiftName: "幽灵班", StartTime: "09:00", EndTime: "18:00", Status: "1",
	}); err == nil {
		t.Fatalf("editing unknown shift should fail")
	}

	// 状态筛选: 停用通宵班后只查启用
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: allNight.Uuid, ShiftName: "通宵班", StartTime: "23:30", EndTime: "07:30", Status: "0",
	}); err != nil {
		t.Fatalf("disable shift: %v", err)
	}
	if enabled := mustList(t, svcCtx, types.ListRequest{Status: "1"}); len(enabled) != 2 || findByName(enabled, "通宵班") != nil {
		t.Fatalf("status filter failed: %+v", enabled)
	}
	if named := mustList(t, svcCtx, types.ListRequest{ShiftName: "夜班"}); len(named) != 1 {
		t.Fatalf("name filter failed: %+v", named)
	}

	// 删除/批量删除
	del := mustList(t, svcCtx, types.ListRequest{})
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{del[0].Uuid, del[1].Uuid}}); err != nil {
		t.Fatalf("batch delete: %v", err)
	}
	if left := mustList(t, svcCtx, types.ListRequest{}); len(left) != 1 {
		t.Fatalf("expected 1 shift after batch delete, got %d", len(left))
	}
}

// TestShiftPureFuncs covers the time math helpers directly (辅接缝: 纯函数单测).
func TestShiftPureFuncs(t *testing.T) {
	if err := validateShiftTimes("09:00", "18:00"); err != nil {
		t.Errorf("valid times rejected: %v", err)
	}
	if err := validateShiftTimes("22:00", "06:00"); err != nil {
		t.Errorf("cross-day times rejected: %v", err)
	}
	for _, bad := range [][2]string{{"9:00", "18:00"}, {"24:00", "18:00"}, {"09:00", "09:00"}} {
		if err := validateShiftTimes(bad[0], bad[1]); err == nil {
			t.Errorf("times %v should be invalid", bad)
		}
	}

	if v := computeCrossDay("09:00", "18:00"); v != 0 {
		t.Errorf("computeCrossDay(09:00,18:00) = %d, want 0", v)
	}
	if v := computeCrossDay("22:00", "06:00"); v != 1 {
		t.Errorf("computeCrossDay(22:00,06:00) = %d, want 1", v)
	}

	cases := []struct {
		start, end, restStart, restEnd string
		want                           float64
	}{
		{"09:00", "18:00", "12:00", "13:00", 8},
		{"09:00", "18:00", "", "", 9},
		{"22:00", "06:00", "00:00", "00:30", 7.5},
		{"23:30", "07:30", "", "", 8},
	}
	for _, c := range cases {
		if got := computeStandardHours(c.start, c.end, c.restStart, c.restEnd); got != c.want {
			t.Errorf("computeStandardHours(%q,%q,%q,%q) = %v, want %v", c.start, c.end, c.restStart, c.restEnd, got, c.want)
		}
	}
}
