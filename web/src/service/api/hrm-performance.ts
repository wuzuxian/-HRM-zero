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
