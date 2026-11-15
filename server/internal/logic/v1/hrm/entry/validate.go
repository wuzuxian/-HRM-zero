package entry

import (
	"fmt"
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
)

// 员工入职单 enums shared with HRM-03 员工档案 (GLOSSARY):
// 性别 1=男 2=女; 婚姻状况 1=未婚 2=已婚 3=离异 4=丧偶.
const (
	genderMale   = "1"
	genderFemale = "2"

	maritalUnmarried = "1"
	maritalMarried   = "2"
	maritalDivorced  = "3"
	maritalWidowed   = "4"
)

var genders = map[string]struct{}{
	genderMale: {}, genderFemale: {},
}

var maritalStatuses = map[string]struct{}{
	maritalUnmarried: {}, maritalMarried: {}, maritalDivorced: {}, maritalWidowed: {},
}

// editableStatuses are the 单据状态 an 入职单 can still be modified/deleted in:
// 未提交/审批不通过/已撤销 (审批中/审批通过 are locked; 审批通过 already 建档).
var editableStatuses = map[string]struct{}{
	bill.StatusUnsubmitted: {}, bill.StatusRejected: {}, bill.StatusRevoked: {},
}

// validateEntry checks the user-supplied 入职单 fields and maps failures to the
// 1047x errcode range. Date formats follow the repo-wide YYYY-MM-DD convention.
func validateEntry(name, gender, maritalStatus, hireDate, expectFormalDate string, probationMonths int64) error {
	if name == "" {
		return status.ErrorMessage(errcodes.EntryNameRequiredCode, "姓名为必填")
	}
	if gender != "" {
		if _, ok := genders[gender]; !ok {
			return status.ErrorMessage(errcodes.EntryInvalidCode, "性别非法（须为 1=男 2=女）")
		}
	}
	if maritalStatus != "" {
		if _, ok := maritalStatuses[maritalStatus]; !ok {
			return status.ErrorMessage(errcodes.EntryInvalidCode, "婚姻状况非法（须为 1=未婚 2=已婚 3=离异 4=丧偶）")
		}
	}
	if probationMonths < 0 {
		return status.ErrorMessage(errcodes.EntryInvalidCode, "试用期(月)不能为负数")
	}
	if hireDate == "" {
		return status.ErrorMessage(errcodes.EntryInvalidCode, "入职日期必填")
	}
	if _, err := time.Parse(time.DateOnly, hireDate); err != nil {
		return status.ErrorMessage(errcodes.EntryInvalidCode,
			fmt.Sprintf("入职日期格式非法（须为 YYYY-MM-DD）: %s", hireDate))
	}
	if expectFormalDate != "" {
		if _, err := time.Parse(time.DateOnly, expectFormalDate); err != nil {
			return status.ErrorMessage(errcodes.EntryInvalidCode,
				fmt.Sprintf("预计转正日期格式非法（须为 YYYY-MM-DD）: %s", expectFormalDate))
		}
	}
	return nil
}

// canEdit reports whether the 单据状态 allows modification/deletion.
func canEdit(s string) bool {
	_, ok := editableStatuses[s]
	return ok
}

// expectFormalOf computes the 预计转正日期: explicit value wins; otherwise
// 入职日期 + 试用期月 (demo-inventory 3.1: 预计转正=入职日期+试用期).
func expectFormalOf(hireDate, expectFormalDate string, probationMonths int64) string {
	if expectFormalDate != "" {
		return expectFormalDate
	}
	day, err := time.Parse(time.DateOnly, hireDate)
	if err != nil {
		return ""
	}
	return day.AddDate(0, int(probationMonths), 0).Format(time.DateOnly)
}
