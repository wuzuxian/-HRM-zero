package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-22 考勤统计业务错误码（1063x 段）。
// 1061x 已被 HRM-34 薪酬方案占用，1060x 已被 HRM-21 加班单占用，
// 1062x 已被 HRM-35 薪资档案占用（开工时 grep 确认 1063x 空闲）。
const (
	// PunchStatisticsInvalidCode 考勤统计参数非法（日期格式/区间越界/员工为空等）
	PunchStatisticsInvalidCode = 10630
)

func init() {
	status.Register(PunchStatisticsInvalidCode)
}
