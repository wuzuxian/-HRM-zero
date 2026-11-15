package overtimerule

import (
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_rule"
)

// 加班当天类型 (day_type, 与 GLOSSARY「当天类型」一致)
const (
	DayTypeWorkday = "1" // 工作日
	DayTypeRestday = "2" // 休息日
	DayTypeHoliday = "3" // 法定节假日
)

// 解析来源 (ResolveResponse.source)
const (
	ruleSourceGroup  = "group"  // 考勤组绑定规则
	ruleSourceOrg    = "org"    // 组织默认加班规则
	ruleSourceGlobal = "global" // 全局兜底规则 (未绑考勤组且未挂组织)
)

// 导出别名（HRM-21 加班申请单等消费方使用）
const (
	RuleSourceGroup  = ruleSourceGroup
	RuleSourceOrg    = ruleSourceOrg
	RuleSourceGlobal = ruleSourceGlobal
)

// status values
const (
	statusEnabled  = "1"
	statusDisabled = "0"
)

// 策略开关值
const (
	policyRequired   = 1 // 需审批 / 需打卡
	policyNotRequired = 0
)

// minUnitDefaultMinutes 最小计算单位默认值（30 分钟）
const minUnitDefaultMinutes = 30

// 合法最小计算单位（分钟）：半小时/一小时
var validMinUnits = map[int64]bool{
	30: true,
	60: true,
}

// conditionNewChainEnabled 查询启用规则的公共条件链.
func conditionNewChainEnabled() []condition.Condition {
	return condition.NewChain().
		Equal(hrm_overtime_rule.Status, statusEnabled).
		OrderByAsc(hrm_overtime_rule.Id).
		Build()
}
