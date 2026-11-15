package cert_template

import (
	"strings"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 证书模板状态: 1=启用 2=停用 (GLOSSARY 启用/停用; 停用后不可发证).
const (
	TemplateStatusEnabled  = "1"
	TemplateStatusDisabled = "2"
)

// 证书模板来源: 1=系统 2=企业.
const (
	TemplateSourceSystem = "1"
	TemplateSourceOrg    = "2"
)

// noPrefixMaxLen is the 编号前缀 upper bound (演示前缀 HF/CE/YX/XF/RZ/JN/SG/AQ/HG/JX 均为 2 字符).
const noPrefixMaxLen = 8

// validateTemplate is the parameter validation shared by Add/Edit: 名称必填、
// 证书类型必填、来源枚举、前缀必填且为字母数字.
func validateTemplate(templateName, certType, source, noPrefix string) error {
	if strings.TrimSpace(templateName) == "" {
		return status.ErrorMessage(errcodes.CertTemplateInvalidCode, "模板名称不能为空")
	}
	if strings.TrimSpace(certType) == "" {
		return status.ErrorMessage(errcodes.CertTemplateInvalidCode, "证书类型不能为空")
	}
	if source != TemplateSourceSystem && source != TemplateSourceOrg {
		return status.ErrorMessage(errcodes.CertTemplateInvalidCode, "来源非法（1=系统 2=企业）")
	}
	prefix := strings.TrimSpace(noPrefix)
	if prefix == "" {
		return status.ErrorMessage(errcodes.CertTemplateInvalidCode, "编号前缀不能为空")
	}
	if len(prefix) > noPrefixMaxLen {
		return status.ErrorMessage(errcodes.CertTemplateInvalidCode, "编号前缀过长（最多 8 字符）")
	}
	for _, r := range prefix {
		if !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') {
			return status.ErrorMessage(errcodes.CertTemplateInvalidCode,
				"编号前缀仅允许字母数字（如 JN/AQ/HF）")
		}
	}
	return nil
}
