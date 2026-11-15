package punch

// 打卡/日结领域核心（HRM-16）：常量 + SettleDay 日结契约 + 打卡行读写 helpers。
// 端点 logic 见同目录 punch.go / list.go / add.go / edit.go / delete.go / export.go / import.go / settle.go。

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendancegroup"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// 打卡/日结领域常量（GLOSSARY「打卡记录」「日结」）。
const (
	// 打卡状态: normal 正常 / field 外勤 / 空=缺卡
	StatusNormal = "normal"
	StatusField  = "field"

	// 记录来源
	SourceApp             = "app"
	SourceAdmin           = "admin"
	SourceImport          = "import"
	SourceCardReplacement = "card_replacement"

	// 打卡类型（Punch 端点）
	PunchTypeClockIn  = "clockIn"
	PunchTypeClockOut = "clockOut"

	OvertimeYes = "1"
	OvertimeNo  = "0"
)

// 日结状态集合（settle_status 列取值 + DaySettleResult.Status）。
const (
	SettleNormal  = "normal"  // 正常
	SettleLate    = "late"    // 迟到
	SettleEarly   = "early"   // 早退
	SettleMissing = "missing" // 缺卡
	SettleRest    = "rest"    // 休息日
	SettleNoRule  = "no_rule" // 无规则兜底（未判定）
)

// DateLayout 统一日期字符串格式。
const DateLayout = "2006-01-02"

// DaySettleResult 是 SettleDay 的返回契约（HRM-17 看板 / HRM-22 考勤统计 / HRM-23 月报消费）。
//
// 语义：
//   - Status ∈ {normal, late, early, missing, rest, no_rule}；
//     迟到且早退并存时 Status=late（迟到优先展示），EarlyMinutes 仍记录具体分钟数。
//   - LateMinutes/EarlyMinutes 已扣除弹性分钟（弹性内的偏差不算迟到/早退）。
//   - WorkHours = 实打上下班时间差（小时，跨天班含次日时段，2 位小数）。
//     考勤组/组织规则无休息时段字段，暂不扣除休息。
//   - rest/no_rule 时打卡数据原样保留，WorkHours 仍按实际打卡计算。
//   - Found=false（无规则兜底）时 Status=no_rule，消费方应视为「未判定」而非异常。
type DaySettleResult struct {
	Found        bool
	EmployeeUuid string
	Date         string
	Status       string
	LateMinutes  int64
	EarlyMinutes int64
	WorkHours    float64
	RuleSource   string // group | org | ""
}

// SettleDay 日结唯一入口（HRM-16 消费契约，HRM-17/22/23 直接复用）：
// 按 attendancegroup.ResolveEffectiveRule 取当日生效规则 + 班次时间，结合当日打卡记录
// 判定 正常/迟到/早退/缺卡/休息日（支持跨天班——下班时间属次日），并把日结快照
// （settle_status/settle_late_minutes/settle_early_minutes/settle_work_hours/settle_time/rule_source）
// 落库到 hrm_punch_record（一人一天一行；当日无打卡记录也会落一行空记录以固化缺卡判定，
// 供下游统计聚合）。
func SettleDay(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid, date string) (*DaySettleResult, error) {
	employee, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, employeeUuid).Build()...)
	if err != nil {
		return nil, err
	}

	result, err := computeDaySettle(ctx, svcCtx, employee, date)
	if err != nil {
		return nil, err
	}

	if err := persistSettle(ctx, svcCtx, employeeUuid, date, result); err != nil {
		return nil, err
	}
	return result, nil
}

// computeDaySettle 判定逻辑（DB 读，无写）：规则解析 + 打卡行 → DaySettleResult。
func computeDaySettle(ctx context.Context, svcCtx *svc.ServiceContext, employee *hrm_employee.HrmEmployee, date string) (*DaySettleResult, error) {
	result := &DaySettleResult{
		EmployeeUuid: employee.Uuid,
		Date:         date,
	}

	rule, err := attendancegroup.ResolveEffectiveRule(ctx, svcCtx, employee)
	if err != nil {
		return nil, err
	}
	if rule == nil || !rule.Found {
		// 无规则兜底: 只按实际打卡算工时，状态未判定。
		result.Found = false
		result.Status = SettleNoRule
		result.WorkHours = actualWorkHours(ctx, svcCtx, employee.Uuid, date)
		return result, nil
	}
	result.Found = true
	result.RuleSource = rule.Source

	// 休息日判定: date 的 weekday 不在 WorkDays。
	weekday := weekdayOf(date)
	isRestDay := !containsInt64(rule.WorkDays, weekday)

	if isRestDay {
		result.Status = SettleRest
		result.WorkHours = actualWorkHours(ctx, svcCtx, employee.Uuid, date)
		return result, nil
	}

	startMinutes, err := parseHHmm(rule.StartTime)
	if err != nil {
		return nil, fmt.Errorf("生效规则上班时间非法: %w", err)
	}
	endMinutes, err := parseHHmm(rule.EndTime)
	if err != nil {
		return nil, fmt.Errorf("生效规则下班时间非法: %w", err)
	}

	punch, err := findPunchRow(ctx, svcCtx, employee.Uuid, date)
	if err != nil {
		return nil, err
	}

	// 缺卡: 上班卡或下班卡缺失（外勤 field 视为有效打卡）。
	inMissing := punch == nil || !punch.ClockInTime.Valid || punch.ClockInStatus == ""
	outMissing := punch == nil || !punch.ClockOutTime.Valid || punch.ClockOutStatus == ""
	if inMissing || outMissing {
		result.Status = SettleMissing
		result.WorkHours = actualWorkHours(ctx, svcCtx, employee.Uuid, date)
		return result, nil
	}

	dayStart, _ := time.ParseInLocation(time.DateTime, date+" 00:00:00", time.Local)

	// 迟到分钟 = 打卡上班时间 - 规则上班时间 - 弹性，下限 0。
	inAt := punch.ClockInTime.Time
	inOffset := minutesSince(inAt, dayStart) - startMinutes - rule.FlexibleMinutes
	if inOffset > 0 {
		result.LateMinutes = inOffset
	}

	// 早退分钟 = 规则下班时间 - 打卡下班时间，下限 0。跨天班下班卡在次日，
	// 以「下班时刻相对上班日零点的分钟数」与 endMinutes(+24h) 比较，天然支持跨天。
	crossDay := endMinutes <= startMinutes
	endAnchor := endMinutes
	if crossDay {
		endAnchor += 24 * 60
	}
	outAt := punch.ClockOutTime.Time
	outOffset := endAnchor - minutesSince(outAt, dayStart)
	if outOffset > 0 {
		result.EarlyMinutes = outOffset
	}

	// 工时 = 实打时间差（小时，跨天自然覆盖），2 位小数。
	result.WorkHours = round2(outAt.Sub(inAt).Hours())

	// 状态判定: 迟到优先展示（早退分钟数仍记录）。
	switch {
	case result.LateMinutes > 0:
		result.Status = SettleLate
	case result.EarlyMinutes > 0:
		result.Status = SettleEarly
	default:
		result.Status = SettleNormal
	}
	return result, nil
}

// persistSettle 把日结快照写回打卡行（无行则落一行空记录固化缺卡判定）。
func persistSettle(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid, date string, result *DaySettleResult) error {
	punch, err := findPunchRow(ctx, svcCtx, employeeUuid, date)
	if err != nil {
		return err
	}
	if punch == nil {
		employee, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
			condition.NewChain().Equal(hrm_employee.Uuid, employeeUuid).Build()...)
		if err != nil {
			return err
		}
		punch = &hrm_punch_record.HrmPunchRecord{
			Uuid:         newUuid(),
			PunchDate:    date,
			EmployeeUuid: employeeUuid,
			UserUuid:     employee.UserUuid,
			Username:     usernameOf(ctx, svcCtx, employee.UserUuid),
			EmployeeNo:   employee.EmployeeNo,
			EmployeeName: employee.Name,
			Source:       SourceApp,
		}
		if _, err := svcCtx.Model.HrmPunchRecord.Insert(ctx, nil, punch); err != nil {
			return err
		}
	}
	return svcCtx.Model.HrmPunchRecord.UpdateFieldsByCondition(ctx, nil,
		map[string]any{
			"settle_status":        result.Status,
			"settle_late_minutes":  result.LateMinutes,
			"settle_early_minutes": result.EarlyMinutes,
			"settle_work_hours":    result.WorkHours,
			"rule_source":          result.RuleSource,
			"settle_time":          time.Now(),
		},
		condition.NewChain().Equal(hrm_punch_record.Uuid, punch.Uuid).Build()...)
}

// actualWorkHours 当日实际打卡工时（缺卡/休息日/无规则时的 WorkHours 兜底值）。
func actualWorkHours(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid, date string) float64 {
	punch, err := findPunchRow(ctx, svcCtx, employeeUuid, date)
	if err != nil || punch == nil || !punch.ClockInTime.Valid || !punch.ClockOutTime.Valid {
		return 0
	}
	return round2(punch.ClockOutTime.Time.Sub(punch.ClockInTime.Time).Hours())
}

// findPunchRow 当日打卡行（一人一天一行），无行返回 nil。
func findPunchRow(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid, date string) (*hrm_punch_record.HrmPunchRecord, error) {
	row, err := svcCtx.Model.HrmPunchRecord.FindOneByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_punch_record.PunchDate, date).
			Equal(hrm_punch_record.EmployeeUuid, employeeUuid).
			Build()...)
	if err != nil {
		if err == hrm_punch_record.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return row, nil
}

// weekdayOf yyyy-MM-dd → 考勤日（仓库约定 1=周一 .. 7=周日，与 hrm_attendance_group.work_days
// / hrm_attendance_config.work_days 的 "1,2,3,4,5" 周一至周五口径一致）；非法日期返回 0（视为非工作日）。
func weekdayOf(date string) int64 {
	t, err := time.ParseInLocation(DateLayout, date, time.Local)
	if err != nil {
		return 0
	}
	wd := int64(t.Weekday()) // Go: 周日=0
	if wd == 0 {
		return 7
	}
	return wd
}

// minutesSince t 相对 anchor 的分钟数（可为负）。
func minutesSince(t time.Time, anchor time.Time) int64 {
	return int64(t.Sub(anchor).Minutes())
}

func containsInt64(list []int64, v int64) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

// parseHHmm "HH:mm" → 分钟数（0..1439）。
func parseHHmm(v string) (int64, error) {
	if len(v) != 5 || v[2] != ':' {
		return 0, fmt.Errorf("时间 %q 必须为 HH:mm 格式", v)
	}
	for i, c := range v {
		if i == 2 {
			continue
		}
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("时间 %q 必须为合法的 HH:mm", v)
		}
	}
	hour := int(v[0]-'0')*10 + int(v[1]-'0')
	minute := int(v[3]-'0')*10 + int(v[4]-'0')
	if hour > 23 || minute > 59 {
		return 0, fmt.Errorf("时间 %q 必须为合法的 HH:mm（00:00-23:59）", v)
	}
	return int64(hour*60 + minute), nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func newUuid() string {
	return uuid.New().String()
}

// usernameOf 登录账号快照（user_uuid 无对应账号时返回空）。
func usernameOf(ctx context.Context, svcCtx *svc.ServiceContext, userUuid string) string {
	if userUuid == "" {
		return ""
	}
	row, err := svcCtx.Model.ManageUser.FindOneByUuid(ctx, nil, userUuid)
	if err != nil || row == nil {
		return ""
	}
	return row.Username
}
