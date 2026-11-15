package resignation

import (
	"fmt"
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
)

// BillTypeCode is the 员工离职 类型码 (HR204, demo-inventory §0).
const BillTypeCode = "204"

// 离职类型（spec 用户故事 22 / demo-inventory §3.4）：自愿离职/被动离职/合同到期/退休.
const (
	ResignTypeVoluntary    = "voluntary"    // 自愿离职
	ResignTypeInvoluntary  = "involuntary"  // 被动离职
	ResignTypeContractEnd  = "contract_end" // 合同到期
	ResignTypeRetirement   = "retirement"   // 退休
)

var resignTypes = map[string]struct{}{
	ResignTypeVoluntary:   {},
	ResignTypeInvoluntary: {},
	ResignTypeContractEnd: {},
	ResignTypeRetirement:  {},
}

// editableStatuses are the 单据状态 a 离职单 can still be modified/deleted in:
// 未提交/审批不通过/已撤销 (审批中/审批通过 are locked, mirroring the demo site).
var editableStatuses = map[string]struct{}{
	bill.StatusUnsubmitted: {}, bill.StatusRejected: {}, bill.StatusRevoked: {},
}

// validateResignation checks the user-supplied 离职单 fields and maps failures
// to the 1049x errcode range. 员工快照 (工号/姓名) is required; 离职类型 must be
// one of the enum values; 离职日期 is mandatory (定时任务生效日) and the other
// date fields must be valid YYYY-MM-DD when given.
func validateResignation(employeeNo, employeeName, resignType, resignDate string,
	lastWorkDate, salarySettleDate string) error {
	if employeeNo == "" || employeeName == "" {
		return status.ErrorMessage(errcodes.ResignationEmployeeRequiredCode, "员工工号与姓名为必填")
	}
	if _, ok := resignTypes[resignType]; !ok {
		return status.ErrorMessage(errcodes.ResignationInvalidCode,
			"离职类型必须是 自愿离职/被动离职/合同到期/退休")
	}
	if resignDate == "" {
		return status.ErrorMessage(errcodes.ResignationInvalidCode, "离职日期必填（定时任务在离职日当天统一置离职）")
	}
	for name, value := range map[string]string{
		"离职日期":     resignDate,
		"最后工作日期":   lastWorkDate,
		"薪资结算日期":   salarySettleDate,
	} {
		if value == "" {
			continue
		}
		if _, err := time.Parse(time.DateOnly, value); err != nil {
			return status.ErrorMessage(errcodes.ResignationInvalidCode,
				fmt.Sprintf("%s格式非法（须为 YYYY-MM-DD）: %s", name, value))
		}
	}
	return nil
}

// canEdit reports whether the 单据状态 allows modification/deletion.
func canEdit(status string) bool {
	_, ok := editableStatuses[status]
	return ok
}
