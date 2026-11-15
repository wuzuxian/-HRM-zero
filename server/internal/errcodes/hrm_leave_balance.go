package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM 假期余额（HRM-18）业务错误码。1051x 段
// （1032x org / 1033x bill / 1034x shift+leavetype / 1035x employee /
// 1036x training-my / 1037x transfer / 1038x attendance / 1040x training-plan+session /
// 1042x 奖惩 / 1043x 招聘需求 / 1044x 候选人 / 1045x 证书 / 1046x 员工账号同步 /
// 1047x 入职 / 1048x 转正 / 1049x 离职 / 10500 perf_scheme，见各 errcode 文件声明）。
const (
	// LeaveBalanceInitSkippedCode 重复初始化被跳过（本年度该员工该假期类型已有余额行）
	LeaveBalanceInitSkippedCode = 10510
	// LeaveBalanceNotFoundCode 余额行不存在
	LeaveBalanceNotFoundCode = 10511
	// LeaveBalanceInvalidCode 假期余额参数非法（年度/天数/原因等）
	LeaveBalanceInvalidCode = 10512
	// LeaveBalanceInsufficientCode 可用余额不足（调整导致可用 < 0 / 请假扣减超可用）
	LeaveBalanceInsufficientCode = 10513
	// LeaveBalanceTypeNotManagedCode 假期类型未启用余额管理
	LeaveBalanceTypeNotManagedCode = 10514
	// LeaveBalanceEmployeeNotFoundCode 员工不存在
	LeaveBalanceEmployeeNotFoundCode = 10515
	// LeaveBalanceTypeNotFoundCode 假期类型不存在
	LeaveBalanceTypeNotFoundCode = 10516
)

func init() {
	status.Register(LeaveBalanceInitSkippedCode)
	status.Register(LeaveBalanceNotFoundCode)
	status.Register(LeaveBalanceInvalidCode)
	status.Register(LeaveBalanceInsufficientCode)
	status.Register(LeaveBalanceTypeNotManagedCode)
	status.Register(LeaveBalanceEmployeeNotFoundCode)
	status.Register(LeaveBalanceTypeNotFoundCode)
}
