package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// HRM 员工↔系统账号双向同步（HRM-04）业务错误码。1046x 段
	// （1040x 培训, 1042x 奖惩, 1043x 招聘需求, 1044x 候选人, 1045x 证书）。
	// EmployeeUserNotFoundCode 生成账号时员工不存在（uuid 无效或已被删除）
	EmployeeUserNotFoundCode = 10460
	// EmployeeUserNoDuplicateCode 系统账号用户名已存在（manage_user.username 与工号冲突）
	EmployeeUserNoDuplicateCode = 10461
	// EmployeeUserReferencedCode 账号被员工档案引用不可删除（HRM-52 删除一致性，
	// 需先在员工档案解除关联；批量删除含任一被引用账号时整批拒绝）
	EmployeeUserReferencedCode = 10469
)

// HRM-51 员工↔账号关联/解绑业务错误码。1072x 段：
// 1046x 已在 10460-10468 用尽（10462-10464 加班、10465-10468 转正单，
// 2026-11-17 grep errcodes/*.go 确认仅 10469 空闲），为保持三个关联错误码
// 连续改取 1072x（1070x 工作台、1071x 薪资确认，grep 确认 1072x 空闲）。
const (
	// EmployeeUserOccupiedCode 账号已被其他员工档案持有（关联一对一校验，报错指明占用者）
	EmployeeUserOccupiedCode = 10720
	// EmployeeUserAccountNotFoundCode 待关联的系统账号不存在（uuid 无效或已被删除）
	EmployeeUserAccountNotFoundCode = 10721
	// EmployeeUserAlreadyLinkedCode 员工已关联其他系统账号（档案已持有不同账号时再次关联/建号）
	EmployeeUserAlreadyLinkedCode = 10722
	// EmployeeUserResignedCode 离职补录员工不开通系统账号（ADR 0007: person_status='5' 跳过开号）
	EmployeeUserResignedCode = 10723
)

func init() {
	status.Register(EmployeeUserNotFoundCode)
	status.Register(EmployeeUserNoDuplicateCode)
	status.Register(EmployeeUserOccupiedCode)
	status.Register(EmployeeUserAccountNotFoundCode)
	status.Register(EmployeeUserAlreadyLinkedCode)
	status.Register(EmployeeUserResignedCode)
	status.Register(EmployeeUserReferencedCode)
}
