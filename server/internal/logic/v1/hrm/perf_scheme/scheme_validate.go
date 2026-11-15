package perf_scheme

import "errors"

// 绩效方案枚举校验（HRM-41，演示站 9.1 绩效方案）。
// 枚举取值与 GLOSSARY.md「绩效方案」条目及演示站清单对齐：
//   - 周期类型 period_type：month=月度 quarter=季度 half_year=半年 year=年度
//   - 评估模式 assessment_mode：kpi=KPI kpi_behavior=KPI+行为混合
//   - 评分模式 score_mode：percent=百分制 grade=等级制 amount=金额制
//   - 联薪方式 pay_link_mode：coefficient=系数制 amount=金额制（HRM-49 联薪契约字段）
var (
	// ErrSchemeNameRequired: 方案名称不能为空。
	ErrSchemeNameRequired = errors.New("方案名称不能为空")
	// ErrSchemeInvalid: 枚举字段非法（周期类型/评估模式/评分模式/联薪方式/状态越界）。
	ErrSchemeInvalid = errors.New("绩效方案枚举字段非法")
	// ErrSchemeTemplateRequired: 必须关联一个考核模板。
	ErrSchemeTemplateRequired = errors.New("绩效方案必须关联考核模板")
	// ErrSchemeNotFound: 绩效方案不存在。
	ErrSchemeNotFound = errors.New("绩效方案不存在")
	// ErrSchemeTemplateNotFound: 关联的考核模板不存在。
	ErrSchemeTemplateNotFound = errors.New("关联的考核模板不存在")
)

// 合法枚举集合。前端与 HRM-49 依赖这些字符串取值，勿改动已有取值。
var (
	// PeriodTypes 周期类型
	PeriodTypes = map[string]struct{}{
		"month":     {}, // 月度
		"quarter":   {}, // 季度
		"half_year": {}, // 半年
		"year":      {}, // 年度
	}
	// AssessmentModes 评估模式
	AssessmentModes = map[string]struct{}{
		"kpi":          {}, // KPI
		"kpi_behavior": {}, // KPI+行为混合
	}
	// ScoreModes 评分模式
	ScoreModes = map[string]struct{}{
		"percent": {}, // 百分制
		"grade":   {}, // 等级制
		"amount":  {}, // 金额制
	}
	// PayLinkModes 联薪方式
	PayLinkModes = map[string]struct{}{
		"coefficient": {}, // 系数制
		"amount":      {}, // 金额制
	}
	// Statuses 启停状态（与 manage 域 EnableStatus 一致）
	Statuses = map[string]struct{}{
		"1": {}, // 启用
		"2": {}, // 停用
	}
)

// enumOk reports whether value is in the allowed set.
func enumOk(value string, allowed map[string]struct{}) bool {
	_, ok := allowed[value]
	return ok
}

// ValidateScheme validates one 绩效方案 request body (name + enums + template reference presence).
// Pure function — template existence check lives in scheme.go (DB lookup).
func ValidateScheme(name, templateUuid, periodType, assessmentMode, scoreMode, payLinkMode, status string) error {
	if name == "" {
		return ErrSchemeNameRequired
	}
	if templateUuid == "" {
		return ErrSchemeTemplateRequired
	}
	if !enumOk(periodType, PeriodTypes) ||
		!enumOk(assessmentMode, AssessmentModes) ||
		!enumOk(scoreMode, ScoreModes) ||
		!enumOk(payLinkMode, PayLinkModes) {
		return ErrSchemeInvalid
	}
	if status != "" && !enumOk(status, Statuses) {
		return ErrSchemeInvalid
	}
	return nil
}
