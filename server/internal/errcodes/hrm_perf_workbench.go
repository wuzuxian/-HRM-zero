package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-44 考核任务工作台业务错误码（1067x 段；1064x 已被考核任务链占用
// 10650-10667 offer/考核链，grep errcodes/*.go 于 2026-11-16 确认 1067x 空闲，
// 本票声明占用 10670）。
const (
	// HrmPerfWorkbenchInvalidCode 工作台请求非法（如未绑定员工档案且无法定位身份）
	HrmPerfWorkbenchInvalidCode = 10670
)

func init() {
	status.Register(HrmPerfWorkbenchInvalidCode)
}
