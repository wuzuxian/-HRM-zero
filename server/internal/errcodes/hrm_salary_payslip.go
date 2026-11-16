// Package errcodes — HRM-39 工资条与薪资统计台账业务错误码（1069x 段 10695 起）。
// 10690-10694 已占 HRM-45 绩效结果（grep errcodes/*.go 确认 10695 起空闲）。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// SalaryPayslipNotFoundCode 工资条不存在
	SalaryPayslipNotFoundCode = 10695
	// SalaryPayslipInvalidCode 工资条参数非法（uuid 空、批次空等）
	SalaryPayslipInvalidCode = 10696
	// SalaryPayslipBatchNotConfirmedCode 批次未确认，拒绝发布工资条（confirm_status != '1'）
	SalaryPayslipBatchNotConfirmedCode = 10697
	// SalaryPayslipBatchNoResultCode 批次尚无试算结果（无明细行可生成工资条）
	SalaryPayslipBatchNoResultCode = 10698
)

func init() {
	status.Register(SalaryPayslipNotFoundCode)
	status.Register(SalaryPayslipInvalidCode)
	status.Register(SalaryPayslipBatchNotConfirmedCode)
	status.Register(SalaryPayslipBatchNoResultCode)
}
