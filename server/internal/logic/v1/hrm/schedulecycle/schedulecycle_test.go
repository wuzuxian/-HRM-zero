package schedulecycle

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero/core/stores/condition"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	shiftlogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/shift"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_cycle"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_shift"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	shifttypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/shift"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/schedulecycle"
)

func newSvcCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

// createShift 通过 shift logic 建一个班次, 返回 uuid.
func createShift(t *testing.T, ctx context.Context, svcCtx *svc.ServiceContext, name, status string) string {
	t.Helper()
	req := &shifttypes.AddRequest{
		ShiftName: name,
		StartTime: "09:00",
		EndTime:   "18:00",
		Status:    status,
	}
	if _, err := shiftlogic.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add shift %s: %v", name, err)
	}
	row, err := svcCtx.Model.HrmShift.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_shift.ShiftName, name).Build()...)
	if err != nil {
		t.Fatalf("find shift %s: %v", name, err)
	}
	return row.Uuid
}

func addReq(name string, shiftUuids ...string) *types.AddRequest {
	return &types.AddRequest{
		CycleName:  name,
		ShiftUuids: shiftUuids,
		Status:     "1",
	}
}

// TestAddValidation 覆盖 Add 入参校验: 名称/空序列/不存在班次/停用班次/空引用/非法状态/不存在考勤组.
func TestAddValidation(t *testing.T) {
	ctx := context.Background()
	svcCtx := newSvcCtx(t)

	onUuid := createShift(t, ctx, svcCtx, "白班", "1")
	offUuid := createShift(t, ctx, svcCtx, "停用班", "0")

	cases := []struct {
		name string
		mut  func(r *types.AddRequest)
	}{
		{"empty name", func(r *types.AddRequest) { r.CycleName = "" }},
		{"empty shift seq", func(r *types.AddRequest) { r.ShiftUuids = nil }},
		{"missing shift", func(r *types.AddRequest) { r.ShiftUuids = []string{onUuid, "no-such-shift"} }},
		{"disabled shift", func(r *types.AddRequest) { r.ShiftUuids = []string{onUuid, offUuid} }},
		{"empty shift ref", func(r *types.AddRequest) { r.ShiftUuids = []string{onUuid, ""} }},
		{"bad status", func(r *types.AddRequest) { r.Status = "2" }},
		{"missing group", func(r *types.AddRequest) { r.GroupUuids = []string{"no-such-group"} }},
		{"empty group ref", func(r *types.AddRequest) { r.GroupUuids = []string{""} }},
	}
	for _, tc := range cases {
		req := addReq("周期-"+tc.name, onUuid)
		tc.mut(req)
		if _, err := NewAdd(ctx, svcCtx, nil).Add(req); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}

	// valid request inserts
	if _, err := NewAdd(ctx, svcCtx, nil).Add(addReq("四班三运转", onUuid, onUuid)); err != nil {
		t.Fatalf("valid add: %v", err)
	}
	row, err := svcCtx.Model.HrmScheduleCycle.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_schedule_cycle.CycleName, "四班三运转").Build()...)
	if err != nil {
		t.Fatalf("find inserted: %v", err)
	}
	if row.Status != "1" {
		t.Errorf("status default = %q, want 1", row.Status)
	}
	// 有序序列落库为 CSV, 顺序保持
	if _, err := NewAdd(ctx, svcCtx, nil).Add(addReq("序列顺序", onUuid, onUuid, onUuid)); err != nil {
		t.Fatalf("seq add: %v", err)
	}
	seqRow, err := svcCtx.Model.HrmScheduleCycle.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_schedule_cycle.CycleName, "序列顺序").Build()...)
	if err != nil {
		t.Fatalf("find seq row: %v", err)
	}
	if seqRow.ShiftUuids != onUuid+","+onUuid+","+onUuid {
		t.Errorf("shift_uuids = %q", seqRow.ShiftUuids)
	}
}

// TestEditDeleteToggle CRUD 轮转: edit 覆盖字段/toggle 启停/delete 批删/编辑不存在报错.
func TestEditDeleteToggle(t *testing.T) {
	ctx := context.Background()
	svcCtx := newSvcCtx(t)

	shiftA := createShift(t, ctx, svcCtx, "早班", "1")
	shiftB := createShift(t, ctx, svcCtx, "夜班", "1")

	if _, err := NewAdd(ctx, svcCtx, nil).Add(addReq("周期A", shiftA)); err != nil {
		t.Fatalf("add: %v", err)
	}
	row, err := svcCtx.Model.HrmScheduleCycle.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_schedule_cycle.CycleName, "周期A").Build()...)
	if err != nil {
		t.Fatalf("find: %v", err)
	}

	// edit not-found
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{Uuid: "missing", CycleName: "x", ShiftUuids: []string{shiftA}, Status: "1"}); err == nil {
		t.Error("edit missing: expected error")
	}

	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid:       row.Uuid,
		CycleName:  "周期A-改",
		ShiftUuids: []string{shiftA, shiftB},
		GroupUuids: []string{},
		Status:     "0",
		Remark:     "remark",
	}); err != nil {
		t.Fatalf("edit: %v", err)
	}
	edited, err := svcCtx.Model.HrmScheduleCycle.FindOneByUuid(ctx, nil, row.Uuid)
	if err != nil {
		t.Fatalf("find edited: %v", err)
	}
	if edited.CycleName != "周期A-改" || edited.Status != "0" || edited.ShiftUuids != shiftA+","+shiftB {
		t.Errorf("edit not persisted: %+v", edited)
	}

	// toggle back on
	if _, err := NewToggleStatus(ctx, svcCtx, nil).ToggleStatus(&types.ToggleStatusRequest{Uuid: row.Uuid, Status: "1"}); err != nil {
		t.Fatalf("toggle: %v", err)
	}
	toggled, err := svcCtx.Model.HrmScheduleCycle.FindOneByUuid(ctx, nil, row.Uuid)
	if err != nil {
		t.Fatalf("find toggled: %v", err)
	}
	if toggled.Status != "1" {
		t.Errorf("toggle not persisted: %q", toggled.Status)
	}
	// toggle invalid status
	if _, err := NewToggleStatus(ctx, svcCtx, nil).ToggleStatus(&types.ToggleStatusRequest{Uuid: row.Uuid, Status: "2"}); err == nil {
		t.Error("toggle invalid status: expected error")
	}

	// delete
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svcCtx.Model.HrmScheduleCycle.FindOneByUuid(ctx, nil, row.Uuid); err != hrm_schedule_cycle.ErrNotFound {
		t.Errorf("deleted row find err = %v, want ErrNotFound", err)
	}
}

// TestListFilter 分页列表 + 名称/状态筛选.
func TestListFilter(t *testing.T) {
	ctx := context.Background()
	svcCtx := newSvcCtx(t)

	shift := createShift(t, ctx, svcCtx, "中班", "1")
	for _, name := range []string{"列表周期1", "列表周期2", "特殊周期"} {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(addReq(name, shift)); err != nil {
			t.Fatalf("add %s: %v", name, err)
		}
	}

	// page all
	resp, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.Total != 3 || len(resp.Records) != 3 {
		t.Errorf("total=%d records=%d, want 3/3", resp.Total, len(resp.Records))
	}
	// record shape: uuids split back into slices
	rec := resp.Records[0]
	if len(rec.ShiftUuids) != 1 || rec.ShiftUuids[0] != shift {
		t.Errorf("record shiftUuids = %v", rec.ShiftUuids)
	}
	if rec.CreateTime == "" || rec.UpdateTime == "" {
		t.Errorf("record times empty: %+v", rec)
	}

	// filter by name
	resp, err = NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, CycleName: "特殊"})
	if err != nil {
		t.Fatalf("list filter: %v", err)
	}
	if resp.Total != 1 || resp.Records[0].CycleName != "特殊周期" {
		t.Errorf("name filter: total=%d", resp.Total)
	}

	// filter by status
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: resp.Records[0].Uuid, CycleName: "特殊周期", ShiftUuids: []string{shift}, Status: "0",
	}); err != nil {
		t.Fatalf("disable: %v", err)
	}
	resp, err = NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, Status: "0"})
	if err != nil {
		t.Fatalf("list status filter: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("status filter total = %d, want 1", resp.Total)
	}
}

// TestExpandCycleWithCycleRow Add→落库→splitCsv 还原有序序列→ExpandCycle 展开为逐日映射
// （HRM-15 将按此管线消费: 读行 → 还原序列 → ExpandCycle）.
func TestExpandCycleWithCycleRow(t *testing.T) {
	ctx := context.Background()
	svcCtx := newSvcCtx(t)

	shiftA := createShift(t, ctx, svcCtx, "甲班", "1")
	shiftB := createShift(t, ctx, svcCtx, "乙班", "1")
	if _, err := NewAdd(ctx, svcCtx, nil).Add(addReq("两班倒", shiftA, shiftB)); err != nil {
		t.Fatalf("add: %v", err)
	}
	row, err := svcCtx.Model.HrmScheduleCycle.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_schedule_cycle.CycleName, "两班倒").Build()...)
	if err != nil {
		t.Fatalf("find: %v", err)
	}

	// 序列引用还原（uuid 场景走 splitCsv; ExpandCycle 演示用数值 id 模拟相位）
	seq := splitCsv(row.ShiftUuids)
	if len(seq) != 2 || seq[0] != shiftA || seq[1] != shiftB {
		t.Fatalf("splitCsv order broken: %v", seq)
	}
	// 周期天数 = 序列长度（派生）
	if len(seq) != 2 {
		t.Errorf("cycle days = %d, want 2", len(seq))
	}
	// 相位展开: 第 1/2/3 天 → A/B/A
	ids := []int64{1, 2}
	expanded := ExpandCycle(ids, date(2026, 3, 1), date(2026, 3, 3))
	if expanded[key(2026, 3, 1)] != 1 || expanded[key(2026, 3, 2)] != 2 || expanded[key(2026, 3, 3)] != 1 {
		t.Errorf("expanded = %+v", expanded)
	}
}
