package tax_opening

import (
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// validYear 年度: 4 位数字（如 2026）.
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

// validMonth 截至月 1-12.
func validMonth(v int64) bool {
	return v >= 1 && v <= 12
}

// validAmount 累计金额非负.
func validAmount(v float64) bool {
	return v >= 0
}

// validateOpeningCommon 期初行的公共校验: 员工/年度/截至月/累计金额.
func validateOpeningCommon(employeeUuid, taxYear string, untilMonth int64, amounts ...float64) error {
	if employeeUuid == "" {
		return status.ErrorMessage(errcodes.TaxOpeningInvalidCode, "关联员工不能为空")
	}
	if !validYear(taxYear) {
		return status.ErrorMessage(errcodes.TaxOpeningInvalidCode, "年度必须是 4 位数字年份")
	}
	if !validMonth(untilMonth) {
		return status.ErrorMessage(errcodes.TaxOpeningInvalidCode, "截至月必须在 1-12 之间")
	}
	for _, a := range amounts {
		if !validAmount(a) {
			return status.ErrorMessage(errcodes.TaxOpeningInvalidCode, "累计金额不能为负数")
		}
	}
	return nil
}
