package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM 考勤打卡与日结（HRM-16）业务错误码。1052x 段
// （1032x org / 1033x bill / 1034x shift+leavetype / 1035x employee /
// 1036x training-my / 1037x transfer / 1038x attendance / 1040x training-plan+session /
// 1042x 奖惩 / 1043x 招聘需求 / 1044x 候选人 / 1045x 证书 / 1046x 员工账号同步 /
// 1047x 入职 / 1048x 转正 / 1049x 离职 / 10500 perf_scheme / 1051x leave_balance，见各 errcode 文件声明）。
const (
	// PunchNotFoundCode 打卡记录不存在
	PunchNotFoundCode = 10520
	// PunchInvalidCode 打卡参数非法（打卡类型/日期/地点等）
	PunchInvalidCode = 10521
	// PunchEmployeeNotFoundCode 打卡员工不存在（当前登录人未建员工档案或指定员工不存在）
	PunchEmployeeNotFoundCode = 10522
	// PunchDuplicateCode 同日重复打卡（幂等：同一打卡日期+同一打卡类型已有记录且策略为拒绝）
	PunchDuplicateCode = 10523
	// PunchImportInvalidCode 导入外部考勤数据行非法（缺关键列/日期格式错等）
	PunchImportInvalidCode = 10524
	// SettleInvalidCode 日结参数非法（日期格式/批量范围等）
	SettleInvalidCode = 10525
)

func init() {
	status.Register(PunchNotFoundCode)
	status.Register(PunchInvalidCode)
	status.Register(PunchEmployeeNotFoundCode)
	status.Register(PunchDuplicateCode)
	status.Register(PunchImportInvalidCode)
	status.Register(SettleInvalidCode)
}
