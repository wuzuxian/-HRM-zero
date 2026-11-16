import { transformRecordToOption } from '@/utils/common';

// 绩效管理（HRM-40）领域枚举：指标类型 / 评分规则 / 数据来源 / 启用状态。
// 状态与 manage 不同（后端 '1'=启用 '0'=停用），故独立定义。

export const perfStatusRecord: Record<Api.Hrm.Performance.CommonStatus, App.I18n.I18nKey> = {
  '1': 'page.hrm.performance.common.status.enable',
  '2': 'page.hrm.performance.common.status.disable'
};

export const perfStatusOptions = transformRecordToOption(perfStatusRecord);

export const perfIndicatorTypeRecord: Record<Api.Hrm.Performance.IndicatorType, App.I18n.I18nKey> = {
  kpi: 'page.hrm.performance.indicatorType.kpi',
  behavior: 'page.hrm.performance.indicatorType.behavior'
};

export const perfIndicatorTypeOptions = transformRecordToOption(perfIndicatorTypeRecord);

export const perfScoreRuleRecord: Record<Api.Hrm.Performance.ScoreRule, App.I18n.I18nKey> = {
  formula: 'page.hrm.performance.scoreRule.formula',
  manual: 'page.hrm.performance.scoreRule.manual'
};

export const perfScoreRuleOptions = transformRecordToOption(perfScoreRuleRecord);

export const perfDataSourceRecord: Record<Api.Hrm.Performance.DataSource, App.I18n.I18nKey> = {
  system: 'page.hrm.performance.dataSource.system',
  manual: 'page.hrm.performance.dataSource.manual'
};

export const perfDataSourceOptions = transformRecordToOption(perfDataSourceRecord);

export function perfIndicatorTypeTag(type: Api.Hrm.Performance.IndicatorType): NaiveUI.ThemeColor {
  return type === 'kpi' ? 'primary' : 'info';
}

export function perfStatusTag(status: Api.Hrm.Performance.CommonStatus | null): NaiveUI.ThemeColor | null {
  if (status === null) return null;
  return status === '1' ? 'success' : 'warning';
}

// ---- 绩效方案（HRM-41）领域枚举：周期类型 / 评估模式 / 评分模式 / 联薪方式 ----

export const perfSchemePeriodTypeRecord: Record<Api.Hrm.Performance.SchemePeriodType, App.I18n.I18nKey> = {
  month: 'page.hrm.performance.schemePeriodType.month',
  quarter: 'page.hrm.performance.schemePeriodType.quarter',
  half_year: 'page.hrm.performance.schemePeriodType.half_year',
  year: 'page.hrm.performance.schemePeriodType.year'
};

export const perfSchemePeriodTypeOptions = transformRecordToOption(perfSchemePeriodTypeRecord);

export const perfSchemeAssessmentModeRecord: Record<Api.Hrm.Performance.SchemeAssessmentMode, App.I18n.I18nKey> = {
  kpi: 'page.hrm.performance.schemeAssessmentMode.kpi',
  kpi_behavior: 'page.hrm.performance.schemeAssessmentMode.kpi_behavior'
};

export const perfSchemeAssessmentModeOptions = transformRecordToOption(perfSchemeAssessmentModeRecord);

export const perfSchemeScoreModeRecord: Record<Api.Hrm.Performance.SchemeScoreMode, App.I18n.I18nKey> = {
  percent: 'page.hrm.performance.schemeScoreMode.percent',
  grade: 'page.hrm.performance.schemeScoreMode.grade',
  amount: 'page.hrm.performance.schemeScoreMode.amount'
};

export const perfSchemeScoreModeOptions = transformRecordToOption(perfSchemeScoreModeRecord);

export const perfSchemePayLinkModeRecord: Record<Api.Hrm.Performance.SchemePayLinkMode, App.I18n.I18nKey> = {
  coefficient: 'page.hrm.performance.schemePayLinkMode.coefficient',
  amount: 'page.hrm.performance.schemePayLinkMode.amount'
};

export const perfSchemePayLinkModeOptions = transformRecordToOption(perfSchemePayLinkModeRecord);

// ---- 绩效批次（HRM-42）领域枚举：批次类型 / 生命周期 ----
// 生命周期 forward-only（GLOSSARY「绩效批次」）：草稿→进行中→待确认→待校准→结果已发布→已发布。

export const perfPeriodTypeRecord: Record<Api.Hrm.Performance.PeriodType, App.I18n.I18nKey> = {
  month: 'page.hrm.performance.periodType.month',
  quarter: 'page.hrm.performance.periodType.quarter'
};

export const perfPeriodTypeOptions = transformRecordToOption(perfPeriodTypeRecord);

export const perfPeriodStatusRecord: Record<Api.Hrm.Performance.PeriodStatus, App.I18n.I18nKey> = {
  '0': 'page.hrm.performance.periodStatus.draft',
  '1': 'page.hrm.performance.periodStatus.ongoing',
  '2': 'page.hrm.performance.periodStatus.confirming',
  '3': 'page.hrm.performance.periodStatus.calibrating',
  '4': 'page.hrm.performance.periodStatus.resultPublished',
  '5': 'page.hrm.performance.periodStatus.published'
};

export const perfPeriodStatusOptions = transformRecordToOption(perfPeriodStatusRecord);

export function perfPeriodStatusTag(status: Api.Hrm.Performance.PeriodStatus | null): NaiveUI.ThemeColor | null {
  if (status === null) return null;
  switch (status) {
    case '0':
      return 'default';
    case '1':
      return 'info';
    case '2':
      return 'warning';
    case '3':
      return 'warning';
    case '4':
      return 'primary';
    case '5':
      return 'success';
    default:
      return null;
  }
}

// ---- 考核任务链（HRM-43）领域枚举：当前节点 / 任务状态 / 申诉状态 ----

export const perfAssessmentNodeRecord: Record<Api.Hrm.Performance.AssessmentNode, App.I18n.I18nKey> = {
  '1': 'page.hrm.performance.assessmentNode.planConfirm',
  '2': 'page.hrm.performance.assessmentNode.selfReview',
  '3': 'page.hrm.performance.assessmentNode.leaderReview',
  '4': 'page.hrm.performance.assessmentNode.calibration',
  '5': 'page.hrm.performance.assessmentNode.resultConfirm'
};

export const perfAssessmentNodeOptions = transformRecordToOption(perfAssessmentNodeRecord);

export function perfAssessmentNodeTag(node: Api.Hrm.Performance.AssessmentNode): NaiveUI.ThemeColor {
  switch (node) {
    case '1':
      return 'info';
    case '2':
      return 'primary';
    case '3':
      return 'warning';
    case '4':
      return 'warning';
    case '5':
      return 'success';
    default:
      return 'default';
  }
}

export const perfAssessmentTaskStatusRecord: Record<Api.Hrm.Performance.AssessmentTaskStatus, App.I18n.I18nKey> = {
  '0': 'page.hrm.performance.assessmentTaskStatus.ongoing',
  '1': 'page.hrm.performance.assessmentTaskStatus.done',
  '2': 'page.hrm.performance.assessmentTaskStatus.appealed'
};

export const perfAssessmentTaskStatusOptions = transformRecordToOption(perfAssessmentTaskStatusRecord);

export function perfAssessmentTaskStatusTag(status: Api.Hrm.Performance.AssessmentTaskStatus): NaiveUI.ThemeColor {
  switch (status) {
    case '0':
      return 'info';
    case '1':
      return 'success';
    case '2':
      return 'error';
    default:
      return 'default';
  }
}

export const perfAssessmentAppealStatusRecord: Record<Api.Hrm.Performance.AssessmentAppealStatus, App.I18n.I18nKey> = {
  '0': 'page.hrm.performance.assessmentAppealStatus.pending',
  '1': 'page.hrm.performance.assessmentAppealStatus.upheld',
  '2': 'page.hrm.performance.assessmentAppealStatus.adjusted'
};

export function perfAssessmentAppealStatusTag(status: Api.Hrm.Performance.AssessmentAppealStatus): NaiveUI.ThemeColor {
  switch (status) {
    case '0':
      return 'warning';
    case '1':
      return 'info';
    case '2':
      return 'success';
    default:
      return 'default';
  }
}
