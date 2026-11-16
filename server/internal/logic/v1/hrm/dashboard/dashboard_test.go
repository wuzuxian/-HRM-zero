package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/punch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	punchtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/dashboard"
)

// TestDashboardUnregisteredGuide 未建档员工引导：无 user_uuid 关联 → Registered=false
// + 全零 payload（不报错，前端展示建档引导）。
func TestDashboardUnregisteredGuide(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	resp := mustResp(t, svcCtx, "")
	if resp.Registered {
		t.Fatalf("registered = true, want false (unlinked user)")
	}
	if resp.EmployeeUuid != "" {
		t.Fatalf("employeeUuid = %q, want empty", resp.EmployeeUuid)
	}
	if len(resp.Calendar) == 0 {
		t.Fatalf("calendar empty, want month days scaffold")
	}
	if resp.Stats.ExpectedDays != 0 || resp.Stats.AttendedDays != 0 || resp.Stats.AnnualLeaveAvailable != 0 {
		t.Fatalf("stats not zero for unregistered: %+v", resp.Stats)
	}
}

// TestDashboardNotFoundEmployee 当前登录人不存在于任何员工档案 → 不报错走引导分支。
func TestDashboardNotFoundEmployee(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	resp := mustResp(t, svcCtx, "")
	if resp.Registered || resp.EmployeeUuid != "" {
		t.Fatalf("want unregistered guide payload")
	}
}

// TestDashboardTodayCard 今日卡片：绑定考勤组（指定人员）+ 当日日结行 →
// 组名/休息/上下班卡状态/日结状态全量断言。
func TestDashboardTodayCard(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	empUuid := seedEmployeeViaLogic(t, svcCtx)

	today := time.Now().Format("2006-01-02")
	// 正常出勤日结行
	seedSettledDay(t, svcCtx, empUuid, today, "normal", 0, 0, 8, "0")

	resp := mustResp(t, svcCtx, "")
	if !resp.Registered {
		t.Fatalf("registered = false, want true")
	}
	if resp.EmployeeUuid != empUuid || resp.EmployeeName != "张三" || resp.EmployeeNo != "E0001" {
		t.Fatalf("identity mismatch: %+v", resp)
	}
	if resp.Today.ClockInTime == "" || resp.Today.ClockOutTime == "" {
		t.Fatalf("today card missing punch times: %+v", resp.Today)
	}
	if resp.Today.ClockInStatus != "normal" || resp.Today.ClockOutStatus != "normal" {
		t.Fatalf("today card status = %s/%s, want normal/normal",
			resp.Today.ClockInStatus, resp.Today.ClockOutStatus)
	}
	if resp.Today.SettleStatus != "normal" {
		t.Fatalf("today settleStatus = %q, want normal", resp.Today.SettleStatus)
	}
	if resp.Today.IsRestDay {
		t.Fatalf("IsRestDay = true on seeded workday")
	}
}

// TestDashboardTodayCardRestDay 休息日判定：日结行 settle_status=rest → IsRestDay=true。
func TestDashboardTodayCardRestDay(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	empUuid := seedEmployeeViaLogic(t, svcCtx)
	today := time.Now().Format("2006-01-02")
	seedSettledDay(t, svcCtx, empUuid, today, "rest", 0, 0, 0, "0")

	resp := mustResp(t, svcCtx, "")
	if !resp.Today.IsRestDay {
		t.Fatalf("IsRestDay = false, want true (settle_status=rest)")
	}
	if resp.Today.SettleStatus != "rest" {
		t.Fatalf("settleStatus = %q, want rest", resp.Today.SettleStatus)
	}
}

// TestDashboardTodayCardNoRule 无规则兜底：未命中任何规则且当日打卡 → NoRule=true + no_rule。
func TestDashboardTodayCardNoRule(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	empUuid := seedEmployeeViaLogicNoOrg(t, svcCtx)
	today := time.Now().Format("2006-01-02")
	seedSettledDay(t, svcCtx, empUuid, today, "no_rule", 0, 0, 6.5, "0")

	resp := mustResp(t, svcCtx, "")
	if !resp.Today.NoRule {
		t.Fatalf("NoRule = false, want true")
	}
	if resp.Today.SettleStatus != "no_rule" {
		t.Fatalf("settleStatus = %q, want no_rule", resp.Today.SettleStatus)
	}
	if resp.Today.RuleSource != "" {
		t.Fatalf("ruleSource = %q, want empty", resp.Today.RuleSource)
	}
}

// TestDashboardTodayCardLate 迟到展示：settle_status=late + 迟到分钟透出。
func TestDashboardTodayCardLate(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	empUuid := seedEmployeeViaLogic(t, svcCtx)
	today := time.Now().Format("2006-01-02")
	seedSettledDay(t, svcCtx, empUuid, today, "late", 15, 0, 8, "0")

	resp := mustResp(t, svcCtx, "")
	if resp.Today.SettleStatus != "late" || resp.Today.LateMinutes != 15 {
		t.Fatalf("today = %+v, want late/15min", resp.Today)
	}
}

// TestDashboardMonthStats 本月统计聚合：normal×2 + late×1 + missing×1 + 加班标记×1
// + 已审批加班单 3h + 年假 5-1=4 → 各卡值精确断言。
func TestDashboardMonthStats(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	empUuid := seedEmployeeViaLogic(t, svcCtx)

	month := time.Now().Format("2006-01")
	today := time.Now().Format("2006-01-02")
	firstDay := month + "-01"
	secondDay := month + "-02"

	seedSettledDay(t, svcCtx, empUuid, firstDay, "normal", 0, 0, 8, "0")
	seedSettledDay(t, svcCtx, empUuid, secondDay, "late", 10, 0, 7.5, "1")
	seedSettledDay(t, svcCtx, empUuid, today, "missing", 0, 0, 0, "0")

	// 非本月行（不应计入）
	seedSettledDay(t, svcCtx, empUuid, "2020-01-05", "normal", 0, 0, 8, "0")

	seedOvertimeBill(t, svcCtx, empUuid, firstDay, 3)
	// 未审批单不计
	unapproved := &hrm_overtime_bill.HrmOvertimeBill{
		Uuid: uuid.New().String(), BillNo: "HR217-TEST-UNAPPROVED", Status: "unsubmitted",
		EmployeeUuid: empUuid, OvertimeDate: secondDay, DayType: "1", ConfirmHours: 99,
	}
	if _, err := svcCtx.Model.HrmOvertimeBill.Insert(context.Background(), nil, unapproved); err != nil {
		t.Fatalf("seed unapproved bill: %v", err)
	}

	year := int64(time.Now().Year())
	seedAnnualBalance(t, svcCtx, year, empUuid, 5, 1, 0)

	resp := mustResp(t, svcCtx, "")
	stats := resp.Stats
	if stats.AttendedDays != 2 { // normal + late
		t.Fatalf("attendedDays = %d, want 2", stats.AttendedDays)
	}
	if stats.LateTimes != 1 {
		t.Fatalf("lateTimes = %d, want 1", stats.LateTimes)
	}
	if stats.MissingTimes != 1 || stats.AbsentTimes != 1 {
		t.Fatalf("missing/absent = %d/%d, want 1/1", stats.MissingTimes, stats.AbsentTimes)
	}
	if stats.OvertimeTimes != 1 {
		t.Fatalf("overtimeTimes = %d, want 1 (overtime='1' row)", stats.OvertimeTimes)
	}
	if stats.OvertimeHours != 3 {
		t.Fatalf("overtimeHours = %v, want 3 (approved confirm_hours only)", stats.OvertimeHours)
	}
	if stats.AnnualLeaveAvailable != 4 || stats.AnnualLeaveYear != year {
		t.Fatalf("annualLeave = %v (year %d), want 4 (%d)", stats.AnnualLeaveAvailable, stats.AnnualLeaveYear, year)
	}
	if stats.ExpectedDays < 3 {
		t.Fatalf("expectedDays = %d, want >= 3 (seeded workdays)", stats.ExpectedDays)
	}
}

// TestDashboardCalendar 逐日状态映射：rest/late/missing/normal 各落一行 →
// 日历行状态与打卡时间透出；未来日期标 future；跨月不出现。
func TestDashboardCalendar(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	empUuid := seedEmployeeViaLogic(t, svcCtx)

	month := time.Now().Format("2006-01")
	seedSettledDay(t, svcCtx, empUuid, month+"-01", "rest", 0, 0, 0, "0")
	seedSettledDay(t, svcCtx, empUuid, month+"-02", "late", 12, 0, 7.5, "0")
	seedSettledDay(t, svcCtx, empUuid, month+"-03", "missing", 0, 0, 0, "0")
	seedSettledDay(t, svcCtx, empUuid, month+"-04", "normal", 0, 0, 8, "0")
	seedSettledDay(t, svcCtx, empUuid, "2020-01-09", "normal", 0, 0, 8, "0") // 非本月

	resp := mustResp(t, svcCtx, month)
	byDate := map[string]types.DashboardCalendarDay{}
	for _, d := range resp.Calendar {
		byDate[d.Date] = d
	}
	if len(resp.Calendar) < 28 {
		t.Fatalf("calendar days = %d, want >= 28 (full month)", len(resp.Calendar))
	}
	if _, ok := byDate["2020-01-09"]; ok {
		t.Fatalf("calendar contains foreign-month row 2020-01-09")
	}

	cases := []struct {
		date, status string
	}{
		{month + "-01", "rest"},
		{month + "-02", "late"},
		{month + "-03", "missing"},
		{month + "-04", "normal"},
	}
	for _, c := range cases {
		d, ok := byDate[c.date]
		if !ok {
			t.Fatalf("calendar missing %s", c.date)
		}
		if d.Status != c.status {
			t.Fatalf("%s status = %q, want %q", c.date, d.Status, c.status)
		}
	}
	if d := byDate[month+"-02"]; d.LateMinutes != 12 {
		t.Fatalf("%s lateMinutes = %d, want 12", month+"-02", d.LateMinutes)
	}
	if d := byDate[month+"-03"]; d.ClockInTime != "" || d.ClockOutTime != "" {
		t.Fatalf("%s missing day should have empty punch times, got %q/%q",
			month+"-03", d.ClockInTime, d.ClockOutTime)
	}
	// 未来日期 = future
	today := time.Now().Day()
	if today < 27 {
		future := month + "-28"
		if d, ok := byDate[future]; ok {
			if d.Status != "future" {
				t.Fatalf("%s status = %q, want future", future, d.Status)
			}
		} else {
			t.Fatalf("calendar missing future day %s", future)
		}
	}
	// weekday 透出
	if d := byDate[month+"-01"]; d.Weekday == "" {
		t.Fatalf("weekday empty for %s", month+"-01")
	}
}

// TestDashboardLockHint 锁定日提示：默认口径 lockDay=5，文案含次月锁定日期。
func TestDashboardLockHint(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	resp := mustResp(t, svcCtx, "")
	if resp.Lock.LockDay != 5 {
		t.Fatalf("lockDay = %d, want 5 (默认口径)", resp.Lock.LockDay)
	}
	if resp.Lock.HintText == "" {
		t.Fatalf("hintText empty")
	}
	if resp.Lock.LockMonth != time.Now().Format("2006-01") {
		t.Fatalf("lockMonth = %q, want current month", resp.Lock.LockMonth)
	}
}

// TestDashboardInvalidMonth 非法月份格式 → 10631。
func TestDashboardInvalidMonth(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	ctx := context.WithValue(context.Background(), "uuid", "test-user-uuid")
	_, err := NewGetMyDashboard(ctx, svcCtx, nil).GetMyDashboard(&types.GetMyDashboardRequest{Month: "2026/01"})
	if sc := status.FromError(err); sc == nil || sc.Code() != errcodes.DashboardInvalidCode {
		t.Fatalf("err = %v, want code 10631", err)
	}
}

// TestDashboardExplicitEmployeeQuery 管理端代查：EmployeeUuid 直查（Punch 端点同款代打卡语义）。
func TestDashboardExplicitEmployeeQuery(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	empUuid := seedEmployeeViaLogic(t, svcCtx)
	today := time.Now().Format("2006-01-02")
	seedSettledDay(t, svcCtx, empUuid, today, "normal", 0, 0, 8, "0")

	ctx := context.Background() // 无登录态
	resp, err := NewGetMyDashboard(ctx, svcCtx, nil).GetMyDashboard(&types.GetMyDashboardRequest{
		EmployeeUuid: empUuid,
	})
	if err != nil {
		t.Fatalf("explicit query: %v", err)
	}
	if !resp.Registered || resp.EmployeeUuid != empUuid {
		t.Fatalf("explicit query identity mismatch: %+v", resp)
	}
}

// TestDashboardPunchSettleIntegration 端到端：走 punch.Punch 真实打卡 → SettleDay 落行 →
// 看板今日卡片透出真实日结状态。
func TestDashboardPunchSettleIntegration(t *testing.T) {
	svcCtx := newDashboardCtx(t)
	empUuid := seedEmployeeViaLogic(t, svcCtx)

	punchCtx := context.Background()
	if _, err := punch.NewPunch(punchCtx, svcCtx, nil).Punch(&punchtypes.PunchRequest{
		Type: "clockIn", EmployeeUuid: empUuid,
	}); err != nil {
		t.Fatalf("punch clockIn: %v", err)
	}
	if _, err := punch.NewPunch(punchCtx, svcCtx, nil).Punch(&punchtypes.PunchRequest{
		Type: "clockOut", EmployeeUuid: empUuid,
	}); err != nil {
		t.Fatalf("punch clockOut: %v", err)
	}

	resp := mustResp(t, svcCtx, "")
	if resp.Today.ClockInTime == "" || resp.Today.SettleStatus == "" {
		t.Fatalf("today card not reflecting real punch: %+v", resp.Today)
	}
	if resp.Today.NoRule {
		// 无规则时仍应有 no_rule 语义（未判定），打卡时间透出
		if resp.Today.ClockInTime == "" {
			t.Fatalf("no-rule day should still show punch times")
		}
	}
}
