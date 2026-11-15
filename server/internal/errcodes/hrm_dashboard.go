// Package errcodes — HRM-17 个人考勤看板业务错误码（1061x 段）。
// 1057x 补卡单 / 1060x 加班单已占用；1061x 起为本票看板段。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// DashboardEmployeeNotFoundCode 当前登录人未关联员工档案（看板给出建档引导）
	DashboardEmployeeNotFoundCode = 10610
	// DashboardInvalidCode 看板参数非法（月份格式等）
	DashboardInvalidCode = 10611
)

func init() {
	status.Register(DashboardEmployeeNotFoundCode)
	status.Register(DashboardInvalidCode)
}
