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
