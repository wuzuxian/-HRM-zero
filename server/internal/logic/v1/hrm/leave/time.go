package leave

import (
	"fmt"
	"time"
)

// billTimeLayout 单据时间字段统一格式。
const billTimeLayout = "2006-01-02 15:04:05"

// parseBillTime 解析 YYYY-MM-DD HH:MM:SS 时间。
func parseBillTime(s string) (time.Time, error) {
	t, err := time.ParseInLocation(billTimeLayout, s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q", ErrInvalidTimeFormat, s)
	}
	return t, nil
}
