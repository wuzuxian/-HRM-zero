package regularization

import (
	"fmt"
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
)

// BillTypeCode is the 员工转正 类型码 (HR202, demo-inventory §0).
const BillTypeCode = "202"

// editableStatuses are the 单据状态 a 转正单 can still be modified/deleted in:
// 未提交/审批不通过/已撤销 (审批中/审批通过 are locked, mirroring the demo site).
var editableStatuses = map[string]struct{}{
	bill.StatusUnsubmitted: {}, bill.StatusRejected: {}, bill.StatusRevoked: {},
}

// validateRegularization checks the user-supplied 转正单 fields and maps
// failures to the 10465+ errcode range. 员工快照 (工号/姓名) is required —
// 选择员工带出档案快照 (spec 用户故事 20); the rest are pass-through snapshot
// fields except 实际转正日期 which must be a valid date when given.
func validateRegularization(employeeNo, employeeName, formalDate string) error {
	if employeeNo == "" || employeeName == "" {
		return status.ErrorMessage(errcodes.RegularizationEmployeeRequiredCode, "员工工号与姓名为必填")
	}
	if formalDate != "" {
		if _, err := time.Parse(time.DateOnly, formalDate); err != nil {
			return status.ErrorMessage(errcodes.RegularizationInvalidCode,
				fmt.Sprintf("实际转正日期格式非法（须为 YYYY-MM-DD）: %s", formalDate))
		}
	}
	return nil
}

// canEdit reports whether the 单据状态 allows modification/deletion.
func canEdit(status string) bool {
	_, ok := editableStatuses[status]
	return ok
}
