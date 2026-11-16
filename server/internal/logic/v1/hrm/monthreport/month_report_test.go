package monthreport

// HRM-23 考勤月报 logic 测试（svctest sqlite）：
//   - 多员工多状态聚合 worked example（应出勤/实出勤/迟到/早退/缺卡/旷工 + 加班三桶 + 调休入账）；
//   - 锁定快照 + 解锁重算 diff 标记（只看锁定后变更）；
//   - GetMonthReportRows 锁定 vs 未锁定口径（HRM-37 消费契约）；
//   - 导出 xlsx round-trip。

import (
	"context"
	"database/sql"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/xuri/excelize/v2"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/leavebalance"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_month_lock"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_month_report"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/monthreport"
)

// newMrSvcCtx 测试 ServiceContext（Model + SqlxConn，与 punch 同款）。
func newMrSvcCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

// seedEmployee 直接插员工行（不走 employee logic，避免组织级联；org 直插）。
func seedEmployee(t *testing.T, svcCtx *svc.ServiceContext, name, employeeNo, orgUuid string) string {
	t.Helper()
	ctx := context.Background()
	_, err := svcCtx.Model.HrmEmployee.Insert(ctx, nil, &hrm_employee.HrmEmployee{
		Uuid:       newUuid(),
		Name:       name,
		EmployeeNo: employeeNo,
		OrgUuid:    orgUuid,
	})
	if err != nil {
		t.Fatalf("seed employee %s: %v", name, err)
	}
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, employeeNo).Build()...)
	if err != nil {
		t.Fatalf("find employee %s: %v", employeeNo, err)
	}
	return row.Uuid
}

// seedOrg 直插组织行。
func seedOrg(t *testing.T, svcCtx *svc.ServiceContext, name, parentUuid string) string {
	t.Helper()
	ctx := context.Background()
	orgUuid := newUuid()
	_, err := svcCtx.Model.HrmOrg.Insert(ctx, nil, &hrm_org.HrmOrg{
		Uuid:       orgUuid,
		ParentUuid: parentUuid,
		OrgName:    name,
		OrgType:    "2",
		OrderNum:   1,
		Status:     "1",
	})
	if err != nil {
		t.Fatalf("seed org %s: %v", name, err)
	}
	return orgUuid
}

// mrPunchRow 落一行已日结打卡记录（模拟 SettleDay 结果）。
func mrPunchRow(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, employeeNo, name, date, settleStatus string, inTime, outTime string) {
	t.Helper()
	row := &hrm_punch_record.HrmPunchRecord{
		Uuid:         newUuid(),
		PunchDate:    date,
		EmployeeUuid: employeeUuid,
		EmployeeNo:   employeeNo,
		EmployeeName: name,
		SettleStatus: settleStatus,
	}
	if inTime != "" {
		row.ClockInTime = sql.NullTime{Valid: true}
	}
	if outTime != "" {
		row.ClockOutTime = sql.NullTime{Valid: true}
	}
	if _, err := svcCtx.Model.HrmPunchRecord.Insert(context.Background(), nil, row); err != nil {
		t.Fatalf("seed punch row %s/%s: %v", employeeNo, date, err)
	}
}

// mrOvertimeBill 落一张已审批加班单（dayType 1=工作日 2=休息日 3=节假日）。
func mrOvertimeBill(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, date, dayType string, confirmHours, applyHours, accrueHours float64) {
	t.Helper()
	_, err := svcCtx.Model.HrmOvertimeBill.Insert(context.Background(), nil, &hrm_overtime_bill.HrmOvertimeBill{
		Uuid:         newUuid(),
		BillNo:       "HR217-TEST-" + newUuid()[:8],
		Status:       BillStatusApproved,
		EmployeeUuid: employeeUuid,
		OvertimeDate: date,
		DayType:      dayType,
		ApplyHours:   applyHours,
		ConfirmHours: confirmHours,
		AccrueHours:  accrueHours,
	})
	if err != nil {
		t.Fatalf("seed overtime bill: %v", err)
	}
}

// TestAggregationMultiEmployeeMultiStatus 多员工多状态聚合 worked example。
// 2026-11：11-16(一)/11-17(二)/11-18(三) 工作日（默认 WorkDays 1-5），11-21(六) 休息。
//   张三: late / normal / missing(双卡空) + 工作日加班 2h(核定) + 休息日加班 3h + 节假日 1h + 调休入账 2.5h
//     → 应出勤 3, 实出勤 2, 迟到 1, 缺卡 1, 旷工 1
//   李四: missing(有上班卡) / normal → 应出勤 21(全月工作日), 实出勤 1, 缺卡 1, 旷工 0
func TestAggregationMultiEmployeeMultiStatus(t *testing.T) {
	svcCtx := newMrSvcCtx(t)
	ctx := context.Background()
	orgUuid := seedOrg(t, svcCtx, "公司A", "")
	zhang := seedEmployee(t, svcCtx, "张三", "E0001", orgUuid)
	li := seedEmployee(t, svcCtx, "李四", "E0002", orgUuid)

	// 张三
	mrPunchRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", "late", "09:30", "18:00")
	mrPunchRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-17", "normal", "09:00", "18:00")
	mrPunchRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-18", "missing", "", "")
	mrOvertimeBill(t, svcCtx, zhang, "2026-11-16", DayTypeWorkday, 2, 3, 2)
	mrOvertimeBill(t, svcCtx, zhang, "2026-11-21", DayTypeRestday, 0, 3, 3)
	mrOvertimeBill(t, svcCtx, zhang, "2026-11-01", DayTypeHoliday, 0, 1, 1)
	// 李四
	mrPunchRow(t, svcCtx, li, "E0002", "李四", "2026-11-16", "missing", "09:00", "")
	mrPunchRow(t, svcCtx, li, "E0002", "李四", "2026-11-17", "normal", "09:00", "18:00")

	rows, _, err := GetMonthReportRows(ctx, svcCtx, "2026-11", nil)
	if err != nil {
		t.Fatalf("GetMonthReportRows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0].EmployeeNo != "E0001" || rows[1].EmployeeNo != "E0002" {
		t.Fatalf("rows not sorted by employee no: %s, %s", rows[0].EmployeeNo, rows[1].EmployeeNo)
	}
	zr, lr := rows[0], rows[1]

	// 张三（无规则兜底周一至周五 → 全月应出勤 21；2026-11 有 21 个周一至周五）
	if zr.ExpectedDays != 21 || zr.AttendDays != 2 || zr.LateCount != 1 || zr.EarlyCount != 0 ||
		zr.MissingCount != 1 || zr.AbsentDays != 1 {
		t.Fatalf("zhang agg wrong: %+v", zr)
	}
	if zr.WorkdayOtH != 2 || zr.RestdayOtH != 3 || zr.HolidayOtH != 1 || zr.CompAccrueH != 6 {
		t.Fatalf("zhang overtime wrong: %+v", zr)
	}
	// 李四：missing 单侧卡 → 缺卡不算旷工
	if lr.ExpectedDays != 21 || lr.AttendDays != 1 || lr.MissingCount != 1 || lr.AbsentDays != 0 {
		t.Fatalf("li agg wrong: %+v", lr)
	}
	if lr.WorkdayOtH != 0 || lr.CompAccrueH != 0 {
		t.Fatalf("li overtime wrong: %+v", lr)
	}
}

// TestLockSnapshotAndChangedDiff 锁定快照 → 锁定后重算 diff 标记 → 解锁重算覆盖。
func TestLockSnapshotAndChangedDiff(t *testing.T) {
	svcCtx := newMrSvcCtx(t)
	ctx := context.Background()
	orgUuid := seedOrg(t, svcCtx, "公司A", "")
	zhang := seedEmployee(t, svcCtx, "张三", "E0001", orgUuid)
	li := seedEmployee(t, svcCtx, "李四", "E0002", orgUuid)

	mrPunchRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", "normal", "09:00", "18:00")
	mrPunchRow(t, svcCtx, li, "E0002", "李四", "2026-11-16", "normal", "09:00", "18:00")

	// 锁定 → 快照落库
	if err := LockMonth(ctx, svcCtx, "2026-11", "op-uuid", "管理员", "月度锁定"); err != nil {
		t.Fatalf("LockMonth: %v", err)
	}
	locked, err := IsMonthLocked(ctx, svcCtx, "2026-11")
	if err != nil || !locked {
		t.Fatalf("month should be locked: %v %v", locked, err)
	}
	snapRows, _, err := GetMonthReportRows(ctx, svcCtx, "2026-11", nil)
	if err != nil || len(snapRows) != 2 {
		t.Fatalf("snapshot rows: %v %v", len(snapRows), err)
	}

	// 锁定后数据变更（李四 11-17 迟到）
	mrPunchRow(t, svcCtx, li, "E0002", "李四", "2026-11-17", "late", "09:30", "18:00")

	// 锁定月重算：快照值不变，李四行 changed=1
	locked, changedCount, err := RecalcMonth(ctx, svcCtx, "2026-11")
	if err != nil || !locked {
		t.Fatalf("RecalcMonth: %v %v", locked, err)
	}
	if changedCount != 1 {
		t.Fatalf("changedCount want 1 got %d", changedCount)
	}
	snaps, err := svcCtx.Model.HrmAttendanceMonthReport.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_attendance_month_report.YearMonth, "2026-11").
			OrderByAsc(hrm_attendance_month_report.EmployeeNo).Build()...)
	if err != nil {
		t.Fatal(err)
	}
	if snaps[0].EmployeeNo != "E0001" || snaps[0].Changed != 0 {
		t.Fatalf("zhang changed should be 0: %+v", snaps[0])
	}
	if snaps[1].EmployeeNo != "E0002" || snaps[1].Changed != 1 {
		t.Fatalf("li changed should be 1: %+v", snaps[1])
	}
	// 锁定月快照数值不覆盖（李四 attend_days 仍为锁定时的 1）
	if snaps[1].AttendDays != 1 {
		t.Fatalf("locked snapshot should not be overwritten: attend=%d", snaps[1].AttendDays)
	}

	// 锁定记录流水（lock + recalc）
	lockRecs, err := svcCtx.Model.HrmAttendanceMonthLock.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_attendance_month_lock.YearMonth, "2026-11").Build()...)
	if err != nil || len(lockRecs) != 2 {
		t.Fatalf("lock records: %v %v", len(lockRecs), err)
	}

	// 解锁 → 快照删除 → 重算覆盖为实时聚合（changed 清零）
	if err := UnlockMonth(ctx, svcCtx, "2026-11", "op-uuid", "管理员", "复核解锁"); err != nil {
		t.Fatalf("UnlockMonth: %v", err)
	}
	locked, err = IsMonthLocked(ctx, svcCtx, "2026-11")
	if err != nil || locked {
		t.Fatalf("month should be unlocked: %v %v", locked, err)
	}
	if _, _, err := RecalcMonth(ctx, svcCtx, "2026-11"); err != nil {
		t.Fatalf("recalc after unlock: %v", err)
	}
	snaps, err = svcCtx.Model.HrmAttendanceMonthReport.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_attendance_month_report.YearMonth, "2026-11").
			OrderByAsc(hrm_attendance_month_report.EmployeeNo).Build()...)
	if err != nil || len(snaps) != 2 {
		t.Fatalf("rows after unlock recalc: %v %v", len(snaps), err)
	}
	if snaps[1].AttendDays != 2 || snaps[1].LateCount != 1 || snaps[1].Changed != 0 {
		t.Fatalf("after unlock recalc li should reflect realtime: %+v", snaps[1])
	}
	// 流水四条 lock/recalc/unlock/recalc(解锁后缓存刷新)
	lockRecs, _ = svcCtx.Model.HrmAttendanceMonthLock.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_attendance_month_lock.YearMonth, "2026-11").Build()...)
	if len(lockRecs) != 4 {
		t.Fatalf("want 4 lock records, got %d", len(lockRecs))
	}
}

// TestGetMonthReportRowsLockedVsUnlocked GetMonthReportRows 锁定/未锁定口径。
func TestGetMonthReportRowsLockedVsUnlocked(t *testing.T) {
	svcCtx := newMrSvcCtx(t)
	ctx := context.Background()
	orgUuid := seedOrg(t, svcCtx, "公司A", "")
	orgChild := seedOrg(t, svcCtx, "技术部", orgUuid)
	zhang := seedEmployee(t, svcCtx, "张三", "E0001", orgChild)

	mrPunchRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", "normal", "09:00", "18:00")

	// 未锁定：unlocked=false，实时聚合
	rows, locked, err := GetMonthReportRows(ctx, svcCtx, "2026-11", nil)
	if err != nil || locked {
		t.Fatalf("unlocked month: locked=%v err=%v", locked, err)
	}
	if len(rows) != 1 || rows[0].AttendDays != 1 {
		t.Fatalf("unlocked rows: %+v err=%v", rows, err)
	}

	// orgUuids 过滤（父部门含下级由调用方展开后传入——此处直接传子部门 uuid）
	rows, _, err = GetMonthReportRows(ctx, svcCtx, "2026-11", []string{orgChild})
	if err != nil || len(rows) != 1 {
		t.Fatalf("org filter: %+v err=%v", rows, err)
	}
	rows, _, err = GetMonthReportRows(ctx, svcCtx, "2026-11", []string{"org-not-match"})
	if err != nil || len(rows) != 0 {
		t.Fatalf("org filter no match: %+v err=%v", rows, err)
	}

	// 锁定后：locked=true，返回快照行
	if err := LockMonth(ctx, svcCtx, "2026-11", "", "", ""); err != nil {
		t.Fatal(err)
	}
	mrPunchRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-17", "late", "09:30", "18:00")
	rows, locked, err = GetMonthReportRows(ctx, svcCtx, "2026-11", nil)
	if err != nil || !locked {
		t.Fatalf("locked month: locked=%v err=%v", locked, err)
	}
	if len(rows) != 1 || rows[0].AttendDays != 1 || rows[0].LateCount != 0 {
		t.Fatalf("locked snapshot rows should not include post-lock change: %+v", rows)
	}

	// 非法月份
	if _, _, err := GetMonthReportRows(ctx, svcCtx, "2026/11", nil); err == nil {
		t.Fatal("invalid month should error")
	}
}

// TestExportRoundTrip 导出 xlsx round-trip（base64 → excelize 读回校验）。
func TestExportRoundTrip(t *testing.T) {
	svcCtx := newMrSvcCtx(t)
	ctx := context.Background()
	orgUuid := seedOrg(t, svcCtx, "公司A", "")
	zhang := seedEmployee(t, svcCtx, "张三", "E0001", orgUuid)
	mrPunchRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", "normal", "09:00", "18:00")

	req := &types.ExportRequest{YearMonth: "2026-11"}
	resp, err := NewExport(ctx, svcCtx, nil).Export(req)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if !strings.HasSuffix(resp.FileName, ".xlsx") {
		t.Fatalf("file name: %s", resp.FileName)
	}
	raw, err := base64.StdEncoding.DecodeString(resp.Content)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	f, err := excelize.OpenReader(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	sheet := f.GetSheetName(0)
	head, _ := f.GetRows(sheet)
	if len(head) == 0 || head[0][2] != "姓名" {
		t.Fatalf("header: %v", head[0])
	}
	rows, _ := f.GetRows(sheet)
	if len(rows) != 2 || rows[1][2] != "张三" {
		t.Fatalf("data rows: %v", rows)
	}
	if rows[1][4] != "21" || rows[1][5] != "1" {
		t.Fatalf("expected/attend: %v", rows[1])
	}
}

// TestListOnlyAbnormalAndChangedFilters 列表过滤：只看异常 / 只看锁定后变更 / 考勤组名。
func TestListOnlyAbnormalAndChangedFilters(t *testing.T) {
	svcCtx := newMrSvcCtx(t)
	ctx := context.Background()
	orgUuid := seedOrg(t, svcCtx, "公司A", "")
	zhang := seedEmployee(t, svcCtx, "张三", "E0001", orgUuid)
	li := seedEmployee(t, svcCtx, "李四", "E0002", orgUuid)
	mrPunchRow(t, svcCtx, zhang, "E0001", "张三", "2026-11-16", "normal", "09:00", "18:00")
	mrPunchRow(t, svcCtx, li, "E0002", "李四", "2026-11-16", "late", "09:30", "18:00")

	listLogic := NewList(ctx, svcCtx, nil)

	// 只看异常 → 仅李四
	resp, err := listLogic.List(&types.ListRequest{YearMonth: "2026-11", OnlyAbnormal: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 || resp.Records[0].EmployeeName != "李四" || !resp.Records[0].Changed == false {
		t.Fatalf("onlyAbnormal: %+v", resp.Records)
	}
	if resp.Locked {
		t.Fatal("should be unlocked")
	}
	if resp.LockHint == "" {
		t.Fatal("lock hint empty")
	}

	// 锁定 → 李四迟到 → 重算 diff → 只看锁定后变更
	if err := LockMonth(ctx, svcCtx, "2026-11", "", "", ""); err != nil {
		t.Fatal(err)
	}
	mrPunchRow(t, svcCtx, li, "E0002", "李四", "2026-11-17", "missing", "", "")
	if _, _, err := RecalcMonth(ctx, svcCtx, "2026-11"); err != nil {
		t.Fatal(err)
	}
	resp, err = listLogic.List(&types.ListRequest{YearMonth: "2026-11", OnlyChanged: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 || resp.Records[0].EmployeeName != "李四" || !resp.Records[0].Changed {
		t.Fatalf("onlyChanged: %+v", resp.Records)
	}
	if !resp.Locked {
		t.Fatal("should be locked")
	}

	// 非法月份 → 10670
	if _, err := listLogic.List(&types.ListRequest{YearMonth: "abc"}); err == nil {
		t.Fatal("invalid month should error")
	}

	_ = leavebalance.CompLeaveCode // 保留 import（调休入账口径来源）
}
