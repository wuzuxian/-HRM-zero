package salary_archive

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
)

// TestSalaryArchiveCrud walks the CRUD seam (HRM-35 user story 57 / demo-inventory 8.2):
// 新增（选员工带出 公司/部门/工号/姓名 快照）→ 编辑（换主体/启停）→ 筛选（姓名/工号/主体/方案/状态）→ 删除.
func TestSalaryArchiveCrud(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	company := seedOrg(t, svcCtx, "甲公司", "1")
	department := seedOrg(t, svcCtx, "研发部", "2")
	employee := seedEmployeeRow(t, svcCtx, "张三", "E0001", company, department)
	subjectA := seedSubject(t, svcCtx, "甲公司主体")
	subjectB := seedSubject(t, svcCtx, "乙公司主体")
	scheme := seedScheme(t, svcCtx, "月度方案", subjectA)

	// 新增: 快照列由员工/主体/方案带出
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subjectA, SchemeUuid: scheme,
		FixedSalary: 8000, FixedAllowance: 500,
		EffectiveDate: "2026-01-01",
	}); err != nil {
		t.Fatalf("add salary archive: %v", err)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 1 {
		t.Fatalf("expected 1 archive after add, got %d", len(records))
	}
	row := records[0]
	if row.EmployeeNo != "E0001" || row.EmployeeName != "张三" ||
		row.CompanyName != "甲公司" || row.OrgName != "研发部" ||
		row.SubjectName != "甲公司主体" || row.SchemeName != "月度方案" ||
		row.Status != "1" || row.FixedSalary != 8000 || row.FixedAllowance != 500 ||
		row.ExpireDate != "" {
		t.Fatalf("add snapshots/defaults not applied: %+v", row)
	}

	// 筛选: 员工姓名 / 员工ID / 主体 / 方案 / 状态
	if got := mustList(t, svcCtx, types.ListRequest{EmployeeName: "张三"}); len(got) != 1 {
		t.Fatalf("employeeName filter failed: %+v", got)
	}
	if got := mustList(t, svcCtx, types.ListRequest{EmployeeNo: "E0001"}); len(got) != 1 {
		t.Fatalf("employeeNo filter failed: %+v", got)
	}
	if got := mustList(t, svcCtx, types.ListRequest{SubjectUuid: subjectA}); len(got) != 1 {
		t.Fatalf("subject filter failed: %+v", got)
	}
	if got := mustList(t, svcCtx, types.ListRequest{SchemeUuid: scheme}); len(got) != 1 {
		t.Fatalf("scheme filter failed: %+v", got)
	}
	if got := mustList(t, svcCtx, types.ListRequest{Status: "0"}); len(got) != 0 {
		t.Fatalf("status filter should be empty: %+v", got)
	}

	// 编辑: 换主体+方案、金额、启停；快照随换随更
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: row.Uuid, EmployeeUuid: employee, SubjectUuid: subjectB, SchemeUuid: scheme,
		FixedSalary: 9000, FixedAllowance: 600, Status: "0",
		EffectiveDate: "2026-01-01", ExpireDate: "2026-12-31",
	}); err != nil {
		t.Fatalf("edit salary archive: %v", err)
	}
	row = *findByName(mustList(t, svcCtx, types.ListRequest{}), "张三")
	if row.SubjectName != "乙公司主体" || row.FixedSalary != 9000 || row.Status != "0" ||
		row.ExpireDate != "2026-12-31" {
		t.Fatalf("edit not applied: %+v", row)
	}

	// 编辑不存在的档案
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: "no-such-uuid", EmployeeUuid: employee, SubjectUuid: subjectA,
		Status: "1", EffectiveDate: "2026-01-01",
	}); err == nil {
		t.Fatalf("editing unknown archive should fail")
	}

	// 删除
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
		t.Fatalf("delete salary archive: %v", err)
	}
	if left := mustList(t, svcCtx, types.ListRequest{}); len(left) != 0 {
		t.Fatalf("expected empty after delete, got %+v", left)
	}
}

// TestSalaryArchiveValidation 非法参数守卫:
// 员工/主体为空、员工不存在、主体不存在、方案不存在、金额负数、状态枚举越界、日期格式非法、失效<生效拒绝.
func TestSalaryArchiveValidation(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	company := seedOrg(t, svcCtx, "甲公司", "1")
	employee := seedEmployeeRow(t, svcCtx, "李四", "E0002", company, "")
	subject := seedSubject(t, svcCtx, "甲公司主体")

	cases := []struct {
		name string
		req  types.AddRequest
		code int
	}{
		{"missing employee", types.AddRequest{SubjectUuid: subject, EffectiveDate: "2026-01-01"}, errcodes.SalaryArchiveInvalidCode},
		{"missing subject", types.AddRequest{EmployeeUuid: employee, EffectiveDate: "2026-01-01"}, errcodes.SalaryArchiveInvalidCode},
		{"unknown employee", types.AddRequest{EmployeeUuid: "no-such-emp", SubjectUuid: subject, EffectiveDate: "2026-01-01"}, errcodes.SalaryArchiveEmployeeNotFoundCode},
		{"unknown subject", types.AddRequest{EmployeeUuid: employee, SubjectUuid: "no-such-subject", EffectiveDate: "2026-01-01"}, errcodes.SalaryArchiveSubjectNotFoundCode},
		{"unknown scheme", types.AddRequest{EmployeeUuid: employee, SubjectUuid: subject, SchemeUuid: "no-such-scheme", EffectiveDate: "2026-01-01"}, errcodes.SalaryArchiveSchemeNotFoundCode},
		{"bad status", types.AddRequest{EmployeeUuid: employee, SubjectUuid: subject, Status: "2", EffectiveDate: "2026-01-01"}, errcodes.SalaryArchiveInvalidCode},
		{"negative salary", types.AddRequest{EmployeeUuid: employee, SubjectUuid: subject, FixedSalary: -1, EffectiveDate: "2026-01-01"}, errcodes.SalaryArchiveInvalidCode},
		{"bad date format", types.AddRequest{EmployeeUuid: employee, SubjectUuid: subject, EffectiveDate: "2026/01/01"}, errcodes.SalaryArchiveInvalidCode},
		{"bad expire format", types.AddRequest{EmployeeUuid: employee, SubjectUuid: subject, EffectiveDate: "2026-01-01", ExpireDate: "2026-13-01"}, errcodes.SalaryArchiveInvalidCode},
		{"expire equals effective", types.AddRequest{EmployeeUuid: employee, SubjectUuid: subject, EffectiveDate: "2026-01-01", ExpireDate: "2026-01-01"}, errcodes.SalaryArchiveDateInvalidCode},
		{"expire before effective", types.AddRequest{EmployeeUuid: employee, SubjectUuid: subject, EffectiveDate: "2026-02-01", ExpireDate: "2026-01-31"}, errcodes.SalaryArchiveDateInvalidCode},
	}
	for _, tc := range cases {
		_, err := NewAdd(ctx, svcCtx, nil).Add(&tc.req)
		assertErrCode(t, err, tc.code, tc.name)
	}

	// 合法路径兜底确认（确保上面的失败不是环境性的）
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subject, EffectiveDate: "2026-01-01",
	}); err != nil {
		t.Fatalf("valid add should pass: %v", err)
	}
}

// TestSalaryArchiveIntervalOverlap 业务守卫 #1: 同员工同主体同期间重叠拒绝（生效区间不可交叉）；
// 同员工不同主体不冲突；停用档案不占区间；Edit 排除自身.
func TestSalaryArchiveIntervalOverlap(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	company := seedOrg(t, svcCtx, "甲公司", "1")
	employee := seedEmployeeRow(t, svcCtx, "王五", "E0003", company, "")
	employee2 := seedEmployeeRow(t, svcCtx, "赵六", "E0004", company, "")
	subjectA := seedSubject(t, svcCtx, "甲公司主体")
	subjectB := seedSubject(t, svcCtx, "乙公司主体")

	// 既有档案: 2026-03-01 ~ 2026-08-31
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subjectA,
		EffectiveDate: "2026-03-01", ExpireDate: "2026-08-31",
	}); err != nil {
		t.Fatalf("seed existing archive: %v", err)
	}

	overlapCases := []struct {
		name          string
		effectiveDate string
		expireDate    string
	}{
		{"完全包含", "2026-04-01", "2026-05-31"},
		{"首部交叉", "2026-02-01", "2026-03-15"},
		{"尾部交叉", "2026-08-15", "2026-12-31"},
		{"完全覆盖", "2026-01-01", "2026-12-31"},
		{"紧贴首日交叉", "2026-02-28", "2026-03-01"},
		{"紧贴末日交叉", "2026-08-31", "2026-09-30"},
		{"无失效日期长档", "2026-05-01", ""},
		{"既有无失效新档落其后", "2026-07-01", ""}, // newExp=∞ > oldEff → 交叉
	}
	for _, tc := range overlapCases {
		_, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
			EmployeeUuid: employee, SubjectUuid: subjectA,
			EffectiveDate: tc.effectiveDate, ExpireDate: tc.expireDate,
		})
		assertErrCode(t, err, errcodes.SalaryArchiveOverlapCode, tc.name)
	}

	// 合法: 失效日 = 既有生效日前一天（首部紧贴不交叉）
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subjectA,
		EffectiveDate: "2026-01-01", ExpireDate: "2026-02-28",
	}); err != nil {
		t.Fatalf("adjacent-before archive should pass: %v", err)
	}
	// 合法: 生效日 = 既有失效日后一天（尾部紧贴不交叉）
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subjectA,
		EffectiveDate: "2026-09-01", ExpireDate: "2026-12-31",
	}); err != nil {
		t.Fatalf("adjacent-after archive should pass: %v", err)
	}
	// 合法: 同员工不同主体同期间不冲突
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subjectB,
		EffectiveDate: "2026-05-01", ExpireDate: "2026-06-30",
	}); err != nil {
		t.Fatalf("same employee different subject should pass: %v", err)
	}
	// 合法: 不同员工同主体同期间不冲突
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee2, SubjectUuid: subjectA,
		EffectiveDate: "2026-05-01", ExpireDate: "2026-06-30",
	}); err != nil {
		t.Fatalf("different employee same subject should pass: %v", err)
	}

	// 停用既有档案后，同区间新档放行（停用档案不占区间）
	rows := mustList(t, svcCtx, types.ListRequest{Status: "1"})
	var firstRow types.HrmSalaryArchive
	for _, r := range rows {
		if r.EffectiveDate == "2026-03-01" {
			firstRow = r
			break
		}
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: firstRow.Uuid, EmployeeUuid: employee, SubjectUuid: subjectA,
		Status: "0", EffectiveDate: "2026-03-01", ExpireDate: "2026-08-31",
	}); err != nil {
		t.Fatalf("disable existing archive: %v", err)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subjectA,
		EffectiveDate: "2026-04-01", ExpireDate: "2026-05-31",
	}); err != nil {
		t.Fatalf("overlap with disabled archive should pass: %v", err)
	}

	// Edit 排除自身: 把既有档案区间原地平移一格不报重叠
	edited := findByName(mustList(t, svcCtx, types.ListRequest{}), "王五")
	if edited == nil {
		t.Fatalf("archive of 王五 not found")
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: edited.Uuid, EmployeeUuid: employee, SubjectUuid: subjectA,
		Status: "1", EffectiveDate: "2026-04-01", ExpireDate: "2026-05-31",
	}); err != nil {
		t.Fatalf("edit keeping own interval should pass: %v", err)
	}
}
