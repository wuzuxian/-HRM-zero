// Package errcodes — HRM-19 请假销假单（HR205）业务错误码（1054x 段）。
// 1051x 已被 HRM-18 假期余额占用（10510-10516），1052x/1053x 已被 HRM-36 个税占用（10520-10533）。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// LeaveNotFoundCode 请假单不存在
	LeaveNotFoundCode = 10560
	// LeaveInvalidCode 请假单参数非法（时间倒序/类型缺失等）
	LeaveInvalidCode = 10561
	// LeaveNotEditableCode 当前单据状态不允许修改/删除/销假
	LeaveNotEditableCode = 10562
	// LeaveRefundOverflowCode 销假退回天数超过该单据已扣减余额（幂等保护）
	LeaveRefundOverflowCode = 10564
)

func init() {
	status.Register(LeaveNotFoundCode)
	status.Register(LeaveInvalidCode)
	status.Register(LeaveNotEditableCode)
	status.Register(LeaveRefundOverflowCode)
}
