package leave

import (
	"errors"
)

// BillTypeCode is the 请假销假 类型码 (HR205, GLOSSARY「单据编号」/ demo-inventory §4.2).
const BillTypeCode = "205"

// sentinel errors of the days calculator (pure function, 单测 seam).
var (
	// ErrInvalidTimeRange: 结束时间必须晚于开始时间。
	ErrInvalidTimeRange = errors.New("结束时间必须晚于开始时间")
	// ErrInvalidTimeFormat: 时间格式非法，须为 YYYY-MM-DD HH:MM:SS。
	ErrInvalidTimeFormat = errors.New("时间格式非法，须为 YYYY-MM-DD HH:MM:SS")
)

// HalfDayUnit 请假天数的最小步进（0.5 天口径，与假期余额 0.5 步进一致）。
const HalfDayUnit = 0.5

// HoursPerDay 天→小时折算口径：1 天 = 8 小时（与 leavebalance.HoursPerDay 一致）。
const HoursPerDay = 8.0

// CalcLeaveDays 预计/实际请假天数自动计算（纯函数，单测 seam）。
// 口径（写单据备注/notes）: 天数 = 时长(小时) / 8，向上取整到 0.5 天步进（不足部分进位）；
// 同一天请假最短 0.5 天（4 小时以内按 0.5 天计）。
//   - 2 小时  → 0.5 天
//   - 5 小时  → 1 天   (5/8=0.625 → 向上取整到 1)
//   - 10 小时 → 1.5 天 (10/8=1.25 → 向上取整到 1.5)
//   - 12 小时（1.5 天整）→ 1.5 天
func CalcLeaveDays(startTime, endTime string) (float64, error) {
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
	hours := end.Sub(start).Hours()
	return CeilHalfUp(hours / HoursPerDay), nil
}

// CeilHalfUp 向上取整到 0.5 天步进（请假天数口径：多算不少数）。
func CeilHalfUp(days float64) float64 {
	steps := int64(days * 2)
	if float64(steps) < days*2 {
		steps++ // 有余数进位
	}
	if steps == 0 {
		steps = 1 // 最短 0.5 天
	}
	return float64(steps) / 2
}

// RoundHalfStep 四舍五入到 0.5 天步进（销假差额浮点漂移防御；计划/实际天数均为
// 0.5 步进的产物，差额理论上也应是 0.5 的倍数，此函数只消浮点误差；负数对称舍入）。
func RoundHalfStep(days float64) float64 {
	if days < 0 {
		return -RoundHalfStep(-days)
	}
	return float64(int64(days*2+0.5)) / 2
}
