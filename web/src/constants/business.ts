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
