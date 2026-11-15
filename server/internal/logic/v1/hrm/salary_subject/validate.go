package salary_subject

import (
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// orgTypeCompany HRM 组织类型: 1=公司 2=部门（GLOSSARY「组织」）。薪酬主体必须关联公司节点。
const orgTypeCompany = "1"

// validSubjectStatus 薪酬主体启停枚举: 1=启用 0=停用.
func validSubjectStatus(v string) bool {
	return v == "1" || v == "0"
}

// validateSubjectCommon 主体名称非空 + 启停状态合法.
func validateSubjectCommon(subjectName, statusStr string) error {
	if subjectName == "" {
		return status.ErrorMessage(errcodes.SalarySubjectInvalidCode, "主体名称不能为空")
	}
	if !validSubjectStatus(statusStr) {
		return status.ErrorMessage(errcodes.SalarySubjectInvalidCode, "薪酬主体状态必须是 启用 或 停用")
	}
	return nil
}
