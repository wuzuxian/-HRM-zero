package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// HRM 培训管理：证书模板（HRM-30）业务错误码。1042x 段
	// （1040x 归 training-plan+session，1041x 预留，integration-warnings.md #4）。
	// CertTemplateNotFoundCode 证书模板不存在
	CertTemplateNotFoundCode = 10420
	// CertTemplateInvalidCode 证书模板参数非法（名称/前缀/来源等）
	CertTemplateInvalidCode = 10421
	// CertTemplatePrefixDuplicateCode 编号前缀已存在（前缀唯一，决定证书号=前缀-年-流水）
	CertTemplatePrefixDuplicateCode = 10422

	// HRM 培训管理：证书发放（HRM-30）业务错误码。
	// CertificateNotFoundCode 证书不存在
	CertificateNotFoundCode = 10423
	// CertificateInvalidCode 证书参数非法（批量发证/导入/上传等参数错误）
	CertificateInvalidCode = 10424
	// CertificateTemplateDisabledCode 证书模板已停用，不可发证
	CertificateTemplateDisabledCode = 10425
	// CertificateStatusInvalidCode 证书状态流转非法（发放要求待发、撤销要求已发）
	CertificateStatusInvalidCode = 10426
	// CertificateNoDuplicateCode 证书号已存在（外部导入证书号唯一）
	CertificateNoDuplicateCode = 10427
	// CertificateFileNotFoundCode 证书扫描件不存在（未上传）
	CertificateFileNotFoundCode = 10428
)

func init() {
	status.Register(CertTemplateNotFoundCode)
	status.Register(CertTemplateInvalidCode)
	status.Register(CertTemplatePrefixDuplicateCode)

	status.Register(CertificateNotFoundCode)
	status.Register(CertificateInvalidCode)
	status.Register(CertificateTemplateDisabledCode)
	status.Register(CertificateStatusInvalidCode)
	status.Register(CertificateNoDuplicateCode)
	status.Register(CertificateFileNotFoundCode)
}
