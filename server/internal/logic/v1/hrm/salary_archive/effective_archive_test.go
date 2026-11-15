package salary_archive

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
)

// TestGetEffectiveArchive HRM-37 试算消费契约:
// 生效当月命中 / 失效当月命中 / 区间断档月无档案 / 跨档案切换取生效日最晚 /
// 停用档案不返回 / 未绑定方案 Items 空 / 有方案串 GetSchemeItemsForScheme（calc_order 升序+覆盖值）.
func TestGetEffectiveArchive(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	company := seedOrg(t, svcCtx, "甲公司", "1")
	employee := seedEmployeeRow(t, svcCtx, "钱七", "E0005", company, "")
	subject := seedSubject(t, svcCtx, "甲公司主体")

	// 月份边界档案: 2026-03-15 ~ 2026-06-10
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subject,
		FixedSalary: 7000, EffectiveDate: "2026-03-15", ExpireDate: "2026-06-10",
	}); err != nil {
		t.Fatalf("add boundary archive: %v", err)
	}

	cases := []struct {
		name      string
		yearMonth string
		wantMiss  bool
	}{
		{"生效前的月份", "2026-02", true},
		{"生效当月（月中生效≤月末）", "2026-03", false},
		{"覆盖月份", "2026-04", false},
		{"失效当月（失效≥月初）", "2026-06", false},
		{"失效后的月份", "2026-07", true},
	}
	for _, tc := range cases {
		resp, err := GetEffectiveArchiveForEmployee(ctx, svcCtx, employee, tc.yearMonth)
		if tc.wantMiss {
			assertErrCode(t, err, errcodes.SalaryArchiveNotFoundCode, tc.name)
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if resp.Archive.EmployeeName != "钱七" || resp.Archive.FixedSalary != 7000 {
			t.Fatalf("%s: wrong archive: %+v", tc.name, resp.Archive)
		}
	}

	// 月末日恰好=生效日 / 月初日恰好=失效日
	if _, err := GetEffectiveArchiveForEmployee(ctx, svcCtx, employee, "2026-05"); err != nil {
		t.Fatalf("mid month should hit: %v", err)
	}

	// 未绑定方案: Scheme 摘要回填主体快照, Items 空
	resp, err := GetEffectiveArchiveForEmployee(ctx, svcCtx, employee, "2026-04")
	if err != nil {
		t.Fatalf("get effective archive: %v", err)
	}
	if resp.Scheme.SubjectUuid != subject || resp.Scheme.Uuid != "" || len(resp.Items) != 0 {
		t.Fatalf("no-scheme contract broken: %+v", resp)
	}
}

// TestGetEffectiveArchiveSchemeItems 有方案: 串 GetSchemeItemsForScheme（calc_order 升序、覆盖值透传）.
func TestGetEffectiveArchiveSchemeItems(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	company := seedOrg(t, svcCtx, "甲公司", "1")
	employee := seedEmployeeRow(t, svcCtx, "孙八", "E0006", company, "")
	subject := seedSubject(t, svcCtx, "甲公司主体")

	// 乱序: LATE(900) 先插, EARLY(20) 后插 → 返回必须按 calc_order 重排
	late := seedSalaryItem(t, svcCtx, "ARCH_TEST_LATE", "测试后项", "formula", "deduction", "formula", 900)
	early := seedSalaryItem(t, svcCtx, "ARCH_TEST_EARLY", "测试先项", "fixed", "earning", "manual", 20)
	scheme := seedScheme(t, svcCtx, "档案方案", subject)
	seedSchemeItem(t, svcCtx, scheme, late)
	seedSchemeItem(t, svcCtx, scheme, early)

	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subject, SchemeUuid: scheme,
		EffectiveDate: "2026-01-01",
	}); err != nil {
		t.Fatalf("add archive with scheme: %v", err)
	}

	resp, err := GetEffectiveArchiveForEmployee(ctx, svcCtx, employee, "2026-02")
	if err != nil {
		t.Fatalf("get effective archive: %v", err)
	}
	if resp.Scheme.Uuid != scheme || resp.Scheme.SchemeName != "档案方案" || resp.Scheme.SettleCycle != "monthly" {
		t.Fatalf("scheme summary wrong: %+v", resp.Scheme)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].ItemCode != "ARCH_TEST_EARLY" || resp.Items[0].CalcOrder != 20 {
		t.Fatalf("items not ordered by calc_order: %+v", resp.Items)
	}
	if resp.Items[1].ItemCode != "ARCH_TEST_LATE" {
		t.Fatalf("second item wrong: %+v", resp.Items[1])
	}

	// 停用项目容忍返回（status 透传, 过滤口径由试算侧定）
	itemRow, err := svcCtx.Model.HrmSalaryItem.FindOneByUuid(ctx, nil, early)
	if err != nil {
		t.Fatalf("find item: %v", err)
	}
	itemRow.Status = "0"
	if err := svcCtx.Model.HrmSalaryItem.Update(ctx, nil, itemRow); err != nil {
		t.Fatalf("disable item: %v", err)
	}
	resp, err = GetEffectiveArchiveForEmployee(ctx, svcCtx, employee, "2026-02")
	if err != nil {
		t.Fatalf("get effective after item disabled: %v", err)
	}
	if len(resp.Items) != 2 || resp.Items[0].Status != "0" {
		t.Fatalf("disabled item tolerance broken: %+v", resp.Items)
	}
}

// TestGetEffectiveArchiveSwitching 跨档案切换 + 停用档案不返回:
// A 档 [01-01, 05-31] 停用后, 5 月试算不返回 A（status='1' 过滤在区间判定之前）.
func TestGetEffectiveArchiveSwitching(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	company := seedOrg(t, svcCtx, "甲公司", "1")
	employee := seedEmployeeRow(t, svcCtx, "周九", "E0007", company, "")
	subject := seedSubject(t, svcCtx, "甲公司主体")

	// A 档 1-5 月（薪资 1000）, B 档 6 月起长期（薪资 2000）
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subject, FixedSalary: 1000,
		EffectiveDate: "2026-01-01", ExpireDate: "2026-05-31",
	}); err != nil {
		t.Fatalf("add archive A: %v", err)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: employee, SubjectUuid: subject, FixedSalary: 2000,
		EffectiveDate: "2026-06-01",
	}); err != nil {
		t.Fatalf("add archive B: %v", err)
	}

	// 切换前取 A, 切换后取 B
	resp, err := GetEffectiveArchiveForEmployee(ctx, svcCtx, employee, "2026-03")
	if err != nil {
		t.Fatalf("get 2026-03: %v", err)
	}
	if resp.Archive.FixedSalary != 1000 {
		t.Fatalf("2026-03 should hit A: %+v", resp.Archive)
	}
	resp, err = GetEffectiveArchiveForEmployee(ctx, svcCtx, employee, "2026-06")
	if err != nil {
		t.Fatalf("get 2026-06: %v", err)
	}
	if resp.Archive.FixedSalary != 2000 {
		t.Fatalf("2026-06 should hit B: %+v", resp.Archive)
	}

	// B 停用 → 6 月无档案（停用档案不返回）
	rows := mustList(t, svcCtx, types.ListRequest{})
	var archiveB types.HrmSalaryArchive
	for _, r := range rows {
		if r.FixedSalary == 2000 {
			archiveB = r
			break
		}
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: archiveB.Uuid, EmployeeUuid: employee, SubjectUuid: subject,
		FixedSalary: 2000, Status: "0", EffectiveDate: "2026-06-01",
	}); err != nil {
		t.Fatalf("disable archive B: %v", err)
	}
	if _, err := GetEffectiveArchiveForEmployee(ctx, svcCtx, employee, "2026-06"); err == nil {
		t.Fatalf("disabled archive should not be returned")
	} else {
		assertErrCode(t, err, errcodes.SalaryArchiveNotFoundCode, "disabled archive miss")
	}

	// 停用 A 也不会 fallback 到 A（6 月依然 miss）
	rows = mustList(t, svcCtx, types.ListRequest{})
	var archiveA types.HrmSalaryArchive
	for _, r := range rows {
		if r.FixedSalary == 1000 {
			archiveA = r
			break
		}
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: archiveA.Uuid, EmployeeUuid: employee, SubjectUuid: subject,
		FixedSalary: 1000, Status: "0", EffectiveDate: "2026-01-01", ExpireDate: "2026-05-31",
	}); err != nil {
		t.Fatalf("disable archive A: %v", err)
	}
	_, err = GetEffectiveArchiveForEmployee(ctx, svcCtx, employee, "2026-03")
	assertErrCode(t, err, errcodes.SalaryArchiveNotFoundCode, "both disabled 2026-03")
}

// TestGetEffectiveArchiveInvalidInput 入参守卫: 空 employee_uuid / yearMonth 格式非法.
func TestGetEffectiveArchiveInvalidInput(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	if _, err := GetEffectiveArchiveForEmployee(ctx, svcCtx, "", "2026-01"); err == nil {
		t.Fatalf("empty employee should fail")
	} else {
		assertErrCode(t, err, errcodes.SalaryArchiveInvalidCode, "empty employee")
	}
	for _, bad := range []string{"", "2026", "2026-1", "202601", "abcd-01", "2026-13"} {
		if _, err := GetEffectiveArchiveForEmployee(ctx, svcCtx, "some-emp", bad); err == nil {
			t.Fatalf("bad yearMonth %q should fail", bad)
		} else {
			assertErrCode(t, err, errcodes.SalaryArchiveInvalidCode, "bad yearMonth "+bad)
		}
	}
}
