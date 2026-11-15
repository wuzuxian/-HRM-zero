package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// HRM 组织管理 error codes (1032x range, manage.go uses 10311).
	OrgParentNotFoundCode = 10321
	OrgNotFoundCode       = 10322
	OrgHasChildrenCode    = 10323
	OrgImportInvalidCode  = 10324

	// HRM 单据审批底座（HRM-06）业务错误码。
	// HrmBillTypeNotRegisteredCode 单据类型未注册审批处理器
	HrmBillTypeNotRegisteredCode = 10330
	// HrmBillNotFoundCode 单据不存在
	HrmBillNotFoundCode = 10331
	// HrmBillInvalidTransitionCode 当前单据状态不允许该操作（单据状态机拒绝）
	HrmBillInvalidTransitionCode = 10332
	// HrmBillInvalidTypeCode 单据类型码非法（须为 3 位数字）
	HrmBillInvalidTypeCode = 10333
	// HrmBillNoSeqOutOfRangeCode 单据编号流水超出范围
	HrmBillNoSeqOutOfRangeCode = 10334

	// HRM 假勤设置：班次管理（HRM-11）业务错误码。
	// ShiftTimeInvalidCode 上下班时间非法（格式须为 HH:mm，且上下班时间不能相同）
	ShiftTimeInvalidCode = 10340
	// ShiftNotFoundCode 班次不存在
	ShiftNotFoundCode = 10341
	// ShiftRestTimeInvalidCode 休息时段非法（须成对填写，且休息开始早于休息结束）
	ShiftRestTimeInvalidCode = 10342
	// ShiftStatusInvalidCode 班次状态非法（须为 启用/停用）
	ShiftStatusInvalidCode = 10343

	// HRM 假勤设置：假期类型（HRM-11）业务错误码。
	// LeaveTypeCodeDuplicateCode 类型编码已存在
	LeaveTypeCodeDuplicateCode = 10344
	// LeaveTypeNotFoundCode 假期类型不存在
	LeaveTypeNotFoundCode = 10345
	// LeaveTypeInvalidCode 类型编码非法（须为小写字母+下划线）或状态非法
	LeaveTypeInvalidCode = 10346
)

func init() {
	status.Register(OrgParentNotFoundCode)
	status.Register(OrgNotFoundCode)
	status.Register(OrgHasChildrenCode)
	status.Register(OrgImportInvalidCode)

	status.Register(HrmBillTypeNotRegisteredCode)
	status.Register(HrmBillNotFoundCode)
	status.Register(HrmBillInvalidTransitionCode)
	status.Register(HrmBillInvalidTypeCode)
	status.Register(HrmBillNoSeqOutOfRangeCode)

	status.Register(ShiftTimeInvalidCode)
	status.Register(ShiftNotFoundCode)
	status.Register(ShiftRestTimeInvalidCode)
	status.Register(ShiftStatusInvalidCode)

	status.Register(LeaveTypeCodeDuplicateCode)
	status.Register(LeaveTypeNotFoundCode)
	status.Register(LeaveTypeInvalidCode)
}
