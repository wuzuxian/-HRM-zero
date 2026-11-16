// Package errcodes — HRM-37 薪资批次与试算引擎业务错误码（1068x 段）。
// 1067x=HRM-23 考勤月报、1066x=HRM-43 考核任务链、1065x 已占（开工 grep 确认 1068x 空闲）。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// SalaryBatchNotFoundCode 薪资批次不存在
	SalaryBatchNotFoundCode = 10680
	// SalaryBatchInvalidCode 薪资批次参数非法（名称空、主体空、年月越界、批次类型枚举、状态枚举）
	SalaryBatchInvalidCode = 10681
	// SalaryBatchSubjectNotFoundCode 批次关联薪酬主体不存在
	SalaryBatchSubjectNotFoundCode = 10682
	// SalaryBatchDuplicateCode 同主体+年月+批次类型的批次已存在（唯一范围冲突）
	SalaryBatchDuplicateCode = 10683
	// SalaryBatchTrialFailedCode 试算失败（无任何有效档案员工等无法产出批次结果）
	SalaryBatchTrialFailedCode = 10684
	// SalaryBatchNoResultCode 批次尚未试算（查看结果/进度推进前置校验）
	SalaryBatchNoResultCode = 10685
)

func init() {
	status.Register(SalaryBatchNotFoundCode)
	status.Register(SalaryBatchInvalidCode)
	status.Register(SalaryBatchSubjectNotFoundCode)
	status.Register(SalaryBatchDuplicateCode)
	status.Register(SalaryBatchTrialFailedCode)
	status.Register(SalaryBatchNoResultCode)
}
