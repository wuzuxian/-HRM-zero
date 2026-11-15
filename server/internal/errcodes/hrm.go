package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM 单据审批底座（HRM-06）业务错误码。
const (
	// HrmBillTypeNotRegisteredCode 单据类型未注册审批处理器
	HrmBillTypeNotRegisteredCode = 10330
	// HrmBillNotFoundCode 单据不存在
	HrmBillNotFoundCode = 10331
	// HrmBillInvalidTransitionCode 当前单据状态不允许该操作（单据状态机拒绝）
	HrmBillInvalidTransitionCode = 10332
	// HrmBillInvalidTypeCode 单据类型码非法（须为 3 位数字）
	HrmBillInvalidTypeCode = 10333
	// HrmBillNoSeqOutOfRangeCode 单据编号流水超出范围
	HrmBillNoSeqOutOfRangeCode = 10334
)

func init() {
	status.Register(HrmBillTypeNotRegisteredCode)
	status.Register(HrmBillNotFoundCode)
	status.Register(HrmBillInvalidTransitionCode)
	status.Register(HrmBillInvalidTypeCode)
	status.Register(HrmBillNoSeqOutOfRangeCode)
}
