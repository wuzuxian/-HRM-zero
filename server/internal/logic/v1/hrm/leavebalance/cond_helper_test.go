package leavebalance

import (
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
)

func conditionEqualTypeCode(code string) condition.Chain {
	return condition.NewChain().Equal(hrm_leave_type.TypeCode, code)
}
