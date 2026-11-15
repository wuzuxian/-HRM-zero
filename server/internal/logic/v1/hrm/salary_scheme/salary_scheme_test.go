package salary_scheme

import (
	"context"
	"testing"

	"github.com/google/uuid"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

// seedSubject 直接插薪酬主体行（方案必须关联已有主体）.
func seedSubject(t *testing.T, svcCtx *svc.ServiceContext, name string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalarySubject.Insert(t.Context(), nil, &hrm_salary_subject.HrmSalarySubject{
		Uuid: u, SubjectName: name, OrgUuid: uuid.New().String(), OrgName: "甲公司", Status: "1",
	}); err != nil {
		t.Fatalf("seed subject %s: %v", name, err)
	}
	return u
}

// seedItem 直接插薪资项目行（is_preset=1 模拟预置项；item_code 全局唯一，code 须避开迁移种子）.
func seedItem(t *testing.T, svcCtx *svc.ServiceContext, code, name string, itemType, itemClass, valueMode string, calcOrder int64) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalaryItem.Insert(t.Context(), nil, &hrm_salary_item.HrmSalaryItem{
		Uuid: u, ItemCode: code, ItemName: name, ItemType: itemType, ItemClass: itemClass,
		ValueMode: valueMode, CalcOrder: calcOrder, OrderNum: calcOrder, IsPreset: 1, Status: "1",
	}); err != nil {
		t.Fatalf("seed item %s: %v", code, err)
	}
	return u
}

// findPresetItem 按编码取迁移种子预置项目（BASIC_SALARY/TAX 等，见 25_salary_setting.up.sql）.
func findPresetItem(t *testing.T, svcCtx *svc.ServiceContext, code string) string {
	t.Helper()
	row, err := svcCtx.Model.HrmSalaryItem.FindOneByItemCode(t.Context(), nil, code)
	if err != nil {
		t.Fatalf("find preset item %s: %v", code, err)
	}
	return row.Uuid
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmSalaryScheme {
	t.Helper()
	req.Current = 1
	req.Size = 100
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list salary schemes: %v", err)
	}
	return resp.Records
}

func findByName(records []types.HrmSalaryScheme, name string) *types.HrmSalaryScheme {
	for i := range records {
		if records[i].SchemeName == name {
			return &records[i]
		}
	}
	return nil
}

// TestSalarySchemeCrud walks the CRUD seam (HRM-34 user story 56):
// 新增（主体名称快照/人员状态多选）→ 编辑（换主体/启停）→ 删除, plus 名称筛选.
func TestSalarySchemeCrud(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	subjectA := seedSubject(t, svcCtx, "甲公司主体")
	subjectB := seedSubject(t, svcCtx, "乙公司主体")

	// 新增: 主体名称快照 + person_status 多选落库
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		SchemeName: "月度标准方案", SubjectUuid: subjectA,
		PersonStatus: []string{"1", "2"}, SettleCycle: "monthly", PayDay: 10, Currency: "CNY",
	}); err != nil {
		t.Fatalf("add salary scheme: %v", err)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 1 {
		t.Fatalf("expected 1 scheme after add, got %d", len(records))
	}
	if records[0].SubjectName != "甲公司主体" || records[0].PersonStatus != "1,2" ||
		records[0].PayDay != 10 || records[0].Currency != "CNY" || records[0].SettleCycle != "monthly" {
		t.Fatalf("add snapshot/defaults not applied: %+v", records[0])
	}

	// 名称筛选
	if named := mustList(t, svcCtx, types.ListRequest{SchemeName: "月度标准"}); len(named) != 1 {
		t.Fatalf("name filter failed: %+v", named)
	}

	// 编辑: 换主体 + 停用 + 人员状态变更
	row := findByName(mustList(t, svcCtx, types.ListRequest{}), "月度标准方案")
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: row.Uuid, SchemeName: "月度标准方案", SubjectUuid: subjectB,
		PersonStatus: []string{"1", "3", "6"}, SettleCycle: "monthly", PayDay: 15, Currency: "USD", Status: "0",
	}); err != nil {
		t.Fatalf("edit salary scheme: %v", err)
	}
	row = findByName(mustList(t, svcCtx, types.ListRequest{}), "月度标准方案")
	if row.SubjectName != "乙公司主体" || row.PersonStatus != "1,3,6" || row.Status != "0" ||
		row.PayDay != 15 || row.Currency != "USD" {
		t.Fatalf("edit not applied: %+v", row)
	}

	// 编辑不存在的方案
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: "no-such-uuid", SchemeName: "幽灵方案", SubjectUuid: subjectA,
		PersonStatus: []string{"1"}, SettleCycle: "monthly", PayDay: 1, Status: "1",
	}); err == nil {
		t.Fatalf("editing unknown scheme should fail")
	}

	// 删除
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
		t.Fatalf("delete salary scheme: %v", err)
	}
	if left := mustList(t, svcCtx, types.ListRequest{}); len(left) != 0 {
		t.Fatalf("expected empty after delete, got %+v", left)
	}
}

// TestSalarySchemeValidation 非法参数守卫: 名称空/人员状态空/状态越界/结算周期/发薪日/主体不存在.
func TestSalarySchemeValidation(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	subject := seedSubject(t, svcCtx, "丙公司主体")

	cases := []struct {
		name string
		req  types.AddRequest
	}{
		{"empty name", types.AddRequest{SchemeName: "", SubjectUuid: subject, PersonStatus: []string{"1"}, SettleCycle: "monthly", PayDay: 1}},
		{"empty person status", types.AddRequest{SchemeName: "X", SubjectUuid: subject, PersonStatus: nil, SettleCycle: "monthly", PayDay: 1}},
		{"bad person status", types.AddRequest{SchemeName: "X", SubjectUuid: subject, PersonStatus: []string{"9"}, SettleCycle: "monthly", PayDay: 1}},
		{"bad settle cycle", types.AddRequest{SchemeName: "X", SubjectUuid: subject, PersonStatus: []string{"1"}, SettleCycle: "weekly", PayDay: 1}},
		{"pay day too big", types.AddRequest{SchemeName: "X", SubjectUuid: subject, PersonStatus: []string{"1"}, SettleCycle: "monthly", PayDay: 32}},
		{"pay day too small", types.AddRequest{SchemeName: "X", SubjectUuid: subject, PersonStatus: []string{"1"}, SettleCycle: "monthly", PayDay: -1}},
		{"bad status", types.AddRequest{SchemeName: "X", SubjectUuid: subject, PersonStatus: []string{"1"}, SettleCycle: "monthly", PayDay: 1, Status: "2"}},
		{"missing subject", types.AddRequest{SchemeName: "X", SubjectUuid: "", PersonStatus: []string{"1"}, SettleCycle: "monthly", PayDay: 1}},
		{"unknown subject", types.AddRequest{SchemeName: "X", SubjectUuid: "no-such-subject", PersonStatus: []string{"1"}, SettleCycle: "monthly", PayDay: 1}},
	}
	for _, tc := range cases {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&tc.req); err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
	}

	// 合法路径兜底确认（默认值: settle_cycle=monthly pay_day=1 currency=CNY）
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		SchemeName: "丙公司方案", SubjectUuid: subject, PersonStatus: []string{"1"},
	}); err != nil {
		t.Fatalf("valid add should pass: %v", err)
	}
	records := mustList(t, svcCtx, types.ListRequest{SchemeName: "丙公司方案"})
	if len(records) != 1 || records[0].SettleCycle != "monthly" || records[0].PayDay != 1 || records[0].Currency != "CNY" {
		t.Fatalf("defaults not applied: %+v", records)
	}
}

// TestSalarySchemeItems 关联表配置: 全量勾选落库 / 编辑整体替换（全删全插事务性）/
// 重复勾选拒绝 / 引用不存在项目拒绝 / 完整度联动.
func TestSalarySchemeItems(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	subject := seedSubject(t, svcCtx, "甲公司主体")
	basic := findPresetItem(t, svcCtx, "BASIC_SALARY")
	tax := findPresetItem(t, svcCtx, "TAX")

	// 无项目方案: 完整度 项目✗ 计薪✗ 个税✗
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		SchemeName: "空方案", SubjectUuid: subject, PersonStatus: []string{"1"}, SettleCycle: "monthly", PayDay: 1,
	}); err != nil {
		t.Fatalf("add empty scheme: %v", err)
	}
	empty := findByName(mustList(t, svcCtx, types.ListRequest{}), "空方案")
	if empty.ItemCount != 0 || empty.Completeness.SchemeItems != 0 ||
		empty.Completeness.SchemeCalc != 0 || empty.Completeness.SchemeTax != 0 {
		t.Fatalf("empty scheme completeness wrong: %+v", empty.Completeness)
	}

	// 引用不存在项目
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: empty.Uuid, SchemeName: "空方案", SubjectUuid: subject, PersonStatus: []string{"1"},
		SettleCycle: "monthly", PayDay: 1, Status: "1",
		Items: []types.SchemeItemPayload{{ItemUuid: "no-such-item"}},
	}); err == nil {
		t.Fatalf("unknown item should fail")
	}

	// 重复勾选同一项目
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: empty.Uuid, SchemeName: "空方案", SubjectUuid: subject, PersonStatus: []string{"1"},
		SettleCycle: "monthly", PayDay: 1, Status: "1",
		Items: []types.SchemeItemPayload{{ItemUuid: basic}, {ItemUuid: basic}},
	}); err == nil {
		t.Fatalf("duplicate item should fail")
	}

	// 配置 2 个项目（个税 formula 项 + 固定项带覆盖金额）→ 项目✓ 计薪✓ 个税✓
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: empty.Uuid, SchemeName: "空方案", SubjectUuid: subject, PersonStatus: []string{"1"},
		SettleCycle: "monthly", PayDay: 1, Status: "1",
		Items: []types.SchemeItemPayload{
			{ItemUuid: basic, OverrideAmount: 5000},
			{ItemUuid: tax},
		},
	}); err != nil {
		t.Fatalf("edit scheme items: %v", err)
	}
	row := findByName(mustList(t, svcCtx, types.ListRequest{}), "空方案")
	if row.ItemCount != 2 || row.Completeness.SchemeItems != 1 ||
		row.Completeness.SchemeCalc != 1 || row.Completeness.SchemeTax != 1 {
		t.Fatalf("completeness wrong after config: count=%d %+v", row.ItemCount, row.Completeness)
	}
	// 未接入项恒 0
	if row.Completeness.SchemePosition != 0 || row.Completeness.SchemeSi != 0 || row.Completeness.SchemeAttendance != 0 {
		t.Fatalf("not-integrated checks must be 0: %+v", row.Completeness)
	}

	// 详情回显: 按 calc_order 排序 + 项目快照 + 覆盖值
	detail, err := NewGetSchemeDetail(ctx, svcCtx, nil).GetSchemeDetail(&types.GetSchemeDetailRequest{Uuid: row.Uuid})
	if err != nil {
		t.Fatalf("get scheme detail: %v", err)
	}
	if len(detail.Items) != 2 {
		t.Fatalf("expected 2 detail items, got %d", len(detail.Items))
	}
	if detail.Items[0].ItemCode != "BASIC_SALARY" || detail.Items[1].ItemCode != "TAX" {
		t.Fatalf("calc_order sort wrong: %+v", detail.Items)
	}
	if detail.Items[0].OverrideAmount != 5000 || detail.Items[0].ItemName != "基本工资" {
		t.Fatalf("override/snapshot wrong: %+v", detail.Items[0])
	}

	// 全删: 编辑为 0 项目 → 完整度归零（全删全插事务性）
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: row.Uuid, SchemeName: "空方案", SubjectUuid: subject, PersonStatus: []string{"1"},
		SettleCycle: "monthly", PayDay: 1, Status: "1", Items: nil,
	}); err != nil {
		t.Fatalf("clear scheme items: %v", err)
	}
	row = findByName(mustList(t, svcCtx, types.ListRequest{}), "空方案")
	if row.ItemCount != 0 || row.Completeness.SchemeItems != 0 {
		t.Fatalf("items not cleared: count=%d %+v", row.ItemCount, row.Completeness)
	}

	// 删除方案后关联表无残留
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
		t.Fatalf("delete scheme: %v", err)
	}
	left, err := GetSchemeItemsForScheme(ctx, svcCtx, row.Uuid)
	if err == nil {
		t.Fatalf("GetSchemeItemsForScheme on deleted scheme should fail")
	}
	if len(left) != 0 {
		t.Fatalf("orphan items left: %+v", left)
	}
}

// TestGetSchemeItemsOrdering HRM-37 消费契约: GetSchemeItemsForScheme 按 calc_order ASC 排序
// （与关联表插入顺序无关），包含覆盖值，容忍停用项目 status='0'。
func TestGetSchemeItemsOrdering(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	subject := seedSubject(t, svcCtx, "甲公司主体")
	// 故意按乱序插入配置: LATE(900) 先于 EARLY(20)（code 避开迁移种子，item_code 唯一）
	late := seedItem(t, svcCtx, "SCHEME_TEST_LATE", "测试后项", "formula", "deduction", "formula", 900)
	early := seedItem(t, svcCtx, "SCHEME_TEST_EARLY", "测试先项", "fixed", "earning", "manual", 20)

	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		SchemeName: "乱序方案", SubjectUuid: subject, PersonStatus: []string{"1"},
		SettleCycle: "monthly", PayDay: 1,
		Items: []types.SchemeItemPayload{
			{ItemUuid: late, OverrideFormula: "应发合计*0.1"},
			{ItemUuid: early},
		},
	}); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	row := findByName(mustList(t, svcCtx, types.ListRequest{}), "乱序方案")

	items, err := GetSchemeItemsForScheme(ctx, svcCtx, row.Uuid)
	if err != nil {
		t.Fatalf("get scheme items: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ItemCode != "SCHEME_TEST_EARLY" || items[0].CalcOrder != 20 {
		t.Fatalf("first item should be SCHEME_TEST_EARLY(20): %+v", items[0])
	}
	if items[1].ItemCode != "SCHEME_TEST_LATE" || items[1].OverrideFormula != "应发合计*0.1" {
		t.Fatalf("second item should be SCHEME_TEST_LATE with override: %+v", items[1])
	}

	// 停用先项后仍容忍展示（status='0'）
	earlyRow, err := svcCtx.Model.HrmSalaryItem.FindOneByUuid(ctx, nil, early)
	if err != nil {
		t.Fatalf("find early item: %v", err)
	}
	earlyRow.Status = "0"
	if err := svcCtx.Model.HrmSalaryItem.Update(ctx, nil, earlyRow); err != nil {
		t.Fatalf("disable early item: %v", err)
	}
	items, err = GetSchemeItemsForScheme(ctx, svcCtx, row.Uuid)
	if err != nil {
		t.Fatalf("get scheme items after disable: %v", err)
	}
	if len(items) != 2 || items[0].Status != "0" {
		t.Fatalf("disabled item tolerance broken: %+v", items)
	}

	// 不存在的方案
	if _, err := GetSchemeItemsForScheme(ctx, svcCtx, "no-such-scheme"); err == nil {
		t.Fatalf("unknown scheme should fail")
	}
}
