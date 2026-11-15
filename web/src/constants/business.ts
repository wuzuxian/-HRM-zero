import { transformRecordToOption } from '@/utils/common';

export const enableStatusRecord: Record<Api.Common.EnableStatus, App.I18n.I18nKey> = {
  '1': 'page.manage.common.status.enable',
  '2': 'page.manage.common.status.disable'
};

export const enableStatusOptions = transformRecordToOption(enableStatusRecord);

export const userGenderRecord: Record<Api.Manage.UserGender, App.I18n.I18nKey> = {
  '1': 'page.manage.user.gender.male',
  '2': 'page.manage.user.gender.female'
};

export const userGenderOptions = transformRecordToOption(userGenderRecord);

export const menuTypeRecord: Record<Api.Manage.MenuType, App.I18n.I18nKey> = {
  '1': 'page.manage.menu.type.directory',
  '2': 'page.manage.menu.type.menu',
  '3': 'page.manage.menu.type.button'
};

export const menuTypeOptions = transformRecordToOption(menuTypeRecord);

export const menuIconTypeRecord: Record<Api.Manage.IconType, App.I18n.I18nKey> = {
  '1': 'page.manage.menu.iconType.iconify',
  '2': 'page.manage.menu.iconType.local'
};

export const menuIconTypeOptions = transformRecordToOption(menuIconTypeRecord);

export const hrmOrgTypeRecord: Record<Api.Hrm.OrgType, App.I18n.I18nKey> = {
  '1': 'page.hrm.org.orgType.company',
  '2': 'page.hrm.org.orgType.department'
};

export const hrmOrgTypeOptions = transformRecordToOption(hrmOrgTypeRecord);

export const hrmOrgStatusRecord: Record<Api.Hrm.OrgStatus, App.I18n.I18nKey> = {
  '1': 'page.hrm.org.status.open',
  '0': 'page.hrm.org.status.closed'
};

export const hrmOrgStatusOptions = transformRecordToOption(hrmOrgStatusRecord);

export const hrmAttendanceStatusRecord: Record<Api.Hrm.AttendanceStatus, App.I18n.I18nKey> = {
  '1': 'page.hrm.attendance.status.enabled',
  '0': 'page.hrm.attendance.status.disabled'
};

export const hrmAttendanceStatusOptions = transformRecordToOption(hrmAttendanceStatusRecord);

/** 单据状态 (shared approval state machine, GLOSSARY) */
export const hrmBillStatusRecord: Record<Api.Hrm.BillStatus, App.I18n.I18nKey> = {
  unsubmitted: 'page.hrm.employeeRelation.billStatus.unsubmitted',
  approving: 'page.hrm.employeeRelation.billStatus.approving',
  approved: 'page.hrm.employeeRelation.billStatus.approved',
  rejected: 'page.hrm.employeeRelation.billStatus.rejected',
  revoked: 'page.hrm.employeeRelation.billStatus.revoked'
};

export const hrmBillStatusOptions = transformRecordToOption(hrmBillStatusRecord);

/** 异动类型 (人事调动 HR203) */
export const hrmTransferChangeTypeRecord: Record<Api.Hrm.TransferChangeType, App.I18n.I18nKey> = {
  post: 'page.hrm.employeeRelation.transfer.changeType.post',
  org: 'page.hrm.employeeRelation.transfer.changeType.org',
  salary: 'page.hrm.employeeRelation.transfer.changeType.salary',
  company: 'page.hrm.employeeRelation.transfer.changeType.company',
  other: 'page.hrm.employeeRelation.transfer.changeType.other'
};

export const hrmTransferChangeTypeOptions = transformRecordToOption(hrmTransferChangeTypeRecord);

/** 奖惩类型 (奖惩单 HR215) */
export const hrmRewardTypeRecord: Record<Api.Hrm.RewardType, App.I18n.I18nKey> = {
  reward: 'page.hrm.reward.rewardType.reward',
  punishment: 'page.hrm.reward.rewardType.punishment'
};

export const hrmRewardTypeOptions = transformRecordToOption(hrmRewardTypeRecord);

/** 奖惩级别 (奖惩单 HR215) */
export const hrmRewardLevelRecord: Record<Api.Hrm.RewardLevel, App.I18n.I18nKey> = {
  company: 'page.hrm.reward.rewardLevel.company',
  dept: 'page.hrm.reward.rewardLevel.dept',
  notice: 'page.hrm.reward.rewardLevel.notice',
  other: 'page.hrm.reward.rewardLevel.other'
};

export const hrmRewardLevelOptions = transformRecordToOption(hrmRewardLevelRecord);

/** 岗位类别 (招聘需求单 HR207) */
export const hrmRequirementPostCategoryRecord: Record<Api.Hrm.RequirementPostCategory, App.I18n.I18nKey> = {
  tech: 'page.hrm.requirement.postCategory.tech',
  product: 'page.hrm.requirement.postCategory.product',
  design: 'page.hrm.requirement.postCategory.design',
  operation: 'page.hrm.requirement.postCategory.operation',
  market: 'page.hrm.requirement.postCategory.market',
  function: 'page.hrm.requirement.postCategory.function',
  other: 'page.hrm.requirement.postCategory.other'
};

export const hrmRequirementPostCategoryOptions = transformRecordToOption(hrmRequirementPostCategoryRecord);

/** 紧急度 (招聘需求单 HR207) */
export const hrmRequirementUrgencyRecord: Record<Api.Hrm.RequirementUrgency, App.I18n.I18nKey> = {
  normal: 'page.hrm.requirement.urgency.normal',
  urgent: 'page.hrm.requirement.urgency.urgent'
};

export const hrmRequirementUrgencyOptions = transformRecordToOption(hrmRequirementUrgencyRecord);

export const hrmTrainingSessionStatusRecord: Record<Api.Hrm.TrainingSessionStatus, App.I18n.I18nKey> = {
  '1': 'page.hrm.training.sessionStatus.notStarted',
  '2': 'page.hrm.training.sessionStatus.ongoing',
  '3': 'page.hrm.training.sessionStatus.finished'
};

export const hrmTrainingSessionStatusOptions = transformRecordToOption(hrmTrainingSessionStatusRecord);

export const hrmTrainingCertificateStatusRecord: Record<Api.Hrm.TrainingCertificateStatus, App.I18n.I18nKey> = {
  '0': 'page.hrm.training.certStatus.pending',
  '1': 'page.hrm.training.certStatus.issued',
  '2': 'page.hrm.training.certStatus.revoked'
};

export const hrmTrainingCertificateStatusOptions = transformRecordToOption(hrmTrainingCertificateStatusRecord);

/** 培训计划 计划类型 (HRM-28 HR216) */
export const hrmTrainingPlanTypeRecord: Record<Api.Hrm.TrainingPlanType, App.I18n.I18nKey> = {
  special: 'page.hrm.training.planType.special',
  new_hire: 'page.hrm.training.planType.newHire',
  safety: 'page.hrm.training.planType.safety',
  skill: 'page.hrm.training.planType.skill'
};

export const hrmTrainingPlanTypeOptions = transformRecordToOption(hrmTrainingPlanTypeRecord);

/** 培训计划 发布状态 (0=草稿 1=已发布, 审批通过即发布) */
export const hrmTrainingPlanPublishedRecord: Record<'0' | '1', App.I18n.I18nKey> = {
  '0': 'page.hrm.training.published.draft',
  '1': 'page.hrm.training.published.published'
};

export const hrmTrainingPlanPublishedOptions = transformRecordToOption(hrmTrainingPlanPublishedRecord);

// ---- HRM 证书模板与证书发放 (HRM-30) ----

/** 证书模板 来源 (1=系统 2=企业) */
export const hrmCertTemplateSourceRecord: Record<Api.Hrm.CertTemplateSource, App.I18n.I18nKey> = {
  '1': 'page.hrm.training.certTemplate.source.system',
  '2': 'page.hrm.training.certTemplate.source.org'
};

export const hrmCertTemplateSourceOptions = transformRecordToOption(hrmCertTemplateSourceRecord);

/** 证书模板 状态 (1=启用 2=停用) */
export const hrmCertTemplateStatusRecord: Record<Api.Hrm.CertTemplateStatus, App.I18n.I18nKey> = {
  '1': 'page.hrm.training.certTemplate.status.enabled',
  '2': 'page.hrm.training.certTemplate.status.disabled'
};

export const hrmCertTemplateStatusOptions = transformRecordToOption(hrmCertTemplateStatusRecord);

/** 证书发放 来源 (1=班次结业 2=外部导入) */
export const hrmCertificateSourceRecord: Record<Api.Hrm.TrainingCertificateSource, App.I18n.I18nKey> = {
  '1': 'page.hrm.training.certificate.source.session',
  '2': 'page.hrm.training.certificate.source.import'
};

export const hrmCertificateSourceOptions = transformRecordToOption(hrmCertificateSourceRecord);

// ---- HRM 员工档案 (HRM-03) ----

/** 性别 1=男 2=女 */
export const hrmEmployeeGenderRecord: Record<Api.Hrm.Gender, App.I18n.I18nKey> = {
  '1': 'page.hrm.employee.gender.male',
  '2': 'page.hrm.employee.gender.female'
};

export const hrmEmployeeGenderOptions = transformRecordToOption(hrmEmployeeGenderRecord);

/** 婚姻状况 1=未婚 2=已婚 3=离异 4=丧偶 */
export const hrmEmployeeMaritalStatusRecord: Record<Api.Hrm.MaritalStatus, App.I18n.I18nKey> = {
  '1': 'page.hrm.employee.maritalStatus.unmarried',
  '2': 'page.hrm.employee.maritalStatus.married',
  '3': 'page.hrm.employee.maritalStatus.divorced',
  '4': 'page.hrm.employee.maritalStatus.widowed'
};

export const hrmEmployeeMaritalStatusOptions = transformRecordToOption(hrmEmployeeMaritalStatusRecord);

/** 人员状态 1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休 */
export const hrmEmployeePersonStatusRecord: Record<Api.Hrm.PersonStatus, App.I18n.I18nKey> = {
  '1': 'page.hrm.employee.personStatus.regular',
  '2': 'page.hrm.employee.personStatus.probation',
  '3': 'page.hrm.employee.personStatus.intern',
  '4': 'page.hrm.employee.personStatus.temp',
  '5': 'page.hrm.employee.personStatus.resigned',
  '6': 'page.hrm.employee.personStatus.retired'
};

export const hrmEmployeePersonStatusOptions = transformRecordToOption(hrmEmployeePersonStatusRecord);

/** 用户状态 1=已生成系统账号 0=未生成 */
export const hrmEmployeeUserStatusRecord: Record<Api.Hrm.UserStatus, App.I18n.I18nKey> = {
  '1': 'page.hrm.employee.userStatus.created',
  '0': 'page.hrm.employee.userStatus.none'
};

/** 候选人状态 (HRM-26 候选人中心): 新线索→已投递→面试中→已发Offer→录用通过; 不通过 */
export const hrmCandidateStatusRecord: Record<Api.Hrm.CandidateStatus, App.I18n.I18nKey> = {
  new: 'page.hrm.candidate.status.new',
  delivered: 'page.hrm.candidate.status.delivered',
  interviewing: 'page.hrm.candidate.status.interviewing',
  offer: 'page.hrm.candidate.status.offer',
  hired: 'page.hrm.candidate.status.hired',
  rejected: 'page.hrm.candidate.status.rejected'
};

export const hrmCandidateStatusOptions = transformRecordToOption(hrmCandidateStatusRecord);

/** 候选人来源渠道 (demo-inventory §5 筛选项) */
export const hrmCandidateSourceChannelRecord: Record<Api.Hrm.CandidateSourceChannel, App.I18n.I18nKey> = {
  recommend: 'page.hrm.candidate.sourceChannel.recommend',
  website: 'page.hrm.candidate.sourceChannel.website',
  campus: 'page.hrm.candidate.sourceChannel.campus',
  headhunter: 'page.hrm.candidate.sourceChannel.headhunter',
  other: 'page.hrm.candidate.sourceChannel.other'
};

export const hrmCandidateSourceChannelOptions = transformRecordToOption(hrmCandidateSourceChannelRecord);

/** 候选人最高学历 (demo-inventory §5 列) */
export const hrmCandidateEducationRecord: Record<Api.Hrm.CandidateEducation, App.I18n.I18nKey> = {
  doctor: 'page.hrm.candidate.education.doctor',
  master: 'page.hrm.candidate.education.master',
  bachelor: 'page.hrm.candidate.education.bachelor',
  associate: 'page.hrm.candidate.education.associate',
  senior: 'page.hrm.candidate.education.senior',
  other: 'page.hrm.candidate.education.other'
};

export const hrmCandidateEducationOptions = transformRecordToOption(hrmCandidateEducationRecord);

/** 薪酬主体/薪资项目 启停状态 (HRM-33, 与考勤/组织 1=启用 0=停用 同构) */
export const hrmSalaryStatusRecord: Record<Api.Hrm.SalarySubjectStatus, App.I18n.I18nKey> = {
  '1': 'page.hrm.attendance.status.enabled',
  '0': 'page.hrm.attendance.status.disabled'
};

export const hrmSalaryStatusOptions = transformRecordToOption(hrmSalaryStatusRecord);

/** 薪资项目类型 (GLOSSARY「薪资项目」): 固定项/取数项/公式项 */
export const hrmSalaryItemTypeRecord: Record<Api.Hrm.SalaryItemType, App.I18n.I18nKey> = {
  fixed: 'page.hrm.salary.item.itemType.fixed',
  data: 'page.hrm.salary.item.itemType.data',
  formula: 'page.hrm.salary.item.itemType.formula'
};

export const hrmSalaryItemTypeOptions = transformRecordToOption(hrmSalaryItemTypeRecord);

/** 薪资项目分类: 应发/应扣/成本 */
export const hrmSalaryItemClassRecord: Record<Api.Hrm.SalaryItemClass, App.I18n.I18nKey> = {
  earning: 'page.hrm.salary.item.itemClass.earning',
  deduction: 'page.hrm.salary.item.itemClass.deduction',
  cost: 'page.hrm.salary.item.itemClass.cost'
};

export const hrmSalaryItemClassOptions = transformRecordToOption(hrmSalaryItemClassRecord);

/** 取值方式: 手工维护/系统取数/公式计算 */
export const hrmSalaryItemValueModeRecord: Record<Api.Hrm.SalaryItemValueMode, App.I18n.I18nKey> = {
  manual: 'page.hrm.salary.item.valueMode.manual',
  system: 'page.hrm.salary.item.valueMode.system',
  formula: 'page.hrm.salary.item.valueMode.formula'
};

export const hrmSalaryItemValueModeOptions = transformRecordToOption(hrmSalaryItemValueModeRecord);

/** HRM-05 合同预警状态 normal/warning(≤90天)/expired/terminated（terminated 优先于天数口径） */
export const contractExpiryStateRecord: Record<Api.Hrm.ContractExpiryState, App.I18n.I18nKey> = {
  normal: 'page.hrm.employee.contractExpiring.stateNormal',
  warning: 'page.hrm.employee.contractExpiring.stateWarning',
  expired: 'page.hrm.employee.contractExpiring.stateExpired',
  terminated: 'page.hrm.employee.contractExpiring.stateTerminated'
};

export const contractExpiryStateOptions = transformRecordToOption(contractExpiryStateRecord);

/** HRM-05 证照预警状态 normal/warning(≤90天)/expired */
export const certExpiryStateRecord: Record<Api.Hrm.CertExpiryState, App.I18n.I18nKey> = {
  normal: 'page.hrm.employee.certExpiring.stateNormal',
  warning: 'page.hrm.employee.certExpiring.stateWarning',
  expired: 'page.hrm.employee.certExpiring.stateExpired'
};

export const certExpiryStateOptions = transformRecordToOption(certExpiryStateRecord);
