package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM 员工入职单（HRM-07，HR201）业务错误码。1047x 段
// （1032x org / 1033x bill / 1034x shift+leavetype / 1035x employee+perf /
// 1036x training / 1037x transfer / 1038x attendance / 1039x perf 预留 /
// 1040x training-plan / 1041x training-session / 1042x 奖惩 / 1043x 招聘需求 /
// 1044x 候选人 / 1045x 证书 / 1046x 员工账号同步，见各 errcode 文件声明）。
const (
	// EntryNotFoundCode 入职单不存在
	EntryNotFoundCode = 10470
	// EntryInvalidCode 入职单参数非法（姓名/入职日期/预计转正日期/试用期月等）
	EntryInvalidCode = 10471
	// EntryNameRequiredCode 姓名为必填
	EntryNameRequiredCode = 10472
	// EntryNotEditableCode 单据当前状态不允许修改/删除（仅 未提交/审批不通过/已撤销 可改）
	EntryNotEditableCode = 10473
	// EntryArchivedCode 入职单已建档（审批通过已自动创建员工档案，不可重复建档/修改）
	EntryArchivedCode = 10474
)

func init() {
	status.Register(EntryNotFoundCode)
	status.Register(EntryInvalidCode)
	status.Register(EntryNameRequiredCode)
	status.Register(EntryNotEditableCode)
	status.Register(EntryArchivedCode)
}
