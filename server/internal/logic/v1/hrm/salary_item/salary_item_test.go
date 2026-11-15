package salary_item

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_item"
)

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{Model: env.Model}
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmSalaryItem {
	t.Helper()
	req.Current = 1
	req.Size = 100
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list salary items: %v", err)
	}
	return resp.Records
}

func findByCode(records []types.HrmSalaryItem, code string) *types.HrmSalaryItem {
	for i := range records {
		if records[i].ItemCode == code {
			return &records[i]
		}
	}
	return nil
}

// TestSalaryItemPresets verifies the 16 预置薪资项目 seeded by migration 25
// (应发 6 → 应扣 6 含实发 → 成本 3), their 项目类型/取值方式 and 计算顺序 ordering.
func TestSalaryItemPresets(t *testing.T) {
	svcCtx := newCtx(t)

	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 16 {
		t.Fatalf("expected 16 preset salary items, got %d: %+v", len(records), records)
	}

	wantOrder := []string{
		"BASIC_SALARY", "POST_SALARY", "MERIT_SALARY", "FIXED_ALLOWANCE", "HEAT_ALLOWANCE", "REWARD_BONUS",
		"ABSENCE_DEDUCTION", "SI_PERSONAL", "HF_PERSONAL", "TAX", "REWARD_PENALTY", "OTHER_DEDUCTION",
		"NET_SALARY",
		"SI_COMPANY", "HF_COMPANY", "LABOR_COST",
	}
	for i, code := range wantOrder {
		if records[i].ItemCode != code {
			t.Fatalf("preset calc order: records[%d].ItemCode = %q, want %q", i, records[i].ItemCode, code)
		}
		if records[i].IsPreset != 1 {
			t.Fatalf("preset %s must have is_preset=1", code)
		}
	}

	basic := findByCode(records, "BASIC_SALARY")
	if basic == nil || basic.ItemName != "基本工资" || basic.ItemType != "fixed" ||
		basic.ItemClass != "earning" || basic.ValueMode != "manual" || basic.CalcOrder != 10 {
		t.Fatalf("unexpected 基本工资 preset: %+v", basic)
	}
	merit := findByCode(records, "MERIT_SALARY")
	if merit == nil || merit.ItemType != "data" || merit.ValueMode != "system" || merit.ItemClass != "earning" {
		t.Fatalf("unexpected 绩效工资 preset (取数项/系统取数/应发): %+v", merit)
	}
	tax := findByCode(records, "TAX")
	if tax == nil || tax.ItemType != "formula" || tax.ValueMode != "formula" || tax.ItemClass != "deduction" {
		t.Fatalf("unexpected 个税 preset (公式项/公式计算/应扣): %+v", tax)
	}
	net := findByCode(records, "NET_SALARY")
	if net == nil || net.ItemClass != "earning" || net.Formula == "" {
		t.Fatalf("unexpected 实发工资 preset: %+v", net)
	}
	labor := findByCode(records, "LABOR_COST")
	if labor == nil || labor.ItemClass != "cost" || labor.ItemType != "formula" {
		t.Fatalf("unexpected 人工成本 preset: %+v", labor)
	}

	// 项目分类筛选: cost 类恰为公司社保/公司公积金/人工成本
	if costs := mustList(t, svcCtx, types.ListRequest{ItemClass: "cost"}); len(costs) != 3 {
		t.Fatalf("expected 3 cost presets, got %+v", costs)
	}
}

// TestSalaryItemCrud walks the CRUD seam: 新增/编辑/删除自定义项 + 编码唯一性
//（新增重复、编辑成别人的编码都不行）+ 名称筛选.
func TestSalaryItemCrud(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	// 新增自定义薪资项目
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		ItemCode: "OTHER_BONUS", ItemName: "其他奖金", ItemType: "fixed", ItemClass: "earning",
		ValueMode: "manual", CalcOrder: 65, OrderNum: 20,
	}); err != nil {
		t.Fatalf("add salary item: %v", err)
	}
	if records := mustList(t, svcCtx, types.ListRequest{}); len(records) != 17 {
		t.Fatalf("expected 17 items after add, got %d", len(records))
	}

	// 编码唯一
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		ItemCode: "BASIC_SALARY", ItemName: "重复基本工资", ItemType: "fixed", ItemClass: "earning", ValueMode: "manual",
	}); err == nil {
		t.Fatalf("duplicate item code should fail")
	}

	// 编码非法（小写/横线/空）
	for _, bad := range []string{"other_bonus", "other-bonus", "其他奖金", ""} {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
			ItemCode: bad, ItemName: "非法编码", ItemType: "fixed", ItemClass: "earning", ValueMode: "manual",
		}); err == nil {
			t.Fatalf("invalid item code %q should fail", bad)
		}
	}

	// 编辑: 调整计算顺序/显示顺序/启停
	bonus := findByCode(mustList(t, svcCtx, types.ListRequest{}), "OTHER_BONUS")
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: bonus.Uuid, ItemCode: "OTHER_BONUS", ItemName: "其他奖金", ItemType: "fixed",
		ItemClass: "earning", ValueMode: "manual", CalcOrder: 45, OrderNum: 21, Status: "0",
	}); err != nil {
		t.Fatalf("edit salary item: %v", err)
	}
	bonus = findByCode(mustList(t, svcCtx, types.ListRequest{}), "OTHER_BONUS")
	if bonus.CalcOrder != 45 || bonus.OrderNum != 21 || bonus.Status != "0" {
		t.Fatalf("edit not applied: %+v", bonus)
	}

	// 编辑不存在的项目
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: "no-such-uuid", ItemCode: "GHOST_ITEM", ItemName: "幽灵项",
		ItemType: "fixed", ItemClass: "earning", ValueMode: "manual", Status: "1",
	}); err == nil {
		t.Fatalf("editing unknown salary item should fail")
	}

	// 编辑成别人的编码也不行 (唯一性)
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: bonus.Uuid, ItemCode: "BASIC_SALARY", ItemName: "其他奖金",
		ItemType: "fixed", ItemClass: "earning", ValueMode: "manual", Status: "1",
	}); err == nil {
		t.Fatalf("editing item code onto an existing one should fail")
	}

	// 名称筛选
	if named := mustList(t, svcCtx, types.ListRequest{ItemName: "其他奖金"}); len(named) != 1 || named[0].ItemCode != "OTHER_BONUS" {
		t.Fatalf("name filter failed: %+v", named)
	}

	// 删除自定义项目 OK; 预置项目保留
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{bonus.Uuid}}); err != nil {
		t.Fatalf("delete custom salary item: %v", err)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 16 || findByCode(records, "OTHER_BONUS") != nil {
		t.Fatalf("expected presets only after delete, got %d records", len(records))
	}
}

// TestSalaryItemEnumAndFormulaValidation 枚举守卫: 项目类型/项目分类/取值方式/启停非法值
// 均拒绝；公式项必须带公式表达式；取值方式 formula ⇒ 项目类型必须为 formula；非公式项公式清空.
func TestSalaryItemEnumAndFormulaValidation(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	base := types.AddRequest{ItemName: "校验用", ItemType: "fixed", ItemClass: "earning", ValueMode: "manual"}

	invalid := []types.AddRequest{
		// 项目类型非法
		func(r types.AddRequest) types.AddRequest { r.ItemType = "other"; return r }(base),
		// 项目分类非法
		func(r types.AddRequest) types.AddRequest { r.ItemClass = "other"; return r }(base),
		// 取值方式非法
		func(r types.AddRequest) types.AddRequest { r.ValueMode = "other"; return r }(base),
		// 状态非法
		func(r types.AddRequest) types.AddRequest { r.Status = "2"; return r }(base),
		// 名称空
		func(r types.AddRequest) types.AddRequest { r.ItemName = ""; return r }(base),
		// value_mode=formula 但项目类型不是公式项
		func(r types.AddRequest) types.AddRequest { r.ValueMode = "formula"; return r }(base),
		// 公式项但没有公式表达式
		func(r types.AddRequest) types.AddRequest { r.ItemType = "formula"; r.ValueMode = "formula"; return r }(base),
	}
	for i, req := range invalid {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&req); err == nil {
			t.Fatalf("invalid case %d should fail: %+v", i, req)
		}
	}

	// 合法公式项: 公式落库
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		ItemCode: "GROSS_TOTAL", ItemName: "应发合计", ItemType: "formula", ItemClass: "earning",
		ValueMode: "formula", CalcOrder: 200, Formula: "基本工资+岗位工资",
	}); err != nil {
		t.Fatalf("valid formula item should pass: %v", err)
	}
	gross := findByCode(mustList(t, svcCtx, types.ListRequest{}), "GROSS_TOTAL")
	if gross.Formula != "基本工资+岗位工资" {
		t.Fatalf("formula not stored: %+v", gross)
	}

	// 非公式项的公式被清空
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		ItemCode: "CLEANUP_ITEM", ItemName: "清理项", ItemType: "fixed", ItemClass: "deduction",
		ValueMode: "manual", Formula: "不应存储",
	}); err != nil {
		t.Fatalf("fixed item with formula should pass but clear formula: %v", err)
	}
	cleanup := findByCode(mustList(t, svcCtx, types.ListRequest{}), "CLEANUP_ITEM")
	if cleanup.Formula != "" {
		t.Fatalf("non-formula item formula must be cleared: %+v", cleanup)
	}
}

// TestSalaryItemPresetDeleteGuard 预置项目删除守卫: 单删/混删（含一个预置项的批量）均拒绝，
// 预置项在拒绝后仍然存在；如需下线请停用（编辑 status=0）.
func TestSalaryItemPresetDeleteGuard(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	// 单删预置项
	basic := findByCode(mustList(t, svcCtx, types.ListRequest{}), "BASIC_SALARY")
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{basic.Uuid}}); err == nil {
		t.Fatalf("deleting preset item should fail")
	}
	if findByCode(mustList(t, svcCtx, types.ListRequest{}), "BASIC_SALARY") == nil {
		t.Fatalf("preset item must survive failed delete")
	}

	// 混合批删（自定义 + 预置）整体拒绝
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		ItemCode: "TEMP_ITEM", ItemName: "临时项", ItemType: "fixed", ItemClass: "earning", ValueMode: "manual",
	}); err != nil {
		t.Fatalf("add temp item: %v", err)
	}
	temp := findByCode(mustList(t, svcCtx, types.ListRequest{}), "TEMP_ITEM")
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{temp.Uuid, basic.Uuid}}); err == nil {
		t.Fatalf("mixed batch delete with preset should fail")
	}
	if findByCode(mustList(t, svcCtx, types.ListRequest{}), "TEMP_ITEM") == nil ||
		findByCode(mustList(t, svcCtx, types.ListRequest{}), "BASIC_SALARY") == nil {
		t.Fatalf("mixed delete must be all-or-nothing")
	}

	// 预置项可以停用（下线口径）
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: basic.Uuid, ItemCode: "BASIC_SALARY", ItemName: "基本工资", ItemType: "fixed",
		ItemClass: "earning", ValueMode: "manual", CalcOrder: 10, OrderNum: 1, Status: "0",
	}); err != nil {
		t.Fatalf("disable preset item: %v", err)
	}
	basic = findByCode(mustList(t, svcCtx, types.ListRequest{}), "BASIC_SALARY")
	if basic.Status != "0" || basic.IsPreset != 1 {
		t.Fatalf("preset disable not applied: %+v", basic)
	}
}

// TestSalaryItemCalcOrderConstraint 计算顺序约束: 列表按 calc_order 升序、同序按 id 升序；
// 新增可插入到任意顺序位置（0 默认排最前），试算引擎按列表顺序消费.
func TestSalaryItemCalcOrderConstraint(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	// 三个相同 calc_order 的自定义项 → 按 id 升序稳定输出
	sameOrder := []string{"SAME_ORDER_A", "SAME_ORDER_B", "SAME_ORDER_C"}
	for i, code := range sameOrder {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
			ItemCode: code, ItemName: "同序" + string(rune('甲'+i)), ItemType: "fixed",
			ItemClass: "deduction", ValueMode: "manual", CalcOrder: 55,
		}); err != nil {
			t.Fatalf("add same-order item: %v", err)
		}
	}

	records := mustList(t, svcCtx, types.ListRequest{})
	prevOrder := int64(-1)
	for i, r := range records {
		if r.CalcOrder < prevOrder {
			t.Fatalf("calc order not ascending at %d: %+v", i, records)
		}
		prevOrder = r.CalcOrder
	}

	// 全表顺序: 应发(10..50) → 55 插入段 → 60 (REWARD_BONUS) → 应扣(70..120) → 实发(130) → 成本(140..160)
	// 即 55 段三个自定义项必须落在 50 (HEAT_ALLOWANCE) 与 60 (REWARD_BONUS) 之间.
	idxHeat, idxFirst55, idxLast55, idxBonus := -1, -1, -1, -1
	for i, r := range records {
		switch {
		case r.ItemCode == "HEAT_ALLOWANCE":
			idxHeat = i
		case r.CalcOrder == 55:
			if idxFirst55 == -1 {
				idxFirst55 = i
			}
			idxLast55 = i
		case r.ItemCode == "REWARD_BONUS":
			idxBonus = i
		}
	}
	if !(idxHeat < idxFirst55 && idxLast55 < idxBonus) {
		t.Fatalf("calc_order=55 items must sit between calc 50 and 60: heat=%d first55=%d last55=%d bonus=%d",
			idxHeat, idxFirst55, idxLast55, idxBonus)
	}
}
