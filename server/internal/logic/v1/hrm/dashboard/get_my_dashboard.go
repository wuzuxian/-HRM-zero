package dashboard

// 个人考勤看板（HRM-17，demo-inventory 4.1 我的假勤-个人看板）：
// 单端点聚合 今日卡片 + 本月统计 + 锁定日提示 + 考勤日历。
// 统计口径（hrm-notes/hrm-16-notes.md 消费契约）：全部从 hrm_punch_record settle_* 列聚合；
// 年假可用从 hrm_leave_balance（granted − used − frozen）聚合；
// 加班时长从 hrm_overtime_bill 已审批通过单据 confirm_hours 聚合。
// 未建档员工不报错，返回 Registered=false + 空身份 + 全月日历骨架（前端展示建档引导）。

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendancegroup"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/punch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/dashboard"
)

// DefaultLockDay 考勤锁定日默认口径：每月 5 号锁定当月月报（HRM-23 月报落地前的展示值）。
const DefaultLockDay = 5

// AnnualLeaveCode 年假类型编码（hrm_leave_type 预置 annual_leave）。
const AnnualLeaveCode = "annual_leave"

type GetMyDashboard struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetMyDashboard(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetMyDashboard {
	return &GetMyDashboard{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *GetMyDashboard) GetMyDashboard(req *types.GetMyDashboardRequest) (resp *types.GetMyDashboardResponse, err error) {
	// 月份参数（缺省=当前月）
	month := req.Month
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	monthTime, parseErr := time.ParseInLocation("2006-01", month, time.Local)
	if parseErr != nil {
		return nil, status.ErrorMessage(errcodes.DashboardInvalidCode,
			"月份参数非法（应为 yyyy-MM）: "+month)
	}
	monthStart := monthTime
	monthEnd := monthTime.AddDate(0, 1, -1)
	today := time.Now().Format(punch.DateLayout)

	resp = &types.GetMyDashboardResponse{
		Calendar: make([]types.DashboardCalendarDay, 0, 31),
	}

	// 定位员工：管理端代查 employeeUuid 优先，否则当前登录人 → 员工档案。
	employeeUuid := req.EmployeeUuid
	if employeeUuid == "" {
		user := hrmuser.Current(l.ctx)
		if user.Uuid != "" {
			row, findErr := l.svcCtx.Model.HrmEmployee.FindOneByCondition(l.ctx, nil,
				condition.NewChain().Equal(hrm_employee.UserUuid, user.Uuid).Build()...)
			if findErr != nil && findErr != hrm_employee.ErrNotFound {
				return nil, findErr
			}
			if findErr == nil {
				employeeUuid = row.Uuid
			}
		}
	}

	// 未建档（无关联员工）：返回全月日历骨架 + 零值统计，前端展示建档引导。
	if employeeUuid == "" {
		resp.Lock = buildLockHint(monthTime)
		resp.Calendar = buildCalendarSkeleton(monthStart, monthEnd, today, nil)
		return resp, nil
	}

	employee, err := l.svcCtx.Model.HrmEmployee.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, employeeUuid).Build()...)
	if err != nil {
		return nil, err
	}
	resp.EmployeeUuid = employee.Uuid
	resp.EmployeeName = employee.Name
	resp.EmployeeNo = employee.EmployeeNo
	resp.Registered = true

	// 生效规则（三级回退唯一入口，HRM-12 契约）。
	rule, err := attendancegroup.ResolveEffectiveRule(l.ctx, l.svcCtx, employee)
	if err != nil {
		return nil, err
	}

	// 当月打卡/日结行（一人一天一行，HRM-16 契约）。
	punchRows, err := l.svcCtx.Model.HrmPunchRecord.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_punch_record.EmployeeUuid, employeeUuid).
			GreaterEqualThan(hrm_punch_record.PunchDate, monthStart.Format(punch.DateLayout)).
			LessEqualThan(hrm_punch_record.PunchDate, monthEnd.Format(punch.DateLayout)).
			OrderByAsc(hrm_punch_record.PunchDate).
			Build()...)
	if err != nil {
		return nil, err
	}
	rowsByDate := make(map[string]*hrm_punch_record.HrmPunchRecord, len(punchRows))
	for _, row := range punchRows {
		rowsByDate[row.PunchDate] = row
	}

	resp.Today = buildTodayCard(rule, rowsByDate[today], today)

	// 今日无日结行且今天在统计月内 → 即时日结固化判定（缺卡/休息/无规则也落行），
	// 失败不阻断看板（记日志降级）。
	if rowsByDate[today] == nil && today >= monthStart.Format(punch.DateLayout) && today <= monthEnd.Format(punch.DateLayout) {
		if result, settleErr := punch.SettleDay(l.ctx, l.svcCtx, employeeUuid, today); settleErr == nil && result != nil {
			resp.Today = buildTodayCard(rule, nil, today)
			applyDaySettleToToday(&resp.Today, result)
		} else {
			l.Errorf("dashboard on-the-fly settle %s: %v", today, settleErr)
		}
	}

	resp.Stats = buildMonthStats(rule, rowsByDate, monthStart, monthEnd, today, employeeUuid, l)
	resp.Lock = buildLockHint(monthTime)
	resp.Calendar = buildCalendarSkeleton(monthStart, monthEnd, today, rowsByDate)
	return resp, nil
}

// buildTodayCard 今日卡片（考勤组名/休息/上下班卡/日结状态）。
func buildTodayCard(rule *attendancegroup.ResolvedRule, row *hrm_punch_record.HrmPunchRecord, today string) types.DashboardTodayCard {
	card := types.DashboardTodayCard{}
	if rule != nil && rule.Found {
		card.GroupName = rule.SourceName
		card.RuleSource = rule.Source
		card.ScheduledStart = rule.StartTime
		card.ScheduledEnd = rule.EndTime
		// 休息日判定与 SettleDay 同口径：weekday ∉ WorkDays。
		if !containsInt64(rule.WorkDays, weekdayOf(today)) {
			card.IsRestDay = true
		}
	} else {
		card.NoRule = true
	}
	if row != nil {
		card.ClockInTime = formatNullTime(row.ClockInTime)
		card.ClockOutTime = formatNullTime(row.ClockOutTime)
		card.ClockInStatus = punchStatusOf(row.ClockInTime, row.ClockInStatus)
		card.ClockOutStatus = punchStatusOf(row.ClockOutTime, row.ClockOutStatus)
		card.SettleStatus = row.SettleStatus
		card.LateMinutes = row.SettleLateMinutes
		card.EarlyMinutes = row.SettleEarlyMinutes
		card.WorkHours = row.SettleWorkHours
		if card.SettleStatus == punch.SettleRest {
			card.IsRestDay = true
		}
		if card.SettleStatus == punch.SettleNoRule {
			card.NoRule = true
		}
	}
	return card
}

// applyDaySettleToToday 即时日结结果透出（无打卡行时 SettleDay 的缺卡/no_rule 判定）。
func applyDaySettleToToday(card *types.DashboardTodayCard, result *punch.DaySettleResult) {
	card.SettleStatus = result.Status
	card.LateMinutes = result.LateMinutes
	card.EarlyMinutes = result.EarlyMinutes
	card.WorkHours = result.WorkHours
	card.RuleSource = result.RuleSource
	if result.Status == punch.SettleRest {
		card.IsRestDay = true
	}
	if result.Status == punch.SettleNoRule || !result.Found {
		card.NoRule = true
	}
}

// buildMonthStats 本月统计（全部 settle_* 聚合口径）。
func buildMonthStats(rule *attendancegroup.ResolvedRule, rowsByDate map[string]*hrm_punch_record.HrmPunchRecord,
	monthStart, monthEnd time.Time, today, employeeUuid string, l *GetMyDashboard) types.DashboardMonthStats {
	stats := types.DashboardMonthStats{}

	// 年假可用（granted − used − frozen，HRM-18 口径）。
	year := int64(monthStart.Year())
	stats.AnnualLeaveYear = year
	balances, err := l.svcCtx.Model.HrmLeaveBalance.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_leave_balance.EmployeeUuid, employeeUuid).
			Equal(hrm_leave_balance.Year, year).
			Equal(hrm_leave_balance.LeaveTypeCode, AnnualLeaveCode).
			Build()...)
	if err != nil {
		l.Errorf("dashboard annual balance: %v", err)
	} else {
		for _, b := range balances {
			stats.AnnualLeaveAvailable += b.Granted - b.Used - b.Frozen
		}
	}

	// 加班时长：已审批通过加班单 confirm_hours 合计（HRM-21）。
	overtimeBills, err := l.svcCtx.Model.HrmOvertimeBill.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_overtime_bill.EmployeeUuid, employeeUuid).
			Equal(hrm_overtime_bill.Status, bill.StatusApproved).
			Like(hrm_overtime_bill.OvertimeDate, monthStart.Format("2006-01")+"%").
			Build()...)
	if err != nil {
		l.Errorf("dashboard overtime bills: %v", err)
	} else {
		for _, b := range overtimeBills {
			stats.OvertimeHours += b.ConfirmHours
		}
	}

	// 逐日聚合 settle_* 行（正常/迟到/早退算出勤；迟到早退并存按 late 行计，早期分钟另列）。
	for _, row := range rowsByDate {
		switch row.SettleStatus {
		case punch.SettleNormal:
			stats.AttendedDays++
		case punch.SettleLate:
			stats.AttendedDays++
			stats.LateTimes++
		case punch.SettleEarly:
			stats.AttendedDays++
			stats.EarlyTimes++
		case punch.SettleMissing:
			// 缺勤口径 = 缺卡天数（HRM-16 契约：无行工作日也落 missing 行）。
			stats.MissingTimes++
			stats.AbsentTimes++
		}
		if row.Overtime == "1" {
			stats.OvertimeTimes++
		}
	}

	// 应出勤天数：有规则 → 当月（至今天）weekday ∈ WorkDays 的天数；
	// 无规则 → 已落行的 settle 行数兜底。
	if rule != nil && rule.Found {
		end := monthEnd
		if today <= monthEnd.Format(punch.DateLayout) {
			end = mustParseDate(today)
		}
		for d := monthStart; !d.After(end); d = d.AddDate(0, 0, 1) {
			if containsInt64(rule.WorkDays, weekdayOf(d.Format(punch.DateLayout))) {
				stats.ExpectedDays++
			}
		}
	} else {
		stats.ExpectedDays = int64(len(rowsByDate))
	}
	return stats
}

// buildLockHint 考勤锁定日提示（HRM-23 月报未落地：展示配置口径值——每月 DefaultLockDay=5 号
// 锁定当月月报，文案按「当月考勤将于次月 LockDay 锁定」口径生成，如「11月考勤将于12-05锁定」）。
func buildLockHint(monthTime time.Time) types.DashboardLockHint {
	nextMonth := monthTime.AddDate(0, 1, 0)
	lockDate := fmt.Sprintf("%02d-%02d", int(nextMonth.Month()), DefaultLockDay)
	return types.DashboardLockHint{
		LockDay:   DefaultLockDay,
		LockMonth: monthTime.Format("2006-01"),
		HintText:  fmt.Sprintf("%d月考勤将于%s锁定", int(monthTime.Month()), lockDate),
	}
}

// buildCalendarSkeleton 全月逐日骨架（rest/异常/缺勤 + 打卡时间）。未来日期 = future。
func buildCalendarSkeleton(monthStart, monthEnd time.Time, today string,
	rowsByDate map[string]*hrm_punch_record.HrmPunchRecord) []types.DashboardCalendarDay {
	days := make([]types.DashboardCalendarDay, 0, 31)
	for d := monthStart; !d.After(monthEnd); d = d.AddDate(0, 0, 1) {
		date := d.Format(punch.DateLayout)
		day := types.DashboardCalendarDay{
			Date:    date,
			Weekday: weekdayName(date),
			Status:  "future",
		}
		if date > today {
			days = append(days, day)
			continue
		}
		if row, ok := rowsByDate[date]; ok && row.SettleStatus != "" {
			day.Status = row.SettleStatus
			if day.Status != punch.SettleMissing {
				day.ClockInTime = formatNullTime(row.ClockInTime)
				day.ClockOutTime = formatNullTime(row.ClockOutTime)
			}
			day.LateMinutes = row.SettleLateMinutes
			day.EarlyMinutes = row.SettleEarlyMinutes
		} else {
			// 无行/未日结的过去日期 → 未判定（前端按空态渲染）。
			day.Status = ""
		}
		days = append(days, day)
	}
	return days
}

// ---- 本文件局部 helpers（punch 包同名 helper 未导出，不跨包引用私有实现） ----

// punchStatusOf 打卡状态展示：normal|field|missing（空卡/空状态 = missing，punch.toRecord 同款语义）。
func punchStatusOf(t sql.NullTime, s string) string {
	if !t.Valid || s == "" {
		return punch.SettleMissing
	}
	return s
}

func formatNullTime(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.Format(time.DateTime)
}

// weekdayOf yyyy-MM-dd → 1=周一..7=周日（与 work_days 口径一致；非法日期 0）。
func weekdayOf(date string) int64 {
	t, err := time.ParseInLocation(punch.DateLayout, date, time.Local)
	if err != nil {
		return 0
	}
	wd := int64(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func weekdayName(date string) string {
	t, err := time.ParseInLocation(punch.DateLayout, date, time.Local)
	if err != nil {
		return ""
	}
	names := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	return names[int(t.Weekday())]
}

func containsInt64(list []int64, v int64) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

func mustParseDate(date string) time.Time {
	t, _ := time.ParseInLocation(punch.DateLayout, date, time.Local)
	return t
}
