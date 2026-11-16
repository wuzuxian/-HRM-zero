package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-22 考勤统计业务错误码（1063x 段，10630-10631 已被 HRM-17 看板占用——合并 feat/hrm
// 时撞码，本票让位改取 10632-10633；1061x 薪酬方案 / 1060x 加班单 / 1062x 薪资档案）。
const (
	// PunchStatisticsInvalidCode 考勤统计参数非法（日期格式/区间越界/员工为空等）
	PunchStatisticsInvalidCode = 10632
)

func init() {
	status.Register(PunchStatisticsInvalidCode)
}
