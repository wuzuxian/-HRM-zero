// Package errcodes — HRM-21 加班申请单（HR217）业务错误码（1057x 段）。
// 1056x 已被 HRM-19 请假销假占用，10580-10585 已被 HRM-25 招聘职位占用（hrm.go）。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// OvertimeNotFoundCode 加班单不存在
	OvertimeNotFoundCode = 10570
	// OvertimeInvalidCode 加班单参数非法（加班日期/当天类型/起止时间等）
	OvertimeInvalidCode = 10571
	// OvertimeNotEditableCode 单据当前状态不允许修改/删除/核定（仅允许状态集之外拒绝）
	OvertimeNotEditableCode = 10572
	// OvertimeConfirmHoursExceededCode 核定时长超过申请时长
	OvertimeConfirmHoursExceededCode = 10573
	// OvertimeMonthlyLimitExceededCode 折算调休小时超出该员工当月加班月上限
	OvertimeMonthlyLimitExceededCode = 10574
)

func init() {
	status.Register(OvertimeNotFoundCode)
	status.Register(OvertimeInvalidCode)
	status.Register(OvertimeNotEditableCode)
	status.Register(OvertimeConfirmHoursExceededCode)
	status.Register(OvertimeMonthlyLimitExceededCode)
}
