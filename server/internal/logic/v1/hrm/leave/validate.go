package leave

import (
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
)

// validateLeave Add/Edit 共用校验：类型编码非空 + 时间可解析且结束晚于开始；返回自动计算的天数.
func validateLeave(leaveTypeCode, planStartTime, planEndTime string) (float64, error) {
	if leaveTypeCode == "" {
		return 0, status.ErrorMessage(errcodes.LeaveInvalidCode, "请假类型不能为空")
	}
	return CalcLeaveDays(planStartTime, planEndTime)
}

// conditionChainByUuid 单据 uuid 条件.
func conditionChainByUuid(billUuid string) []condition.Condition {
	return condition.NewChain().Equal(hrm_leave_bill.Uuid, billUuid).Build()
}

// conditionByTypeCode 假期类型编码条件.
func conditionByTypeCode(code string) condition.Chain {
	return condition.NewChain().Equal(hrm_leave_type.TypeCode, code)
}
