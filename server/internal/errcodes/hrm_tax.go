package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-36 个税管理三件套业务错误码（1052x；10500-10502 perf_scheme 已用，1051x 留给 HRM-18 假期余额）。
const (
	// TaxProfileNotFoundCode 员工个税档案不存在
	TaxProfileNotFoundCode = 10520
	// TaxProfileInvalidCode 员工个税档案字段非法（员工为空、居民类型/启停枚举越界、专项附加编码非法）
	TaxProfileInvalidCode = 10521
	// TaxProfileDuplicateCode 该员工已存在个税档案（员工唯一）
	TaxProfileDuplicateCode = 10522
	// TaxProfileEmployeeNotFoundCode 关联员工不存在
	TaxProfileEmployeeNotFoundCode = 10523

	// TaxSpecialDeductionNotFoundCode 专项附加扣除行不存在
	TaxSpecialDeductionNotFoundCode = 10524
	// TaxSpecialDeductionInvalidCode 专项附加扣除字段非法（年度/月份/扣除项枚举越界、金额为负）
	TaxSpecialDeductionInvalidCode = 10525
	// TaxSpecialDeductionDuplicateCode 该员工年度月份扣除项已存在（唯一键冲突）
	TaxSpecialDeductionDuplicateCode = 10526
	// TaxSpecialDeductionEmployeeNotFoundCode 关联员工不存在
	TaxSpecialDeductionEmployeeNotFoundCode = 10527
	// TaxSpecialDeductionImportInvalidCode 专项附加扣除导入文件非法
	TaxSpecialDeductionImportInvalidCode = 10528

	// TaxOpeningNotFoundCode 个税累计期初不存在
	TaxOpeningNotFoundCode = 10529
	// TaxOpeningInvalidCode 个税累计期初字段非法（年度/截至月越界、金额为负）
	TaxOpeningInvalidCode = 10530
	// TaxOpeningDuplicateCode 该员工主体年度累计期初已存在（唯一键冲突）
	TaxOpeningDuplicateCode = 10531
	// TaxOpeningEmployeeNotFoundCode 关联员工不存在
	TaxOpeningEmployeeNotFoundCode = 10532
	// TaxOpeningImportInvalidCode 个税累计期初导入文件非法
	TaxOpeningImportInvalidCode = 10533
)

func init() {
	status.Register(TaxProfileNotFoundCode)
	status.Register(TaxProfileInvalidCode)
	status.Register(TaxProfileDuplicateCode)
	status.Register(TaxProfileEmployeeNotFoundCode)
	status.Register(TaxSpecialDeductionNotFoundCode)
	status.Register(TaxSpecialDeductionInvalidCode)
	status.Register(TaxSpecialDeductionDuplicateCode)
	status.Register(TaxSpecialDeductionEmployeeNotFoundCode)
	status.Register(TaxSpecialDeductionImportInvalidCode)
	status.Register(TaxOpeningNotFoundCode)
	status.Register(TaxOpeningInvalidCode)
	status.Register(TaxOpeningDuplicateCode)
	status.Register(TaxOpeningEmployeeNotFoundCode)
	status.Register(TaxOpeningImportInvalidCode)
}
