package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-23 考勤月报业务错误码（1067x 段——grep errcodes/*.go：1066x 考核任务链已占，
// 1065x 空、1067x 空闲；票面预授权 1067x，本票取 10670-10671）。
const (
	// MonthReportInvalidCode 考勤月报参数非法（月份格式非 yyyy-MM 等）
	MonthReportInvalidCode = 10670
	// MonthReportNotFoundCode 月报月份不存在（锁定/解锁/重算时该月无任何月报数据）
	MonthReportNotFoundCode = 10671
)

func init() {
	status.Register(MonthReportInvalidCode)
	status.Register(MonthReportNotFoundCode)
}
