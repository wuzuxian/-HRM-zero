package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// HRM 员工↔系统账号双向同步（HRM-04）业务错误码。1046x 段
	// （1040x 培训, 1042x 奖惩, 1043x 招聘需求, 1044x 候选人, 1045x 证书）。
	// EmployeeUserNotFoundCode 生成账号时员工不存在（uuid 无效或已被删除）
	EmployeeUserNotFoundCode = 10460
	// EmployeeUserNoDuplicateCode 系统账号用户名已存在（manage_user.username 与工号冲突）
	EmployeeUserNoDuplicateCode = 10461
)

func init() {
	status.Register(EmployeeUserNotFoundCode)
	status.Register(EmployeeUserNoDuplicateCode)
}
