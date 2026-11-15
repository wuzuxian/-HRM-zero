package tax_special_deduction

import (
	"fmt"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_special_deduction"
)

// deductionKey 提交内 (月份, 扣除项) 去重键.
func deductionKey(month int64, deductType string) string {
	return fmt.Sprintf("%d|%s", month, deductType)
}

// deductionUniqueCondition 表级唯一键 (员工, 年度, 月份, 扣除项) 条件.
func deductionUniqueCondition(employeeUuid, year string, month int64, deductType string) []condition.Condition {
	return []condition.Condition{
		{Field: hrm_tax_special_deduction.EmployeeUuid, Operator: condition.Equal, Value: employeeUuid},
		{Field: hrm_tax_special_deduction.DeductYear, Operator: condition.Equal, Value: year},
		{Field: hrm_tax_special_deduction.DeductMonth, Operator: condition.Equal, Value: month},
		{Field: hrm_tax_special_deduction.DeductType, Operator: condition.Equal, Value: deductType},
	}
}

// deductionUuidCondition 按 uuid 定位单行.
func deductionUuidCondition(uuid string) []condition.Condition {
	return []condition.Condition{{
		Field:    hrm_tax_special_deduction.Uuid,
		Operator: condition.Equal,
		Value:    uuid,
	}}
}
