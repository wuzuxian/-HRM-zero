package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// HRM 组织管理 error codes (1032x range, manage.go uses 10311).
	OrgParentNotFoundCode = 10321
	OrgNotFoundCode       = 10322
	OrgHasChildrenCode    = 10323
	OrgImportInvalidCode  = 10324
	// OrgMoveCycleCode 编辑组织时把父组织改到自己或自己的子孙上（会成环）
	OrgMoveCycleCode = 10325

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

	// HRM 薪酬设置：薪酬主体与薪资项目（HRM-33）业务错误码。
	// （10440-10443 已被候选人占用；薪资项目参数/预置守卫已重编号至 10590-10591。）
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
	// （code-review F3 重编号：原 10450/10451 与证书模板 10450-10451 撞号，
	// 10452-10458 已被证书发放占用 → 取空闲 1059x 段）。
	SalaryItemInvalidCode = 10590
	// SalaryItemPresetGuardCode 预置薪资项目不可删除（可停用）
	SalaryItemPresetGuardCode = 10591

	// HRM 假勤设置：加班规则（HRM-13）业务错误码。10462-10464 段
	// （1044x 已被候选人/薪酬占用，10450-10458 证书模板/发放，10460-10461 员工账号同步）。
	// OvertimeRuleNotFoundCode 加班规则不存在
	OvertimeRuleNotFoundCode = 10462
	// OvertimeRuleInvalidCode 加班规则参数非法（名称/当天策略/调休倍数/最小单位/月上限/考勤组绑定/状态等）
	OvertimeRuleInvalidCode = 10463
	// OvertimeRuleDayTypeInvalidCode 加班当天类型非法（须为 工作日/休息日/法定节假日）
	OvertimeRuleDayTypeInvalidCode = 10464

	// HRM 排班管理：排班周期（HRM-14）业务错误码。10480-10483 段（1047x 已被入职单占用；
	// 劳动合同/证照已重编号至 10485-10489，本段独占 1048x 前段）。
	// ScheduleCycleNotFoundCode 排班周期不存在
	ScheduleCycleNotFoundCode = 10480
	// ScheduleCycleInvalidCode 排班周期参数非法（名称/班次序列/考勤组绑定/状态等）
	ScheduleCycleInvalidCode = 10481
	// ScheduleCycleShiftInvalidCode 班次序列非法（班次不存在或已停用）
	ScheduleCycleShiftInvalidCode = 10482
	// ScheduleCycleGroupInvalidCode 适用考勤组不存在
	ScheduleCycleGroupInvalidCode = 10483
	// HRM 员工转正单（HRM-08，HR202）业务错误码。10465+ 段
	// （10462-10464 加班规则，10450-10458 证书模板/发放，10460-10461 员工账号同步）。
	// RegularizationNotFoundCode 转正单不存在
	RegularizationNotFoundCode = 10465
	// RegularizationInvalidCode 转正单参数非法（员工快照/转正日期格式等）
	RegularizationInvalidCode = 10466
	// RegularizationEmployeeRequiredCode 未指定员工（工号与姓名为必填快照）
	RegularizationEmployeeRequiredCode = 10467
	// RegularizationNotEditableCode 单据当前状态不允许修改/删除（仅 未提交/审批不通过/已撤销 可改）
	RegularizationNotEditableCode = 10468

	// HRM 劳动合同与证照（HRM-05）业务错误码。10485-10489 段
	// （code-review F3 重编号：原 10480-10484 与排班周期 10480-10483 撞号，
	// 10485+ 空闲且紧邻原段，保持语义连续；1049x 离职单）。
	// ContractNotFoundCode 劳动合同不存在
	ContractNotFoundCode = 10485
	// ContractInvalidCode 劳动合同参数非法（员工档案必填/日期格式/到期日早于生效日/状态非法等）
	ContractInvalidCode = 10486
	// ContractNotRenewableCode 合同当前状态不允许续签（仅 有效/已到期 可续签，已解除不可续签）
	ContractNotRenewableCode = 10487
	// EmployeeCertificateNotFoundCode 员工证照不存在
	EmployeeCertificateNotFoundCode = 10488
	// EmployeeCertificateInvalidCode 员工证照参数非法（员工档案/证书类型/证号必填等）
	EmployeeCertificateInvalidCode = 10489
	// HRM 员工离职单（HRM-10，HR204）业务错误码。1049x 段
	// （1048x 排班周期 + 劳动合同/证照（10485-10489），1047x 入职单，10465-10468 转正单）。
	// ResignationNotFoundCode 离职单不存在
	ResignationNotFoundCode = 10490
	// ResignationInvalidCode 离职单参数非法（离职类型/离职日期格式等）
	ResignationInvalidCode = 10491
	// ResignationEmployeeRequiredCode 未指定员工（工号与姓名为必填快照）
	ResignationEmployeeRequiredCode = 10492
	// ResignationNotEditableCode 单据当前状态不允许修改/删除（仅 未提交/审批不通过/已撤销 可改）
	ResignationNotEditableCode = 10493

	// HRM 招聘职位（HRM-25）业务错误码。1058x 段
	// （1056x 请假销假，1057x 预留；uuid 段 25xx，迁移 39）。
	// PositionNotFoundCode 招聘职位不存在
	PositionNotFoundCode = 10580
	// PositionInvalidCode 招聘职位参数非法（岗位名称/岗位类别/招聘人数/发布日期等）
	PositionInvalidCode = 10581
	// PositionRequirementNotApproved 来源需求单未审批通过，不可发布职位（仅 status=approved 可引用）
	PositionRequirementNotApprovedCode = 10582
	// PositionRequirementRequired 来源需求单必填
	PositionRequirementRequiredCode = 10583
	// PositionStatusTransitionInvalidCode 职位状态流转非法（已满/已关闭 不可手工改动；closed 终态）
	PositionStatusTransitionInvalidCode = 10584
	// PositionHiredCountExceeded 已录用人数超出招聘人数（递增后将超过 headcount）
	PositionHiredCountExceededCode = 10585
)

func init() {
	status.Register(OrgParentNotFoundCode)
	status.Register(OrgNotFoundCode)
	status.Register(OrgHasChildrenCode)
	status.Register(OrgImportInvalidCode)
	status.Register(OrgMoveCycleCode)

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

	status.Register(OvertimeRuleNotFoundCode)
	status.Register(OvertimeRuleInvalidCode)
	status.Register(OvertimeRuleDayTypeInvalidCode)

	status.Register(ScheduleCycleNotFoundCode)
	status.Register(ScheduleCycleInvalidCode)
	status.Register(ScheduleCycleShiftInvalidCode)
	status.Register(ScheduleCycleGroupInvalidCode)
	status.Register(RegularizationNotFoundCode)
	status.Register(RegularizationInvalidCode)
	status.Register(RegularizationEmployeeRequiredCode)
	status.Register(RegularizationNotEditableCode)

	status.Register(ContractNotFoundCode)
	status.Register(ContractInvalidCode)
	status.Register(ContractNotRenewableCode)
	status.Register(EmployeeCertificateNotFoundCode)
	status.Register(EmployeeCertificateInvalidCode)

	status.Register(ResignationNotFoundCode)
	status.Register(ResignationInvalidCode)
	status.Register(ResignationEmployeeRequiredCode)
	status.Register(ResignationNotEditableCode)

	status.Register(PositionNotFoundCode)
	status.Register(PositionInvalidCode)
	status.Register(PositionRequirementNotApprovedCode)
	status.Register(PositionRequirementRequiredCode)
	status.Register(PositionStatusTransitionInvalidCode)
	status.Register(PositionHiredCountExceededCode)
}
