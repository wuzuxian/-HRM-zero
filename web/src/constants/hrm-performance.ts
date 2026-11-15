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
