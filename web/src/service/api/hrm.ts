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

/** get 员工档案 paginated list (筛选: 工号/姓名/部门/人员状态) */
export function GetEmployeeList(params?: Api.Hrm.EmployeeSearchParams) {
  return request<Api.Hrm.EmployeeList>({
    url: '/api/v1/hrm/getEmployeeList',
    method: 'get',
    params
  });
}

/** get 员工档案 detail incl. 教育经历/工作经历/家属 sub-tables */
export function GetEmployeeDetail(uuid: string) {
  return request<Api.Hrm.EmployeeDetailResponse>({
    url: '/api/v1/hrm/getEmployeeDetail',
    method: 'get',
    params: { uuid }
  });
}

/** add 员工档案; employeeNo 为空时后端按 最大工号+1 (8位补零) 自动生成 */
export function AddEmployee(req: Api.Hrm.AddEmployeeRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addEmployee',
    method: 'post',
    data: req
  });
}

/** edit 员工档案 (从表全删全插) */
export function EditEmployee(req: Api.Hrm.EditEmployeeRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editEmployee',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 员工档案 (级联删除三个从表) */
export function DeleteEmployee(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteEmployee',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** import 员工档案 from a base64-encoded xlsx file */
export function ImportEmployee(req: Api.Hrm.ImportEmployeeRequest) {
  return request<Api.Hrm.ImportEmployeeResponse>({
    url: '/api/v1/hrm/importEmployee',
    method: 'post',
    data: req
  });
}

/** export 员工档案 as a base64-encoded xlsx file */
export function ExportEmployee() {
  return request<Api.Hrm.ExportEmployeeResponse>({
    url: '/api/v1/hrm/exportEmployee',
    method: 'get'
  });
}

// ---- HRM 培训计划单 (HRM-28, HR216) ----

/** get 培训计划单 paginated list */
export function GetTrainingPlanList(params?: Api.Hrm.TrainingPlanSearchParams) {
  return request<Api.Hrm.TrainingPlanList>({
    url: '/api/v1/hrm/getTrainingPlanList',
    method: 'get',
    params
  });
}

/** add 培训计划单 (单据编号 HR216-... 后端自动生成, 初始 未提交/草稿) */
export function AddTrainingPlan(req: Api.Hrm.AddTrainingPlanRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addTrainingPlan',
    method: 'post',
    data: req
  });
}

/** edit 培训计划单 (仅未发布且单据状态可编辑时) */
export function EditTrainingPlan(req: Api.Hrm.EditTrainingPlanRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editTrainingPlan',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 培训计划单 */
export function DeleteTrainingPlan(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteTrainingPlan',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

// ---- HRM 培训班次 (HRM-28, 管理端 CRUD) ----

/** get 培训班次 paginated list */
export function GetTrainingSessionList(params?: Api.Hrm.TrainingSessionSearchParams) {
  return request<Api.Hrm.TrainingSessionList>({
    url: '/api/v1/hrm/getTrainingSessionList',
    method: 'get',
    params
  });
}

/** add 培训班次 (班次编号 PX-年-流水 后端自动生成; 关联计划须已发布) */
export function AddTrainingSession(req: Api.Hrm.AddTrainingSessionRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addTrainingSession',
    method: 'post',
    data: req
  });
}

/** edit 培训班次 */
export function EditTrainingSession(req: Api.Hrm.EditTrainingSessionRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editTrainingSession',
    method: 'post',
    data: req
  });
}

/** finish 培训班次 (行操作: 状态置已结束) */
export function FinishTrainingSession(uuid: string) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/finishTrainingSession',
    method: 'post',
    data: { uuid }
  });
}

/** delete/batch delete 培训班次 (级联清理学员绑定) */
export function DeleteTrainingSession(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteTrainingSession',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get 班次学员 list (选人弹窗回显) */
export function GetSessionParticipantList(sessionUuid: string) {
  return request<Api.Hrm.SessionParticipantList>({
    url: '/api/v1/hrm/getSessionParticipantList',
    method: 'get',
    params: { sessionUuid }
  });
}

/** set 班次学员 (选人, 全量覆盖) */
export function SetSessionParticipants(req: Api.Hrm.SetSessionParticipantsRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/setSessionParticipants',
    method: 'post',
    data: req
  });
}

/** get 考勤组 paginated list (筛选: 名称/班制/状态) */
export function GetAttendanceGroupList(params?: Api.Hrm.AttendanceGroupSearchParams) {
  return request<Api.Hrm.AttendanceGroupList>({
    url: '/api/v1/hrm/getAttendanceGroupList',
    method: 'get',
    params
  });
}

/** add 考勤组 */
export function AddAttendanceGroup(req: Api.Hrm.AddAttendanceGroupRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addAttendanceGroup',
    method: 'post',
    data: req
  });
}

/** edit 考勤组 */
export function EditAttendanceGroup(req: Api.Hrm.EditAttendanceGroupRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editAttendanceGroup',
    method: 'post',
    data: req
  });
}

/** toggle 考勤组 启用/停用 (停用后三级回退跳过该考勤组) */
export function ToggleAttendanceGroupStatus(req: { uuid: string; status: Api.Hrm.AttendanceStatus }) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/toggleAttendanceGroupStatus',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 考勤组 */
export function DeleteAttendanceGroup(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteAttendanceGroup',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get 考勤组 生效人数 (部门含下级 + 指定人员, 去重); uuid 缺省返回所有启用考勤组并集 */
export function GetAttendanceGroupMembers(uuid?: string) {
  return request<Api.Hrm.AttendanceGroupMembersResponse>({
    url: '/api/v1/hrm/getAttendanceGroupMembers',
    method: 'get',
    params: { uuid }
  });
}

/** 生效规则查询 (三级回退: 考勤组 > 本部门规则 > 最近上级组织规则); 按员工 uuid 或工号 */
export function GetEffectiveAttendanceRule(params: { employeeUuid?: string; employeeNo?: string }) {
  return request<Api.Hrm.EffectiveAttendanceRuleResponse>({
    url: '/api/v1/hrm/getEffectiveAttendanceRule',
    method: 'get',
    params
  });
}

/** 班次引用查询 (支撑班次管理页 引用考勤组 列) */
export function GetAttendanceGroupShiftUsage(shiftUuid?: string) {
  return request<Api.Hrm.AttendanceGroupShiftUsageResponse>({
    url: '/api/v1/hrm/getAttendanceGroupShiftUsage',
    method: 'get',
    params: { shiftUuid }
  });
}

/** get 组织默认规则 paginated list (筛选: 组织名称) */
export function GetAttendanceConfigList(params?: Api.Hrm.AttendanceConfigSearchParams) {
  return request<Api.Hrm.AttendanceConfigList>({
    url: '/api/v1/hrm/getAttendanceConfigList',
    method: 'get',
    params
  });
}

/** save 组织默认规则 (upsert: 快速开始/新增/编辑共用) */
export function SaveAttendanceConfig(req: Api.Hrm.SaveAttendanceConfigRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/saveAttendanceConfig',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 组织默认规则 */
export function DeleteAttendanceConfig(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteAttendanceConfig',
    method: 'post',
    data: {
      uuids: req
    }
  });
}
