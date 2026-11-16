package perf_period

import (
	"errors"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 绩效批次 sentinel errors（HRM-42，GLOSSARY「绩效批次」）。
// 映射到注册业务错误码见 errors.go wrapBizErr。
var (
	// ErrPeriodNameRequired: 批次名称不能为空。
	ErrPeriodNameRequired = errors.New("批次名称不能为空")
	// ErrPeriodSchemeRequired: 必须关联一个绩效方案。
	ErrPeriodSchemeRequired = errors.New("绩效批次必须关联绩效方案")
	// ErrPeriodInvalid: 字段非法（批次类型/周期日期/适用范围 JSON/人员状态越界）。
	ErrPeriodInvalid = errors.New("绩效批次字段非法")
	// ErrPeriodScopeInvalid: 适用范围 scope 不是合法 JSON 或结构越界。
	ErrPeriodScopeInvalid = errors.New("适用范围条件非法")
	// ErrPeriodNotFound: 绩效批次不存在。
	ErrPeriodNotFound = errors.New("绩效批次不存在")
	// ErrPeriodSchemeNotFound: 关联的绩效方案不存在。
	ErrPeriodSchemeNotFound = errors.New("关联的绩效方案不存在")
	// ErrPeriodNotEditable: 批次已发布，仅草稿期可编辑/删除/重新圈人。
	ErrPeriodNotEditable = errors.New("批次已发布，仅草稿期可编辑")
	// ErrPeriodNoParticipants: 发布守卫：发布前须已圈定参与人。
	ErrPeriodNoParticipants = errors.New("发布前须已圈定参与人")
	// ErrPeriodTransitionInvalid: 生命周期状态机非法流转（forward-only）。
	ErrPeriodTransitionInvalid = errors.New("批次状态不允许该操作")
	// ErrPeriodNoSeqOutOfRange: PERF 编号 4 位流水超出 1..9999。
	ErrPeriodNoSeqOutOfRange = errors.New("批次编号流水超出范围")
)

// wrapBizErr maps 绩效批次 sentinel errors to registered business error codes;
// unknown errors pass through unchanged.
func wrapBizErr(err error) error {
	switch {
	case errors.Is(err, ErrPeriodNotFound):
		return status.Wrap(errcodes.HrmPerfPeriodNotFoundCode, err)
	case errors.Is(err, ErrPeriodNameRequired):
		return status.Wrap(errcodes.HrmPerfPeriodNameRequiredCode, err)
	case errors.Is(err, ErrPeriodSchemeNotFound):
		return status.Wrap(errcodes.HrmPerfPeriodSchemeNotFoundCode, err)
	case errors.Is(err, ErrPeriodNoParticipants):
		return status.Wrap(errcodes.HrmPerfPeriodNoParticipantsCode, err)
	case errors.Is(err, ErrPeriodTransitionInvalid):
		return status.Wrap(errcodes.HrmPerfPeriodTransitionInvalidCode, err)
	case errors.Is(err, ErrPeriodSchemeRequired),
		errors.Is(err, ErrPeriodInvalid),
		errors.Is(err, ErrPeriodScopeInvalid),
		errors.Is(err, ErrPeriodNotEditable),
		errors.Is(err, ErrPeriodNoSeqOutOfRange):
		return status.Wrap(errcodes.HrmPerfPeriodInvalidCode, err)
	default:
		return err
	}
}
