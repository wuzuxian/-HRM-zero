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

	// HRM 培训管理：我的培训（HRM-31）业务错误码。
	// TrainingCertificateNotFoundCode 证书不存在或无权访问（我的培训仅限证书本人）
	TrainingCertificateNotFoundCode = 10360
	// TrainingCertificateFileNotFoundCode 证书扫描件不存在（未上传或已清除）
	TrainingCertificateFileNotFoundCode = 10361

	// HRM 人事档案：员工档案（HRM-03）业务错误码。
	// EmployeeNotFoundCode 员工不存在
	EmployeeNotFoundCode = 10350
	// EmployeeNoExhaustedCode 工号流水已超出 8 位上限
	EmployeeNoExhaustedCode = 10351
	// EmployeeNoDuplicateCode 工号已存在
	EmployeeNoDuplicateCode = 10352
	// EmployeeInvalidCode 员工档案参数非法（姓名/性别/婚姻状况/人员状态等）
	EmployeeInvalidCode = 10353
	// EmployeeImportInvalidCode 导入员工文件非法
	EmployeeImportInvalidCode = 10354

	// HRM 人事调动单（HRM-09，HR203）业务错误码。
	// TransferNotFoundCode 调动单不存在
	TransferNotFoundCode = 10370
	// TransferInvalidCode 调动单参数非法（异动类型/生效设置等）
	TransferInvalidCode = 10371
	// TransferEmployeeRequiredCode 未指定员工（工号与姓名为必填快照）
	TransferEmployeeRequiredCode = 10372
	// TransferNotEditableCode 单据当前状态不允许修改/删除（仅 未提交/审批不通过/已撤销 可改）
	TransferNotEditableCode = 10373

	// HRM 考勤组与组织默认规则（HRM-12）业务错误码。
	// AttendanceGroupNotFoundCode 考勤组不存在
	AttendanceGroupNotFoundCode = 10380
	// AttendanceGroupInvalidCode 考勤组参数非法（名称/班制/考勤时间/考勤日/状态等）
	AttendanceGroupInvalidCode = 10381
	// AttendanceConfigOrgNotFoundCode 组织默认规则对应的组织不存在
	AttendanceConfigOrgNotFoundCode = 10382
	// AttendanceConfigInvalidCode 组织默认规则参数非法（考勤时间/考勤日等）
	AttendanceConfigInvalidCode = 10383
	// AttendanceRuleNotFoundCode 未找到生效规则（员工不在考勤组且组织链上无默认规则）
	AttendanceRuleNotFoundCode = 10384

	// HRM 奖惩单（HRM-07，HR215）业务错误码。1041x 段（1039x 为绩效重编号预留，1040x 跳过以防并行工单撞车）。
	// RewardNotFoundCode 奖惩单不存在
	RewardNotFoundCode = 10410
	// RewardInvalidCode 奖惩单参数非法（奖惩类型/级别/发生日期/入账年月等）
	RewardInvalidCode = 10411
	// RewardEmployeeRequiredCode 未指定员工（工号与姓名为必填快照）
	RewardEmployeeRequiredCode = 10412
	// RewardNotEditableCode 单据当前状态不允许修改/删除（仅 未提交/审批不通过/已撤销 可改）
	RewardNotEditableCode = 10413
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

	status.Register(AttendanceGroupNotFoundCode)
	status.Register(AttendanceGroupInvalidCode)
	status.Register(AttendanceConfigOrgNotFoundCode)
	status.Register(AttendanceConfigInvalidCode)
	status.Register(AttendanceRuleNotFoundCode)

	status.Register(TrainingCertificateNotFoundCode)
	status.Register(TrainingCertificateFileNotFoundCode)

	status.Register(EmployeeNotFoundCode)
	status.Register(EmployeeNoExhaustedCode)
	status.Register(EmployeeNoDuplicateCode)
	status.Register(EmployeeInvalidCode)
	status.Register(EmployeeImportInvalidCode)

	status.Register(TransferNotFoundCode)
	status.Register(TransferInvalidCode)
	status.Register(TransferEmployeeRequiredCode)
	status.Register(TransferNotEditableCode)

	status.Register(RewardNotFoundCode)
	status.Register(RewardInvalidCode)
	status.Register(RewardEmployeeRequiredCode)
	status.Register(RewardNotEditableCode)
}
