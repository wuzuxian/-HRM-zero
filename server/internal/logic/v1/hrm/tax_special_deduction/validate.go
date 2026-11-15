package tax_special_deduction

import (
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// DeductType 专项附加扣除项 7 类（demo-inventory 8.3）.
const (
	DeductTypeChildrenEdu     = "CHILDREN_EDU"     // 子女教育
	DeductTypeContinuingEdu   = "CONTINUING_EDU"   // 继续教育
	DeductTypeCriticalIllness = "CRITICAL_ILLNESS" // 大病医疗
	DeductTypeHouseLoan       = "HOUSE_LOAN"       // 住房贷款利息
	DeductTypeHouseRent       = "HOUSE_RENT"       // 住房租金
	DeductTypeSupportElderly  = "SUPPORT_ELDERLY"  // 赡养老人
	DeductTypeInfantCare      = "INFANT_CARE"      // 3岁以下婴幼儿照护
)

// Source 来源: manual=手工维护 import=导入
const (
	SourceManual = "manual"
	SourceImport = "import"
)

// validDeductType 扣除项枚举校验.
func validDeductType(v string) bool {
	switch v {
	case DeductTypeChildrenEdu, DeductTypeContinuingEdu, DeductTypeCriticalIllness,
		DeductTypeHouseLoan, DeductTypeHouseRent, DeductTypeSupportElderly, DeductTypeInfantCare:
		return true
	}
	return false
}

// validYear 扣除年度: 4 位数字（如 2026）.
func validYear(v string) bool {
	if len(v) != 4 {
		return false
	}
	for _, c := range v {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// validMonth 月份 1-12.
func validMonth(v int64) bool {
	return v >= 1 && v <= 12
}

// validAmount 扣除金额非负.
func validAmount(v float64) bool {
	return v >= 0
}

// validateDeductionCommon 单行扣除的公共校验.
func validateDeductionCommon(employeeUuid, year string, month int64, deductType string, amount float64) error {
	if employeeUuid == "" {
		return status.ErrorMessage(errcodes.TaxSpecialDeductionInvalidCode, "关联员工不能为空")
	}
	if !validYear(year) {
		return status.ErrorMessage(errcodes.TaxSpecialDeductionInvalidCode, "扣除年度必须是 4 位数字年份")
	}
	if !validMonth(month) {
		return status.ErrorMessage(errcodes.TaxSpecialDeductionInvalidCode, "扣除月份必须在 1-12 之间")
	}
	if !validDeductType(deductType) {
		return status.ErrorMessage(errcodes.TaxSpecialDeductionInvalidCode, "扣除项不是合法的专项附加扣除类型")
	}
	if !validAmount(amount) {
		return status.ErrorMessage(errcodes.TaxSpecialDeductionInvalidCode, "扣除金额不能为负数")
	}
	return nil
}
