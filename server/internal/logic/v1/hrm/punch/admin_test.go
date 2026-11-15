package punch

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	employeelogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	employeetypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// seedTwoEmployees 张三(E0001)/李四(E0002) 同在技术部（组织规则 09:00-18:00 周一至周五）。
func seedTwoEmployees(t *testing.T, svcCtx *svc.ServiceContext) (string, string) {
	t.Helper()
	s := seedOrgAndEmployee(t, svcCtx)
	saveOrgConfig(t, svcCtx, s.orgTech, "09:00", "18:00", []int64{1, 2, 3, 4, 5})
	req := &employeetypes.AddRequest{Name: "李四", EmployeeNo: "E0002", OrgUuid: s.orgTech, PersonStatus: "1"}
	if _, err := employeelogic.NewAdd(context.Background(), svcCtx, nil).Add(req); err != nil {
		t.Fatalf("add employee2: %v", err)
	}
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(context.Background(), nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, "E0002").Build()...)
	if err != nil {
		t.Fatalf("find employee2: %v", err)
	}
	return s.empZhang, row.Uuid
}

// TestAdminListFilters 管理端列表: 日期区间 / 姓名(模糊) / 签到状态 筛选 + 分页。
func TestAdminListFilters(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	zhang, li := seedTwoEmployees(t, svcCtx)

	seed := func(employeeUuid, name, date, inStatus string) {
		row := &hrm_punch_record.HrmPunchRecord{
			Uuid: newUuid(), PunchDate: date, EmployeeUuid: employeeUuid,
			EmployeeNo: map[string]string{zhang: "E0001", li: "E0002"}[employeeUuid],
			EmployeeName: name, ClockInStatus: inStatus, Source: SourceImport,
		}
		if inStatus != "" {
			row.ClockInTime = mustTime(date + " 09:00:00")
		}
		if _, err := svcCtx.Model.HrmPunchRecord.Insert(ctx, nil, row); err != nil {
			t.Fatalf("seed row: %v", err)
		}
	}
	seed(zhang, "张三", "2026-11-16", StatusNormal)
	seed(zhang, "张三", "2026-11-17", StatusField)
	seed(li, "李四", "2026-11-16", StatusNormal)
	seed(li, "李四", "2026-11-18", "") // 缺卡

	// 全量 4 行
	resp, err := NewList(ctx, svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if resp.Total != 4 {
		t.Fatalf("total = %d, want 4", resp.Total)
	}

	// 日期区间 [11-16, 11-17]
	resp, err = NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, StartDate: "2026-11-16", EndDate: "2026-11-17",
	})
	if err != nil {
		t.Fatalf("list range: %v", err)
	}
	if resp.Total != 3 {
		t.Fatalf("range total = %d, want 3", resp.Total)
	}

	// 姓名 模糊 "张"
	resp, err = NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, Name: "张",
	})
	if err != nil {
		t.Fatalf("list name: %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("name total = %d, want 2", resp.Total)
	}

	// 签到状态 外勤
	resp, err = NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, ClockInStatus: StatusField,
	})
	if err != nil {
		t.Fatalf("list status: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("status total = %d, want 1", resp.Total)
	}

	// 组合: 姓名 李四 + 缺卡（clockInStatus=missing 约定筛空状态行）
	resp, err = NewList(ctx, svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, Name: "李四", ClockInStatus: SettleMissing,
	})
	if err != nil {
		t.Fatalf("list missing: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("missing total = %d, want 1", resp.Total)
	}
}

// TestAdminAddEditDelete 管理端新增/修改/批量删除打卡记录（补录外部数据单行）。
func TestAdminAddEditDelete(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	zhang, _ := seedTwoEmployees(t, svcCtx)

	// Add: 完整一行（含缺卡签退）
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: zhang, PunchDate: "2026-11-16",
		ClockInTime: "2026-11-16 09:05:00", ClockInStatus: StatusNormal,
	}); err != nil {
		t.Fatalf("add: %v", err)
	}

	row, err := svcCtx.Model.HrmPunchRecord.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_punch_record.PunchDate, "2026-11-16").
			Equal(hrm_punch_record.EmployeeUuid, zhang).Build()...)
	if err != nil {
		t.Fatalf("find added row: %v", err)
	}
	if !row.ClockInTime.Valid || row.ClockInStatus != StatusNormal {
		t.Fatalf("added row in = %v/%q", row.ClockInTime, row.ClockInStatus)
	}
	if row.Source != SourceAdmin {
		t.Fatalf("source = %q, want admin", row.Source)
	}
	// Add 后自动日结
	if row.SettleStatus == "" {
		t.Fatalf("added row should be settled immediately")
	}

	// Add 重复 (punch_date, employee_uuid) → 10543
	_, err = NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: zhang, PunchDate: "2026-11-16", ClockInTime: "2026-11-16 09:00:00",
	})
	if sc := status.FromError(err); sc == nil || sc.Code() != errcodes.PunchDuplicateCode {
		t.Fatalf("duplicate add error = %v, want 10543", err)
	}

	// Edit: 补签退
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: row.Uuid, ClockOutTime: "2026-11-16 18:10:00", ClockOutStatus: StatusNormal,
	}); err != nil {
		t.Fatalf("edit: %v", err)
	}
	row, err = svcCtx.Model.HrmPunchRecord.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_punch_record.Uuid, row.Uuid).Build()...)
	if err != nil {
		t.Fatalf("find edited row: %v", err)
	}
	if !row.ClockOutTime.Valid {
		t.Fatalf("edited row clockOut not written")
	}
	// 上班卡 09:05（规则 09:00 无弹性）→ 迟到 5 分钟
	if row.SettleStatus != SettleLate || row.SettleLateMinutes != 5 {
		t.Fatalf("edited row settle = %q/%d, want late/5", row.SettleStatus, row.SettleLateMinutes)
	}

	// Delete: 批量（uuids），并验证行消失
	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{row.Uuid}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = svcCtx.Model.HrmPunchRecord.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_punch_record.Uuid, row.Uuid).Build()...)
	if err != hrm_punch_record.ErrNotFound {
		t.Fatalf("row should be deleted, err = %v", err)
	}
}

// TestManualSettle 手动日结端点: 批量员工 + 日期区间，落库 settle_* 列。
func TestManualSettle(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	zhang, li := seedTwoEmployees(t, svcCtx)

	seedRow := func(employeeUuid, name, employeeNo, date, in, out string) {
		row := &hrm_punch_record.HrmPunchRecord{
			Uuid: newUuid(), PunchDate: date, EmployeeUuid: employeeUuid,
			EmployeeNo: employeeNo, EmployeeName: name, Source: SourceImport,
		}
		if in != "" {
			row.ClockInTime = mustTime(date + " " + in)
			row.ClockInStatus = StatusNormal
		}
		if out != "" {
			row.ClockOutTime = mustTime(date + " " + out)
			row.ClockOutStatus = StatusNormal
		}
		if _, err := svcCtx.Model.HrmPunchRecord.Insert(ctx, nil, row); err != nil {
			t.Fatalf("seed row: %v", err)
		}
	}
	seedRow(zhang, "张三", "E0001", "2026-11-16", "09:30:00", "18:05:00")
	seedRow(li, "李四", "E0002", "2026-11-17", "09:00:00", "18:00:00")

	resp, err := NewSettle(ctx, svcCtx, nil).Settle(&types.SettleRequest{
		EmployeeUuids: []string{zhang, li}, StartDate: "2026-11-16", EndDate: "2026-11-17",
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	// 2 员工 × 2 天 = 4 行结果（无打卡行也产出 missing）
	if len(resp.Records) != 4 {
		t.Fatalf("settle records = %d, want 4", len(resp.Records))
	}
	byKey := map[string]types.HrmSettleResult{}
	for _, r := range resp.Records {
		byKey[r.EmployeeUuid+r.PunchDate] = r
	}
	if r := byKey[zhang+"2026-11-16"]; r.SettleStatus != SettleLate || r.LateMinutes != 30 {
		t.Fatalf("zhang 11-16 = %+v, want late 30", r)
	}
	if r := byKey[li+"2026-11-17"]; r.SettleStatus != SettleNormal {
		t.Fatalf("li 11-17 = %+v, want normal", r)
	}

	// 落库断言
	rows, err := svcCtx.Model.HrmPunchRecord.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_punch_record.PunchDate, "2026-11-16").
			Equal(hrm_punch_record.EmployeeUuid, zhang).Build()...)
	if err != nil || len(rows) != 1 {
		t.Fatalf("find zhang row: %v", err)
	}
	if rows[0].SettleStatus != SettleLate || rows[0].SettleLateMinutes != 30 {
		t.Fatalf("persisted = %q/%d, want late/30", rows[0].SettleStatus, rows[0].SettleLateMinutes)
	}

	// 无打卡行的工作日也产出 missing 结果（2026-11-18）
	resp, err = NewSettle(ctx, svcCtx, nil).Settle(&types.SettleRequest{
		EmployeeUuids: []string{zhang}, StartDate: "2026-11-18",
	})
	if err != nil {
		t.Fatalf("settle missing day: %v", err)
	}
	if len(resp.Records) != 1 || resp.Records[0].SettleStatus != SettleMissing {
		t.Fatalf("missing day result = %+v, want missing", resp.Records)
	}

	// 日期校验: 非法日期 → 10545
	_, err = NewSettle(ctx, svcCtx, nil).Settle(&types.SettleRequest{StartDate: "2026-13-99"})
	if sc := status.FromError(err); sc == nil || sc.Code() != errcodes.SettleInvalidCode {
		t.Fatalf("bad date error = %v, want 10545", err)
	}
}

// TestImportPunch 导入外部考勤数据（CSV）: 成功计数 + 非法行整批拒绝（10544）。
func TestImportPunch(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	_, li := seedTwoEmployees(t, svcCtx)

	csv := "date,employee_no,clock_in_time,clock_in_status,clock_in_location,clock_out_time,clock_out_status,clock_out_location,overtime\n" +
		"2026-11-16,E0001,2026-11-16 09:00:00,normal,,2026-11-16 18:00:00,normal,,0\n" +
		"2026-11-16,E0002,2026-11-16 09:45:00,field,客户现场,,2026-11-16 18:30:00,field,1\n"
	resp, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "punch.csv", FileContent: csv})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.SuccessCount != 2 {
		t.Fatalf("successCount = %d, want 2", resp.SuccessCount)
	}

	// 行内容验证（李四 外勤 + 加班）
	row, err := svcCtx.Model.HrmPunchRecord.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_punch_record.PunchDate, "2026-11-16").
			Equal(hrm_punch_record.EmployeeUuid, li).Build()...)
	if err != nil {
		t.Fatalf("find imported row: %v", err)
	}
	if row.ClockInStatus != StatusField || row.ClockInLocation != "客户现场" || row.Overtime != OvertimeYes {
		t.Fatalf("imported li row = %+v", row)
	}
	if row.Source != SourceImport {
		t.Fatalf("source = %q, want import", row.Source)
	}

	// 非法行（未知工号）→ 整批拒绝 10544，且不落任何行
	badCsv := "date,employee_no,clock_in_time,clock_in_status,clock_in_location,clock_out_time,clock_out_status,clock_out_location,overtime\n" +
		"2026-11-17,E0001,2026-11-17 09:00:00,normal,,2026-11-17 18:00:00,normal,,0\n" +
		"2026-11-17,E9999,2026-11-17 09:00:00,normal,,,,,\n"
	_, err = NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{FileName: "bad.csv", FileContent: badCsv})
	if sc := status.FromError(err); sc == nil || sc.Code() != errcodes.PunchImportInvalidCode {
		t.Fatalf("bad import error = %v, want 10544", err)
	}
	rows, err := svcCtx.Model.HrmPunchRecord.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_punch_record.PunchDate, "2026-11-17").Build()...)
	if err != nil {
		t.Fatalf("find rows after bad import: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("bad import should not persist any row, got %d", len(rows))
	}
}

// TestExportPunch 导出: 返回 base64 xlsx，含筛选行。
func TestExportPunch(t *testing.T) {
	svcCtx := newSvcCtx(t)
	ctx := context.Background()
	zhang, li := seedTwoEmployees(t, svcCtx)

	seed := func(employeeUuid, name, date string) {
		row := &hrm_punch_record.HrmPunchRecord{
			Uuid: newUuid(), PunchDate: date, EmployeeUuid: employeeUuid,
			EmployeeNo: map[string]string{zhang: "E0001", li: "E0002"}[employeeUuid],
			EmployeeName: name, ClockInStatus: StatusNormal, Source: SourceAdmin,
		}
		row.ClockInTime = mustTime(date + " 09:00:00")
		if _, err := svcCtx.Model.HrmPunchRecord.Insert(ctx, nil, row); err != nil {
			t.Fatalf("seed row: %v", err)
		}
	}
	seed(zhang, "张三", "2026-11-16")
	seed(li, "李四", "2026-11-17")

	resp, err := NewExport(ctx, svcCtx, nil).Export(&types.ExportRequest{Name: "张三"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if resp.FileName == "" || resp.Content == "" {
		t.Fatalf("export file/content empty")
	}
}

func mustTime(s string) (nt sql.NullTime) {
	tm, err := time.ParseInLocation(time.DateTime, s, time.Local)
	if err != nil {
		panic(err)
	}
	return sql.NullTime{Time: tm, Valid: true}
}
