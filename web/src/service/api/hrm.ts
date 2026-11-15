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

/** get hrm regularization bill (员工转正单 HR202) paginated list — defaults to 「我的单据」 server-side */
export function GetRegularizationList(params?: Api.Hrm.RegularizationSearchParams) {
  return request<Api.Hrm.RegularizationList>({
    url: '/api/v1/hrm/getRegularizationList',
    method: 'get',
    params
  });
}

/** add hrm regularization bill (选择员工带出档案快照) */
export function AddRegularization(req: Api.Hrm.AddRegularizationRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addRegularization',
    method: 'post',
    data: req
  });
}

/** edit hrm regularization bill (only in editable 单据状态) */
export function EditRegularization(req: Api.Hrm.EditRegularizationRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editRegularization',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm regularization bills (only in editable 单据状态) */
export function DeleteRegularization(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteRegularization',
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

/** 生成系统账号 (HRM-04): 单个/批量; 用户名=工号, 昵称=姓名, 初始密码取配置; 已生成的员工跳过 */
export function GenerateEmployeeUser(req: Api.Hrm.GenerateUserRequest) {
  return request<Api.Hrm.GenerateUserResponse>({
    url: '/api/v1/hrm/generateUser',
    method: 'post',
    data: req
  });
}

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

/** get 证书模板 paginated list (筛选: 模板名称/证书类型/来源/状态) */
export function GetCertTemplateList(params?: Api.Hrm.CertTemplateSearchParams) {
  return request<Api.Hrm.CertTemplateList>({
    url: '/api/v1/hrm/getCertTemplateList',
    method: 'get',
    params
  });
}

/** add 证书模板 (编号前缀唯一, 初始启用) */
export function AddCertTemplate(req: Api.Hrm.AddCertTemplateRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addCertTemplate',
    method: 'post',
    data: req
  });
}

/** edit 证书模板 (前缀变更时保持唯一) */
export function EditCertTemplate(req: Api.Hrm.EditCertTemplateRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editCertTemplate',
    method: 'post',
    data: req
  });
}

/** toggle 证书模板 status (1=启用 2=停用, 停用后不可发证) */
export function ToggleCertTemplateStatus(req: Api.Hrm.ToggleCertTemplateStatusRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/toggleCertTemplateStatus',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 证书模板 */
export function DeleteCertTemplate(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteCertTemplate',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get 证书发放 paginated list (筛选: 证书号/名称/类型/状态/来源/班次) */
export function GetCertificateList(params?: Api.Hrm.TrainingCertificateSearchParams) {
  return request<Api.Hrm.TrainingCertificateList>({
    url: '/api/v1/hrm/getCertificateList',
    method: 'get',
    params
  });
}

/** 批量发证: 班次结业学员按模板生成证书号=前缀-年-流水 (幂等, 跳过已有有效证书) */
export function BatchIssueCertificates(req: Api.Hrm.BatchIssueRequest) {
  return request<Api.Hrm.BatchIssueResponse>({
    url: '/api/v1/hrm/batchIssueCertificates',
    method: 'post',
    data: req
  });
}

/** 发放: 待发(0) → 已发(1), 置发证时间 */
export function IssueCertificates(uuids: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/issueCertificates',
    method: 'post',
    data: { uuids }
  });
}

/** 撤销: 已发(1) → 已撤销(2) */
export function RevokeCertificates(uuids: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/revokeCertificates',
    method: 'post',
    data: { uuids }
  });
}

/** 导入外部证书 (source='2', 证书号手动录入且唯一) */
export function ImportExternalCertificate(req: Api.Hrm.ImportExternalCertificateRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/importExternalCertificate',
    method: 'post',
    data: req
  });
}

/** 上传/替换证书扫描件 (base64) */
export function UploadCertificateFile(req: Api.Hrm.UploadCertificateFileRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/uploadCertificateFile',
    method: 'post',
    data: req
  });
}

/** 证书扫描件预览/下载 (base64, 与我的培训同一返回模式) */
export function DownloadCertificateFile(uuid: string) {
  return request<Api.Hrm.DownloadCertificateFileResponse>({
    url: '/api/v1/hrm/downloadCertificateFile',
    method: 'get',
    params: { uuid }
  });
}

/** get hrm reward bill (奖惩单 HR215) paginated list — defaults to 「我的单据」 server-side */
export function GetRewardList(params?: Api.Hrm.RewardSearchParams) {
  return request<Api.Hrm.RewardList>({
    url: '/api/v1/hrm/getRewardList',
    method: 'get',
    params
  });
}

/** add hrm reward bill */
export function AddReward(req: Api.Hrm.AddRewardRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addReward',
    method: 'post',
    data: req
  });
}

/** edit hrm reward bill (only in editable 单据状态) */
export function EditReward(req: Api.Hrm.EditRewardRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editReward',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm reward bills (only in editable 单据状态) */
export function DeleteReward(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteReward',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get recruitment requirement (招聘需求单 HR207) paginated list — defaults to 「我的单据」 server-side */
export function GetRequirementList(params?: Api.Hrm.RequirementSearchParams) {
  return request<Api.Hrm.RequirementList>({
    url: '/api/v1/hrm/getRequirementList',
    method: 'get',
    params
  });
}

/** add recruitment requirement bill */
export function AddRequirement(req: Api.Hrm.AddRequirementRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addRequirement',
    method: 'post',
    data: req
  });
}

/** edit recruitment requirement bill (only in editable 单据状态) */
export function EditRequirement(req: Api.Hrm.EditRequirementRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editRequirement',
    method: 'post',
    data: req
  });
}

/** delete/batch delete recruitment requirement bills (only in editable 单据状态) */
export function DeleteRequirement(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteRequirement',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get 候选人 paginated list (筛选 姓名/手机号/来源渠道/状态) */
export function GetCandidateList(params?: Api.Hrm.CandidateSearchParams) {
  return request<Api.Hrm.CandidateList>({
    url: '/api/v1/hrm/getCandidateList',
    method: 'get',
    params
  });
}

/** add 候选人 (初始状态 新线索) */
export function AddCandidate(req: Api.Hrm.AddCandidateRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addCandidate',
    method: 'post',
    data: req
  });
}

/** edit 候选人 (状态走 updateCandidateStatus 流转) */
export function EditCandidate(req: Api.Hrm.EditCandidateRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editCandidate',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 候选人 (普通实体, 不限状态) */
export function DeleteCandidate(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteCandidate',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** 候选人状态流转 (新线索→已投递→…→录用通过; 不可跳级/不可回退) */
export function UpdateCandidateStatus(req: Api.Hrm.UpdateCandidateStatusRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/updateCandidateStatus',
    method: 'post',
    data: req
  });
}

/** import 候选人 from a base64-encoded xlsx file (AI 入库通道按 ADR-0002 裁剪) */
export function ImportCandidate(req: Api.Hrm.ImportCandidateRequest) {
  return request<Api.Hrm.ImportCandidateResponse>({
    url: '/api/v1/hrm/importCandidate',
    method: 'post',
    data: req
  });
}

/** get hrm entry bill (员工入职单 HR201) paginated list — defaults to 「我的单据」 server-side */
export function GetEntryList(params?: Api.Hrm.EntrySearchParams) {
  return request<Api.Hrm.EntryList>({
    url: '/api/v1/hrm/getEntryList',
    method: 'get',
    params
  });
}

// ===== HRM-33 薪酬主体与薪资项目 (薪酬管理域 薪酬设置) =====

/** get 薪酬主体 paginated list */
export function GetSalarySubjectList(params?: Api.Hrm.SalarySubjectSearchParams) {
  return request<Api.Hrm.SalarySubjectList>({
    url: '/api/v1/hrm/getSalarySubjectList',
    method: 'get',
    params
  });
}

/** get hrm entry bill detail (带三子表快照) */
export function GetEntryDetail(uuid: string) {
  return request<Api.Hrm.EntryDetail>({
    url: '/api/v1/hrm/getEntryDetail',
    method: 'get',
    params: { uuid }
  });
}

/** add hrm entry bill */
export function AddEntry(req: Api.Hrm.AddEntryRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addEntry',
    method: 'post',
    data: req
  });
}

/** add 薪酬主体 (关联公司须为 org_type=1 的公司节点) */
export function AddSalarySubject(req: Api.Hrm.AddSalarySubjectRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addSalarySubject',
    method: 'post',
    data: req
  });
}

/** edit hrm entry bill (only in editable 单据状态; 已建档的不可改) */
export function EditEntry(req: Api.Hrm.EditEntryRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editEntry',
    method: 'post',
    data: req
  });
}

/** edit 薪酬主体 (公司名称快照随换随更) */
export function EditSalarySubject(req: Api.Hrm.EditSalarySubjectRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editSalarySubject',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm entry bills (only in editable 单据状态; 已建档的不可删) */
export function DeleteEntry(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteEntry',
    method: 'post',
    data: {
      uuids: req
    }
  });
}
/** delete/batch delete 薪酬主体 */
export function DeleteSalarySubject(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteSalarySubject',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get 薪资项目 paginated list (默认按计算顺序升序, 同序按 id) */
export function GetSalaryItemList(params?: Api.Hrm.SalaryItemSearchParams) {
  return request<Api.Hrm.SalaryItemList>({
    url: '/api/v1/hrm/getSalaryItemList',
    method: 'get',
    params
  });
}

/** add 薪资项目 (编码唯一; 预置项目只能由种子迁移产生) */
export function AddSalaryItem(req: Api.Hrm.AddSalaryItemRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addSalaryItem',
    method: 'post',
    data: req
  });
}

/** edit 薪资项目 */
export function EditSalaryItem(req: Api.Hrm.EditSalaryItemRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editSalaryItem',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 薪资项目 (预置项目 is_preset=1 不可删除) */
export function DeleteSalaryItem(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteSalaryItem',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get hrm overtime rule (加班规则) paginated list */
export function GetOvertimeRuleList(params?: Api.Hrm.OvertimeRuleSearchParams) {
  return request<Api.Hrm.OvertimeRuleList>({
    url: '/api/v1/hrm/getOvertimeRuleList',
    method: 'get',
    params
  });
}

/** add hrm overtime rule */
export function AddOvertimeRule(req: Api.Hrm.AddOvertimeRuleRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addOvertimeRule',
    method: 'post',
    data: req
  });
}

/** edit hrm overtime rule */
export function EditOvertimeRule(req: Api.Hrm.EditOvertimeRuleRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editOvertimeRule',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm overtime rules */
export function DeleteOvertimeRule(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteOvertimeRule',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** toggle hrm overtime rule status (启用/停用) */
export function ToggleOvertimeRuleStatus(req: { uuid: string; status: Api.Hrm.AttendanceStatus }) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/toggleOvertimeRuleStatus',
    method: 'post',
    data: req
  });
}

// ---- HRM-05 劳动合同（到期预警/待续签共享端点，spec 用户故事 14/15/17） ----

/** get 劳动合同 paginated list (筛选: 姓名/工号/到期范围; 出参含 距到期天数/预警状态) */
export function GetEmployeeContractList(params?: Api.Hrm.ContractSearchParams) {
  return request<Api.Hrm.ContractList>({
    url: '/api/v1/hrm/getEmployeeContractList',
    method: 'get',
    params
  });
}

/** add 劳动合同 (员工档案上下文维护, 选择员工带出快照) */
export function AddEmployeeContract(req: Api.Hrm.AddContractRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addEmployeeContract',
    method: 'post',
    data: req
  });
}

/** edit 劳动合同 */
export function EditEmployeeContract(req: Api.Hrm.EditContractRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editEmployeeContract',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 劳动合同 */
export function DeleteEmployeeContract(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteEmployeeContract',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** renew 劳动合同 (续签: 续签次数+1、生成新到期日; 缺省=原到期日+1年) */
export function RenewEmployeeContract(req: Api.Hrm.RenewContractRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/renewEmployeeContract',
    method: 'post',
    data: req
  });
}

// ---- HRM-05 员工证照（证照到期预警，spec 用户故事 16/17；证号出参脱敏） ----

/** get 员工证照 paginated list (筛选: 姓名/工号/证书类型; 证号脱敏、距到期天数) */
export function GetEmployeeCertificateList(params?: Api.Hrm.EmployeeCertificateSearchParams) {
  return request<Api.Hrm.EmployeeCertificateList>({
    url: '/api/v1/hrm/getEmployeeCertificateList',
    method: 'get',
    params
  });
}

/** add 员工证照 (员工档案上下文维护, 选择员工带出快照) */
export function AddEmployeeCertificate(req: Api.Hrm.AddEmployeeCertificateRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addEmployeeCertificate',
    method: 'post',
    data: req
  });
}

/** edit 员工证照 (回显脱敏证号原样提交视为未修改) */
export function EditEmployeeCertificate(req: Api.Hrm.EditEmployeeCertificateRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editEmployeeCertificate',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 员工证照 */
export function DeleteEmployeeCertificate(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteEmployeeCertificate',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get hrm schedule cycle (排班周期) paginated list */
export function GetScheduleCycleList(params?: Api.Hrm.ScheduleCycleSearchParams) {
  return request<Api.Hrm.ScheduleCycleList>({
    url: '/api/v1/hrm/getScheduleCycleList',
    method: 'get',
    params
  });
}

/** add hrm schedule cycle */
export function AddScheduleCycle(req: Api.Hrm.AddScheduleCycleRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addScheduleCycle',
    method: 'post',
    data: req
  });
}

/** edit hrm schedule cycle */
export function EditScheduleCycle(req: Api.Hrm.EditScheduleCycleRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editScheduleCycle',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm schedule cycles */
export function DeleteScheduleCycle(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteScheduleCycle',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** toggle hrm schedule cycle status (启用/停用) */
export function ToggleScheduleCycleStatus(req: { uuid: string; status: Api.Hrm.AttendanceStatus }) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/toggleScheduleCycleStatus',
    method: 'post',
    data: req
  });
}

/** get hrm resignation bill (员工离职单 HR204) paginated list — defaults to 「我的单据」 server-side */
export function GetResignationList(params?: Api.Hrm.ResignationSearchParams) {
  return request<Api.Hrm.ResignationList>({
    url: '/api/v1/hrm/getResignationList',
    method: 'get',
    params
  });
}

/** add hrm resignation bill (选择员工带出档案快照) */
export function AddResignation(req: Api.Hrm.AddResignationRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addResignation',
    method: 'post',
    data: req
  });
}

/** edit hrm resignation bill (only in editable 单据状态) */
export function EditResignation(req: Api.Hrm.EditResignationRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editResignation',
    method: 'post',
    data: req
  });
}

/** delete/batch delete hrm resignation bills (only in editable 单据状态) */
export function DeleteResignation(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteResignation',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** get hrm leave balance (假期余额) paginated list — 部门含下级由后端展开 */
export function GetLeaveBalanceList(params?: Api.Hrm.LeaveBalanceSearchParams) {
  return request<Api.Hrm.LeaveBalanceList>({
    url: '/api/v1/hrm/getLeaveBalanceList',
    method: 'get',
    params
  });
}

/** get hrm leave balance log (余额流水) paginated list */
export function GetLeaveBalanceLogList(params?: Api.Hrm.LeaveBalanceLogSearchParams) {
  return request<Api.Hrm.LeaveBalanceLogList>({
    url: '/api/v1/hrm/getLeaveBalanceLogList',
    method: 'get',
    params
  });
}

/** preview hrm leave balance init (初始化预览：按默认天数 × 员工 × 启用余额管理的类型) */
export function PreviewLeaveBalanceInit(req: Api.Hrm.LeaveBalanceInitPreviewRequest) {
  return request<{ records: Api.Hrm.LeaveBalanceInitPreviewItem[] }>({
    url: '/api/v1/hrm/previewLeaveBalanceInit',
    method: 'post',
    data: req
  });
}

/** init hrm leave balance for a year (初始化本年度额度，重复初始化的行自动跳过) */
export function InitLeaveBalance(req: Api.Hrm.LeaveBalanceInitRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/initLeaveBalance',
    method: 'post',
    data: req
  });
}

/** adjust hrm leave balance (手工增减应发，记原因；可用不得为负) */
export function AdjustLeaveBalance(req: Api.Hrm.LeaveBalanceAdjustRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/adjustLeaveBalance',
    method: 'post',
    data: req
  });
}

// ===== HRM-36 个税管理三件套 (薪酬管理域 个税管理) =====

/** get 员工个税档案 paginated list (员工唯一) */
export function GetTaxProfileList(params?: Api.Hrm.TaxProfileSearchParams) {
  return request<Api.Hrm.TaxProfileList>({
    url: '/api/v1/hrm/getTaxProfileList',
    method: 'get',
    params
  });
}

/** add 员工个税档案 (姓名/身份证以快照落库) */
export function AddTaxProfile(req: Api.Hrm.AddTaxProfileRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addTaxProfile',
    method: 'post',
    data: req
  });
}

/** edit 员工个税档案 */
export function EditTaxProfile(req: Api.Hrm.EditTaxProfileRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editTaxProfile',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 员工个税档案 */
export function DeleteTaxProfile(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteTaxProfile',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** export 员工个税档案 as a base64-encoded xlsx file */
export function ExportTaxProfile() {
  return request<Api.Hrm.ExportTaxProfileResponse>({
    url: '/api/v1/hrm/exportTaxProfile',
    method: 'get'
  });
}

/** get 专项附加扣除 paginated list (按 员工+年度+月份+扣除项 唯一) */
export function GetTaxSpecialDeductionList(params?: Api.Hrm.TaxSpecialDeductionSearchParams) {
  return request<Api.Hrm.TaxSpecialDeductionList>({
    url: '/api/v1/hrm/getTaxSpecialDeductionList',
    method: 'get',
    params
  });
}

/** 维护弹窗提交: 员工+年度的一批 (月份, 扣除项, 金额) 格子; 已存在行覆盖金额, 部分月份更新 */
export function MaintainTaxSpecialDeduction(req: Api.Hrm.MaintainTaxSpecialDeductionRequest) {
  return request<Api.Hrm.MaintainTaxSpecialDeductionResponse>({
    url: '/api/v1/hrm/maintainTaxSpecialDeduction',
    method: 'post',
    data: req
  });
}

/** 年度维护: 扣除项默认金额套用到 1-12 月 (months 为空=全年) */
export function MaintainTaxSpecialDeductionYear(req: Api.Hrm.MaintainTaxSpecialDeductionYearRequest) {
  return request<Api.Hrm.MaintainTaxSpecialDeductionResponse>({
    url: '/api/v1/hrm/maintainTaxSpecialDeductionYear',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 专项附加扣除行 */
export function DeleteTaxSpecialDeduction(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteTaxSpecialDeduction',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** import 专项附加扣除 (xlsx, 姓名+身份证 定位员工) */
export function ImportTaxSpecialDeduction(req: Api.Hrm.TaxSpecialDeductionImportRequest) {
  return request<Api.Hrm.TaxSpecialDeductionImportResponse>({
    url: '/api/v1/hrm/importTaxSpecialDeduction',
    method: 'post',
    data: req
  });
}

/** export 专项附加扣除 as a base64-encoded xlsx file */
export function ExportTaxSpecialDeduction() {
  return request<Api.Hrm.ExportTaxSpecialDeductionResponse>({
    url: '/api/v1/hrm/exportTaxSpecialDeduction',
    method: 'get'
  });
}

/** get 个税累计期初 paginated list (员工+主体+年度 唯一; HRM-37 试算消费) */
export function GetTaxOpeningList(params?: Api.Hrm.TaxOpeningSearchParams) {
  return request<Api.Hrm.TaxOpeningList>({
    url: '/api/v1/hrm/getTaxOpeningList',
    method: 'get',
    params
  });
}

/** add 个税累计期初 */
export function AddTaxOpening(req: Api.Hrm.AddTaxOpeningRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addTaxOpening',
    method: 'post',
    data: req
  });
}

/** edit 个税累计期初 */
export function EditTaxOpening(req: Api.Hrm.EditTaxOpeningRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editTaxOpening',
    method: 'post',
    data: req
  });
}

/** delete/batch delete 个税累计期初 */
export function DeleteTaxOpening(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deleteTaxOpening',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** import 个税累计期初 (xlsx, 主体按名称解析可空) */
export function ImportTaxOpening(req: Api.Hrm.TaxOpeningImportRequest) {
  return request<Api.Hrm.TaxOpeningImportResponse>({
    url: '/api/v1/hrm/importTaxOpening',
    method: 'post',
    data: req
  });
}

/** export 个税累计期初 as a base64-encoded xlsx file */
export function ExportTaxOpening() {
  return request<Api.Hrm.ExportTaxOpeningResponse>({
    url: '/api/v1/hrm/exportTaxOpening',
    method: 'get'
  });
}

/** punch (上/下班打卡 HRM-16): type=clockIn|clockOut, location 外勤打卡地点 */
export function Punch(req: Api.Hrm.PunchRequest) {
  return request<Api.Hrm.PunchResponse>({
    url: '/api/v1/hrm/punch',
    method: 'post',
    data: req
  });
}

/** get my punch records (我的打卡记录, 考勤打卡页) paginated list */
export function GetMyPunchList(params?: Api.Hrm.PunchSearchParams) {
  return request<Api.Hrm.PunchList>({
    url: '/api/v1/hrm/getMyPunchList',
    method: 'get',
    params
  });
}

/** get punch records (管理端打卡记录页) paginated list — 筛选日期区间/姓名/签到状态 */
export function GetPunchList(params?: Api.Hrm.PunchSearchParams) {
  return request<Api.Hrm.PunchList>({
    url: '/api/v1/hrm/getPunchList',
    method: 'get',
    params
  });
}

/** add punch record (管理端手动新增/补录) */
export function AddPunchRecord(req: Api.Hrm.AddPunchRecordRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/addPunchRecord',
    method: 'post',
    data: req
  });
}

/** edit punch record (管理端修改) */
export function EditPunchRecord(req: Api.Hrm.EditPunchRecordRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/editPunchRecord',
    method: 'post',
    data: req
  });
}

/** delete/batch delete punch records */
export function DeletePunchRecord(req: string[]) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/deletePunchRecord',
    method: 'post',
    data: {
      uuids: req
    }
  });
}

/** export punch records (同筛选条件导出 xlsx, base64) */
export function ExportPunchRecord(req: Api.Hrm.PunchSearchParams) {
  return request<Api.Hrm.PunchExportResponse>({
    url: '/api/v1/hrm/exportPunchRecord',
    method: 'post',
    data: req
  });
}

/** import external punch records (CSV, 整批原子) */
export function ImportPunchRecord(req: Api.Hrm.PunchImportRequest) {
  return request<Api.Hrm.PunchImportResponse>({
    url: '/api/v1/hrm/importPunchRecord',
    method: 'post',
    data: req
  });
}

/** manual day settle (手动日结: 批量员工 × 日期区间) */
export function SettlePunchDay(req: Api.Hrm.PunchSettleRequest) {
  return request<Api.Hrm.PunchSettleResponse>({
    url: '/api/v1/hrm/settlePunchDay',
    method: 'post',
    data: req
  });
}

/** get schedule board matrix (排班表矩阵: 员工×当月每日班次, 按考勤组维护) */
export function GetScheduleMatrix(params: Api.Hrm.ScheduleMatrixRequest) {
  return request<Api.Hrm.ScheduleMatrixResponse>({
    url: '/api/v1/hrm/getScheduleMatrix',
    method: 'get',
    params
  });
}

/** save one schedule board cell (矩阵单元格保存; shiftUuid 空=休/清空) */
export function SaveScheduleCell(req: Api.Hrm.ScheduleCellRequest) {
  return request<Api.Common.Empty>({
    url: '/api/v1/hrm/saveScheduleCell',
    method: 'post',
    data: req
  });
}

/** batch save schedule board cells (整表保存, 事务) */
export function BatchSaveSchedule(req: Api.Hrm.ScheduleBatchSaveRequest) {
  return request<Api.Hrm.ScheduleBatchSaveResponse>({
    url: '/api/v1/hrm/batchSaveSchedule',
    method: 'post',
    data: req
  });
}

/** generate schedule from cycle (周期生成: ExpandCycle 相位展开落当月) */
export function GenerateScheduleFromCycle(req: Api.Hrm.ScheduleGenerateRequest) {
  return request<Api.Hrm.ScheduleGenerateResponse>({
    url: '/api/v1/hrm/generateScheduleFromCycle',
    method: 'post',
    data: req
  });
}

/** copy week schedule (复制周: 源周按星期对齐复制到目标周) */
export function CopyScheduleWeek(req: Api.Hrm.ScheduleCopyWeekRequest) {
  return request<Api.Hrm.ScheduleCopyWeekResponse>({
    url: '/api/v1/hrm/copyScheduleWeek',
    method: 'post',
    data: req
  });
}

/** import schedule board matrix template (xlsx: 行=工号, 列=1..31 班次编码) */
export function ImportScheduleBoard(req: Api.Hrm.ScheduleImportRequest) {
  return request<Api.Hrm.ScheduleImportResponse>({
    url: '/api/v1/hrm/importScheduleBoard',
    method: 'post',
    data: req
  });
}

/** export schedule board matrix (xlsx base64) */
export function ExportScheduleBoard(req: Api.Hrm.ScheduleExportRequest) {
  return request<Api.Hrm.ScheduleExportResponse>({
    url: '/api/v1/hrm/exportScheduleBoard',
    method: 'post',
    data: req
  });
}

/** publish/unpublish schedule month (发布本月/撤销发布) */
export function PublishScheduleMonth(req: Api.Hrm.SchedulePublishRequest) {
  return request<Api.Hrm.SchedulePublishResponse>({
    url: '/api/v1/hrm/publishScheduleMonth',
    method: 'post',
    data: req
  });
}
