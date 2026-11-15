package overtimebill

import (
	"errors"
	"time"
)

// BillTypeCode is the 加班申请 类型码 (HR217, GLOSSARY「单据编号」/ demo-inventory §4.2).
const BillTypeCode = "217"

// 补偿方式 (comp_type, demo 列「补偿方式」).
const (
	CompTypeLeave = "comp_leave" // 调休
	CompTypeOther = "other"      // 其他
)

// 来源 (source, demo 列「来源」).
const SourceEmployee = "employee" // 员工申请

// sentinel errors of the hours calculator (pure function, 单测 seam).
var (
	// ErrInvalidTimeRange: 结束时间必须晚于开始时间。
	ErrInvalidTimeRange = errors.New("结束时间必须晚于开始时间")
	// ErrInvalidTimeFormat: 时间格式非法，须为 YYYY-MM-DD HH:MM:SS。
	ErrInvalidTimeFormat = errors.New("时间格式非法，须为 YYYY-MM-DD HH:MM:SS")
)

// MinutesPerHour 小时↔分钟折算.
const MinutesPerHour = 60.0

// CalcOvertimeHours 申请时长自动计算（纯函数，单测 seam）:
// 时长(小时) = (end-start)，向下取整到 minUnitMinutes 的整数倍（加班最小计算单位口径，
// 不足一个最小单位的不计；与请假「向上取整」口径相反——加班不多算）。
// minUnitMinutes <= 0 时按 30 分钟处理（规则默认值）。
//   - 2h05m, unit 30min → 2h
//   - 2h29m, unit 60min → 2h
//   - 45m,   unit 30min → 0.75h
//   - 29m,   unit 30min → 0（不足最小单位不计）
func CalcOvertimeHours(startTime, endTime string, minUnitMinutes int64) (float64, error) {
	start, err := parseBillTime(startTime)
	if err != nil {
		return 0, err
	}
	end, err := parseBillTime(endTime)
	if err != nil {
		return 0, err
	}
	if !end.After(start) {
		return 0, ErrInvalidTimeRange
	}
	if minUnitMinutes <= 0 {
		minUnitMinutes = 30
	}
	minutes := int64(end.Sub(start).Minutes())
	units := minutes / minUnitMinutes // 向下取整到最小单位
	return float64(units*minUnitMinutes) / MinutesPerHour, nil
}

// RoundHoursToUnit 核定时长也按最小计算单位向下取整（审批人填任意值，折算口径统一）.
func RoundHoursToUnit(hours float64, minUnitMinutes int64) float64 {
	if minUnitMinutes <= 0 {
		minUnitMinutes = 30
	}
	minutes := int64(hours*MinutesPerHour + 1e-9)
	units := minutes / minUnitMinutes
	return float64(units*minUnitMinutes) / MinutesPerHour
}

// ConvertAccrueHours 规则折算（纯函数，单测 seam）:
// 折算调休小时 = 基准小时(核定时长，未核定用申请时长) × CompMultiplier，保留一位小数.
func ConvertAccrueHours(baseHours, compMultiplier float64) float64 {
	if baseHours <= 0 {
		return 0
	}
	if compMultiplier <= 0 {
		compMultiplier = 1
	}
	return round1(baseHours * compMultiplier)
}

// round1 保留一位小数（负数对称）.
func round1(v float64) float64 {
	if v < 0 {
		return -round1(-v)
	}
	return float64(int64(v*10+0.5)) / 10
}

// billTimeLayout 单据时间字段统一格式.
const billTimeLayout = "2006-01-02 15:04:05"

// parseBillTime 解析 YYYY-MM-DD HH:MM:SS 时间.
func parseBillTime(s string) (time.Time, error) {
	t, err := time.ParseInLocation(billTimeLayout, s, time.Local)
	if err != nil {
		return time.Time{}, errors.Join(ErrInvalidTimeFormat, err)
	}
	return t, nil
}
