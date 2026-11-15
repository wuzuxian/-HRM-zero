package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-41 绩效方案业务错误码（1038x；1037x 预留给在途的人事调动单）。
const (
	// HrmPerfSchemeNotFoundCode 绩效方案不存在
	HrmPerfSchemeNotFoundCode = 10380
	// HrmPerfSchemeInvalidCode 绩效方案字段非法（名称为空、周期类型/评估模式/评分模式/联薪方式/状态枚举越界）
	HrmPerfSchemeInvalidCode = 10381
	// HrmPerfSchemeTemplateNotFoundCode 关联考核模板不存在（template_uuid 须指向 hrm_perf_template 已有行）
	HrmPerfSchemeTemplateNotFoundCode = 10382
)

func init() {
	status.Register(HrmPerfSchemeNotFoundCode)
	status.Register(HrmPerfSchemeInvalidCode)
	status.Register(HrmPerfSchemeTemplateNotFoundCode)
}
