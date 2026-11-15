package tax_opening

import (
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_opening"
)

// openingUniqueCondition 表级唯一键 (员工, 主体, 年度) 条件.
func openingUniqueCondition(employeeUuid, subjectUuid, taxYear string) []condition.Condition {
	return []condition.Condition{
		{Field: hrm_tax_opening.EmployeeUuid, Operator: condition.Equal, Value: employeeUuid},
		{Field: hrm_tax_opening.SubjectUuid, Operator: condition.Equal, Value: subjectUuid},
		{Field: hrm_tax_opening.TaxYear, Operator: condition.Equal, Value: taxYear},
	}
}

// openingUuidCondition 按 uuid 定位单行.
func openingUuidCondition(uuid string) []condition.Condition {
	return []condition.Condition{{
		Field:    hrm_tax_opening.Uuid,
		Operator: condition.Equal,
		Value:    uuid,
	}}
}
