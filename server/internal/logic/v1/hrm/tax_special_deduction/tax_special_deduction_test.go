package tax_special_deduction

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_special_deduction"
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

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmTaxSpecialDeductionRow {
	t.Helper()
	req.Current = 1
	req.Size = 200
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list special deductions: %v", err)
	}
	return resp.Records
}

// findCell 找 (月份, 扣除项) 行，找不到返回 nil.
func findCell(records []types.HrmTaxSpecialDeductionRow, month int64, deductType string) *types.HrmTaxSpecialDeductionRow {
	for i := range records {
		if records[i].DeductMonth == month && records[i].DeductType == deductType {
			return &records[i]
		}
	}
	return nil
}

// buildImportXlsx builds the shared import/export template content.
func buildImportXlsx(t *testing.T, rows [][]string) string {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if err := f.SetSheetRow(sheet, "A1", &importHeader); err != nil {
		t.Fatalf("set header: %v", err)
	}
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+2)
		if err != nil {
			t.Fatalf("cell: %v", err)
		}
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			t.Fatalf("set row: %v", err)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// TestTaxSpecialDeductionMaintain 维护弹窗 seam: 首次落格 → 部分月份更新（未提交格子不动）
// → 提交内重复拒绝 → 校验拒绝（月份/扣除项/金额）。
func TestTaxSpecialDeductionMaintain(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := seedEmployee(t, svcCtx, "王五", "420106199303039012")

	// 首次维护 3 月+6 月的子女教育
	resp, err := NewMaintain(ctx, svcCtx, nil).Maintain(&types.MaintainRequest{
		EmployeeUuid: emp, DeductYear: "2026",
		Rows: []types.MaintainRow{
			{DeductMonth: 3, DeductType: DeductTypeChildrenEdu, Amount: 2000},
			{DeductMonth: 6, DeductType: DeductTypeChildrenEdu, Amount: 1000},
		},
	})
	if err != nil {
		t.Fatalf("maintain: %v", err)
	}
	if resp.Saved != 2 {
		t.Fatalf("saved = %d, want 2", resp.Saved)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 2 || findCell(records, 3, DeductTypeChildrenEdu) == nil ||
		findCell(records, 6, DeductTypeChildrenEdu).Amount != 1000 {
		t.Fatalf("after first maintain: %+v", records)
	}
	// 快照: 姓名落库、身份证从员工表带出、来源 manual
	if records[0].EmployeeName != "王五" || records[0].IdCard != "420106199303039012" ||
		records[0].Source != SourceManual || records[0].DeductYear != "2026" {
		t.Fatalf("snapshot/source wrong: %+v", records[0])
	}

	// 部分月份更新: 只提交 3 月, 6 月不动
	if _, err := NewMaintain(ctx, svcCtx, nil).Maintain(&types.MaintainRequest{
		EmployeeUuid: emp, DeductYear: "2026",
		Rows: []types.MaintainRow{{DeductMonth: 3, DeductType: DeductTypeChildrenEdu, Amount: 2500}},
	}); err != nil {
		t.Fatalf("partial maintain: %v", err)
	}
	records = mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 2 {
		t.Fatalf("partial maintain should not add rows, got %d", len(records))
	}
	if findCell(records, 3, DeductTypeChildrenEdu).Amount != 2500 {
		t.Fatalf("march amount not updated: %+v", findCell(records, 3, DeductTypeChildrenEdu))
	}
	if findCell(records, 6, DeductTypeChildrenEdu).Amount != 1000 {
		t.Fatalf("june amount should stay: %+v", findCell(records, 6, DeductTypeChildrenEdu))
	}

	// 跨年度: 2025 年的同一格是独立行
	if _, err := NewMaintain(ctx, svcCtx, nil).Maintain(&types.MaintainRequest{
		EmployeeUuid: emp, DeductYear: "2025",
		Rows: []types.MaintainRow{{DeductMonth: 3, DeductType: DeductTypeChildrenEdu, Amount: 1800}},
	}); err != nil {
		t.Fatalf("prior year maintain: %v", err)
	}
	if got := len(mustList(t, svcCtx, types.ListRequest{})); got != 3 {
		t.Fatalf("expected 3 rows across years, got %d", got)
	}

	// 提交内重复 (月份, 扣除项)
	if _, err := NewMaintain(ctx, svcCtx, nil).Maintain(&types.MaintainRequest{
		EmployeeUuid: emp, DeductYear: "2026",
		Rows: []types.MaintainRow{
			{DeductMonth: 5, DeductType: DeductTypeHouseRent, Amount: 1500},
			{DeductMonth: 5, DeductType: DeductTypeHouseRent, Amount: 1600},
		},
	}); err == nil {
		t.Fatalf("duplicate month+type in one submit should fail")
	}

	// 校验拒绝: 月份越界 / 扣除项非法 / 金额为负
	for _, bad := range []types.MaintainRow{
		{DeductMonth: 13, DeductType: DeductTypeChildrenEdu, Amount: 100},
		{DeductMonth: 1, DeductType: "NOT_A_TYPE", Amount: 100},
		{DeductMonth: 1, DeductType: DeductTypeChildrenEdu, Amount: -1},
	} {
		if _, err := NewMaintain(ctx, svcCtx, nil).Maintain(&types.MaintainRequest{
			EmployeeUuid: emp, DeductYear: "2026", Rows: []types.MaintainRow{bad},
		}); err == nil {
			t.Fatalf("bad row %+v should fail", bad)
		}
	}

	// 员工不存在
	if _, err := NewMaintain(ctx, svcCtx, nil).Maintain(&types.MaintainRequest{
		EmployeeUuid: "no-such", DeductYear: "2026",
		Rows: []types.MaintainRow{{DeductMonth: 1, DeductType: DeductTypeChildrenEdu, Amount: 1}},
	}); err == nil {
		t.Fatalf("unknown employee should fail")
	}
}

// TestTaxSpecialDeductionMaintainYear 年度维护: 全年 12 格套用 / 指定月份套用 / 已有月份覆盖。
func TestTaxSpecialDeductionMaintainYear(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := seedEmployee(t, svcCtx, "赵六", "420106199404041234")

	// 全年套用
	resp, err := NewMaintainYear(ctx, svcCtx, nil).MaintainYear(&types.MaintainYearRequest{
		EmployeeUuid: emp, DeductYear: "2026", DeductType: DeductTypeSupportElderly, Amount: 3000,
	})
	if err != nil {
		t.Fatalf("maintain year: %v", err)
	}
	if resp.Saved != 12 {
		t.Fatalf("saved = %d, want 12", resp.Saved)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 12 {
		t.Fatalf("expected 12 rows, got %d", len(records))
	}

	// 先手工维护 3 月住房租金, 再年度维护只套 3 月 → 覆盖为赡养老人金额, 其余 11 格赡养不变
	if _, err := NewMaintain(ctx, svcCtx, nil).Maintain(&types.MaintainRequest{
		EmployeeUuid: emp, DeductYear: "2026",
		Rows: []types.MaintainRow{{DeductMonth: 3, DeductType: DeductTypeHouseRent, Amount: 1500}},
	}); err != nil {
		t.Fatalf("manual march: %v", err)
	}
	resp, err = NewMaintainYear(ctx, svcCtx, nil).MaintainYear(&types.MaintainYearRequest{
		EmployeeUuid: emp, DeductYear: "2026", DeductType: DeductTypeSupportElderly, Amount: 2500, Months: []int64{3},
	})
	if err != nil {
		t.Fatalf("maintain year march: %v", err)
	}
	if resp.Saved != 1 {
		t.Fatalf("saved = %d, want 1", resp.Saved)
	}
	records = mustList(t, svcCtx, types.ListRequest{})
	// 12 格赡养 + 手工维护的 3 月住房租金 = 13 行（年度维护不新增、只覆盖 3 月赡养）
	if len(records) != 13 {
		t.Fatalf("expected 13 rows (12 support + 1 house rent), got %d", len(records))
	}
	march := findCell(records, 3, DeductTypeSupportElderly)
	if march == nil || march.Amount != 2500 {
		t.Fatalf("march support elderly not applied: %+v", march)
	}
	// 月份越界
	if _, err := NewMaintainYear(ctx, svcCtx, nil).MaintainYear(&types.MaintainYearRequest{
		EmployeeUuid: emp, DeductYear: "2026", DeductType: DeductTypeSupportElderly, Amount: 1, Months: []int64{0},
	}); err == nil {
		t.Fatalf("invalid month should fail")
	}
}

// TestTaxSpecialDeductionImportExportRoundTrip 导入→导出闭环: 导入 2 行（含覆盖已有行）→
// 导出 → 回读校验；坏行（未知扣除项/坏金额）整体回滚。
func TestTaxSpecialDeductionImportExportRoundTrip(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := seedEmployee(t, svcCtx, "孙七", "420106199505055678")

	// 预置一行手工数据（导入同格应覆盖并把来源置 import）
	if _, err := NewMaintain(ctx, svcCtx, nil).Maintain(&types.MaintainRequest{
		EmployeeUuid: emp, DeductYear: "2026",
		Rows: []types.MaintainRow{{DeductMonth: 1, DeductType: DeductTypeChildrenEdu, Amount: 999}},
	}); err != nil {
		t.Fatalf("preset: %v", err)
	}

	content := buildImportXlsx(t, [][]string{
		{"孙七", "420106199505055678", "2026", "1", "子女教育", "2000", ""},
		{"孙七", "420106199505055678", "2026", "2", "住房租金", "1500", ""},
	})
	resp, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "扣除.xlsx", Content: content})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.SuccessCount != 2 {
		t.Fatalf("successCount = %d, want 2", resp.SuccessCount)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 2 {
		t.Fatalf("expected 2 rows after import, got %d", len(records))
	}
	jan := findCell(records, 1, DeductTypeChildrenEdu)
	if jan.Amount != 2000 || jan.Source != SourceImport {
		t.Fatalf("january overwrite failed: %+v", jan)
	}

	// 导出 → 内容里应能看到 2 行数据
	exportResp, err := NewExport(ctx, svcCtx, nil).Export()
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(exportResp.Content)
	if err != nil {
		t.Fatalf("decode export: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("open export: %v", err)
	}
	outRows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		t.Fatalf("read export rows: %v", err)
	}
	if len(outRows) != 3 { // header + 2
		t.Fatalf("export rows = %d, want 3", len(outRows))
	}
	if outRows[1][4] != "子女教育" || outRows[2][4] != "住房租金" {
		t.Fatalf("export deduct type names wrong: %v / %v", outRows[1], outRows[2])
	}

	// 坏行整体回滚: 第二行扣除项非法 → 第一行也不落
	content = buildImportXlsx(t, [][]string{
		{"孙七", "420106199505055678", "2026", "3", "赡养老人", "3000", ""},
		{"孙七", "420106199505055678", "2026", "4", "不明扣除", "100", ""},
	})
	if _, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "扣除.xlsx", Content: content}); err == nil {
		t.Fatalf("import with bad row should fail")
	}
	records = mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 2 || findCell(records, 3, DeductTypeSupportElderly) != nil {
		t.Fatalf("bad import should roll back entirely: %+v", records)
	}

	// 同名员工靠身份证定位: 加一个同名员工, 不带身份证导入应失败
	seedEmployee(t, svcCtx, "孙七", "420106199606061234")
	content = buildImportXlsx(t, [][]string{
		{"孙七", "", "2026", "5", "赡养老人", "3000", ""},
	})
	if _, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "扣除.xlsx", Content: content}); err == nil {
		t.Fatalf("ambiguous employee should fail")
	}

	// 删除
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{
		records[0].Uuid, records[1].Uuid,
	}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if got := len(mustList(t, svcCtx, types.ListRequest{})); got != 0 {
		t.Fatalf("expected 0 rows after delete, got %d", got)
	}
}
