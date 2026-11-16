import { request } from '../request';

// ---- 绩效指标库（HRM-40，演示站 9.1 指标库） ----

/** get performance indicator list */
export function GetPerfIndicatorList(params?: Api.Hrm.Performance.IndicatorSearchParams) {
  return request<Api.Hrm.Performance.IndicatorList>({
    url: '/api/v1/hrm/getPerfIndicatorList',
    method: 'get',
    params
  });
}

/**
 * get all enabled performance indicators
 *
 * 考核模板配置指标时的指标选择器
 */
export function GetAllPerfIndicators() {
  return request<{ records: Api.Hrm.Performance.Indicator[] }>({
    url: '/api/v1/hrm/getAllPerfIndicator',
    method: 'get'
  });
}

/** add performance indicator */
export function AddPerfIndicator(req: Api.Hrm.Performance.AddIndicatorRequest) {
  return request<Api.Hrm.Performance.AddIndicatorResponse>({
    url: '/api/v1/hrm/addPerfIndicator',
    method: 'post',
    data: req
  });
}

/** edit performance indicator */
export function EditPerfIndicator(req: Api.Hrm.Performance.EditIndicatorRequest) {
  return request<Api.Hrm.Performance.EditIndicatorResponse>({
    url: '/api/v1/hrm/editPerfIndicator',
    method: 'post',
    data: req
  });
}

/** delete performance indicators */
export function DeletePerfIndicators(uuids: string[]) {
  return request<Api.Hrm.Performance.DeleteIndicatorResponse>({
    url: '/api/v1/hrm/deletePerfIndicator',
    method: 'post',
    data: { uuids }
  });
}

// ---- 考核模板（HRM-40，演示站 9.1 模板管理） ----

/** get performance template list */
export function GetPerfTemplateList(params?: Api.Hrm.Performance.TemplateSearchParams) {
  return request<Api.Hrm.Performance.TemplateList>({
    url: '/api/v1/hrm/getPerfTemplateList',
    method: 'get',
    params
  });
}

/** get performance template detail with 指标权重配置 */
export function GetPerfTemplateDetail(uuid: string) {
  return request<Api.Hrm.Performance.TemplateDetail>({
    url: '/api/v1/hrm/getPerfTemplateDetail',
    method: 'get',
    params: { uuid }
  });
}

/** add performance template（模板内指标权重之和须为 100%） */
export function AddPerfTemplate(req: Api.Hrm.Performance.AddTemplateRequest) {
  return request<Api.Hrm.Performance.AddTemplateResponse>({
    url: '/api/v1/hrm/addPerfTemplate',
    method: 'post',
    data: req
  });
}

/** edit performance template（整体替换指标权重配置，权重之和须为 100%） */
export function EditPerfTemplate(req: Api.Hrm.Performance.EditTemplateRequest) {
  return request<Api.Hrm.Performance.EditTemplateResponse>({
    url: '/api/v1/hrm/editPerfTemplate',
    method: 'post',
    data: req
  });
}

/** delete performance templates */
export function DeletePerfTemplates(uuids: string[]) {
  return request<Api.Hrm.Performance.DeleteTemplateResponse>({
    url: '/api/v1/hrm/deletePerfTemplate',
    method: 'post',
    data: { uuids }
  });
}

// ---- 绩效方案（HRM-41，演示站 9.1 绩效方案） ----

/** get performance scheme list */
export function GetPerfSchemeList(params?: Api.Hrm.Performance.SchemeSearchParams) {
  return request<Api.Hrm.Performance.SchemeList>({
    url: '/api/v1/hrm/getPerfSchemeList',
    method: 'get',
    params
  });
}

/** add performance scheme（关联模板须已存在；周期/评估/评分/联薪为枚举字段） */
export function AddPerfScheme(req: Api.Hrm.Performance.AddSchemeRequest) {
  return request<Api.Hrm.Performance.AddSchemeResponse>({
    url: '/api/v1/hrm/addPerfScheme',
    method: 'post',
    data: req
  });
}

/** edit performance scheme */
export function EditPerfScheme(req: Api.Hrm.Performance.EditSchemeRequest) {
  return request<Api.Hrm.Performance.EditSchemeResponse>({
    url: '/api/v1/hrm/editPerfScheme',
    method: 'post',
    data: req
  });
}

/** delete performance schemes */
export function DeletePerfSchemes(uuids: string[]) {
  return request<Api.Hrm.Performance.DeleteSchemeResponse>({
    url: '/api/v1/hrm/deletePerfScheme',
    method: 'post',
    data: { uuids }
  });
}

// ---- 绩效批次（HRM-42，演示站 9.2 考核批次工作台） ----

/** get perf period list（工作台列表：状态列 + 参与人数/绩效方案名称快照） */
export function GetPerfPeriodList(params?: Api.Hrm.Performance.PerfPeriodSearchParams) {
  return request<Api.Hrm.Performance.PerfPeriodList>({
    url: '/api/v1/hrm/getPerfPeriodList',
    method: 'get',
    params
  });
}

/** add perf period（PERF 编号自动生成；按适用范围圈定初始参与人） */
export function AddPerfPeriod(req: Api.Hrm.Performance.AddPerfPeriodRequest) {
  return request<Api.Hrm.Performance.AddPerfPeriodResponse>({
    url: '/api/v1/hrm/addPerfPeriod',
    method: 'post',
    data: req
  });
}

/** edit perf period（仅草稿期） */
export function EditPerfPeriod(req: Api.Hrm.Performance.EditPerfPeriodRequest) {
  return request<Api.Hrm.Performance.EditPerfPeriodResponse>({
    url: '/api/v1/hrm/editPerfPeriod',
    method: 'post',
    data: req
  });
}

/** delete perf periods（仅草稿期） */
export function DeletePerfPeriods(uuids: string[]) {
  return request<Api.Hrm.Performance.DeletePerfPeriodResponse>({
    url: '/api/v1/hrm/deletePerfPeriod',
    method: 'post',
    data: { uuids }
  });
}

/** publish perf period（草稿→进行中；发布前须已圈定参与人） */
export function PublishPerfPeriod(uuid: string) {
  return request<Api.Hrm.Performance.PublishPerfPeriodResponse>({
    url: '/api/v1/hrm/publishPerfPeriod',
    method: 'post',
    data: { uuid }
  });
}

/** advance perf period lifecycle（进行中→待确认→待校准→结果已发布→已发布，forward-only） */
export function AdvancePerfPeriod(uuid: string) {
  return request<Api.Hrm.Performance.AdvancePerfPeriodResponse>({
    url: '/api/v1/hrm/advancePerfPeriod',
    method: 'post',
    data: { uuid }
  });
}

// ---- 批次结果审批与绩效结果（HRM-45，演示站 9.3 结果） ----

/** get perf result approval list（HR211 批次结果审批单列表） */
export function GetPerfResultApprovalList(params?: Api.Hrm.Performance.PerfResultApprovalSearchParams) {
  return request<Api.Hrm.Performance.PerfResultApprovalList>({
    url: '/api/v1/hrm/getPerfResultApprovalList',
    method: 'get',
    params
  });
}

/** submit perf result approval（HR211 提交；批次须 结果已发布；一批次一张单） */
export function SubmitPerfResultApproval(req: Api.Hrm.Performance.SubmitPerfResultApprovalRequest) {
  return request<Api.Hrm.Performance.SubmitPerfResultApprovalResponse>({
    url: '/api/v1/hrm/submitPerfResultApproval',
    method: 'post',
    data: req
  });
}

/** get perf result list（绩效结果；mine=仅本人） */
export function GetPerfResultList(params?: Api.Hrm.Performance.PerfResultSearchParams) {
  return request<Api.Hrm.Performance.PerfResultList>({
    url: '/api/v1/hrm/getPerfResultList',
    method: 'get',
    params
  });
}

/** sign perf result（员工本人签收；幂等） */
export function SignPerfResult(uuid: string) {
  return request<Api.Hrm.Performance.SignPerfResultResponse>({
    url: '/api/v1/hrm/signPerfResult',
    method: 'post',
    data: { uuid }
  });
}

// ---- 绩效驾驶舱（HRM-46，演示站 9.3 /hrm/performance/dashboard） ----

/** get perf dashboard（单聚合端点：指标卡 + 风险预警明细） */
export function GetPerfDashboard(params?: Api.Hrm.Performance.PerfDashboardSearchParams) {
  return request<Api.Hrm.Performance.PerfDashboardResponse>({
    url: '/api/v1/hrm/getPerfDashboard',
    method: 'get',
    params
  });
}
