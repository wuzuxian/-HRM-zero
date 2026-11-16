import { request } from '../request';

// ---- 考核任务链（HRM-43，演示站 9.2 五个填报列表页共用） ----

/** get perf assessment list（当前节点/任务状态/当前得分/当前等级列） */
export function GetPerfAssessmentList(params?: Api.Hrm.Performance.AssessmentSearchParams) {
  return request<Api.Hrm.Performance.AssessmentList>({
    url: '/api/v1/hrm/getPerfAssessmentList',
    method: 'get',
    params
  });
}

/** get perf assessment detail（模板快照 items + 各阶段填报记录） */
export function GetPerfAssessmentDetail(uuid: string) {
  return request<Api.Hrm.Performance.AssessmentDetail>({
    url: '/api/v1/hrm/getPerfAssessmentDetail',
    method: 'get',
    params: { uuid }
  });
}

/** generate perf assessments（批次发布后按参与人幂等生成） */
export function GeneratePerfAssessments(periodUuid: string) {
  return request<{ generatedCount: number }>({
    url: '/api/v1/hrm/generatePerfAssessment',
    method: 'post',
    data: { periodUuid }
  });
}

/** 阶段1：目标确认（被考核人） */
export function ConfirmPerfAssessmentPlan(uuid: string, comment?: string) {
  return request<Api.Hrm.Performance.ConfirmPlanResponse>({
    url: '/api/v1/hrm/confirmPerfAssessmentPlan',
    method: 'post',
    data: { uuid, comment }
  });
}

/** 阶段2：自评填报（被考核人，按模板快照逐项评分） */
export function SubmitPerfAssessmentSelfReview(uuid: string, items: Api.Hrm.Performance.ScoreItem[], comment?: string) {
  return request<Api.Hrm.Performance.SubmitSelfReviewResponse>({
    url: '/api/v1/hrm/submitPerfAssessmentSelfReview',
    method: 'post',
    data: { uuid, items, comment }
  });
}

/** 阶段3：上级评审（有 hrm:perf_assessment:review 码者） */
export function SubmitPerfAssessmentLeaderReview(
  uuid: string,
  items: Api.Hrm.Performance.ScoreItem[],
  comment?: string
) {
  return request<Api.Hrm.Performance.SubmitLeaderReviewResponse>({
    url: '/api/v1/hrm/submitPerfAssessmentLeaderReview',
    method: 'post',
    data: { uuid, items, comment }
  });
}

/** 阶段4：校准填报（校准人调整最终得分） */
export function SubmitPerfAssessmentCalibration(
  uuid: string,
  finalScore: number,
  finalGrade?: string,
  comment?: string
) {
  return request<Api.Hrm.Performance.SubmitCalibrationResponse>({
    url: '/api/v1/hrm/submitPerfAssessmentCalibration',
    method: 'post',
    data: { uuid, finalScore, finalGrade, comment }
  });
}

/** 阶段5：结果确认（被考核人确认） */
export function ConfirmPerfAssessmentResult(uuid: string) {
  return request<Api.Hrm.Performance.ConfirmResultResponse>({
    url: '/api/v1/hrm/confirmPerfAssessmentResult',
    method: 'post',
    data: { uuid }
  });
}

/** 发起申诉（被考核人，仅结果确认阶段） */
export function AppealPerfAssessment(assessmentUuid: string, reason: string) {
  return request<Api.Hrm.Performance.AppealResponse>({
    url: '/api/v1/hrm/appealPerfAssessment',
    method: 'post',
    data: { assessmentUuid, reason }
  });
}

/** 申诉裁决（校准人：维持/调整得分→关闭） */
export function HandlePerfAssessmentAppeal(req: Api.Hrm.Performance.HandleAppealRequest) {
  return request<Api.Hrm.Performance.HandleAppealResponse>({
    url: '/api/v1/hrm/handlePerfAssessmentAppeal',
    method: 'post',
    data: req
  });
}

// ---- 考核任务工作台（HRM-44，单聚合端点，只读） ----

/** 考核任务工作台：全周期待办汇总卡 + 最近待办 + 申诉单列 */
export function GetMyAssessmentWorkbench() {
  return request<Api.Hrm.Performance.AssessmentWorkbench>({
    url: '/api/v1/hrm/getMyAssessmentWorkbench',
    method: 'get'
  });
}
