package shift

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// HRM 班次状态（GLOSSARY「班次」）: 1=启用 0=停用.
const (
	shiftStatusEnabled  = "1"
	shiftStatusDisabled = "0"
)

func validShiftStatus(v string) bool {
	return v == shiftStatusEnabled || v == shiftStatusDisabled
}

// parseHHmm parses an "HH:mm" time into minutes since midnight.
func parseHHmm(v string) (int, error) {
	parts := strings.Split(v, ":")
	if len(parts) != 2 || len(parts[0]) != 2 || len(parts[1]) != 2 {
		return 0, fmt.Errorf("时间 %q 必须为 HH:mm 格式", v)
	}
	hour, err1 := strconv.Atoi(parts[0])
	minute, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, fmt.Errorf("时间 %q 必须为合法的 HH:mm（00:00-23:59）", v)
	}
	return hour*60 + minute, nil
}

// validateShiftTimes checks 上下班时间: both must be HH:mm and must not be
// identical (a zero-length shift is meaningless; earlier end time means 跨天).
func validateShiftTimes(start, end string) error {
	startMinute, err := parseHHmm(start)
	if err != nil {
		return err
	}
	endMinute, err := parseHHmm(end)
	if err != nil {
		return err
	}
	if startMinute == endMinute {
		return fmt.Errorf("上班时间 %s 与下班时间 %s 不能相同", start, end)
	}
	return nil
}

// validateRestTimes checks 休息时段: either both empty (no rest) or a valid
// HH:mm pair with rest start strictly before rest end.
func validateRestTimes(restStart, restEnd string) error {
	if restStart == "" && restEnd == "" {
		return nil
	}
	if restStart == "" || restEnd == "" {
		return fmt.Errorf("休息时段必须成对填写")
	}
	startMinute, err := parseHHmm(restStart)
	if err != nil {
		return err
	}
	endMinute, err := parseHHmm(restEnd)
	if err != nil {
		return err
	}
	if startMinute >= endMinute {
		return fmt.Errorf("休息时段开始 %s 必须早于休息时段结束 %s", restStart, restEnd)
	}
	return nil
}

// computeCrossDay derives 跨天: 下班时间早于等于上班时间表示跨天（22:00-次日06:00）.
func computeCrossDay(start, end string) int64 {
	startMinute, errStart := parseHHmm(start)
	endMinute, errEnd := parseHHmm(end)
	if errStart != nil || errEnd != nil || endMinute <= startMinute {
		return 1
	}
	return 0
}

// computeStandardHours computes 标准工时（小时）＝上下班时长－休息时长, 跨天按
// 次日计算; result rounded to 2 decimals.
func computeStandardHours(start, end, restStart, restEnd string) float64 {
	startMinute, _ := parseHHmm(start)
	endMinute, _ := parseHHmm(end)
	total := endMinute - startMinute
	if total <= 0 {
		total += 24 * 60
	}
	if restStart != "" && restEnd != "" {
		restStartMinute, _ := parseHHmm(restStart)
		restEndMinute, _ := parseHHmm(restEnd)
		if restEndMinute > restStartMinute {
			total -= restEndMinute - restStartMinute
		}
	}
	return math.Round(float64(total)/60*100) / 100
}
