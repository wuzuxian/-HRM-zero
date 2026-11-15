package attendancegroup

// 考勤组班制 (group_type)
const (
	groupTypeFixed    = "1" // 固定班制
	groupTypeScheduling = "2" // 排班制
	groupTypeFlexible = "3" // 自由工时
)

// 生效规则来源 (GetEffectiveRuleResponse.source)
const (
	ruleSourceGroup = "group" // 考勤组
	ruleSourceOrg   = "org"   // 组织默认规则 (本部门或最近上级组织)
)

// status values
const (
	statusEnabled  = "1"
	statusDisabled = "0"
)
