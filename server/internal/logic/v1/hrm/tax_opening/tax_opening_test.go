package tax_opening

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_opening"
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

func seedSubject(t *testing.T, svcCtx *svc.ServiceContext, name string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalarySubject.Insert(t.Context(), nil, &hrm_salary_subject.HrmSalarySubject{
		Uuid: u, SubjectName: name, OrgUuid: "org-x", OrgName: "公司", Status: "1",
	}); err != nil {
		t.Fatalf("seed subject %s: %v", name, err)
	}
	return u
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmTaxOpening {
	t.Helper()
	req.Current = 1
	req.Size = 100
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list tax openings: %v", err)
	}
	return resp.Records
}

func findByName(records []types.HrmTaxOpening, name string) *types.HrmTaxOpening {
	for i := range records {
		if records[i].EmployeeName == name {
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

// TestTaxOpeningCrud walks the CRUD seam: 新增（员工/主体快照 + 唯一守卫 员工+主体+年度）
// → 编辑（换员工/改累计数）→ 筛选（姓名/主体/年度）→ 删除。
func TestTaxOpeningCrud(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	empA := seedEmployee(t, svcCtx, "周八", "420106199707079012")
	empB := seedEmployee(t, svcCtx, "吴九", "420106199808083456")
	subject := seedSubject(t, svcCtx, "甲公司薪酬主体")

	// 新增（带主体）
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: empA, SubjectUuid: subject, TaxYear: "2026", UntilMonth: 6,
		IncomeTotal: 60000, SpecialDeductionTotal: 9600, SpecialExtraTotal: 9000,
		TaxableIncomeTotal: 2400, TaxPaidTotal: 72,
	}); err != nil {
		t.Fatalf("add tax opening: %v", err)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 1 {
		t.Fatalf("expected 1 opening after add, got %d", len(records))
	}
	if records[0].EmployeeName != "周八" || records[0].SubjectName != "甲公司薪酬主体" ||
		records[0].TaxYear != "2026" || records[0].UntilMonth != 6 || records[0].IncomeTotal != 60000 ||
		records[0].TaxPaidTotal != 72 {
		t.Fatalf("add snapshot/fields not applied: %+v", records[0])
	}

	// 唯一守卫: 同员工同主体同年度
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: empA, SubjectUuid: subject, TaxYear: "2026", UntilMonth: 7,
	}); err == nil {
		t.Fatalf("duplicate opening should fail")
	}

	// 同员工同年度不同主体 → 允许
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: empA, TaxYear: "2026", UntilMonth: 6,
	}); err != nil {
		t.Fatalf("same employee different subject should pass: %v", err)
	}

	// 校验拒绝: 截至月越界 / 金额为负 / 年度格式
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: empB, TaxYear: "2026", UntilMonth: 13,
	}); err == nil {
		t.Fatalf("month 13 should fail")
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: empB, TaxYear: "2026", UntilMonth: 1, IncomeTotal: -1,
	}); err == nil {
		t.Fatalf("negative income should fail")
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: empB, TaxYear: "26", UntilMonth: 1,
	}); err == nil {
		t.Fatalf("bad year should fail")
	}

	// 编辑: 换员工 + 改累计数
	row := findByName(mustList(t, svcCtx, types.ListRequest{}), "周八")
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: row.Uuid, EmployeeUuid: empB, TaxYear: "2026", UntilMonth: 9,
		IncomeTotal: 90000, TaxPaidTotal: 500,
	}); err != nil {
		t.Fatalf("edit tax opening: %v", err)
	}
	row = findByName(mustList(t, svcCtx, types.ListRequest{}), "吴九")
	if row == nil || row.UntilMonth != 9 || row.IncomeTotal != 90000 || row.TaxPaidTotal != 500 {
		t.Fatalf("edit not applied: %+v", row)
	}

	// 编辑成唯一键冲突（换回 empA 的 2026/subject）应拒绝
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: row.Uuid, EmployeeUuid: empA, SubjectUuid: subject, TaxYear: "2026", UntilMonth: 9,
	}); err == nil {
		t.Fatalf("edit to duplicate unique key should fail")
	}

	// 筛选
	if named := mustList(t, svcCtx, types.ListRequest{EmployeeName: "吴九"}); len(named) != 1 {
		t.Fatalf("name filter failed: %+v", named)
	}
	if bySubject := mustList(t, svcCtx, types.ListRequest{SubjectUuid: subject}); len(bySubject) != 1 {
		t.Fatalf("subject filter failed: %+v", bySubject)
	}
	if byYear := mustList(t, svcCtx, types.ListRequest{TaxYear: "2025"}); len(byYear) != 0 {
		t.Fatalf("year filter should be empty: %+v", byYear)
	}

	// 删除
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
		t.Fatalf("delete tax opening: %v", err)
	}
	if got := len(mustList(t, svcCtx, types.ListRequest{})); got != 1 {
		t.Fatalf("expected 1 opening after delete, got %d", got)
	}
}

// TestTaxOpeningImportExportRoundTrip 导入→导出闭环: 导入 2 行（员工同名单靠身份证区分、
// 主体按名称解析、空主体=不限）→ 同键重复导入整行覆盖 → 导出回读。
func TestTaxOpeningImportExportRoundTrip(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	subject := seedSubject(t, svcCtx, "乙公司薪酬主体")
	// 同名员工两个, 靠身份证区分
	empA := seedEmployee(t, svcCtx, "陈十", "420106199909099012")
	seedEmployee(t, svcCtx, "陈十", "420106200001010054")

	content := buildImportXlsx(t, [][]string{
		{"陈十", "420106199909099012", "乙公司薪酬主体", "2026", "5", "50000", "8000", "7500", "0", "0", "0", "0", "备注A"},
		{"陈十", "420106200001010054", "", "2026", "5", "30000", "4800", "0", "0", "0", "0", "0", ""},
	})
	resp, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "期初.xlsx", Content: content})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.SuccessCount != 2 {
		t.Fatalf("successCount = %d, want 2", resp.SuccessCount)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 2 {
		t.Fatalf("expected 2 openings after import, got %d", len(records))
	}
	var withSubject *types.HrmTaxOpening
	for i := range records {
		if records[i].SubjectUuid == subject {
			withSubject = &records[i]
		}
	}
	if withSubject == nil || withSubject.EmployeeUuid != empA || withSubject.IncomeTotal != 50000 ||
		withSubject.SubjectName != "乙公司薪酬主体" {
		t.Fatalf("imported row with subject wrong: %+v", withSubject)
	}

	// 同键重复导入 → 覆盖
	content = buildImportXlsx(t, [][]string{
		{"陈十", "420106199909099012", "乙公司薪酬主体", "2026", "8", "80000", "12800", "12000", "0", "0", "0", "300", ""},
	})
	resp, err = NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "期初.xlsx", Content: content})
	if err != nil {
		t.Fatalf("re-import: %v", err)
	}
	if resp.SuccessCount != 1 {
		t.Fatalf("re-import successCount = %d, want 1", resp.SuccessCount)
	}
	records = mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 2 {
		t.Fatalf("re-import should not add rows, got %d", len(records))
	}
	for i := range records {
		if records[i].SubjectUuid == subject {
			if records[i].UntilMonth != 8 || records[i].IncomeTotal != 80000 || records[i].TaxPaidTotal != 300 {
				t.Fatalf("re-import overwrite failed: %+v", records[i])
			}
		}
	}

	// 导出 → 回读 3 行（header + 2）
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
	if len(outRows) != 3 {
		t.Fatalf("export rows = %d, want 3", len(outRows))
	}
	if outRows[1][2] != "乙公司薪酬主体" || outRows[1][4] != "8" {
		t.Fatalf("export content wrong: %v", outRows[1])
	}

	// 坏行: 主体不存在 → 整体回滚
	content = buildImportXlsx(t, [][]string{
		{"陈十", "420106199909099012", "不存在的主体", "2026", "9", "1", "0", "0", "0", "0", "0", "0", ""},
	})
	if _, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "期初.xlsx", Content: content}); err == nil {
		t.Fatalf("import with unknown subject should fail")
	}
	if got := len(mustList(t, svcCtx, types.ListRequest{})); got != 2 {
		t.Fatalf("bad import should roll back entirely, got %d rows", got)
	}
}
