package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-42 绩效批次业务错误码（1063x 段；10630-10631 HRM-17 看板、10632 HRM-22 考勤统计已占用，
// 本票取 10633-10638）。
const (
	// HrmPerfPeriodNotFoundCode 绩效批次不存在
	HrmPerfPeriodNotFoundCode = 10633
	// HrmPerfPeriodInvalidCode 绩效批次字段非法（名称为空、批次类型/周期日期/适用范围 JSON 越界）
	HrmPerfPeriodInvalidCode = 10634
	// HrmPerfPeriodSchemeNotFoundCode 关联绩效方案不存在（scheme_uuid 须指向 hrm_perf_scheme 已有行）
	HrmPerfPeriodSchemeNotFoundCode = 10635
	// HrmPerfPeriodNameRequiredCode 批次名称不能为空
	HrmPerfPeriodNameRequiredCode = 10636
	// HrmPerfPeriodNoParticipantsCode 发布守卫：发布前须已圈定参与人（参与人数 > 0）
	HrmPerfPeriodNoParticipantsCode = 10637
	// HrmPerfPeriodTransitionInvalidCode 生命周期状态机：非法流转（forward-only，终态不可推进）
	HrmPerfPeriodTransitionInvalidCode = 10638
)

func init() {
	status.Register(HrmPerfPeriodNotFoundCode)
	status.Register(HrmPerfPeriodInvalidCode)
	status.Register(HrmPerfPeriodSchemeNotFoundCode)
	status.Register(HrmPerfPeriodNameRequiredCode)
	status.Register(HrmPerfPeriodNoParticipantsCode)
	status.Register(HrmPerfPeriodTransitionInvalidCode)
}
