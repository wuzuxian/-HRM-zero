// Package errcodes — HRM-20 补卡单（HR206）业务错误码（1057x 段）。
// grep errcodes/: 1032x org / 1033x bill / 1034x shift+leavetype / 1035x employee /
// 1036x-1037x training+transfer / 1038x attendance / 1040x-1041x training-plan /
// 1042x reward / 1043x-1044x recruitment / 10450-58 cert / 10460-68 user-sync+regularization /
// 10470-74 entry / 10480-84 schedule-cycle+contract / 10490-93 resignation / 10500-02 perf_scheme /
// 10510-16 leavebalance / 10520-33 tax / 10540-45 punch / 10550-56 schedule-board / 10560-64 leave。
// 1057x 空闲，本票取 10570-10574。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// CardReplacementNotFoundCode 补卡单不存在
	CardReplacementNotFoundCode = 10570
	// CardReplacementInvalidCode 补卡单参数非法（日期/类型/时间缺失或格式错误等）
	CardReplacementInvalidCode = 10571
	// CardReplacementNotEditableCode 当前单据状态不允许修改/删除
	CardReplacementNotEditableCode = 10572
	// CardReplacementDateMismatchCode 补卡时间不在补卡日期当天
	CardReplacementDateMismatchCode = 10573
	// CardReplacementEmployeeNotFoundCode 员工不存在
	CardReplacementEmployeeNotFoundCode = 10574
)

func init() {
	status.Register(CardReplacementNotFoundCode)
	status.Register(CardReplacementInvalidCode)
	status.Register(CardReplacementNotEditableCode)
	status.Register(CardReplacementDateMismatchCode)
	status.Register(CardReplacementEmployeeNotFoundCode)
}
