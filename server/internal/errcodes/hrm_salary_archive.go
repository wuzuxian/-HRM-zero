// Package errcodes — HRM-35 员工薪资档案业务错误码（1062x 段）。
// 1061x 已被 HRM-34 薪酬方案占用，1060x 已被 HRM-21 加班单占用（开工时 grep 确认 1062x 空闲）。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// SalaryArchiveNotFoundCode 薪资档案不存在
	SalaryArchiveNotFoundCode = 10620
	// SalaryArchiveInvalidCode 薪资档案参数非法（员工/主体为空、金额负数、日期格式非法、状态枚举越界等）
	SalaryArchiveInvalidCode = 10621
	// SalaryArchiveEmployeeNotFoundCode 档案关联员工不存在
	SalaryArchiveEmployeeNotFoundCode = 10622
	// SalaryArchiveSubjectNotFoundCode 档案关联薪酬主体不存在
	SalaryArchiveSubjectNotFoundCode = 10623
	// SalaryArchiveSchemeNotFoundCode 档案关联薪酬方案不存在
	SalaryArchiveSchemeNotFoundCode = 10624
	// SalaryArchiveDateInvalidCode 失效日期必须大于生效日期
	SalaryArchiveDateInvalidCode = 10625
	// SalaryArchiveOverlapCode 同员工同薪酬主体下生效区间重叠（期间交叉拒绝）
	SalaryArchiveOverlapCode = 10626
)

func init() {
	status.Register(SalaryArchiveNotFoundCode)
	status.Register(SalaryArchiveInvalidCode)
	status.Register(SalaryArchiveEmployeeNotFoundCode)
	status.Register(SalaryArchiveSubjectNotFoundCode)
	status.Register(SalaryArchiveSchemeNotFoundCode)
	status.Register(SalaryArchiveDateInvalidCode)
	status.Register(SalaryArchiveOverlapCode)
}
