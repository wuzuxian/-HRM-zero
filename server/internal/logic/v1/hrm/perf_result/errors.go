package perf_result

import (
	"errors"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 批次结果审批域常量（HRM-45，GLOSSARY「绩效结果」）。

// BillTypeCode is the 单据类型码 for 批次结果审批（HR211，demo-inventory 9.3）。
const BillTypeCode = "211"

// 结果状态（hrm_perf_result.result_status）。
const (
	ResultStatusNormal  = "1" // 正常
	ResultStatusVoided  = "2" // 已作废
)

// sentinel errors（wrapBizErr 映射到注册错误码 10690-10694）。
var (
	// ErrResultApprovalNotFound: 批次结果审批单不存在。
	ErrResultApprovalNotFound = errors.New("批次结果审批单不存在")
	// ErrResultApprovalInvalid: 结果审批字段非法 / 批次状态不符 / 重复提交被拒。
	ErrResultApprovalInvalid = errors.New("批次结果审批字段非法")
	// ErrResultNotFound: 绩效结果行不存在。
	ErrResultNotFound = errors.New("绩效结果不存在")
	// ErrResultInvalid: 绩效结果字段非法。
	ErrResultInvalid = errors.New("绩效结果字段非法")
	// ErrResultSignDenied: 签收越权（仅员工本人）。
	ErrResultSignDenied = errors.New("仅员工本人可签收本人绩效结果")
)

// wrapBizErr maps sentinel errors to registered business error codes.
func wrapBizErr(err error) error {
	switch {
	case errors.Is(err, ErrResultApprovalNotFound):
		return status.Wrap(errcodes.ResultApprovalNotFoundCode, err)
	case errors.Is(err, ErrResultApprovalInvalid):
		return status.Wrap(errcodes.ResultApprovalInvalidCode, err)
	case errors.Is(err, ErrResultNotFound):
		return status.Wrap(errcodes.ResultNotFoundCode, err)
	case errors.Is(err, ErrResultInvalid):
		return status.Wrap(errcodes.ResultInvalidCode, err)
	case errors.Is(err, ErrResultSignDenied):
		return status.Wrap(errcodes.ResultSignDeniedCode, err)
	default:
		return err
	}
}
