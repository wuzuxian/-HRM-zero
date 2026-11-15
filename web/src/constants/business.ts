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
