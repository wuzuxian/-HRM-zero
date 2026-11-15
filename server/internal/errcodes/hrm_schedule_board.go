package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-15 排班表业务错误码。1055x 段
// （1032x org / 1033x bill / 1034x shift+leavetype / 1035x employee /
// 1036x training-my / 1037x transfer / 1038x attendance / 1040x training-plan+session /
// 1042x 奖惩 / 1043x 招聘需求 / 1044x 候选人 / 1045x 证书 / 1046x 员工账号同步 /
// 1047x 入职 / 1048x 转正 / 1049x 离职 / 10500 perf_scheme / 1051x leave_balance /
// 1052x tax / 1054x punch（原 1052x，与个税撞车已改号）/ 1055x 排班表(本票, 1053x 实际被 tax 占用至 10533)，见各 errcode 文件声明）。
const (
	// ScheduleBoardInvalidCode 排班表参数非法（月份/日期格式、区间、周起止非周一、考勤组为空等）
	ScheduleBoardInvalidCode = 10550
	// ScheduleBoardGroupNotFoundCode 考勤组不存在或未启用（排班表按考勤组维护）
	ScheduleBoardGroupNotFoundCode = 10551
	// ScheduleBoardShiftInvalidCode 班次不存在或已停用
	ScheduleBoardShiftInvalidCode = 10552
	// ScheduleBoardCycleNotFoundCode 排班周期不存在或未启用
	ScheduleBoardCycleNotFoundCode = 10553
	// ScheduleBoardEmployeeNotFoundCode 排班表成员不存在（导入/单元格指定的工号或员工 uuid 无效）
	ScheduleBoardEmployeeNotFoundCode = 10554
	// ScheduleBoardImportInvalidCode 排班导入文件非法（缺列/列数错/班次编码不存在等）
	ScheduleBoardImportInvalidCode = 10555
	// ScheduleBoardPublishedLockedCode 已发布排班禁止覆盖修改（先撤销发布）
	ScheduleBoardPublishedLockedCode = 10556
)

func init() {
	status.Register(ScheduleBoardInvalidCode)
	status.Register(ScheduleBoardGroupNotFoundCode)
	status.Register(ScheduleBoardShiftInvalidCode)
	status.Register(ScheduleBoardCycleNotFoundCode)
	status.Register(ScheduleBoardEmployeeNotFoundCode)
	status.Register(ScheduleBoardImportInvalidCode)
	status.Register(ScheduleBoardPublishedLockedCode)
}
