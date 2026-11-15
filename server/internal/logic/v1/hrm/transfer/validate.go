package transfer

import (
	"fmt"
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
)

// 异动类型（GLOSSARY/spec 用户故事 21）：调岗/调部门/调薪/调公司/其他.
const (
	ChangeTypePost    = "post"    // 调岗
	ChangeTypeOrg     = "org"     // 调部门
	ChangeTypeSalary  = "salary"  // 调薪
	ChangeTypeCompany = "company" // 调公司
	ChangeTypeOther   = "other"   // 其他
)

var changeTypes = map[string]struct{}{
	ChangeTypePost: {}, ChangeTypeOrg: {}, ChangeTypeSalary: {},
	ChangeTypeCompany: {}, ChangeTypeOther: {},
}

// editableStatuses are the 单据状态 a 调动单 can still be modified/deleted in:
// 未提交/审批不通过/已撤销 (审批中/审批通过 are locked, mirroring the demo site).
var editableStatuses = map[string]struct{}{
	bill.StatusUnsubmitted: {}, bill.StatusRejected: {}, bill.StatusRevoked: {},
}

// validateTransfer checks the user-supplied 调动单 fields and maps failures to
// the 1036x errcode range.
func validateTransfer(employeeNo, employeeName, changeType string, effectiveImmediately int64, effectiveDate string) error {
	if employeeNo == "" || employeeName == "" {
		return status.ErrorMessage(errcodes.TransferEmployeeRequiredCode, "员工工号与姓名为必填")
	}
	if _, ok := changeTypes[changeType]; !ok {
		return status.ErrorMessage(errcodes.TransferInvalidCode,
			"异动类型必须是 调岗/调部门/调薪/调公司/其他")
	}
	if effectiveImmediately != 1 && effectiveImmediately != 0 {
		return status.ErrorMessage(errcodes.TransferInvalidCode,
			"是否立即生效必须是 是/否")
	}
	if effectiveImmediately == 0 {
		if effectiveDate == "" {
			return status.ErrorMessage(errcodes.TransferInvalidCode,
				"非立即生效时生效日期必填")
		}
		if _, err := time.Parse(time.DateOnly, effectiveDate); err != nil {
			return status.ErrorMessage(errcodes.TransferInvalidCode,
				fmt.Sprintf("生效日期格式非法（须为 YYYY-MM-DD）: %s", effectiveDate))
		}
	}
	return nil
}

// canEdit reports whether the 单据状态 allows modification/deletion.
func canEdit(status string) bool {
	_, ok := editableStatuses[status]
	return ok
}
