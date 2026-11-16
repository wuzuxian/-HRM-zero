// Package errcodes — HRM-38 薪资确认审批与发薪登记业务错误码（1071x 段）。
// 2026-11-16 grep errcodes/*.go 确认 1071x 空闲（1068x=HRM-37 薪资批次、
// 1069x=HRM-45 绩效结果审批、10700=HRM-42 工作台）。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// SalaryConfirmNotFoundCode 确认审批单不存在
	SalaryConfirmNotFoundCode = 10710
	// SalaryConfirmInvalidCode 确认审批字段非法 / 批次状态不符（未试算、已确认）/ 重复提交被拒
	SalaryConfirmInvalidCode = 10711
	// SalaryPaymentNotFoundCode 发薪登记行不存在
	SalaryPaymentNotFoundCode = 10712
	// SalaryPaymentInvalidCode 发薪登记字段非法（发薪方式枚举）
	SalaryPaymentInvalidCode = 10713
	// SalaryPaymentBatchNotConfirmedCode 批次未确认或已发薪（发薪登记前置守卫：
	// 须 confirm_status='1' 且 pay_status='0'）
	SalaryPaymentBatchNotConfirmedCode = 10714
)

func init() {
	status.Register(SalaryConfirmNotFoundCode)
	status.Register(SalaryConfirmInvalidCode)
	status.Register(SalaryPaymentNotFoundCode)
	status.Register(SalaryPaymentInvalidCode)
	status.Register(SalaryPaymentBatchNotConfirmedCode)
}
