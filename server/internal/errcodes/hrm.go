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

	// HRM 奖惩单（HRM-07，HR215）业务错误码。1042x 段（1039x 为绩效重编号预留；
	// 1040x/1041x 已被 HRM-28 培训计划/班次占用，见 hrm_training.go）。
	// RewardNotFoundCode 奖惩单不存在
	RewardNotFoundCode = 10420
	// RewardInvalidCode 奖惩单参数非法（奖惩类型/级别/发生日期/入账年月等）
	RewardInvalidCode = 10421
	// RewardEmployeeRequiredCode 未指定员工（工号与姓名为必填快照）
	RewardEmployeeRequiredCode = 10422
	// RewardNotEditableCode 单据当前状态不允许修改/删除（仅 未提交/审批不通过/已撤销 可改）
	RewardNotEditableCode = 10423

	// HRM 招聘需求单（HRM-24，HR207）业务错误码。1043x 段（1042x 已被奖惩单占用）。
	// RequirementNotFoundCode 招聘需求单不存在
	RequirementNotFoundCode = 10430
	// RequirementInvalidCode 招聘需求单参数非法（岗位名称/类别/人数/紧急度/日期等）
	RequirementInvalidCode = 10431
	// RequirementOrgRequiredCode 未指定所属部门
	RequirementOrgRequiredCode = 10432
	// RequirementNotEditableCode 单据当前状态不允许修改/删除（仅 未提交/审批不通过/已撤销 可改）
	RequirementNotEditableCode = 10433

	// HRM 候选人中心（HRM-26）业务错误码。10440-10441（10441 与 HRM-24 seam 注释中
	// 「需求未审批通过不可发布职位」候选码重叠，最终归属以合并两侧声明为准；
	// 10442+ 预留给 HRM-25 录用审批）。
	// CandidateNotFoundCode 候选人不存在
	CandidateNotFoundCode = 10440
	// CandidateInvalidCode 候选人参数非法（姓名/来源渠道/最高学历等）
	CandidateInvalidCode = 10441
	// CandidateStatusTransitionInvalidCode 候选人状态流转非法（不可跳级/不可回退）
	CandidateStatusTransitionInvalidCode = 10442
	// CandidateImportInvalidCode 导入候选人文件非法
	CandidateImportInvalidCode = 10443

	// HRM 薪酬设置：薪酬主体与薪资项目（HRM-33）业务错误码。1044x 段
	// （10440-10443 已被候选人占用，本段自 10444 起）。
	// SalarySubjectNotFoundCode 薪酬主体不存在
	SalarySubjectNotFoundCode = 10444
	// SalarySubjectInvalidCode 薪酬主体参数非法（名称/关联公司/启停状态等）
	SalarySubjectInvalidCode = 10445
	// SalarySubjectOrgNotFoundCode 关联公司不存在（须为 hrm_org 中 org_type=1 的公司节点）
	SalarySubjectOrgNotFoundCode = 10446
	// SalarySubjectOrgInvalidCode 关联组织不是公司节点（org_type != 1）
	SalarySubjectOrgInvalidCode = 10447
	// SalaryItemNotFoundCode 薪资项目不存在
	SalaryItemNotFoundCode = 10448
	// SalaryItemCodeDuplicateCode 薪资项目编码已存在
	SalaryItemCodeDuplicateCode = 10449
	// SalaryItemInvalidCode 薪资项目参数非法（名称/编码/项目类型/项目分类/取值方式/状态等）
	SalaryItemInvalidCode = 10450
	// SalaryItemPresetGuardCode 预置薪资项目不可删除（可停用）
	SalaryItemPresetGuardCode = 10451
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

	status.Register(RequirementNotFoundCode)
	status.Register(RequirementInvalidCode)
	status.Register(RequirementOrgRequiredCode)
	status.Register(RequirementNotEditableCode)

	status.Register(CandidateNotFoundCode)
	status.Register(CandidateInvalidCode)
	status.Register(CandidateStatusTransitionInvalidCode)
	status.Register(CandidateImportInvalidCode)

	status.Register(SalarySubjectNotFoundCode)
	status.Register(SalarySubjectInvalidCode)
	status.Register(SalarySubjectOrgNotFoundCode)
	status.Register(SalarySubjectOrgInvalidCode)
	status.Register(SalaryItemNotFoundCode)
	status.Register(SalaryItemCodeDuplicateCode)
	status.Register(SalaryItemInvalidCode)
	status.Register(SalaryItemPresetGuardCode)
}
