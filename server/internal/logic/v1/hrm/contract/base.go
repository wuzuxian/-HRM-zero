package contract

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// uuidNew 合同行 uuid。
func uuidNew() string {
	return uuid.New().String()
}

// HRM-05 劳动合同预警口径（demo-inventory §2.2/§2.3，spec 用户故事 14/15）：
//   - 距到期天数 = 到期日 - 今天（纯函数，负数=已过期 N 天，0=今天到期）。
//   - 预警状态列（与演示站状态列语义对齐）：
//       terminated 已解除（合同状态=3，人工解除，不参与预警口径）
//       expired    已到期（距到期天数 < 0）
//       warning    即将到期（0 <= 距到期天数 <= 90，30/60/90 三档共享一个 warning 段，明细由天数列呈现）
//       normal     正常（> 90 天）
const (
	ContractStatusActive     = "1" // 有效
	ContractStatusExpired    = "2" // 已到期
	ContractStatusTerminated = "3" // 已解除

	ExpiryStateNormal     = "normal"
	ExpiryStateWarning    = "warning"
	ExpiryStateExpired    = "expired"
	ExpiryStateTerminated = "terminated"

	// WarningWindowDays 即将到期窗口（≤90 天进 warning，含到期当天）
	WarningWindowDays = 90

	// RenewDefaultYears 续签未指定新到期日时的默认延长年数（到期日+1 年）
	RenewDefaultYears = 1

	dateFormat = "2006-01-02"
)

// ErrInvalidDate 日期格式非法（YYYY-MM-DD）。
var ErrInvalidDate = errors.New("invalid date, expect YYYY-MM-DD")

// DaysUntilDate 计算今天到目标日期的天数（目标-今天；负数=已过期）。
// today 由调用方注入以便测试；目标日期格式 YYYY-MM-DD。
func DaysUntilDate(today time.Time, date string) (int64, error) {
	target, err := time.ParseInLocation(dateFormat, date, time.Local)
	if err != nil {
		return 0, ErrInvalidDate
	}
	today = truncateToDate(today)
	return int64(target.Sub(today).Hours() / 24), nil
}

// ExpiryStateByDays 将距到期天数映射为预警状态。
func ExpiryStateByDays(days int64) string {
	switch {
	case days < 0:
		return ExpiryStateExpired
	case days <= WarningWindowDays:
		return ExpiryStateWarning
	default:
		return ExpiryStateNormal
	}
}

// ContractExpiryState 合同预警状态：已解除合同优先返回 terminated，其余按距到期天数映射。
func ContractExpiryState(status string, days int64) string {
	if status == ContractStatusTerminated {
		return ExpiryStateTerminated
	}
	return ExpiryStateByDays(days)
}

// NextExpiryDate 续签后的新到期日：显式指定则原样返回，否则=原到期日+1 年。
// 同月同日不存在（如 02-29 逢平年）时回退到 3 月 1 日（time.AddDate 语义）。
func NextExpiryDate(currentEndDate, specified string) (string, error) {
	if specified != "" {
		if _, err := time.ParseInLocation(dateFormat, specified, time.Local); err != nil {
			return "", ErrInvalidDate
		}
		return specified, nil
	}
	current, err := time.ParseInLocation(dateFormat, currentEndDate, time.Local)
	if err != nil {
		return "", ErrInvalidDate
	}
	return current.AddDate(RenewDefaultYears, 0, 0).Format(dateFormat), nil
}

// MaskCertNo 证号脱敏：保留前 2 位与后 2 位，中间以等长 * 打码；
// 长度 <= 6 的短证号只保留首 1 位；空串返回空串。
func MaskCertNo(no string) string {
	runes := []rune(no)
	n := len(runes)
	if n == 0 {
		return ""
	}
	if n <= 6 {
		out := []rune(no)
		for i := 1; i < n; i++ {
			out[i] = '*'
		}
		return string(out)
	}
	out := make([]rune, n)
	for i := 2; i < n-2; i++ {
		out[i] = '*'
	}
	out[0], out[1] = runes[0], runes[1]
	out[n-2], out[n-1] = runes[n-2], runes[n-1]
	return string(out)
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// timeParseDate 供 helpers.go 校验 YYYY-MM-DD。
func timeParseDate(v string) (time.Time, error) {
	return time.ParseInLocation(dateFormat, v, time.Local)
}
