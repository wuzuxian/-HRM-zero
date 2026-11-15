package tax_profile

import (
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 居民类型枚举（GLOSSARY「居民类型」）: resident=居民个人 non_resident=非居民个人
const (
	ResidentTypeResident     = "resident"
	ResidentTypeNonResident  = "non_resident"
)

// validResidentType 居民类型枚举校验.
func validResidentType(v string) bool {
	return v == ResidentTypeResident || v == ResidentTypeNonResident
}

// validStatus 启停枚举: 1=启用 0=停用（HRM 约定 '1'/'0'）.
func validStatus(v string) bool {
	return v == "1" || v == "0"
}

// validateProfileCommon 员工 + 居民类型 + 启停 的公共校验.
func validateProfileCommon(employeeUuid, residentType, statusStr string) error {
	if employeeUuid == "" {
		return status.ErrorMessage(errcodes.TaxProfileInvalidCode, "关联员工不能为空")
	}
	if !validResidentType(residentType) {
		return status.ErrorMessage(errcodes.TaxProfileInvalidCode, "居民类型必须是 居民个人 或 非居民个人")
	}
	if !validStatus(statusStr) {
		return status.ErrorMessage(errcodes.TaxProfileInvalidCode, "个税档案状态必须是 启用 或 停用")
	}
	return nil
}
