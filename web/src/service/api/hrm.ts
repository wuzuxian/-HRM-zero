import { request } from '../request';

/** get hrm org tree */
export function GetOrgTree(params?: Api.Hrm.OrgSearchParams) {
  return request<Api.Hrm.OrgTree>({
    url: '/api/v1/hrm/getOrgTree',
    method: 'get',
    params
  });
}

/** add hrm org (root or with parent) */
export function AddOrg(req: Api.Hrm.AddOrgRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addOrg',
    method: 'post',
    data: req
  });
}

/** add a child hrm org under an existing one */
export function AddOrgChild(req: Api.Hrm.AddOrgChildRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addOrgChild',
    method: 'post',
    data: req
  });
}

/** edit hrm org */
export function EditOrg(req: Api.Hrm.EditOrgRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editOrg',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm orgs */
export function DeleteOrg(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteOrg',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** import hrm orgs from a base64-encoded xlsx file */
export function ImportOrg(req: Api.Hrm.ImportOrgRequest) {
  return request<Api.Hrm.ImportOrgResponse>({
    url: '/api/v1/hrm/importOrg',
    method: 'post',
    data: req
  });
}

/** export hrm orgs as a base64-encoded xlsx file */
export function ExportOrg() {
  return request<Api.Hrm.ExportOrgResponse>({
    url: '/api/v1/hrm/exportOrg',
    method: 'get'
  });
}

/** get the 可视化组织架构图 tree (负责人/本级人数/汇总人数) */
export function GetOrgChartTree() {
  return request<Api.Hrm.OrgChartTree>({
    url: '/api/v1/hrm/getOrgChartTree',
    method: 'get'
  });
}

/** get hrm shift (班次) paginated list */
export function GetShiftList(params?: Api.Hrm.ShiftSearchParams) {
  return request<Api.Hrm.ShiftList>({
    url: '/api/v1/hrm/getShiftList',
    method: 'get',
    params
  });
}

/** add hrm shift */
export function AddShift(req: Api.Hrm.AddShiftRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addShift',
    method: 'post',
    data: req
  });
}

/** edit hrm shift */
export function EditShift(req: Api.Hrm.EditShiftRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editShift',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm shifts */
export function DeleteShift(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteShift',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get hrm leave type (假期类型) paginated list */
export function GetLeaveTypeList(params?: Api.Hrm.LeaveTypeSearchParams) {
  return request<Api.Hrm.LeaveTypeList>({
    url: '/api/v1/hrm/getLeaveTypeList',
    method: 'get',
    params
  });
}

/** add hrm leave type */
export function AddLeaveType(req: Api.Hrm.AddLeaveTypeRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addLeaveType',
    method: 'post',
    data: req
  });
}

/** edit hrm leave type */
export function EditLeaveType(req: Api.Hrm.EditLeaveTypeRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editLeaveType',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm leave types */
export function DeleteLeaveType(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteLeaveType',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get hrm transfer bill (人事调动单 HR203) paginated list — defaults to 「我的单据」 server-side */
export function GetTransferList(params?: Api.Hrm.TransferSearchParams) {
  return request<Api.Hrm.TransferList>({
    url: '/api/v1/hrm/getTransferList',
    method: 'get',
    params
  });
}

/** get my training session (我的班次) paginated list, scoped to current login user */
export function GetMyTrainingSessionList(params?: Api.Hrm.MyTrainingSessionSearchParams) {
  return request<Api.Hrm.MyTrainingSessionList>({
    url: '/api/v1/hrm/getMyTrainingSessionList',
    method: 'get',
    params
  });
}

/** add hrm transfer bill */
export function AddTransfer(req: Api.Hrm.AddTransferRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addTransfer',
    method: 'post',
    data: req
  });
}

/** edit hrm transfer bill (only in editable 单据状态) */
export function EditTransfer(req: Api.Hrm.EditTransferRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editTransfer',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm transfer bills (only in editable 单据状态) */
export function DeleteTransfer(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteTransfer',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** submit a bill (单据提交, shared approval base; billType "203" = 人事调动) */
export function SubmitBill(req: Api.Hrm.BillActionRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/submitBill',
    method: 'post',
    data: req
  });
}

/** approve a bill (审批通过) */
export function ApproveBill(req: Api.Hrm.BillActionRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/approveBill',
    method: 'post',
    data: req
  });
}

/** reject a bill (审批驳回) */
export function RejectBill(req: Api.Hrm.BillActionRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/rejectBill',
    method: 'post',
    data: req
  });
}

/** revoke a bill (撤销) */
export function RevokeBill(req: Api.Hrm.BillActionRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/revokeBill',
    method: 'post',
    data: req
  });
}

/** get shared approval records (审批记录) of one bill */
export function GetApprovalRecordList(params: { billType: string; billUuid: string }) {
  return request<{ records: Api.Hrm.ApprovalRecord[] }>({
    url: '/api/v1/hrm/getApprovalRecordList',
    method: 'get',
    params
  });
}

/** get my training certificate (我的培训证书) paginated list, scoped to current login user */
export function GetMyTrainingCertificateList(params?: Api.Hrm.MyTrainingCertificateSearchParams) {
  return request<Api.Hrm.MyTrainingCertificateList>({
    url: '/api/v1/hrm/getMyTrainingCertificateList',
    method: 'get',
    params
  });
}

/** download/preview my certificate scan file as base64 content */
export function DownloadMyCertificateFile(uuid: string) {
  return request<Api.Hrm.DownloadMyCertificateFileResponse>({
    url: '/api/v1/hrm/downloadMyTrainingCertificateFile',
    method: 'get',
    params: { uuid }
  });
}
