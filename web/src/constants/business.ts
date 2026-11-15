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
